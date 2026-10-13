package com.JIeeiroSst

import com.sun.net.httpserver.HttpServer
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import io.ktor.server.testing.*
import java.net.InetSocketAddress
import java.util.concurrent.atomic.AtomicInteger
import kotlin.test.AfterTest
import kotlin.test.BeforeTest
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class SecurityTest {
    private lateinit var users: HttpServer
    private val calls = AtomicInteger()
    private var down = false

    @BeforeTest
    fun startUserService() {
        users = HttpServer.create(InetSocketAddress("127.0.0.1", 0), 0)
        users.createContext("/api/v1/validate") { ex ->
            calls.incrementAndGet()
            val body = ex.requestBody.readBytes().decodeToString()
            val (status, reply) = when {
                down -> 500 to "{}"
                body.contains("\"access-7\"") -> 200 to """{"valid":true,"user_id":"7"}"""
                else -> 200 to """{"valid":false}"""
            }
            val bytes = reply.encodeToByteArray()
            ex.sendResponseHeaders(status, bytes.size.toLong())
            ex.responseBody.use { it.write(bytes) }
        }
        users.start()
    }

    @AfterTest
    fun stopUserService() = users.stop(0)

    private fun authenticator(tokens: List<String>) =
        AccessTokenAuthenticator(UserServiceSessionValidator("http://127.0.0.1:${users.address.port}"), tokens)

    private fun ApplicationTestBuilder.protectedApp(tokens: List<String>) {
        application {
            configureSerialization()
            configureSecurity(authenticator(tokens))
            routing {
                authenticate(ARRANGE_AUTH) {
                    get("/shifts") {
                        val p = call.principal<StorePrincipal>()!!
                        call.respondText("user=${p.userId} service=${p.service}")
                    }
                }
            }
        }
    }

    @Test
    fun acceptsUserServiceAccessToken() = testApplication {
        protectedApp(listOf("svc"))
        val first = client.get("/shifts") { header(HttpHeaders.Authorization, "Bearer access-7") }
        assertEquals(HttpStatusCode.OK, first.status)
        assertEquals("user=7 service=false", first.bodyAsText())
        client.get("/shifts") { header(HttpHeaders.Authorization, "Bearer access-7") }
        assertEquals(1, calls.get(), "validated sessions are cached")
    }

    @Test
    fun rejectsMissingOrInvalidCredentialsWhenEnforcing() = testApplication {
        protectedApp(listOf("svc"))
        assertEquals(HttpStatusCode.Unauthorized, client.get("/shifts").status)
        assertEquals(HttpStatusCode.Unauthorized, client.get("/shifts") { header(HttpHeaders.Authorization, "Bearer forged") }.status)
        assertEquals(HttpStatusCode.Unauthorized, client.get("/shifts") { basicAuth("arrange", "arrange") }.status)
        assertEquals(HttpStatusCode.Unauthorized, client.get("/shifts") { header(INTERNAL_TOKEN_HEADER, "wrong") }.status)
        val service = client.get("/shifts") { header(INTERNAL_TOKEN_HEADER, "svc") }
        assertEquals(HttpStatusCode.OK, service.status)
        assertEquals("user= service=true", service.bodyAsText())
    }

    @Test
    fun userServiceOutageIsReportedAsUnavailable() = testApplication {
        protectedApp(listOf("svc"))
        down = true
        val res = client.get("/shifts") { header(HttpHeaders.Authorization, "Bearer access-7") }
        assertEquals(HttpStatusCode.ServiceUnavailable, res.status)
        assertTrue(res.bodyAsText().contains("authentication service unavailable"))
    }

    @Test
    fun rolloutModeAcceptsAnonymousButStillChecksTokens() = testApplication {
        protectedApp(emptyList())
        assertEquals("user= service=true", client.get("/shifts").bodyAsText())
        assertEquals(HttpStatusCode.Unauthorized, client.get("/shifts") { header(HttpHeaders.Authorization, "Bearer forged") }.status)
    }
}
