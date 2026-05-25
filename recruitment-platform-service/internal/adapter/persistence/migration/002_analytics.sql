-- +goose Up
-- +goose StatementBegin

-- ─── Recruitment Funnel View ──────────────────────────────────────────────────
-- Provides at-a-glance stage conversion metrics per job
CREATE OR REPLACE VIEW v_recruitment_funnel AS
SELECT
    j.id                                                        AS job_id,
    j.title                                                     AS job_title,
    j.code                                                      AS job_code,
    COUNT(a.id)                                                 AS total_applications,
    COUNT(a.id) FILTER (WHERE a.status = 'cv_review')           AS cv_review,
    COUNT(a.id) FILTER (WHERE a.status = 'phone_screen')        AS phone_screen,
    COUNT(a.id) FILTER (WHERE a.status = 'technical')           AS technical,
    COUNT(a.id) FILTER (WHERE a.status = 'final_round')         AS final_round,
    COUNT(a.id) FILTER (WHERE a.status = 'offer')               AS offer,
    COUNT(a.id) FILTER (WHERE a.status IN ('offer_accepted','hired')) AS hired,
    COUNT(a.id) FILTER (WHERE a.status = 'rejected')            AS rejected,
    COUNT(a.id) FILTER (WHERE a.status = 'withdrawn')           AS withdrawn,
    ROUND(
        COUNT(a.id) FILTER (WHERE a.status IN ('offer_accepted','hired'))::numeric
        / NULLIF(COUNT(a.id), 0) * 100, 2
    )                                                           AS offer_rate_pct,
    AVG(
        EXTRACT(DAY FROM NOW() - a.created_at)
    ) FILTER (WHERE a.status IN ('offer_accepted','hired'))     AS avg_days_to_hire
FROM jobs j
LEFT JOIN applications a ON a.job_id = j.id AND a.deleted_at IS NULL
WHERE j.deleted_at IS NULL
GROUP BY j.id, j.title, j.code;

-- ─── Recruiter Performance View ───────────────────────────────────────────────
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

-- ─── Referral Network Stats View ──────────────────────────────────────────────
CREATE OR REPLACE VIEW v_referral_network_stats AS
WITH RECURSIVE tree AS (
    SELECT id, full_name, referred_by_partner_id, hired_referrals, 0 AS depth
    FROM partners
    WHERE referred_by_partner_id IS NULL AND deleted_at IS NULL
    UNION ALL
    SELECT p.id, p.full_name, p.referred_by_partner_id, p.hired_referrals, t.depth + 1
    FROM partners p
    JOIN tree t ON p.referred_by_partner_id = t.id
    WHERE p.deleted_at IS NULL
)
SELECT
    id              AS partner_id,
    full_name,
    depth           AS network_level,
    hired_referrals,
    (SELECT COUNT(*) FROM tree t2 WHERE t2.referred_by_partner_id = tree.id) AS direct_downline
FROM tree;

-- ─── SLA Breach Alert View ────────────────────────────────────────────────────
-- Applications stuck in a stage beyond SLA thresholds
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

-- ─── AI Score Distribution ────────────────────────────────────────────────────
CREATE OR REPLACE VIEW v_ai_score_distribution AS
SELECT
    j.id                                        AS job_id,
    j.title                                     AS job_title,
    COUNT(a.id)                                 AS scored_applications,
    ROUND(AVG((a.match_score->>'score')::float), 2) AS avg_score,
    ROUND(MIN((a.match_score->>'score')::float), 2) AS min_score,
    ROUND(MAX((a.match_score->>'score')::float), 2) AS max_score,
    COUNT(a.id) FILTER (WHERE (a.match_score->>'score')::float >= 80) AS high_match,
    COUNT(a.id) FILTER (WHERE (a.match_score->>'score')::float BETWEEN 60 AND 79) AS medium_match,
    COUNT(a.id) FILTER (WHERE (a.match_score->>'score')::float < 60) AS low_match
FROM jobs j
JOIN applications a ON a.job_id = j.id
WHERE a.match_score IS NOT NULL AND j.deleted_at IS NULL
GROUP BY j.id, j.title;

-- ─── Indexes for reporting queries ────────────────────────────────────────────
CREATE INDEX IF NOT EXISTS idx_applications_last_moved ON applications(last_moved_at);
CREATE INDEX IF NOT EXISTS idx_applications_status_recruiter ON applications(recruiter_id, status);

-- +goose StatementEnd

-- +goose Down
DROP VIEW IF EXISTS v_ai_score_distribution;
DROP VIEW IF EXISTS v_sla_breaches;
DROP VIEW IF EXISTS v_referral_network_stats;
DROP VIEW IF EXISTS v_recruiter_performance;
DROP VIEW IF EXISTS v_recruitment_funnel;
