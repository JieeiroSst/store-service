package com.JIeeiroSst

import com.ucasoft.ktor.simpleCache.SimpleCache
import com.ucasoft.ktor.simpleRedisCache.*
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.plugins.cors.routing.*
import io.ktor.server.plugins.swagger.*
import io.ktor.server.routing.*
import kotlin.time.Duration.Companion.seconds

fun Application.configureHTTP() {
    routing {
        swaggerUI(path = "openapi")
    }
    install(SimpleCache) {
        redisCache {
            invalidateAt = 10.seconds
            host = "localhost"
            port = 6379
        }
    }
    install(CORS) {
        allowMethod(HttpMethod.Options)
        allowMethod(HttpMethod.Put)
        allowMethod(HttpMethod.Patch)
        allowMethod(HttpMethod.Delete)
        allowHeader(HttpHeaders.Authorization)
        anyHost() // @TODO: Don't do this in production if possible. Try to limit it.
    }
}
