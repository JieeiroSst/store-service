package com.JIeeiroSst

import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.response.*
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.booleanOrNull
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.contentOrNull
import kotlinx.serialization.json.jsonPrimitive
import kotlinx.serialization.json.put
import org.slf4j.LoggerFactory
import java.net.URI
import java.net.http.HttpClient
import java.net.http.HttpRequest
import java.net.http.HttpResponse
import java.security.MessageDigest
import java.time.Duration
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.atomic.AtomicBoolean

const val ARRANGE_AUTH = "arrangeAuth"
const val INTERNAL_TOKEN_HEADER = "X-Internal-Token"

data class StorePrincipal(val userId: String, val service: Boolean)

sealed interface AuthOutcome {
    data class Authenticated(val principal: StorePrincipal) : AuthOutcome
    data class Rejected(val status: HttpStatusCode, val message: String) : AuthOutcome
}

fun interface SessionValidator {
    suspend fun validate(token: String): String?
}

class UserServiceSessionValidator(
    baseUrl: String,
    timeout: Duration = Duration.ofSeconds(3),
    private val http: HttpClient = HttpClient.newBuilder().connectTimeout(timeout).build(),
) : SessionValidator {
    private val endpoint = URI.create(baseUrl.trimEnd('/') + "/api/v1/validate")
    private val requestTimeout = timeout

    override suspend fun validate(token: String): String? {
        val body = buildJsonObject { put("session_token", token) }.toString()
        val request = HttpRequest.newBuilder(endpoint)
            .timeout(requestTimeout)
            .header(HttpHeaders.ContentType, ContentType.Application.Json.toString())
            .POST(HttpRequest.BodyPublishers.ofString(body))
            .build()
        val response = try {
            withContext(Dispatchers.IO) { http.send(request, HttpResponse.BodyHandlers.ofString()) }
        } catch (e: Exception) {
            throw UserServiceUnavailable("validate session: ${e.message}")
        }
        when (response.statusCode()) {
            HttpStatusCode.OK.value -> Unit
            HttpStatusCode.Unauthorized.value, HttpStatusCode.Forbidden.value -> return null
            else -> throw UserServiceUnavailable("validate session returned ${response.statusCode()}")
        }
        val json = try {
            Json.parseToJsonElement(response.body()) as? JsonObject
        } catch (e: Exception) {
            null
        } ?: throw UserServiceUnavailable("validate session returned an unreadable body")
        val valid = json["valid"]?.jsonPrimitive?.booleanOrNull ?: false
        val userId = (json["user_id"] ?: json["userId"])?.jsonPrimitive?.contentOrNull.orEmpty()
        return if (valid && userId.isNotEmpty()) userId else null
    }
}

class UserServiceUnavailable(message: String) : Exception(message)

class AccessTokenAuthenticator(
    private val sessions: SessionValidator,
    internalTokens: List<String>,
    private val ttl: Duration = Duration.ofSeconds(30),
    private val now: () -> Long = System::currentTimeMillis,
) {
    private val log = LoggerFactory.getLogger(AccessTokenAuthenticator::class.java)
    private val internal = internalTokens.map { it.trim() }.filter { it.isNotEmpty() }.map { it.toByteArray() }
    private val cache = ConcurrentHashMap<String, Pair<StorePrincipal, Long>>()
    private val warned = AtomicBoolean(false)

    val enforcing: Boolean get() = internal.isNotEmpty()

    suspend fun authenticate(authorization: String?, internalToken: String?): AuthOutcome {
        val service = internalToken?.trim().orEmpty()
        if (service.isNotEmpty()) {
            return if (matchesInternal(service)) {
                AuthOutcome.Authenticated(StorePrincipal(userId = "", service = true))
            } else {
                unauthenticated
            }
        }
        val bearer = authorization?.trim()?.takeIf { it.startsWith("Bearer ", ignoreCase = true) }?.substring(7)?.trim().orEmpty()
        if (bearer.isNotEmpty()) {
            return user(bearer)
        }
        if (!enforcing) {
            if (warned.compareAndSet(false, true)) {
                log.warn("INTERNAL_API_TOKENS is not set, requests without credentials are accepted as internal calls")
            }
            return AuthOutcome.Authenticated(StorePrincipal(userId = "", service = true))
        }
        return unauthenticated
    }

    private suspend fun user(token: String): AuthOutcome {
        val cached = cache[token]
        if (cached != null && cached.second > now()) {
            return AuthOutcome.Authenticated(cached.first)
        }
        val userId = try {
            sessions.validate(token)
        } catch (e: UserServiceUnavailable) {
            log.warn("authn: {}", e.message)
            return AuthOutcome.Rejected(HttpStatusCode.ServiceUnavailable, "authentication service unavailable")
        } ?: return unauthenticated
        val principal = StorePrincipal(userId = userId, service = false)
        if (cache.size > MAX_CACHE) {
            cache.clear()
        }
        cache[token] = principal to now() + ttl.toMillis()
        return AuthOutcome.Authenticated(principal)
    }

    private fun matchesInternal(token: String): Boolean {
        val candidate = token.toByteArray()
        var match = false
        for (want in internal) {
            match = MessageDigest.isEqual(candidate, want) or match
        }
        return match
    }

    companion object {
        private const val MAX_CACHE = 10_000
        private val unauthenticated = AuthOutcome.Rejected(HttpStatusCode.Unauthorized, "authentication required")

        fun fromEnv(env: (String) -> String? = System::getenv): AccessTokenAuthenticator {
            val url = env("USER_SERVICE_URL")?.takeIf { it.isNotBlank() } ?: "http://user-api-svc"
            val tokens = env("INTERNAL_API_TOKENS").orEmpty().split(",")
            return AccessTokenAuthenticator(UserServiceSessionValidator(url), tokens)
        }
    }
}

class AccessTokenAuthenticationProvider(config: Config) : AuthenticationProvider(config) {
    private val authenticator = config.authenticator

    class Config(name: String?) : AuthenticationProvider.Config(name) {
        lateinit var authenticator: AccessTokenAuthenticator
    }

    override suspend fun onAuthenticate(context: AuthenticationContext) {
        val headers = context.call.request.headers
        when (val outcome = authenticator.authenticate(headers[HttpHeaders.Authorization], headers[INTERNAL_TOKEN_HEADER])) {
            is AuthOutcome.Authenticated -> context.principal(name, outcome.principal)
            is AuthOutcome.Rejected -> context.challenge(ARRANGE_AUTH, AuthenticationFailedCause.InvalidCredentials) { challenge, call ->
                call.respond(outcome.status, mapOf("error" to outcome.message))
                challenge.complete()
            }
        }
    }
}

fun Application.configureSecurity(authenticator: AccessTokenAuthenticator = AccessTokenAuthenticator.fromEnv()) {
    authentication {
        register(AccessTokenAuthenticationProvider(AccessTokenAuthenticationProvider.Config(ARRANGE_AUTH).apply {
            this.authenticator = authenticator
        }))
    }
}
