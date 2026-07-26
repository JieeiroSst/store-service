package com.JIeeiroSst

import com.JIeeiroSst.repository.jdbc.JdbcDispatchRepository
import com.JIeeiroSst.repository.jdbc.JdbcShiftRepository
import com.JIeeiroSst.repository.jdbc.JdbcSpaceRepository
import com.JIeeiroSst.routes.dispatchRoutes
import com.JIeeiroSst.routes.shiftRoutes
import com.JIeeiroSst.routes.spaceRoutes
import com.JIeeiroSst.service.DispatchArrangementService
import com.JIeeiroSst.service.ShiftArrangementService
import com.JIeeiroSst.service.SpaceArrangementService
import io.ktor.http.*
import io.ktor.server.application.*
import io.ktor.server.auth.*
import io.ktor.server.resources.*
import io.ktor.server.response.*
import io.ktor.server.routing.*
import java.sql.Connection

/**
 * Composition root: wires the JDBC adapters, arrangement services and HTTP
 * routes for the three bounded contexts this service owns:
 *  - shift    -- employee shift scheduling
 *  - space    -- table/room arrangement
 *  - dispatch -- delivery order/driver route sequencing
 */
fun Application.configureRouting(connection: Connection) {
    install(Resources)

    val shiftService = ShiftArrangementService(JdbcShiftRepository(connection))
    val spaceService = SpaceArrangementService(JdbcSpaceRepository(connection))
    val dispatchService = DispatchArrangementService(JdbcDispatchRepository(connection))

    routing {
        get("/") {
            call.respondText("arrange-service is up")
        }
        get("/health") {
            call.respond(HttpStatusCode.OK, mapOf("status" to "UP"))
        }

        authenticate(ARRANGE_AUTH) {
            shiftRoutes(shiftService)
            spaceRoutes(spaceService)
            dispatchRoutes(dispatchService)
        }
    }
}
