-- +goose Up
-- +goose StatementBegin

-- Users (recruiters, hiring managers, interviewers, partners) live in user-service, whose
-- ids are BIGINT. Candidates are not accounts and keep their UUIDs.
--
-- Fresh databases convert directly. A database with existing rows needs a
-- legacy_user_map(legacy_id UUID PRIMARY KEY, user_id BIGINT NOT NULL) that
-- maps every old UUID to its user-service id (e.g. matched by email) before
-- this migration runs; without it the migration stops instead of guessing.
DO $$
DECLARE
    has_rows boolean;
    has_map  boolean;
BEGIN
    SELECT EXISTS (SELECT 1 FROM jobs) OR EXISTS (SELECT 1 FROM applications)
        OR EXISTS (SELECT 1 FROM partners)
      INTO has_rows;
    SELECT to_regclass('legacy_user_map') IS NOT NULL INTO has_map;

    IF has_rows AND NOT has_map THEN
        RAISE EXCEPTION 'create and fill legacy_user_map(legacy_id uuid, user_id bigint) before migrating existing user references';
    END IF;
    IF NOT has_map THEN
        CREATE TEMP TABLE legacy_user_map (legacy_id UUID PRIMARY KEY, user_id BIGINT NOT NULL) ON COMMIT DROP;
    END IF;
END $$;

DROP VIEW IF EXISTS v_sla_breaches;
DROP VIEW IF EXISTS v_recruiter_performance;
DROP INDEX IF EXISTS idx_applications_status_recruiter;
DROP INDEX IF EXISTS idx_applications_recruiter;
DROP INDEX IF EXISTS idx_partners_user_id;

ALTER TABLE applications ADD COLUMN recruiter_id_new BIGINT;
UPDATE applications t SET recruiter_id_new = m.user_id FROM legacy_user_map m WHERE m.legacy_id = t.recruiter_id;
ALTER TABLE applications DROP COLUMN recruiter_id;
ALTER TABLE applications RENAME COLUMN recruiter_id_new TO recruiter_id;
-- NOT NULL fails here if any old id had no mapping.
ALTER TABLE applications ALTER COLUMN recruiter_id SET NOT NULL;

ALTER TABLE jobs ADD COLUMN hiring_manager_id_new BIGINT;
UPDATE jobs t SET hiring_manager_id_new = m.user_id FROM legacy_user_map m WHERE m.legacy_id = t.hiring_manager_id;
ALTER TABLE jobs DROP COLUMN hiring_manager_id;
ALTER TABLE jobs RENAME COLUMN hiring_manager_id_new TO hiring_manager_id;
-- NOT NULL fails here if any old id had no mapping.
ALTER TABLE jobs ALTER COLUMN hiring_manager_id SET NOT NULL;

ALTER TABLE partners ADD COLUMN user_id_new BIGINT;
UPDATE partners t SET user_id_new = m.user_id FROM legacy_user_map m WHERE m.legacy_id = t.user_id;
ALTER TABLE partners DROP COLUMN user_id;
ALTER TABLE partners RENAME COLUMN user_id_new TO user_id;
-- NOT NULL fails here if any old id had no mapping.
ALTER TABLE partners ALTER COLUMN user_id SET NOT NULL;

UPDATE interviews SET interviewer_ids = COALESCE((
    SELECT jsonb_agg(m.user_id)
    FROM jsonb_array_elements_text(interviews.interviewer_ids) e
    JOIN legacy_user_map m ON m.legacy_id = e::uuid
), '[]'::jsonb);
UPDATE interviews i SET feedback = jsonb_set(i.feedback, '{submitted_by}', to_jsonb(m.user_id))
FROM legacy_user_map m
WHERE i.feedback ? 'submitted_by' AND m.legacy_id = (i.feedback->>'submitted_by')::uuid;
UPDATE jobs SET recruiter_ids = COALESCE((
    SELECT jsonb_agg(m.user_id)
    FROM jsonb_array_elements_text(jobs.recruiter_ids) e
    JOIN legacy_user_map m ON m.legacy_id = e::uuid
), '[]'::jsonb);

CREATE INDEX idx_applications_recruiter ON applications(recruiter_id);
CREATE INDEX IF NOT EXISTS idx_applications_status_recruiter ON applications(recruiter_id, status);
CREATE UNIQUE INDEX idx_partners_user_id ON partners(user_id) WHERE deleted_at IS NULL;

DROP TABLE IF EXISTS legacy_user_map;

CREATE OR REPLACE VIEW v_recruiter_performance AS
SELECT
    a.recruiter_id,
    COUNT(DISTINCT a.id)                                        AS total_managed,
    COUNT(DISTINCT a.id) FILTER (WHERE a.status = 'hired')     AS hired,
    COUNT(DISTINCT a.job_id)                                    AS active_jobs,
    ROUND(AVG(
        EXTRACT(DAY FROM a.last_moved_at - a.created_at)
    ), 1)                                                       AS avg_time_to_close_days,
    COUNT(DISTINCT a.id) FILTER (
        WHERE a.status NOT IN ('rejected','withdrawn','hired')
        AND NOW() - a.last_moved_at > INTERVAL '7 days'
    )                                                           AS stale_applications
FROM applications a
WHERE a.deleted_at IS NULL
GROUP BY a.recruiter_id;

CREATE OR REPLACE VIEW v_sla_breaches AS
SELECT
    a.id            AS application_id,
    a.job_id,
    a.candidate_id,
    a.recruiter_id,
    a.status,
    a.last_moved_at,
    EXTRACT(DAY FROM NOW() - a.last_moved_at)::int  AS days_in_stage,
    CASE a.status
        WHEN 'applied'       THEN 3
        WHEN 'cv_review'     THEN 3
        WHEN 'phone_screen'  THEN 5
        WHEN 'technical'     THEN 7
        WHEN 'final_round'   THEN 5
        WHEN 'offer'         THEN 3
        ELSE 999
    END                                             AS sla_days,
    CASE WHEN EXTRACT(DAY FROM NOW() - a.last_moved_at) >
        CASE a.status
            WHEN 'applied'       THEN 3
            WHEN 'cv_review'     THEN 3
            WHEN 'phone_screen'  THEN 5
            WHEN 'technical'     THEN 7
            WHEN 'final_round'   THEN 5
            WHEN 'offer'         THEN 3
            ELSE 999
        END THEN true ELSE false
    END                                             AS is_breached
FROM applications a
WHERE a.deleted_at IS NULL
  AND a.status NOT IN ('hired', 'rejected', 'withdrawn', 'offer_accepted', 'offer_declined');

-- +goose StatementEnd

-- +goose Down
-- Irreversible: the old UUIDs are gone once mapped to user-service ids.
SELECT 1;
