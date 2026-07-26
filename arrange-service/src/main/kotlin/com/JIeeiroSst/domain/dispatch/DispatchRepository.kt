package com.JIeeiroSst.domain.dispatch

/** Driven port for dispatch persistence, implemented by the JDBC adapter. */
interface DispatchRepository {
    suspend fun countActiveForDriver(driverId: Int): Int
    suspend fun createAssignment(assignment: DispatchAssignment): DispatchAssignment
    suspend fun listByDriver(driverId: Int): List<DispatchAssignment>
    suspend fun updateStatus(id: Int, status: String)
}
