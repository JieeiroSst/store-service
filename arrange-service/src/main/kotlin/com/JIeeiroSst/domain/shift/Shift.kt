package com.JIeeiroSst.domain.shift

import kotlinx.serialization.Serializable

/**
 * A work shift assigned to an employee at a store.
 *
 * [startTime]/[endTime] are ISO-8601 UTC instants (e.g. "2024-01-01T09:00:00Z"),
 * kept as fixed-width strings so overlap checks can compare them lexically
 * without pulling in a date/time library.
 */
@Serializable
data class Shift(
    val id: Int = 0,
    val employeeId: Int,
    val storeId: Int,
    val role: String,
    val startTime: String,
    val endTime: String,
    val status: String = ShiftStatus.SCHEDULED
)

object ShiftStatus {
    const val SCHEDULED = "SCHEDULED"
    const val CANCELLED = "CANCELLED"
}
