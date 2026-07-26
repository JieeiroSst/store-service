package com.JIeeiroSst.routes

import com.JIeeiroSst.service.ShiftArrangementService
import com.JIeeiroSst.service.ShiftConflictException
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.request.*
import io.ktor.server.response.*
import io.ktor.server.routing.*

fun Route.shiftRoutes(service: ShiftArrangementService) {
    route("/shifts") {
        post {
            val shift = call.receive<com.JIeeiroSst.domain.shift.Shift>()
            try {
                val arranged = service.arrange(shift)
                call.respond(HttpStatusCode.Created, arranged)
            } catch (e: ShiftConflictException) {
                call.respond(HttpStatusCode.Conflict, mapOf("error" to e.message))
            }
        }

        get {
            val storeId = call.request.queryParameters["storeId"]?.toIntOrNull()
                ?: return@get call.respond(HttpStatusCode.BadRequest, mapOf("error" to "storeId is required"))
            call.respond(service.listByStore(storeId))
        }

        delete("/{id}") {
            val id = call.parameters["id"]?.toIntOrNull()
                ?: return@delete call.respond(HttpStatusCode.BadRequest, mapOf("error" to "invalid id"))
            service.cancel(id)
            call.respond(HttpStatusCode.OK)
        }
    }
}
