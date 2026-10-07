CREATE TABLE IF NOT EXISTS payment_authentications (
    id               text PRIMARY KEY,
    card_id          text        NOT NULL,
    account_id       text        NOT NULL DEFAULT '',
    customer_id      text        NOT NULL DEFAULT '',
    user_id          bigint      NOT NULL,
    username         text        NOT NULL,
    masked_pan       text        NOT NULL,
    amount           bigint      NOT NULL,
    currency         text        NOT NULL,
    merchant         text        NOT NULL,
    mcc              text        NOT NULL DEFAULT '',
    otp_expires_at   timestamptz NOT NULL,
    attempts         integer     NOT NULL DEFAULT 0,
    max_attempts     integer     NOT NULL,
    resends          integer     NOT NULL DEFAULT 0,
    status           text        NOT NULL,
    failure_reason   text        NOT NULL DEFAULT '',
    expires_at       timestamptz NOT NULL,
    authenticated_at timestamptz,
    valid_until      timestamptz,
    used_at          timestamptz,
    created_at       timestamptz NOT NULL,
    updated_at       timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS payment_authentications_user_pending_idx
    ON payment_authentications (user_id, created_at DESC)
    WHERE status = 'PENDING';
