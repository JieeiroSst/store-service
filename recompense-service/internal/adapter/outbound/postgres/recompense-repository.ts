import type { Pool } from "pg";

import type { Recompense } from "../../../domain/recompense";
import type { RecompenseRepository } from "../../../application/port/outbound/recompense-repository";

const columns = `id, member_id, source, source_id, tier, points_per_purchase, minimum_points,
    code_de_reduction, reduction_description, reduction_type, reduction_value, max_reduction, expiry_date`;

function toModel(row: any): Recompense {
    return {
        ID: String(row.id),
        MemberID: Number(row.member_id),
        Source: row.source,
        SourceID: row.source_id,
        Tier: row.tier ?? undefined,
        PointsPerPurchase: row.points_per_purchase ?? undefined,
        MinimumPoints: row.minimum_points ?? undefined,
        CodeDeReduction: row.code_de_reduction ?? undefined,
        ReductionDescription: row.reduction_description ?? undefined,
        ReductionType: row.reduction_type,
        MaxReduction: row.max_reduction === null ? undefined : Number(row.max_reduction),
        ReductionValue: row.reduction_value === null ? undefined : Number(row.reduction_value),
        ExpiryDate: row.expiry_date === null ? undefined : Number(row.expiry_date),
    };
}

export class PostgresRecompenseRepository implements RecompenseRepository {
    constructor(private readonly pool: Pool) { }

    async insert(p: Recompense): Promise<Recompense> {
        const result = await this.pool.query(
            `INSERT INTO recompense
                (id, member_id, source, source_id, tier, points_per_purchase, minimum_points,
                 code_de_reduction, reduction_description, reduction_type, reduction_value, max_reduction, expiry_date)
             VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
             RETURNING ${columns}`,
            [p.ID, p.MemberID, p.Source ?? '', p.SourceID ?? '', p.Tier, p.PointsPerPurchase,
            p.MinimumPoints, p.CodeDeReduction, p.ReductionDescription, p.ReductionType ?? 'fixed', p.ReductionValue,
            p.MaxReduction, p.ExpiryDate],
        );
        return toModel(result.rows[0]);
    }

    async update(id: string, p: Recompense): Promise<Recompense | undefined> {
        const result = await this.pool.query(
            `UPDATE recompense SET member_id = $1, source = $2, source_id = $3, tier = $4,
                points_per_purchase = $5, minimum_points = $6, code_de_reduction = $7,
                reduction_description = $8, reduction_type = $9, reduction_value = $10,
                max_reduction = $11, expiry_date = $12
             WHERE id = $13
             RETURNING ${columns}`,
            [p.MemberID, p.Source ?? '', p.SourceID ?? '', p.Tier, p.PointsPerPurchase, p.MinimumPoints,
            p.CodeDeReduction, p.ReductionDescription, p.ReductionType ?? 'fixed', p.ReductionValue, p.MaxReduction,
            p.ExpiryDate, id],
        );
        return result.rows[0] ? toModel(result.rows[0]) : undefined;
    }

    async getByID(id: string): Promise<Recompense | undefined> {
        const result = await this.pool.query(`SELECT ${columns} FROM recompense WHERE id = $1`, [id]);
        return result.rows[0] ? toModel(result.rows[0]) : undefined;
    }

    async getByMemberID(memberId: number, limit: number, offset: number): Promise<Recompense[]> {
        const result = await this.pool.query(
            `SELECT ${columns} FROM recompense WHERE member_id = $1 ORDER BY id DESC LIMIT $2 OFFSET $3`,
            [memberId, limit, offset],
        );
        return result.rows.map(toModel);
    }

    async delete(id: string): Promise<boolean> {
        const result = await this.pool.query(`DELETE FROM recompense WHERE id = $1`, [id]);
        return (result.rowCount ?? 0) > 0;
    }

    async list(limit: number, offset: number): Promise<Recompense[]> {
        const result = await this.pool.query(
            `SELECT ${columns} FROM recompense ORDER BY id DESC LIMIT $1 OFFSET $2`,
            [limit, offset],
        );
        return result.rows.map(toModel);
    }

    async ping(): Promise<void> {
        await this.pool.query('SELECT 1');
    }
}

