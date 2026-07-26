package com.JIeeiroSst.routes

import com.JIeeiroSst.service.NoSpaceAvailableException
import com.JIeeiroSst.service.SpaceArrangementService
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import kotlinx.serialization.Serializable

@Serializable
data class ArrangeSpaceRequest(
    val storeId: Int,
    val referenceId: Int,
    val partySize: Int,
    val from: String,
    val to: String
)

fun Route.spaceRoutes(service: SpaceArrangementService) {
    route("/spaces") {
        get {
            val storeId = call.request.queryParameters["storeId"]?.toIntOrNull()
                ?: return@get call.respond(HttpStatusCode.BadRequest, mapOf("error" to "storeId is required"))
            val minCapacity = call.request.queryParameters["minCapacity"]?.toIntOrNull() ?: 1
            call.respond(service.listAvailable(storeId, minCapacity))
        }

        route("/assignments") {
            post {
                val request = call.receive<ArrangeSpaceRequest>()
                try {
                    val assignment = service.arrange(
                        storeId = request.storeId,
                        referenceId = request.referenceId,
                        partySize = request.partySize,
                        from = request.from,
                        to = request.to
                    )
                    call.respond(HttpStatusCode.Created, assignment)
                } catch (e: NoSpaceAvailableException) {
                    call.respond(HttpStatusCode.Conflict, mapOf("error" to e.message))
                }
            }

            get {
                val storeId = call.request.queryParameters["storeId"]?.toIntOrNull()
                    ?: return@get call.respond(HttpStatusCode.BadRequest, mapOf("error" to "storeId is required"))
                call.respond(service.listAssignments(storeId))
            }

            delete("/{id}") {
                val id = call.parameters["id"]?.toIntOrNull()
                    ?: return@delete call.respond(HttpStatusCode.BadRequest, mapOf("error" to "invalid id"))
                service.release(id)
                call.respond(HttpStatusCode.OK)
            }
        }
    }
}
