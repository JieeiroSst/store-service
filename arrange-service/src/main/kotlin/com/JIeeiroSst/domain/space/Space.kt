package com.JIeeiroSst.domain.space

import kotlinx.serialization.Serializable

/** A physical table or room at a store that can be arranged for a party. */
@Serializable
data class Space(
    val id: Int = 0,
    val storeId: Int,
    val name: String,
    val type: String,
    val capacity: Int
)

object SpaceType {
    const val TABLE = "TABLE"
    const val ROOM = "ROOM"
}

/**
 * A time-boxed assignment of a [Space] to a reservation/order (identified by
 * [referenceId]). [arrangedFrom]/[arrangedTo] are ISO-8601 UTC instants, see
 * the note on [com.JIeeiroSst.domain.shift.Shift] for why they're strings.
 */
@Serializable
data class SpaceAssignment(
    val id: Int = 0,
    val spaceId: Int,
    val referenceId: Int,
    val partySize: Int,
    val arrangedFrom: String,
    val arrangedTo: String,
    val status: String = SpaceAssignmentStatus.ARRANGED
)

object SpaceAssignmentStatus {
    const val ARRANGED = "ARRANGED"
    const val RELEASED = "RELEASED"
}
