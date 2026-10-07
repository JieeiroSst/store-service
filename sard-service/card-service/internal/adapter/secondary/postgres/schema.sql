CREATE TABLE IF NOT EXISTS accounts (
    id            text PRIMARY KEY,
    customer_id   text        NOT NULL,
    user_id       bigint      NOT NULL DEFAULT 0,
    program_code  text        NOT NULL,
    mode          text        NOT NULL,
    currency      text        NOT NULL,
    holder_name   text        NOT NULL DEFAULT '',
    status        text        NOT NULL,
    status_reason text        NOT NULL DEFAULT '',
    credit_limit  bigint      NOT NULL DEFAULT 0,
    balance       bigint      NOT NULL DEFAULT 0,
    held          bigint      NOT NULL DEFAULT 0 CHECK (held >= 0),
    created_at    timestamptz NOT NULL,
    updated_at    timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS accounts_customer_idx ON accounts (customer_id, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS accounts_customer_program_open_idx
    ON accounts (customer_id, program_code)
    WHERE status <> 'CANCELLED';

CREATE TABLE IF NOT EXISTS cards (
    id                    text PRIMARY KEY,
    account_id            text        NOT NULL REFERENCES accounts (id),
    customer_id           text        NOT NULL,
    program_code          text        NOT NULL,
    card_type             text        NOT NULL,
    cardholder_name       text        NOT NULL,
    pan_hash              text        NOT NULL UNIQUE,
    pan_cipher            bytea       NOT NULL,
    bin                   text        NOT NULL,
    last4                 text        NOT NULL,
    expiry_month          integer     NOT NULL,
    expiry_year           integer     NOT NULL,
    valid_until           timestamptz,
    service_code          text        NOT NULL,
    status                text        NOT NULL,
    status_reason         text        NOT NULL DEFAULT '',
    limit_per_transaction bigint      NOT NULL,
    limit_daily           bigint      NOT NULL,
    control_pos           boolean     NOT NULL,
    control_contactless   boolean     NOT NULL,
    control_ecommerce     boolean     NOT NULL,
    control_atm           boolean     NOT NULL,
    pin_hash              text        NOT NULL DEFAULT '',
    pin_failures          integer     NOT NULL DEFAULT 0,
    replaces_card_id      text        NOT NULL DEFAULT '',
    replaced_by_card_id   text        NOT NULL DEFAULT '',
    created_at            timestamptz NOT NULL,
    updated_at            timestamptz NOT NULL,
    activated_at          timestamptz
);

CREATE INDEX IF NOT EXISTS cards_account_idx ON cards (account_id, created_at);

CREATE TABLE IF NOT EXISTS authorizations (
    id                text PRIMARY KEY,
    account_id        text        NOT NULL REFERENCES accounts (id),
    card_id           text        NOT NULL REFERENCES cards (id),
    processing_code   text        NOT NULL,
    status            text        NOT NULL,
    amount            bigint      NOT NULL,
    confirmed_amount  bigint      NOT NULL DEFAULT 0,
    currency          text        NOT NULL,
    channel           text        NOT NULL,
    merchant          text        NOT NULL DEFAULT '',
    mcc               text        NOT NULL DEFAULT '',
    response_code     text        NOT NULL,
    reason            text        NOT NULL DEFAULT '',
    auth_code         text        NOT NULL DEFAULT '',
    authentication_id text        NOT NULL DEFAULT '',
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS authorizations_card_idx ON authorizations (card_id, created_at DESC);

CREATE TABLE IF NOT EXISTS transactions (
    id               text PRIMARY KEY,
    account_id       text        NOT NULL REFERENCES accounts (id),
    card_id          text        NOT NULL DEFAULT '',
    authorization_id text        NOT NULL DEFAULT '',
    type             text        NOT NULL,
    processing_code  text        NOT NULL,
    amount           bigint      NOT NULL,
    balance_after    bigint      NOT NULL,
    description      text        NOT NULL DEFAULT '',
    created_at       timestamptz NOT NULL
);

CREATE INDEX IF NOT EXISTS transactions_account_idx ON transactions (account_id, created_at DESC);
