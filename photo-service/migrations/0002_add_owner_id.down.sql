DROP INDEX IF EXISTS idx_composition_jobs_owner_created;
ALTER TABLE composition_jobs DROP COLUMN IF EXISTS owner_id;
