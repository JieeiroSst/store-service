import { Pool } from 'pg';

import type { Postgres } from '../../../../config';

const schema = `
CREATE TABLE IF NOT EXISTS recompense (
    id                    BIGINT PRIMARY KEY,
    member_id             BIGINT NOT NULL,
    source                TEXT NOT NULL DEFAULT '',
    source_id             TEXT NOT NULL DEFAULT '',
    tier                  INT,
    points_per_purchase   INT,
    minimum_points        INT,
    code_de_reduction     TEXT,
    reduction_description TEXT,
    reduction_value       NUMERIC,
    expiry_date           BIGINT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
ALTER TABLE recompense ADD COLUMN IF NOT EXISTS reduction_type TEXT NOT NULL DEFAULT 'fixed';
ALTER TABLE recompense ADD COLUMN IF NOT EXISTS max_reduction NUMERIC;
CREATE INDEX IF NOT EXISTS recompense_member_id_idx ON recompense (member_id);
CREATE INDEX IF NOT EXISTS recompense_source_idx ON recompense (source, source_id);

CREATE TABLE IF NOT EXISTS points_ledger (
    id              BIGINT PRIMARY KEY,
    member_id       BIGINT NOT NULL,
    idempotency_key TEXT NOT NULL,
    points          BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (member_id, idempotency_key)
);
`;

export function newPool(cfg: Postgres, database: string = cfg.Database): Pool {
    return new Pool({
        host: cfg.Host,
        port: cfg.Port,
        user: cfg.User,
        password: cfg.Password,
        database,
        max: 10,
    });
}

export async function migrate(cfg: Postgres, pool: Pool): Promise<void> {
    const admin = newPool(cfg, 'postgres');
    try {
        const exists = await admin.query('SELECT 1 FROM pg_database WHERE datname = $1', [cfg.Database]);
        if (exists.rowCount === 0) {
            await admin.query(`CREATE DATABASE "${cfg.Database.replace(/"/g, '""')}"`);
        }
    } catch (error) {
        console.warn('ensure database:', error);
    } finally {
        await admin.end();
    }
    const client = await pool.connect();
    try {
        await client.query('SELECT pg_advisory_lock(727001)');
        await client.query(schema);
    } finally {
        await client.query('SELECT pg_advisory_unlock(727001)').catch(() => { });
        client.release();
    }
}
