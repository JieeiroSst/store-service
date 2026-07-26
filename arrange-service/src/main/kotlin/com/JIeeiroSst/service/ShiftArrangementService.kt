package com.JIeeiroSst.service

import com.JIeeiroSst.domain.shift.Shift
import com.JIeeiroSst.domain.shift.ShiftRepository
import com.JIeeiroSst.domain.shift.ShiftStatus

class ShiftConflictException(message: String) : Exception(message)

/** Arranges employee shifts, rejecting anything that overlaps a shift the employee already has. */
class ShiftArrangementService(private val repository: ShiftRepository) {

    suspend fun arrange(shift: Shift): Shift {
        val overlapping = repository.findOverlapping(shift.employeeId, shift.startTime, shift.endTime)
        if (overlapping.isNotEmpty()) {
            throw ShiftConflictException(
                "Employee ${shift.employeeId} already has an overlapping shift (id=${overlapping.first().id})"
            )
        }
        return repository.create(shift)
    }

    suspend fun listByStore(storeId: Int): List<Shift> = repository.listByStore(storeId)

    suspend fun cancel(id: Int) {
        repository.updateStatus(id, ShiftStatus.CANCELLED)
    }
}
