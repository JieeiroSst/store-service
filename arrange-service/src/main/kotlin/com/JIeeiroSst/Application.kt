package com.JIeeiroSst

import io.ktor.server.application.*

fun main(args: Array<String>) {
    io.ktor.server.netty.EngineMain.main(args)
}

fun Application.module() {
    configureSerialization()
    val connection = configureDatabases()
    configureHTTP()
    configureSecurity()
    configureAdministration()
    configureRouting(connection)
}
