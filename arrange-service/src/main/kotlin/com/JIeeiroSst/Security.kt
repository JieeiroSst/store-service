package com.JIeeiroSst

import io.ktor.server.application.*
import io.ktor.server.auth.*

// Internal service-to-service auth for the arrangement endpoints. Callers
// (other store-service backends) authenticate with Basic auth; credentials
// are validated against the ARRANGE_SERVICE_USERNAME / ARRANGE_SERVICE_PASSWORD
// environment variables so they can be rotated without a redeploy.
const val ARRANGE_AUTH = "arrangeAuth"

fun Application.configureSecurity() {
    val username = System.getenv("ARRANGE_SERVICE_USERNAME") ?: "arrange"
    val password = System.getenv("ARRANGE_SERVICE_PASSWORD") ?: "arrange"

    authentication {
        basic(name = ARRANGE_AUTH) {
            realm = "arrange-service"
            validate { credentials ->
                if (credentials.name == username && credentials.password == password) {
                    UserIdPrincipal(credentials.name)
                } else {
                    null
                }
            }
        }
    }
}
