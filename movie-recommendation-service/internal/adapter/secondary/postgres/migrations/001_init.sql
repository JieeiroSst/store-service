-- Catalog mirrored from video-service. source says who owns the row:
-- 'sync' rows are pruned when video-service no longer lists them, 'manual'
-- rows (PUT/POST /api/recommendations/videos) are left alone.
CREATE TABLE IF NOT EXISTS videos (
    id          TEXT PRIMARY KEY,
    title       TEXT             NOT NULL,
    description TEXT             NOT NULL DEFAULT '',
    tags        TEXT[]           NOT NULL DEFAULT '{}',
    status      TEXT             NOT NULL DEFAULT 'ready',
    duration    DOUBLE PRECISION NOT NULL DEFAULT 0,
    views       BIGINT           NOT NULL DEFAULT 0,
    source      TEXT             NOT NULL DEFAULT 'sync',
    created_at  TIMESTAMPTZ      NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ      NOT NULL DEFAULT now()
);

-- Append-only user signals. No FK to videos: events may arrive before the
-- catalog sync has seen the video.
CREATE TABLE IF NOT EXISTS interactions (
    id         BIGSERIAL PRIMARY KEY,
    user_id    TEXT             NOT NULL,
    video_id   TEXT             NOT NULL,
    type       TEXT             NOT NULL CHECK (type IN ('view', 'watch', 'like', 'dislike', 'rating')),
    value      DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ      NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_interactions_user  ON interactions (user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_interactions_video ON interactions (video_id);
CREATE INDEX IF NOT EXISTS idx_interactions_time  ON interactions (created_at);

-- Rankings handed out page by page, shared by all replicas so a client can
-- keep paging one ranking whichever replica answers.
CREATE TABLE IF NOT EXISTS ranking_snapshots (
    id         TEXT PRIMARY KEY,
    scope      TEXT        NOT NULL,
    items      JSONB       NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_ranking_snapshots_expires ON ranking_snapshots (expires_at);
