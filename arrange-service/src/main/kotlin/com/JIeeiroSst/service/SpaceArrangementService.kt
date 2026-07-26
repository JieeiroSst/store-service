package com.JIeeiroSst.service

import com.JIeeiroSst.domain.space.Space
import com.JIeeiroSst.domain.space.SpaceAssignment
import com.JIeeiroSst.domain.space.SpaceAssignmentStatus
import com.JIeeiroSst.domain.space.SpaceRepository

class NoSpaceAvailableException(message: String) : Exception(message)

/**
 * Arranges a table/room for a party: picks the smallest available space that
 * fits the party size and has no conflicting assignment in the requested
 * time window.
 */
class SpaceArrangementService(private val repository: SpaceRepository) {

    suspend fun arrange(
        storeId: Int,
        referenceId: Int,
        partySize: Int,
        from: String,
        to: String
    ): SpaceAssignment {
        val candidates = repository.listByStore(storeId, partySize)
        for (space in candidates) {
            val overlapping = repository.findAssignmentsOverlapping(space.id, from, to)
            if (overlapping.isEmpty()) {
                return repository.createAssignment(
                    SpaceAssignment(
                        spaceId = space.id,
                        referenceId = referenceId,
                        partySize = partySize,
                        arrangedFrom = from,
                        arrangedTo = to
                    )
                )
            }
        }
        throw NoSpaceAvailableException("No space available for party of $partySize at store $storeId between $from and $to")
    }

    suspend fun listAvailable(storeId: Int, minCapacity: Int): List<Space> = repository.listByStore(storeId, minCapacity)

    suspend fun listAssignments(storeId: Int) = repository.listAssignmentsByStore(storeId)

    suspend fun release(id: Int) {
        repository.updateAssignmentStatus(id, SpaceAssignmentStatus.RELEASED)
    }
}
