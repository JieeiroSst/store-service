-- bot-service: content-management fields
-- Rounds out the "single source of truth" record per docs/content-management
-- spec: hashtags, media classification, timezone display, campaign grouping,
-- and actor tracking (who created/approved/last changed a post's status).

ALTER TABLE posts
    ADD COLUMN hashtags JSONB NOT NULL DEFAULT '[]',
    -- text_only | single_image | multi_image | video | reel - see
    -- model.MediaKind; "reel" is never auto-derived from the media alone.
    ADD COLUMN media_kind VARCHAR(20),
    -- IANA zone name (e.g. "Asia/Ho_Chi_Minh") for displaying scheduled_at to
    -- a human; scheduled_at itself is already a real timezone-aware value.
    ADD COLUMN timezone VARCHAR(64),
    -- Groups posts for filtering (GET /v1/posts?campaign=...); purely
    -- organizational, never sent to a channel.
    ADD COLUMN campaign VARCHAR(255),
    -- Free-form actor identifiers (usernames) - bot-service has no auth
    -- system of its own, so these are trusted as-is from the caller.
    ADD COLUMN created_by VARCHAR(255),
    ADD COLUMN approved_by VARCHAR(255),
    ADD COLUMN status_changed_by VARCHAR(255);

CREATE INDEX idx_posts_campaign ON posts(campaign) WHERE campaign IS NOT NULL AND campaign <> '';
