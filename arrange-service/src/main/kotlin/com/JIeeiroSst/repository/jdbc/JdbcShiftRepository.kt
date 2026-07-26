package com.JIeeiroSst.repository.jdbc

import com.JIeeiroSst.domain.shift.Shift
import com.JIeeiroSst.domain.shift.ShiftRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.sql.Connection
import java.sql.Statement

class JdbcShiftRepository(private val connection: Connection) : ShiftRepository {

    companion object {
        private const val CREATE_TABLE = """
            CREATE TABLE IF NOT EXISTS shifts (
                id SERIAL PRIMARY KEY,
                employee_id INT NOT NULL,
                store_id INT NOT NULL,
                role VARCHAR(64) NOT NULL,
                start_time VARCHAR(32) NOT NULL,
                end_time VARCHAR(32) NOT NULL,
                status VARCHAR(16) NOT NULL
            );
        """
        private const val INSERT = """
            INSERT INTO shifts (employee_id, store_id, role, start_time, end_time, status)
            VALUES (?, ?, ?, ?, ?, ?)
        """
        private const val SELECT_BY_ID =
            "SELECT id, employee_id, store_id, role, start_time, end_time, status FROM shifts WHERE id = ?"
        private const val SELECT_OVERLAPPING = """
            SELECT id, employee_id, store_id, role, start_time, end_time, status FROM shifts
            WHERE employee_id = ? AND status = 'SCHEDULED' AND start_time < ? AND end_time > ?
        """
        private const val SELECT_BY_STORE =
            "SELECT id, employee_id, store_id, role, start_time, end_time, status FROM shifts WHERE store_id = ?"
        private const val UPDATE_STATUS = "UPDATE shifts SET status = ? WHERE id = ?"
    }

    init {
        connection.createStatement().use { it.executeUpdate(CREATE_TABLE) }
    }

    override suspend fun create(shift: Shift): Shift = withContext(Dispatchers.IO) {
        connection.prepareStatement(INSERT, Statement.RETURN_GENERATED_KEYS).use { statement ->
            statement.setInt(1, shift.employeeId)
            statement.setInt(2, shift.storeId)
            statement.setString(3, shift.role)
            statement.setString(4, shift.startTime)
            statement.setString(5, shift.endTime)
            statement.setString(6, shift.status)
            statement.executeUpdate()

            val keys = statement.generatedKeys
            if (keys.next()) {
                shift.copy(id = keys.getInt(1))
            } else {
                throw IllegalStateException("Unable to retrieve the id of the newly inserted shift")
            }
        }
    }

    override suspend fun findById(id: Int): Shift? = withContext(Dispatchers.IO) {
        connection.prepareStatement(SELECT_BY_ID).use { statement ->
            statement.setInt(1, id)
            statement.executeQuery().use { rs -> if (rs.next()) rs.toShift() else null }
        }
    }

    override suspend fun findOverlapping(employeeId: Int, startTime: String, endTime: String): List<Shift> =
        withContext(Dispatchers.IO) {
            connection.prepareStatement(SELECT_OVERLAPPING).use { statement ->
                statement.setInt(1, employeeId)
                statement.setString(2, endTime)
                statement.setString(3, startTime)
                statement.executeQuery().use { rs ->
                    val shifts = mutableListOf<Shift>()
                    while (rs.next()) shifts.add(rs.toShift())
                    shifts
                }
            }
        }

    override suspend fun listByStore(storeId: Int): List<Shift> = withContext(Dispatchers.IO) {
        connection.prepareStatement(SELECT_BY_STORE).use { statement ->
            statement.setInt(1, storeId)
            statement.executeQuery().use { rs ->
                val shifts = mutableListOf<Shift>()
                while (rs.next()) shifts.add(rs.toShift())
                shifts
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

    private fun java.sql.ResultSet.toShift() = Shift(
        id = getInt("id"),
        employeeId = getInt("employee_id"),
        storeId = getInt("store_id"),
        role = getString("role"),
        startTime = getString("start_time"),
        endTime = getString("end_time"),
        status = getString("status")
    )
}
