-- +goose Up
-- +goose StatementBegin

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "vector";      -- pgvector for AI embeddings
CREATE EXTENSION IF NOT EXISTS "pg_trgm";     -- trigram for fuzzy search

-- ─── Candidates ───────────────────────────────────────────────────────────────
CREATE TABLE candidates (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    full_name           TEXT        NOT NULL,
    email               TEXT        NOT NULL,
    phone               TEXT,
    avatar_url          TEXT,
    linkedin_url        TEXT,
    resume_url          TEXT,
    location            JSONB       DEFAULT '{}',
    current_title       TEXT,
    current_company     TEXT,
    years_of_experience INT         DEFAULT 0,
    experience_level    TEXT        NOT NULL DEFAULT 'fresher'
                            CHECK (experience_level IN ('fresher','junior','mid','senior','lead','manager')),
    skills              JSONB       DEFAULT '[]',
    expected_salary     JSONB,
    notice_period_days  INT         DEFAULT 30,
    status              TEXT        NOT NULL DEFAULT 'new'
                            CHECK (status IN ('new','screening','interview','offer','hired','rejected','withdrawn','blacklist')),
    source              TEXT        NOT NULL DEFAULT 'direct'
                            CHECK (source IN ('linkedin','referral','job_board','direct','agency')),
    referred_by_id      UUID        REFERENCES candidates(id),
    tags                JSONB       DEFAULT '[]',
    notes               TEXT,
    ai_score            JSONB,
    embedding           vector(1536),         -- OpenAI ada-002 / 1536 dims
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_candidates_email ON candidates(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_candidates_status       ON candidates(status);
CREATE INDEX idx_candidates_source       ON candidates(source);
CREATE INDEX idx_candidates_skills       ON candidates USING GIN(skills);
CREATE INDEX idx_candidates_fts          ON candidates USING GIN(to_tsvector('english', full_name || ' ' || COALESCE(current_title,'') || ' ' || COALESCE(current_company,'')));
CREATE INDEX idx_candidates_embedding    ON candidates USING ivfflat(embedding vector_cosine_ops) WITH (lists = 100);

-- ─── Jobs ─────────────────────────────────────────────────────────────────────
CREATE TABLE jobs (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    title               TEXT        NOT NULL,
    code                TEXT        NOT NULL,
    department_id       UUID        NOT NULL,
    hiring_manager_id   UUID        NOT NULL,
    recruiter_ids       JSONB       DEFAULT '[]',
    description         TEXT,
    requirements        JSONB       DEFAULT '[]',
    nice_to_have        JSONB       DEFAULT '[]',
    skills              JSONB       DEFAULT '[]',
    job_type            TEXT        NOT NULL DEFAULT 'full_time'
                            CHECK (job_type IN ('full_time','part_time','contract','freelance','internship')),
    work_mode           TEXT        NOT NULL DEFAULT 'onsite'
                            CHECK (work_mode IN ('onsite','remote','hybrid')),
    location            JSONB       DEFAULT '{}',
    salary_min          JSONB,
    salary_max          JSONB,
    salary_visible      BOOLEAN     DEFAULT false,
    headcount           INT         DEFAULT 1,
    pipeline_stages     JSONB       DEFAULT '[]',
    status              TEXT        NOT NULL DEFAULT 'draft'
                            CHECK (status IN ('draft','open','paused','closed','cancelled')),
    priority            INT         DEFAULT 3 CHECK (priority BETWEEN 1 AND 5),
    tags                JSONB       DEFAULT '[]',
    total_applications  INT         DEFAULT 0,
    open_applications   INT         DEFAULT 0,
    embedding           vector(1536),
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at          TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_jobs_code    ON jobs(code) WHERE deleted_at IS NULL;
CREATE INDEX idx_jobs_status         ON jobs(status);
CREATE INDEX idx_jobs_department     ON jobs(department_id);
CREATE INDEX idx_jobs_skills         ON jobs USING GIN(skills);
CREATE INDEX idx_jobs_embedding      ON jobs USING ivfflat(embedding vector_cosine_ops) WITH (lists = 100);

-- ─── Applications ─────────────────────────────────────────────────────────────
CREATE TABLE applications (
    id                      UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_id                  UUID        NOT NULL REFERENCES jobs(id),
    candidate_id            UUID        NOT NULL REFERENCES candidates(id),
    recruiter_id            UUID        NOT NULL,
    status                  TEXT        NOT NULL DEFAULT 'applied'
                                CHECK (status IN ('applied','cv_review','phone_screen','technical','final_round',
                                                  'offer','offer_accepted','offer_declined','hired','rejected','withdrawn')),
    current_stage_id        UUID,
    rejection_reason        TEXT,
    withdraw_reason         TEXT,
    match_score             JSONB,
    priority                INT         DEFAULT 0,
    referred_by_partner_id  UUID,
    days_in_stage           INT         DEFAULT 0,
    last_moved_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ,
    UNIQUE(job_id, candidate_id)       -- prevent duplicates
);

CREATE INDEX idx_applications_job       ON applications(job_id);
CREATE INDEX idx_applications_candidate ON applications(candidate_id);
CREATE INDEX idx_applications_recruiter ON applications(recruiter_id);
CREATE INDEX idx_applications_status    ON applications(status);

-- ─── Interviews ───────────────────────────────────────────────────────────────
CREATE TABLE interviews (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id  UUID        NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    round           INT         NOT NULL,
    title           TEXT        NOT NULL,
    interviewer_ids JSONB       DEFAULT '[]',
    scheduled_at    TIMESTAMPTZ NOT NULL,
    duration_min    INT         DEFAULT 60,
    meeting_url     TEXT,
    type            TEXT        DEFAULT 'online' CHECK (type IN ('online','onsite')),
    feedback        JSONB,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_interviews_application ON interviews(application_id);
CREATE INDEX idx_interviews_scheduled   ON interviews(scheduled_at);

-- ─── Offers ───────────────────────────────────────────────────────────────────
CREATE TABLE offers (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    application_id  UUID        NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    salary          JSONB       NOT NULL,
    start_date      DATE        NOT NULL,
    title           TEXT        NOT NULL,
    benefits        JSONB       DEFAULT '[]',
    offer_letter_url TEXT,
    expires_at      TIMESTAMPTZ NOT NULL,
    sent_at         TIMESTAMPTZ,
    responded_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- ─── Partners (Cộng tác viên) ────────────────────────────────────────────────
CREATE TABLE partners (
    id                      UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id                 UUID        NOT NULL,
    full_name               TEXT        NOT NULL,
    email                   TEXT        NOT NULL,
    phone                   TEXT,
    company                 TEXT,
    bio                     TEXT,
    referred_by_partner_id  UUID        REFERENCES partners(id),
    network_depth           INT         DEFAULT 0,
    status                  TEXT        NOT NULL DEFAULT 'active'
                                CHECK (status IN ('active','inactive','suspended')),
    tier                    TEXT        NOT NULL DEFAULT 'bronze'
                                CHECK (tier IN ('bronze','silver','gold','platinum')),
    commission              JSONB       DEFAULT '{}',
    total_referrals         INT         DEFAULT 0,
    hired_referrals         INT         DEFAULT 0,
    conversion_rate         FLOAT       DEFAULT 0,
    total_earned            JSONB       DEFAULT '{"amount":0,"currency":"VND"}',
    pending_payout          JSONB       DEFAULT '{"amount":0,"currency":"VND"}',
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at              TIMESTAMPTZ
);

CREATE UNIQUE INDEX idx_partners_user_id ON partners(user_id) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_partners_email   ON partners(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_partners_tier           ON partners(tier);
CREATE INDEX idx_partners_network        ON partners(referred_by_partner_id);

-- ─── Referrals ────────────────────────────────────────────────────────────────
CREATE TABLE referrals (
    id                  UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id          UUID        NOT NULL REFERENCES partners(id),
    candidate_id        UUID        REFERENCES candidates(id),
    job_id              UUID        REFERENCES jobs(id),
    application_id      UUID        REFERENCES applications(id),
    token               TEXT        NOT NULL UNIQUE,
    status              TEXT        NOT NULL DEFAULT 'pending'
                            CHECK (status IN ('pending','applied','hired','rejected','expired')),
    expires_at          TIMESTAMPTZ NOT NULL,
    commission_due      JSONB,
    commission_paid_at  TIMESTAMPTZ,
    click_count         INT         DEFAULT 0,
    last_clicked_at     TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_referrals_partner ON referrals(partner_id);
CREATE INDEX idx_referrals_token   ON referrals(token);
CREATE INDEX idx_referrals_status  ON referrals(status);

-- ─── Payouts ──────────────────────────────────────────────────────────────────
CREATE TABLE payouts (
    id              UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    partner_id      UUID        NOT NULL REFERENCES partners(id),
    referral_ids    JSONB       DEFAULT '[]',
    amount          JSONB       NOT NULL,
    status          TEXT        NOT NULL DEFAULT 'pending'
                        CHECK (status IN ('pending','approved','processed','failed')),
    note            TEXT,
    processed_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payouts_partner ON payouts(partner_id);
CREATE INDEX idx_payouts_status  ON payouts(status);

-- ─── Audit Log ────────────────────────────────────────────────────────────────
CREATE TABLE audit_log (
    id          UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_type  TEXT        NOT NULL,
    entity_type TEXT        NOT NULL,
    entity_id   UUID        NOT NULL,
    actor_id    UUID,
    payload     JSONB       DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_entity    ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_actor     ON audit_log(actor_id);
CREATE INDEX idx_audit_occurred  ON audit_log(occurred_at DESC);

-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS payouts;
DROP TABLE IF EXISTS referrals;
DROP TABLE IF EXISTS partners;
DROP TABLE IF EXISTS offers;
DROP TABLE IF EXISTS interviews;
DROP TABLE IF EXISTS applications;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS candidates;
