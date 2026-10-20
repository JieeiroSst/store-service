CREATE EXTENSION IF NOT EXISTS btree_gist;

CREATE TABLE IF NOT EXISTS lucky_campaigns (
    id                BIGSERIAL    PRIMARY KEY,
    game_type         VARCHAR(32)  NOT NULL,
    name              VARCHAR(255) NOT NULL,
    description       TEXT         NOT NULL DEFAULT '',
    start_at          TIMESTAMPTZ  NOT NULL,
    end_at            TIMESTAMPTZ  NOT NULL,
    timezone          VARCHAR(64)  NOT NULL,
    status            VARCHAR(16)  NOT NULL DEFAULT 'draft',
    win_rate          NUMERIC(5,2) NOT NULL DEFAULT 0,
    daily_spins       INTEGER      NOT NULL DEFAULT 1,
    max_spins_per_day INTEGER      NOT NULL DEFAULT 0,
    max_points_per_user_day BIGINT NOT NULL DEFAULT 0,
    budget_points     BIGINT       NOT NULL DEFAULT 0,
    points_awarded    BIGINT       NOT NULL DEFAULT 0,
    review_points     BIGINT       NOT NULL DEFAULT 0,
    spin_rules        JSONB        NOT NULL DEFAULT '[]',
    checkin_rewards   JSONB        NOT NULL DEFAULT '[]',
    settings          JSONB        NOT NULL DEFAULT '{}',
    created_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at        TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT lucky_campaigns_game_check CHECK (game_type IN ('wheel', 'scratch_card', 'gift_box', 'shake', 'flip_card', 'slot_machine', 'capsule', 'checkin')),
    CONSTRAINT lucky_campaigns_window_check CHECK (end_at > start_at),
    CONSTRAINT lucky_campaigns_status_check CHECK (status IN ('draft', 'active', 'paused')),
    CONSTRAINT lucky_campaigns_win_rate_check CHECK (win_rate >= 0 AND win_rate <= 100),
    CONSTRAINT lucky_campaigns_daily_spins_check CHECK (daily_spins >= 0),
    CONSTRAINT lucky_campaigns_budget_check CHECK (budget_points = 0 OR points_awarded <= budget_points),
    CONSTRAINT lucky_campaigns_one_active EXCLUDE USING gist (game_type WITH =, tstzrange(start_at, end_at) WITH &&) WHERE (status = 'active')
);

CREATE INDEX IF NOT EXISTS lucky_campaigns_status_idx ON lucky_campaigns (status, start_at);

CREATE TABLE IF NOT EXISTS lucky_prizes (
    id          BIGSERIAL    PRIMARY KEY,
    campaign_id BIGINT       NOT NULL REFERENCES lucky_campaigns (id) ON DELETE CASCADE,
    name        VARCHAR(255) NOT NULL,
    type        VARCHAR(16)  NOT NULL,
    points      BIGINT       NOT NULL DEFAULT 0,
    weight      INTEGER      NOT NULL DEFAULT 0,
    quantity    BIGINT,
    awarded     BIGINT       NOT NULL DEFAULT 0,
    position    INTEGER      NOT NULL,
    image_url   TEXT         NOT NULL DEFAULT '',
    CONSTRAINT lucky_prizes_type_check CHECK (type IN ('points', 'none')),
    CONSTRAINT lucky_prizes_stock_check CHECK (quantity IS NULL OR awarded <= quantity),
    CONSTRAINT lucky_prizes_position_key UNIQUE (campaign_id, position) DEFERRABLE INITIALLY DEFERRED
);

CREATE UNIQUE INDEX IF NOT EXISTS lucky_prizes_one_blank ON lucky_prizes (campaign_id) WHERE type = 'none';

CREATE TABLE IF NOT EXISTS lucky_plays (
    id              BIGSERIAL    PRIMARY KEY,
    campaign_id     BIGINT       NOT NULL REFERENCES lucky_campaigns (id),
    game_type       VARCHAR(32)  NOT NULL,
    user_id         VARCHAR(64)  NOT NULL,
    prize_id        BIGINT       NOT NULL REFERENCES lucky_prizes (id),
    prize_name      VARCHAR(255) NOT NULL,
    prize_type      VARCHAR(16)  NOT NULL,
    points          BIGINT       NOT NULL DEFAULT 0,
    position        INTEGER      NOT NULL,
    choice          INTEGER      NOT NULL DEFAULT 0,
    source          VARCHAR(16)  NOT NULL,
    day             DATE         NOT NULL,
    ip              VARCHAR(64)  NOT NULL DEFAULT '',
    device_id       VARCHAR(128) NOT NULL DEFAULT '',
    user_agent      VARCHAR(255) NOT NULL DEFAULT '',
    idempotency_key VARCHAR(128),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT lucky_plays_source_check CHECK (source IN ('daily', 'credit'))
);

