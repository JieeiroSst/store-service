-- wallet-service: Initial Schema Migration
-- Supports the full VISA payment flow (Authorization + Capture & Settlement)

BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- ============================================================
-- Banks (both issuing and acquiring)
-- ============================================================
CREATE TABLE banks (
    id         UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    name       VARCHAR(100) NOT NULL,
    code       VARCHAR(20)  NOT NULL UNIQUE,
    country    CHAR(2)      NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Merchants (Steps 1-2: POS terminal owner, sends batch to acquirer)
-- ============================================================
CREATE TABLE merchants (
    id               UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    name             VARCHAR(200) NOT NULL,
    mcc              VARCHAR(4)   NOT NULL,
    acquirer_bank_id UUID         NOT NULL REFERENCES banks(id),
    country          CHAR(2)      NOT NULL,
    created_at       TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

-- ============================================================
-- Wallets (issuing bank cardholder accounts)
-- balance and frozen_amount stored as NUMERIC for decimal precision
-- version enables optimistic locking to prevent concurrent write conflicts
-- ============================================================
CREATE TABLE wallets (
    id            UUID         PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id       UUID         NOT NULL UNIQUE,
    balance       NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (balance >= 0),
    frozen_amount NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (frozen_amount >= 0),
    currency      CHAR(3)      NOT NULL DEFAULT 'USD',
    status        VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE'
                      CHECK (status IN ('ACTIVE','FROZEN','CLOSED','SUSPENDED')),
    version       INT          NOT NULL DEFAULT 1,
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    CONSTRAINT balance_gte_frozen CHECK (balance >= frozen_amount)
);

-- ============================================================
-- Cards (Step 0: issuing bank issues cards to cardholders)
-- card_number is stored masked; card_number_hash (SHA-256) is used for lookup
-- ============================================================
CREATE TABLE cards (
    id               UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    wallet_id        UUID        NOT NULL REFERENCES wallets(id),
    issuer_bank_id   UUID        NOT NULL REFERENCES banks(id),
    card_number      VARCHAR(25) NOT NULL,              -- masked: **** **** **** 1234
    card_number_hash VARCHAR(64) NOT NULL UNIQUE,       -- SHA-256 for lookup
    holder_name      VARCHAR(100) NOT NULL,
    network          VARCHAR(20) NOT NULL CHECK (network IN ('VISA','MASTERCARD','AMEX')),
    expiry_month     SMALLINT    NOT NULL CHECK (expiry_month BETWEEN 1 AND 12),
    expiry_year      SMALLINT    NOT NULL,
    status           VARCHAR(20) NOT NULL DEFAULT 'ACTIVE'
                         CHECK (status IN ('ACTIVE','BLOCKED','EXPIRED','CANCELLED')),
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_cards_wallet_id ON cards(wallet_id);
CREATE INDEX idx_cards_hash ON cards(card_number_hash);

-- ============================================================
-- Transactions (full authorization + settlement lifecycle)
-- ============================================================
CREATE TABLE transactions (
    id                 UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    idempotency_key    VARCHAR(64) NOT NULL UNIQUE,
    wallet_id          UUID        NOT NULL REFERENCES wallets(id),
    merchant_id        UUID        NOT NULL REFERENCES merchants(id),
    acquirer_bank_id   UUID        NOT NULL REFERENCES banks(id),
    issuer_bank_id     UUID        NOT NULL REFERENCES banks(id),
    card_id            UUID        NOT NULL REFERENCES cards(id),
    card_network       VARCHAR(20) NOT NULL,
    amount             NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    currency           CHAR(3)     NOT NULL DEFAULT 'USD',
    fee                NUMERIC(18,4) NOT NULL DEFAULT 0,
    type               VARCHAR(20) NOT NULL
                           CHECK (type IN ('AUTHORIZATION','CAPTURE','SETTLEMENT','REFUND','VOID')),
    status             VARCHAR(20) NOT NULL
                           CHECK (status IN ('PENDING','AUTHORIZED','CAPTURED','SETTLED','DECLINED','VOIDED')),
    description        TEXT,
    authorization_code VARCHAR(10),
    batch_id           UUID        NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_transactions_wallet_id    ON transactions(wallet_id);
CREATE INDEX idx_transactions_merchant_id  ON transactions(merchant_id);
CREATE INDEX idx_transactions_status       ON transactions(status);
CREATE INDEX idx_transactions_batch_id     ON transactions(batch_id);

-- ============================================================
-- Settlement Batches (Steps 1-2 of settlement flow)
-- ============================================================
CREATE TABLE settlement_batches (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    acquirer_id  UUID          NOT NULL REFERENCES banks(id),
    merchant_id  UUID          NOT NULL REFERENCES merchants(id),
    total_amount NUMERIC(18,4) NOT NULL DEFAULT 0,
    total_fee    NUMERIC(18,4) NOT NULL DEFAULT 0,
    txn_count    INT           NOT NULL DEFAULT 0,
    status       VARCHAR(20)   NOT NULL DEFAULT 'CAPTURED',
    processed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_settlement_batches_merchant ON settlement_batches(merchant_id);

-- ============================================================
-- Clearing Records (Step 3: card network performs netting)
-- One record per unique (issuer, acquirer) pair per batch after netting.
-- This is the result of bilateral netting: reduces total inter-bank transfers.
-- ============================================================
CREATE TABLE clearing_records (
    id           UUID          PRIMARY KEY DEFAULT uuid_generate_v4(),
    batch_id     UUID          NOT NULL REFERENCES settlement_batches(id),
    card_network VARCHAR(20)   NOT NULL,
    acquirer_id  UUID          NOT NULL REFERENCES banks(id),
    issuer_id    UUID          NOT NULL REFERENCES banks(id),
    net_amount   NUMERIC(18,4) NOT NULL,
    cleared_at   TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_clearing_batch_id ON clearing_records(batch_id);

COMMIT;
