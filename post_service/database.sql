-- post_service manages posts ("news"), categories, and their uploaded
-- media. It does not own user identity/auth - author_id is a plain
-- string reference to whatever issued the caller's JWT (e.g.
-- user_service), with no foreign key into that service's database (each
-- service in this cluster owns its own database).
--
-- Table names predate this Go-level refactor (New -> Post, NewCategory ->
-- PostCategory) and are kept as-is (news, new_categories) to avoid a data
-- migration - see model/post.go's doc comment.
--
-- Applied automatically on startup by internal/infrastructure/database.
-- CREATE TABLE IF NOT EXISTS only: safe to re-run, but only creates
-- missing tables. MySQL has no IF NOT EXISTS for ALTER TABLE/CREATE INDEX
-- before 8.0.29/MariaDB 10.0, so every index below is declared inline in
-- its CREATE TABLE instead of as a separate statement - a deployment with
-- pre-existing tables from before this refactor won't get new columns
-- (e.g. created_at/updated_at) or indexes automatically and needs a
-- one-time manual migration for those.

CREATE TABLE IF NOT EXISTS categories (
    id VARCHAR(191) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS media (
    id VARCHAR(191) PRIMARY KEY,
    url VARCHAR(1024) NOT NULL,
    description VARCHAR(255),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS news (
    id VARCHAR(191) PRIMARY KEY,
    author_id VARCHAR(191) NOT NULL,
    name VARCHAR(255) NOT NULL,
    content TEXT,
    description TEXT,
    media_id VARCHAR(191),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_news_author_id (author_id),
    INDEX idx_news_created_at (created_at),
    CONSTRAINT fk_news_media FOREIGN KEY (media_id) REFERENCES media(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS new_categories (
    new_id VARCHAR(191) NOT NULL,
    category_id VARCHAR(191) NOT NULL,
    PRIMARY KEY (new_id, category_id),
    INDEX idx_new_categories_category_id (category_id),
    CONSTRAINT fk_new_categories_news FOREIGN KEY (new_id) REFERENCES news(id) ON DELETE CASCADE,
    CONSTRAINT fk_new_categories_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE CASCADE
);
