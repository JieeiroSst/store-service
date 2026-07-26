package com.JIeeiroSst.routes

import com.JIeeiroSst.service.DispatchArrangementService
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import kotlinx.serialization.Serializable

@Serializable
data class ArrangeDispatchRequest(val orderId: Int, val driverId: Int)

fun Route.dispatchRoutes(service: DispatchArrangementService) {
    route("/dispatch/assignments") {
        post {
            val request = call.receive<ArrangeDispatchRequest>()
            val assignment = service.arrange(request.orderId, request.driverId)
            call.respond(HttpStatusCode.Created, assignment)
        }

        get {
            val driverId = call.request.queryParameters["driverId"]?.toIntOrNull()
                ?: return@get call.respond(HttpStatusCode.BadRequest, mapOf("error" to "driverId is required"))
            call.respond(service.listForDriver(driverId))
        }

        patch("/{id}/delivered") {
            val id = call.parameters["id"]?.toIntOrNull()
                ?: return@patch call.respond(HttpStatusCode.BadRequest, mapOf("error" to "invalid id"))
            service.markDelivered(id)
            call.respond(HttpStatusCode.OK)
        }
    }
}
