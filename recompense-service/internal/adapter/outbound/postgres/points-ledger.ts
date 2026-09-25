import type { Pool } from "pg";

import type { EarnResult, PointsLedger } from "../../../application/port/outbound/points-ledger";

export class PostgresPointsLedger implements PointsLedger {
    constructor(private readonly pool: Pool) { }

    async earn(id: string, memberId: number, key: string, points: number): Promise<EarnResult> {
        const client = await this.pool.connect();
        try {
            await client.query("BEGIN");
            await client.query("SELECT pg_advisory_xact_lock($1)", [memberId]);
            const inserted = await client.query(
                `INSERT INTO points_ledger (id, member_id, idempotency_key, points)
                 VALUES ($1, $2, $3, $4) ON CONFLICT (member_id, idempotency_key) DO NOTHING`,
                [id, memberId, key, points],
            );
            const duplicate = (inserted.rowCount ?? 0) === 0;
            const credited = duplicate
                ? Number((await client.query(
                    "SELECT points FROM points_ledger WHERE member_id = $1 AND idempotency_key = $2", [memberId, key],
                )).rows[0].points)
                : points;
            const sum = await client.query("SELECT COALESCE(SUM(points), 0) AS balance FROM points_ledger WHERE member_id = $1", [memberId]);
            await client.query("COMMIT");
            return { points: credited, balance: Number(sum.rows[0].balance), duplicate };
        } catch (error) {
            await client.query("ROLLBACK").catch(() => { });
            throw error;
        } finally {
            client.release();
        }
    }

    async balance(memberId: number): Promise<number> {
        const result = await this.pool.query("SELECT COALESCE(SUM(points), 0) AS balance FROM points_ledger WHERE member_id = $1", [memberId]);
        return Number(result.rows[0].balance);
    }
}
