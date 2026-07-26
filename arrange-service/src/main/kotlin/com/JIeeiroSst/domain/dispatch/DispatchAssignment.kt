package com.JIeeiroSst.domain.dispatch

import kotlinx.serialization.Serializable

/**
 * The delivery-route position ([sequence]) assigned to an order for a given
 * driver, used to order/coordinate multiple deliveries assigned to the same
 * driver.
 */
@Serializable
data class DispatchAssignment(
    val id: Int = 0,
    val orderId: Int,
    val driverId: Int,
    val sequence: Int,
    val status: String = DispatchStatus.ASSIGNED
)

object DispatchStatus {
    const val ASSIGNED = "ASSIGNED"
    const val DELIVERED = "DELIVERED"
}
