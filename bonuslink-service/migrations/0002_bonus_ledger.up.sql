-- Rewards granted by other services (referral-service) arrive through the queue.
-- user_id is the caller's own id (an opaque string), so there is no FK to users.
-- event_id is the idempotency key: a redelivered message never pays twice.
CREATE TABLE bonus_rewards (
    reward_id    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id     VARCHAR(128) UNIQUE NOT NULL,
    user_id      VARCHAR(64) NOT NULL,
    ref_code     VARCHAR(64),
    source       VARCHAR(50) NOT NULL DEFAULT 'referral',
    reward_type  VARCHAR(50) NOT NULL CHECK (reward_type IN ('POINTS', 'BADGE', 'DISCOUNT', 'EXPERIENCE', 'PREMIUM_CONTENT')),
    reward_value NUMERIC(15, 2) NOT NULL CHECK (reward_value > 0),
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_bonus_rewards_user_created ON bonus_rewards(user_id, created_at DESC);

CREATE TABLE bonus_balances (
    user_id      VARCHAR(64) NOT NULL,
    reward_type  VARCHAR(50) NOT NULL,
    total_value  NUMERIC(15, 2) NOT NULL DEFAULT 0,
    reward_count BIGINT NOT NULL DEFAULT 0,
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, reward_type)
);