CREATE INDEX IF NOT EXISTS lucky_plays_player_day_idx ON lucky_plays (campaign_id, user_id, day);
CREATE INDEX IF NOT EXISTS lucky_plays_user_idx ON lucky_plays (user_id, id DESC);
CREATE INDEX IF NOT EXISTS lucky_plays_wins_idx ON lucky_plays (campaign_id, id DESC) WHERE points > 0;
CREATE INDEX IF NOT EXISTS lucky_plays_device_idx ON lucky_plays (device_id, id DESC) WHERE device_id <> '';
CREATE INDEX IF NOT EXISTS lucky_plays_ip_idx ON lucky_plays (ip, id DESC) WHERE ip <> '';
CREATE UNIQUE INDEX IF NOT EXISTS lucky_plays_idempotency_key ON lucky_plays (user_id, idempotency_key) WHERE idempotency_key IS NOT NULL;

CREATE TABLE IF NOT EXISTS lucky_spin_credits (
    id           BIGSERIAL    PRIMARY KEY,
    campaign_id  BIGINT       NOT NULL REFERENCES lucky_campaigns (id),
    user_id      VARCHAR(64)  NOT NULL,
    event        VARCHAR(64)  NOT NULL,
    reference_id VARCHAR(128) NOT NULL,
    spins        INTEGER      NOT NULL,
    day          DATE         NOT NULL,
    created_at   TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT lucky_spin_credits_spins_check CHECK (spins > 0),
    CONSTRAINT lucky_spin_credits_reference_key UNIQUE (campaign_id, user_id, event, reference_id)
);

CREATE INDEX IF NOT EXISTS lucky_spin_credits_player_day_idx ON lucky_spin_credits (campaign_id, user_id, event, day);

CREATE TABLE IF NOT EXISTS lucky_checkins (
    id          BIGSERIAL   PRIMARY KEY,
    campaign_id BIGINT      NOT NULL REFERENCES lucky_campaigns (id),
    user_id     VARCHAR(64) NOT NULL,
    day         DATE        NOT NULL,
    streak      INTEGER     NOT NULL,
    points      BIGINT      NOT NULL,
    ip          VARCHAR(64)  NOT NULL DEFAULT '',
    device_id   VARCHAR(128) NOT NULL DEFAULT '',
    user_agent  VARCHAR(255) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT lucky_checkins_streak_check CHECK (streak > 0),
    CONSTRAINT lucky_checkins_day_key UNIQUE (campaign_id, user_id, day)
);

CREATE INDEX IF NOT EXISTS lucky_checkins_user_idx ON lucky_checkins (user_id, id DESC);

CREATE TABLE IF NOT EXISTS lucky_reward_outbox (
    id           BIGSERIAL    PRIMARY KEY,
    event_id     VARCHAR(64)  NOT NULL,
    source       VARCHAR(32)  NOT NULL,
    game_type    VARCHAR(32)  NOT NULL,
    user_id      VARCHAR(64)  NOT NULL,
    campaign_id  BIGINT       NOT NULL,
    reference_id BIGINT       NOT NULL,
    title        VARCHAR(255) NOT NULL,
    points       BIGINT       NOT NULL,
    occurred_at  TIMESTAMPTZ  NOT NULL,
    status       VARCHAR(16)  NOT NULL DEFAULT 'pending',
    hold_reason  VARCHAR(32)  NOT NULL DEFAULT '',
    reviewed_by  VARCHAR(64)  NOT NULL DEFAULT '',
    reviewed_at  TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    CONSTRAINT lucky_reward_outbox_event_key UNIQUE (event_id),
    CONSTRAINT lucky_reward_outbox_points_check CHECK (points > 0),
    CONSTRAINT lucky_reward_outbox_status_check CHECK (status IN ('pending', 'held', 'rejected'))
);

CREATE INDEX IF NOT EXISTS lucky_reward_outbox_pending_idx ON lucky_reward_outbox (id) WHERE published_at IS NULL AND status = 'pending';
CREATE INDEX IF NOT EXISTS lucky_reward_outbox_review_idx ON lucky_reward_outbox (status, id DESC);
CREATE INDEX IF NOT EXISTS lucky_reward_outbox_user_idx ON lucky_reward_outbox (user_id, id DESC);

CREATE TABLE IF NOT EXISTS lucky_user_blocks (
    user_id    VARCHAR(64)  PRIMARY KEY,
    reason     VARCHAR(255) NOT NULL,
    blocked_by VARCHAR(64)  NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT now()
);
