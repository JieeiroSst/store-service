import sql from "../pkg/postgres";

import { Recompense } from "../model/recompense";
import id from "../pkg/snowflake";

async function insertRecompense(params: Recompense) {
    const insertQuery = `
    insert into recompense
        (member_id, tier, points_per_purchase,minimum_points,
            code_de_reduction,reduction_description,
            reduction_value,expiry_date
        ) values($1, $2, $3, $4, $5, $6, $7, $8)
    `;
    try {
        await sql.connect();
        const result = await sql.query<Recompense>(insertQuery, [id, params.Tier, params.PointsPerPurchase, params.MinimumPoints, params.CodeDeReduction, params.ReductionDescription, params.ReductionValue, params.ExpiryDate]);
        await sql.end();
        return result.rows[0];
    } catch (error) {
        console.error('Error inserting Recompense:', error);
    }
}

async function updateRecompense(params: Recompense) {
    const updateQuery = `UPDATE recompense SET tier = $1, points_per_purchase = $2,minimum_points =$3,
    code_de_reduction = $4 ,reduction_description = $5,
    reduction_value = $6, expiry_date = $7 WHERE member_id = $8`;
    try {
        await sql.connect();
        const result = await sql.query<Recompense>(updateQuery, [params.Tier, params.PointsPerPurchase, params.MinimumPoints, params.CodeDeReduction, params.ReductionDescription, params.ReductionValue, params.ExpiryDate, params.MemberID]);
        await sql.end();
        return result.rows[0];
    } catch (error) {
        console.error('Error update Recompense:', error);
    }
}

async function getRecompenseByMemberID(memberId: number) {
    const getMemberIDQuery = `select * from recompense where member_id = $1`;
    try {
        await sql.connect();
        const result = await sql.query<Recompense>(getMemberIDQuery, [memberId]);
        await sql.end();
        return result.rows[0];
    } catch (error) {
        console.error('Error get member id Recompense:', error);
    }
}

async function deleteRecompenseByMemberID(memberId: number) {
    const deleteMemberIDQuery = `DELETE FROM recompense WHERE member_id = $1`;
    try {
        await sql.connect();
        const result = await sql.query<Recompense>(deleteMemberIDQuery, [memberId]);
        await sql.end();
        return result.rows[0];
    } catch (error) {
        console.error('Error get member id Recompense:', error);
    }
}

async function getRecompensePagination(limit: number, page: number) {
    const getMemberIDQuery = `SELECT * FROM recompense LIMIT $1 OFFSET $2`;
    try {
        await sql.connect();
        const result = await sql.query<Recompense>(getMemberIDQuery, [limit, page]);
        await sql.end();
        return result.rows;
    } catch (error) {
        console.error('Error get pagination Recompense:', error);
    }
}

export { insertRecompense, updateRecompense, getRecompenseByMemberID, deleteRecompenseByMemberID, getRecompensePagination }