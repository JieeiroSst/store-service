package com.JIeeiroSst.repository.jdbc

import com.JIeeiroSst.domain.space.Space
import com.JIeeiroSst.domain.space.SpaceAssignment
import com.JIeeiroSst.domain.space.SpaceRepository
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.sql.Connection
import java.sql.ResultSet
import java.sql.Statement

class JdbcSpaceRepository(private val connection: Connection) : SpaceRepository {

    companion object {
        private const val CREATE_SPACES_TABLE = """
            CREATE TABLE IF NOT EXISTS spaces (
                id SERIAL PRIMARY KEY,
                store_id INT NOT NULL,
                name VARCHAR(128) NOT NULL,
                type VARCHAR(16) NOT NULL,
                capacity INT NOT NULL
            );
        """
        private const val CREATE_ASSIGNMENTS_TABLE = """
            CREATE TABLE IF NOT EXISTS space_assignments (
                id SERIAL PRIMARY KEY,
                space_id INT NOT NULL,
                reference_id INT NOT NULL,
                party_size INT NOT NULL,
                arranged_from VARCHAR(32) NOT NULL,
                arranged_to VARCHAR(32) NOT NULL,
                status VARCHAR(16) NOT NULL
            );
        """
        private const val SELECT_SPACES_BY_STORE =
            "SELECT id, store_id, name, type, capacity FROM spaces WHERE store_id = ? AND capacity >= ? ORDER BY capacity ASC"
        private const val SELECT_ASSIGNMENTS_OVERLAPPING = """
            SELECT id, space_id, reference_id, party_size, arranged_from, arranged_to, status FROM space_assignments
            WHERE space_id = ? AND status = 'ARRANGED' AND arranged_from < ? AND arranged_to > ?
        """
        private const val INSERT_ASSIGNMENT = """
            INSERT INTO space_assignments (space_id, reference_id, party_size, arranged_from, arranged_to, status)
            VALUES (?, ?, ?, ?, ?, ?)
        """
        private const val UPDATE_ASSIGNMENT_STATUS = "UPDATE space_assignments SET status = ? WHERE id = ?"
        private const val SELECT_ASSIGNMENTS_BY_STORE = """
            SELECT a.id, a.space_id, a.reference_id, a.party_size, a.arranged_from, a.arranged_to, a.status
            FROM space_assignments a JOIN spaces s ON s.id = a.space_id
            WHERE s.store_id = ?
        """
    }

    init {
        connection.createStatement().use { it.executeUpdate(CREATE_SPACES_TABLE) }
        connection.createStatement().use { it.executeUpdate(CREATE_ASSIGNMENTS_TABLE) }
    }

    override suspend fun listByStore(storeId: Int, minCapacity: Int): List<Space> = withContext(Dispatchers.IO) {
        connection.prepareStatement(SELECT_SPACES_BY_STORE).use { statement ->
            statement.setInt(1, storeId)
            statement.setInt(2, minCapacity)
            statement.executeQuery().use { rs ->
                val spaces = mutableListOf<Space>()
                while (rs.next()) spaces.add(rs.toSpace())
                spaces
            }
        }
    }

    override suspend fun findAssignmentsOverlapping(spaceId: Int, from: String, to: String): List<SpaceAssignment> =
        withContext(Dispatchers.IO) {
            connection.prepareStatement(SELECT_ASSIGNMENTS_OVERLAPPING).use { statement ->
                statement.setInt(1, spaceId)
                statement.setString(2, to)
                statement.setString(3, from)
                statement.executeQuery().use { rs ->
                    val assignments = mutableListOf<SpaceAssignment>()
                    while (rs.next()) assignments.add(rs.toAssignment())
                    assignments
                }
            }
        }

    override suspend fun createAssignment(assignment: SpaceAssignment): SpaceAssignment = withContext(Dispatchers.IO) {
        connection.prepareStatement(INSERT_ASSIGNMENT, Statement.RETURN_GENERATED_KEYS).use { statement ->
            statement.setInt(1, assignment.spaceId)
            statement.setInt(2, assignment.referenceId)
            statement.setInt(3, assignment.partySize)
            statement.setString(4, assignment.arrangedFrom)
            statement.setString(5, assignment.arrangedTo)
            statement.setString(6, assignment.status)
            statement.executeUpdate()

            val keys = statement.generatedKeys
            if (keys.next()) {
                assignment.copy(id = keys.getInt(1))
            } else {
                throw IllegalStateException("Unable to retrieve the id of the newly inserted space assignment")
            }
        }
    }

    override suspend fun updateAssignmentStatus(id: Int, status: String): Unit = withContext(Dispatchers.IO) {
        connection.prepareStatement(UPDATE_ASSIGNMENT_STATUS).use { statement ->
            statement.setString(1, status)
            statement.setInt(2, id)
            statement.executeUpdate()
        }
    }

    override suspend fun listAssignmentsByStore(storeId: Int): List<SpaceAssignment> = withContext(Dispatchers.IO) {
        connection.prepareStatement(SELECT_ASSIGNMENTS_BY_STORE).use { statement ->
            statement.setInt(1, storeId)
            statement.executeQuery().use { rs ->
                val assignments = mutableListOf<SpaceAssignment>()
                while (rs.next()) assignments.add(rs.toAssignment())
                assignments
            }
        }
    }

    private fun ResultSet.toSpace() = Space(
        id = getInt("id"),
        storeId = getInt("store_id"),
        name = getString("name"),
        type = getString("type"),
        capacity = getInt("capacity")
    )

    private fun ResultSet.toAssignment() = SpaceAssignment(
        id = getInt("id"),
        spaceId = getInt("space_id"),
        referenceId = getInt("reference_id"),
        partySize = getInt("party_size"),
        arrangedFrom = getString("arranged_from"),
        arrangedTo = getString("arranged_to"),
        status = getString("status")
    )
}
