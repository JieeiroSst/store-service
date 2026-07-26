package com.JIeeiroSst.domain.space

/** Driven port for space/table persistence, implemented by the JDBC adapter. */
interface SpaceRepository {
    suspend fun listByStore(storeId: Int, minCapacity: Int): List<Space>
    suspend fun findAssignmentsOverlapping(spaceId: Int, from: String, to: String): List<SpaceAssignment>
    suspend fun createAssignment(assignment: SpaceAssignment): SpaceAssignment
    suspend fun updateAssignmentStatus(id: Int, status: String)
    suspend fun listAssignmentsByStore(storeId: Int): List<SpaceAssignment>
}
