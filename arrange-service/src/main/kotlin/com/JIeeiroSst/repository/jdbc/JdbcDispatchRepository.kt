package com.JIeeiroSst.repository.jdbc

import com.JIeeiroSst.domain.dispatch.DispatchAssignment
import com.JIeeiroSst.domain.dispatch.DispatchRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.sql.Connection
import java.sql.ResultSet
import java.sql.Statement

class JdbcDispatchRepository(private val connection: Connection) : DispatchRepository {

    companion object {
        private const val CREATE_TABLE = """
            CREATE TABLE IF NOT EXISTS dispatch_assignments (
                id SERIAL PRIMARY KEY,
                order_id INT NOT NULL,
                driver_id INT NOT NULL,
                sequence_no INT NOT NULL,
                status VARCHAR(16) NOT NULL
            );
        """
        private const val COUNT_ACTIVE_FOR_DRIVER =
            "SELECT COUNT(*) FROM dispatch_assignments WHERE driver_id = ? AND status = 'ASSIGNED'"
        private const val INSERT = """
            INSERT INTO dispatch_assignments (order_id, driver_id, sequence_no, status)
            VALUES (?, ?, ?, ?)
        """
        private const val SELECT_BY_DRIVER =
            "SELECT id, order_id, driver_id, sequence_no, status FROM dispatch_assignments WHERE driver_id = ? ORDER BY sequence_no ASC"
        private const val UPDATE_STATUS = "UPDATE dispatch_assignments SET status = ? WHERE id = ?"
    }

    init {
        connection.createStatement().use { it.executeUpdate(CREATE_TABLE) }
    }

    override suspend fun countActiveForDriver(driverId: Int): Int = withContext(Dispatchers.IO) {
        connection.prepareStatement(COUNT_ACTIVE_FOR_DRIVER).use { statement ->
            statement.setInt(1, driverId)
            statement.executeQuery().use { rs -> if (rs.next()) rs.getInt(1) else 0 }
        }
    }

    override suspend fun createAssignment(assignment: DispatchAssignment): DispatchAssignment =
        withContext(Dispatchers.IO) {
            connection.prepareStatement(INSERT, Statement.RETURN_GENERATED_KEYS).use { statement ->
                statement.setInt(1, assignment.orderId)
                statement.setInt(2, assignment.driverId)
                statement.setInt(3, assignment.sequence)
                statement.setString(4, assignment.status)
                statement.executeUpdate()

                val keys = statement.generatedKeys
                if (keys.next()) {
                    assignment.copy(id = keys.getInt(1))
                } else {
                    throw IllegalStateException("Unable to retrieve the id of the newly inserted dispatch assignment")
                }
            }
        }

    override suspend fun listByDriver(driverId: Int): List<DispatchAssignment> = withContext(Dispatchers.IO) {
        connection.prepareStatement(SELECT_BY_DRIVER).use { statement ->
            statement.setInt(1, driverId)
            statement.executeQuery().use { rs ->
                val assignments = mutableListOf<DispatchAssignment>()
                while (rs.next()) assignments.add(rs.toAssignment())
                assignments
            }
        }
    }

    override suspend fun updateStatus(id: Int, status: String): Unit = withContext(Dispatchers.IO) {
        connection.prepareStatement(UPDATE_STATUS).use { statement ->
            statement.setString(1, status)
            statement.setInt(2, id)
            statement.executeUpdate()
        }
    }

    private fun ResultSet.toAssignment() = DispatchAssignment(
        id = getInt("id"),
        orderId = getInt("order_id"),
        driverId = getInt("driver_id"),
        sequence = getInt("sequence_no"),
        status = getString("status")
    )
}
