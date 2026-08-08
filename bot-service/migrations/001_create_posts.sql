-- bot-service: initial posts table
-- Content management's single source of truth: every one-off or recurring
-- post, its schedule, and the outcome of its most recent publish attempt.

CREATE TABLE posts (
    id UUID PRIMARY KEY,
    title TEXT,
    text TEXT,
    media JSONB NOT NULL DEFAULT '[]',
    channels JSONB NOT NULL DEFAULT '[]',

    -- One-off posts (cron_expr IS NULL/empty) fire once at scheduled_at.
    -- Drafts may not have this set yet - it stays the Go zero value
    -- ("0001-01-01T00:00:00Z") until UpdatePost/SubmitForReview fills it in.
    scheduled_at TIMESTAMP WITH TIME ZONE NOT NULL,

    -- Recurring posts set cron_expr (standard 5-field cron, or a
    -- robfig/cron "@every <duration>"/"@daily" style descriptor) and stay
    -- 'active' until cancelled, firing repeatedly on that schedule.
    cron_expr VARCHAR(255),
    max_runs_per_day INT NOT NULL DEFAULT 0,   -- 0 = unlimited
    max_runs_per_month INT NOT NULL DEFAULT 0, -- 0 = unlimited
    runs_today INT NOT NULL DEFAULT 0,
    runs_this_month INT NOT NULL DEFAULT 0,
    last_run_date VARCHAR(10),   -- "2006-01-02", tracks the day runs_today belongs to
    last_run_month VARCHAR(7),   -- "2006-01", tracks the month runs_this_month belongs to
    last_run_at TIMESTAMP WITH TIME ZONE,

    -- draft -> pending_review -> scheduled/active -> published/failed, with
    -- reject_reason set whenever an editor bounces pending_review -> draft.
    -- 'publishing' is a short-lived lock held only while a publish attempt
    -- is in flight (see contentService.publishOneOff/runRecurring).
    status VARCHAR(20) NOT NULL CHECK (status IN (
        'draft', 'pending_review', 'scheduled', 'publishing', 'published', 'failed', 'active', 'cancelled'
    )),
    reject_reason TEXT,
    results JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_posts_status_scheduled_at ON posts(status, scheduled_at);
CREATE INDEX idx_posts_created_at ON posts(created_at);
-- Speeds up ActiveRecurring's lookup of active posts with a cron schedule.
CREATE INDEX idx_posts_active_recurring ON posts(status) WHERE cron_expr IS NOT NULL AND cron_expr <> '';
-- Speeds up a "review queue" listing (GET /v1/posts?status=pending_review).
CREATE INDEX idx_posts_pending_review ON posts(created_at) WHERE status = 'pending_review';
