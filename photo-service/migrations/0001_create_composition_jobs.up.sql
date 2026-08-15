CREATE TABLE IF NOT EXISTS composition_jobs (
    id                 UUID PRIMARY KEY,
    status             TEXT NOT NULL,
    layout             JSONB NOT NULL,
    sources            JSONB NOT NULL,
    idempotency_key    TEXT,
    result_object_key  TEXT,
    result_url         TEXT,
    width              INT,
    height             INT,
    format             TEXT,
    size_bytes         BIGINT,
    error_message      TEXT,
    created_at         TIMESTAMPTZ NOT NULL,
    updated_at         TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_composition_jobs_idempotency_key
    ON composition_jobs (idempotency_key)
    WHERE idempotency_key IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_composition_jobs_status
    ON composition_jobs (status);
