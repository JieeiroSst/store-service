-- threads-service owns the social graph: posts, comments, likes,
-- follows, and tags. It does NOT own user identity/auth - accounts,
-- credentials, and profile data belong to user_service, which this
-- service calls over HTTP to enrich responses with author info (see
-- internal/adapter/secondary/userclient and internal/domain/port.UserClient).
--
-- user_id / follower_id / followed_id below are plain UUID references
-- with no cross-service foreign key: each service in this cluster owns
-- its own database, so a FK into user_service's table isn't possible (or
-- desirable - it would couple the two services' deploys/migrations).
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema. It's plain idempotent DDL
-- (IF NOT EXISTS everywhere), so re-running it on every deploy is safe.
-- New columns on an existing table use ALTER TABLE ... ADD COLUMN IF NOT
-- EXISTS rather than relying on CREATE TABLE IF NOT EXISTS alone, which is
-- a no-op against a table that already exists and would silently skip
-- them on a deployment that predates the column.
--
-- Hot reads (post-by-ID, author lookups) are cached in Redis - see
-- internal/adapter/secondary/cache. Postgres stays authoritative for that
-- data; Redis entries carry a TTL and are invalidated on writes, so a
-- cache/DB mismatch is bounded and self-heals rather than being load-
-- bearing. Trending tags are the one exception: Redis is the *only* store
-- for that ranking (see port.TrendingTagsStore).

CREATE TABLE IF NOT EXISTS posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    content TEXT NOT NULL,
    media_urls TEXT[],
    like_count INT NOT NULL DEFAULT 0,
    comment_count INT NOT NULL DEFAULT 0,
    repost_count INT NOT NULL DEFAULT 0,
    repost_of_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
ALTER TABLE posts ADD COLUMN IF NOT EXISTS repost_count INT NOT NULL DEFAULT 0;
ALTER TABLE posts ADD COLUMN IF NOT EXISTS repost_of_id UUID REFERENCES posts(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts(user_id);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_posts_repost_of_id ON posts(repost_of_id);

-- Comments table (threaded via parent_comment_id)
CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL,
    parent_comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    media_urls TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments(post_id);
CREATE INDEX IF NOT EXISTS idx_comments_parent_comment_id ON comments(parent_comment_id);
CREATE INDEX IF NOT EXISTS idx_comments_user_id ON comments(user_id);

-- Likes table (targets exactly one of post_id/comment_id)
CREATE TABLE IF NOT EXISTS likes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id UUID REFERENCES posts(id) ON DELETE CASCADE,
    comment_id UUID REFERENCES comments(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT like_target_check CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR
        (post_id IS NULL AND comment_id IS NOT NULL)
    )
);
-- Two partial unique indexes, not one UNIQUE(user_id, post_id, comment_id):
-- Postgres treats every NULL as distinct for uniqueness purposes, and
-- post_id/comment_id are nullable (exactly one is set per row per the
-- check above), so a plain composite UNIQUE would never actually reject
-- a duplicate like.
CREATE UNIQUE INDEX IF NOT EXISTS idx_likes_user_post ON likes(user_id, post_id) WHERE post_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_likes_user_comment ON likes(user_id, comment_id) WHERE comment_id IS NOT NULL;

-- Follows table
CREATE TABLE IF NOT EXISTS follows (
    id UUID PRIMARY KEY,
    follower_id UUID NOT NULL,
    followed_id UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Prevent self-follows and duplicates
    CONSTRAINT no_self_follow CHECK (follower_id != followed_id),
    UNIQUE (follower_id, followed_id)
);
CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows(follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_followed_id ON follows(followed_id);

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Post tags mapping
CREATE TABLE IF NOT EXISTS post_tags (
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag_id ON post_tags(tag_id);

-- Bookmarks table (a user privately saving a post - never a public signal,
-- unlike likes/reposts)
CREATE TABLE IF NOT EXISTS bookmarks (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
-- Both columns are NOT NULL here (unlike likes' post_id/comment_id), so a
-- plain composite UNIQUE index works correctly - no NULL-distinctness
-- caveat to work around.
CREATE UNIQUE INDEX IF NOT EXISTS idx_bookmarks_user_post ON bookmarks(user_id, post_id);
