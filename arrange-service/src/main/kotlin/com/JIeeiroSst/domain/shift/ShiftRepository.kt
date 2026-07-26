package com.JIeeiroSst.domain.shift

/** Driven port for shift persistence, implemented by the JDBC adapter. */
interface ShiftRepository {
    suspend fun create(shift: Shift): Shift
    suspend fun findById(id: Int): Shift?
    suspend fun findOverlapping(employeeId: Int, startTime: String, endTime: String): List<Shift>
    suspend fun listByStore(storeId: Int): List<Shift>
    suspend fun updateStatus(id: Int, status: String)
}
