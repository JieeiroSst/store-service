ALTER TABLE composition_jobs ADD COLUMN IF NOT EXISTS owner_id TEXT;

CREATE INDEX IF NOT EXISTS idx_composition_jobs_owner_created
    ON composition_jobs (owner_id, created_at DESC)
    WHERE owner_id IS NOT NULL;
