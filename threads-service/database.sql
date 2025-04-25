-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    bio TEXT,
    profile_image_url VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Posts (threads) table
CREATE TABLE posts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    content TEXT NOT NULL,
    media_urls TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Comments table
CREATE TABLE comments (
    id UUID PRIMARY KEY,
    post_id UUID NOT NULL REFERENCES posts(id),
    user_id UUID NOT NULL REFERENCES users(id),
    parent_comment_id UUID REFERENCES comments(id), -- For nested comments
    content TEXT NOT NULL,
    media_urls TEXT[],
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Likes table
CREATE TABLE likes (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    post_id UUID REFERENCES posts(id),
    comment_id UUID REFERENCES comments(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Constraint to ensure either post_id or comment_id is set, not both
    CONSTRAINT like_target_check CHECK (
        (post_id IS NOT NULL AND comment_id IS NULL) OR
        (post_id IS NULL AND comment_id IS NOT NULL)
    ),
    -- Prevent duplicate likes
    UNIQUE (user_id, post_id, comment_id)
);

-- Follows table
CREATE TABLE follows (
    id UUID PRIMARY KEY,
    follower_id UUID NOT NULL REFERENCES users(id),
    followed_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    -- Prevent self-follows and duplicates
    CONSTRAINT no_self_follow CHECK (follower_id != followed_id),
    UNIQUE (follower_id, followed_id)
);

-- Tags table
CREATE TABLE tags (
    id UUID PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Post tags mapping
CREATE TABLE post_tags (
    post_id UUID NOT NULL REFERENCES posts(id),
    tag_id UUID NOT NULL REFERENCES tags(id),
    PRIMARY KEY (post_id, tag_id)
);