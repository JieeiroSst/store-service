package com.JIeeiroSst.service

import com.JIeeiroSst.domain.dispatch.DispatchAssignment
import com.JIeeiroSst.domain.dispatch.DispatchRepository
import com.JIeeiroSst.domain.dispatch.DispatchStatus

/**
 * Coordinates delivery order dispatch: assigns an order to a driver and gives
 * it the next sequence position in that driver's route.
 */
class DispatchArrangementService(private val repository: DispatchRepository) {

    suspend fun arrange(orderId: Int, driverId: Int): DispatchAssignment {
        val nextSequence = repository.countActiveForDriver(driverId) + 1
        return repository.createAssignment(
            DispatchAssignment(orderId = orderId, driverId = driverId, sequence = nextSequence)
        )
    }

    suspend fun listForDriver(driverId: Int): List<DispatchAssignment> = repository.listByDriver(driverId)

    suspend fun markDelivered(id: Int) {
        repository.updateStatus(id, DispatchStatus.DELIVERED)
    }
}
