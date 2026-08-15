CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE merchants (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    provider_type   TEXT NOT NULL CHECK (provider_type IN ('api','stock','self')),
    config          JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (name)
);
CREATE INDEX idx_merchants_provider_type ON merchants (provider_type);

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email           TEXT NOT NULL,
    password_hash   TEXT NOT NULL,
    role            TEXT NOT NULL CHECK (role IN ('retail','corporate_admin','system_admin')),
    corporate_id    UUID NULL,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (email)
);

CREATE TABLE corporates (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            TEXT NOT NULL,
    tax_code        TEXT,
    contact_email   TEXT,
    budget_limit    NUMERIC(18,2),
    budget_currency TEXT,
    status          TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','inactive')),
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

ALTER TABLE users
    ADD CONSTRAINT fk_users_corporate FOREIGN KEY (corporate_id) REFERENCES corporates (id);

CREATE TABLE wallets (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_type      TEXT NOT NULL CHECK (owner_type IN ('user','corporate')),
    owner_id        UUID NOT NULL,
    balance         NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'VND',
    version         INT NOT NULL DEFAULT 1,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (owner_type, owner_id)
);

CREATE TABLE wallet_transactions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id       UUID NOT NULL REFERENCES wallets (id),
    type            TEXT NOT NULL CHECK (type IN ('credit','debit')),
    amount          NUMERIC(18,2) NOT NULL,
    balance_after   NUMERIC(18,2) NOT NULL,
    ref_type        TEXT,
    ref_id          TEXT,
    idempotency_key TEXT UNIQUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wallet_transactions_wallet_id ON wallet_transactions (wallet_id);

CREATE TABLE orders (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_type      TEXT NOT NULL CHECK (buyer_type IN ('retail','corporate')),
    buyer_id        UUID NOT NULL,
    status          TEXT NOT NULL CHECK (status IN ('pending','awaiting_payment','paid','fulfilling','completed','cancelled','failed')),
    total_amount    NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency        TEXT NOT NULL DEFAULT 'VND',
    version         INT NOT NULL DEFAULT 1,
    idempotency_key TEXT UNIQUE,
    payment_ref     TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_orders_buyer_id ON orders (buyer_id);
CREATE INDEX idx_orders_status ON orders (status);

CREATE TABLE order_items (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id        UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    merchant_id     UUID NOT NULL REFERENCES merchants (id),
    product_sku     TEXT NOT NULL,
    quantity        INT NOT NULL,
    unit_price      NUMERIC(18,2) NOT NULL,
    line_total      NUMERIC(18,2) NOT NULL,
    issued_voucher_ids JSONB NOT NULL DEFAULT '[]',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_order_items_order_id ON order_items (order_id);

CREATE TABLE vouchers (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id         UUID NOT NULL REFERENCES merchants (id),
    order_id            UUID NULL REFERENCES orders (id),
    owner_type          TEXT NULL CHECK (owner_type IN ('user','corporate')),
    owner_id            TEXT NULL,
    product_sku         TEXT NOT NULL,
    denomination        NUMERIC(18,2) NOT NULL DEFAULT 0,
    currency            TEXT NOT NULL DEFAULT 'VND',
    code                TEXT NOT NULL DEFAULT '',
    pin_hash            TEXT NOT NULL DEFAULT '',
    status              TEXT NOT NULL CHECK (status IN ('created','issued','active','redeemed','expired','revoked')),
    version             INT NOT NULL DEFAULT 1,
    idempotency_key     TEXT,
    redeemed_amount     NUMERIC(18,2),
    provider_txn_ref    TEXT,
    issued_at           TIMESTAMPTZ,
    activated_at        TIMESTAMPTZ,
    redeemed_at         TIMESTAMPTZ,
    expires_at          TIMESTAMPTZ,
    revoked_at          TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (merchant_id, code)
);
CREATE UNIQUE INDEX idx_vouchers_idempotency_key ON vouchers (idempotency_key) WHERE idempotency_key IS NOT NULL;
CREATE INDEX idx_vouchers_status ON vouchers (status);
CREATE INDEX idx_vouchers_order_id ON vouchers (order_id);
CREATE INDEX idx_vouchers_owner ON vouchers (owner_type, owner_id);
CREATE INDEX idx_vouchers_expiry_sweep ON vouchers (expires_at) WHERE status IN ('issued','active');

CREATE TABLE voucher_stock (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    merchant_id             UUID NOT NULL REFERENCES merchants (id),
    product_sku             TEXT NOT NULL,
    code                    TEXT NOT NULL,
    pin                     TEXT,
    status                  TEXT NOT NULL DEFAULT 'available' CHECK (status IN ('available','claimed','void')),
    claimed_by_voucher_id   UUID NULL REFERENCES vouchers (id),
    batch_id                UUID NULL,
    imported_at             TIMESTAMPTZ NOT NULL DEFAULT now(),
    claimed_at              TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (merchant_id, code)
);
CREATE INDEX idx_voucher_stock_available ON voucher_stock (merchant_id, product_sku) WHERE status = 'available';
CREATE INDEX idx_voucher_stock_batch_id ON voucher_stock (batch_id);

CREATE TABLE distribution_jobs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    corporate_id        UUID NOT NULL REFERENCES corporates (id),
    order_id            UUID NULL REFERENCES orders (id),
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','completed','failed')),
    total_recipients    INT NOT NULL DEFAULT 0,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE distribution_claims (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    distribution_job_id     UUID NOT NULL REFERENCES distribution_jobs (id),
    voucher_id              UUID NULL REFERENCES vouchers (id),
    recipient_identifier    TEXT NOT NULL,
    claim_token              TEXT NOT NULL,
    claim_token_expires_at   TIMESTAMPTZ NOT NULL,
    status                  TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','sent','claimed','expired')),
    claimed_at              TIMESTAMPTZ,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (claim_token)
);
CREATE INDEX idx_distribution_claims_job_id ON distribution_claims (distribution_job_id);
CREATE INDEX idx_distribution_claims_status ON distribution_claims (status);

CREATE TABLE payments (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id            UUID NOT NULL REFERENCES orders (id),
    provider            TEXT NOT NULL CHECK (provider IN ('vnpay','momo','wallet')),
    amount              NUMERIC(18,2) NOT NULL,
    currency            TEXT NOT NULL DEFAULT 'VND',
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','succeeded','failed','refunded')),
    provider_txn_ref    TEXT,
    idempotency_key     TEXT UNIQUE,
    signature           TEXT,
    raw_callback        JSONB,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_payments_order_id ON payments (order_id);
CREATE INDEX idx_payments_provider_txn_ref ON payments (provider_txn_ref);

CREATE TABLE notifications (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    recipient_type  TEXT NOT NULL,
    recipient_id    TEXT NOT NULL,
    channel         TEXT NOT NULL CHECK (channel IN ('email','sms')),
    template_code   TEXT NOT NULL,
    payload         JSONB NOT NULL DEFAULT '{}',
    status          TEXT NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','sent','failed')),
    error           TEXT,
    sent_at         TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_notifications_status ON notifications (status);
CREATE INDEX idx_notifications_recipient_id ON notifications (recipient_id);

-- Generic HTTP-middleware idempotency cache. Distinct from the domain-level
-- idempotency_key columns above: this one is keyed by the client-supplied
-- Idempotency-Key header and caches the raw HTTP response; the domain-level
-- columns keep use cases idempotent independent of the HTTP layer.
CREATE TABLE idempotency_keys (
    key             TEXT PRIMARY KEY,
    request_hash    TEXT NOT NULL,
    status          TEXT NOT NULL DEFAULT 'in_progress' CHECK (status IN ('in_progress','completed','failed')),
    response_status INT,
    response_body   JSONB,
    locked_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at      TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_idempotency_keys_expires_at ON idempotency_keys (expires_at);

-- Transactional outbox: domain services write events here in the same DB
-- transaction as their state change; a background relay polls and publishes
-- to Kafka, marking published=true on success.
CREATE TABLE outbox_events (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type  TEXT NOT NULL,
    aggregate_id    TEXT NOT NULL,
    event_type      TEXT NOT NULL,
    payload         JSONB NOT NULL,
    topic           TEXT NOT NULL,
    published       BOOLEAN NOT NULL DEFAULT false,
    published_at    TIMESTAMPTZ,
    attempts        INT NOT NULL DEFAULT 0,
    last_error      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_outbox_events_unpublished ON outbox_events (created_at) WHERE published = false;
CREATE INDEX idx_outbox_events_aggregate ON outbox_events (aggregate_type, aggregate_id);

CREATE TABLE audit_log (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_type      TEXT,
    actor_id        TEXT,
    action          TEXT NOT NULL,
    entity_type     TEXT NOT NULL,
    entity_id       TEXT NOT NULL,
    before          JSONB,
    after           JSONB,
    ip_address      TEXT,
    user_agent      TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_log_entity ON audit_log (entity_type, entity_id);
CREATE INDEX idx_audit_log_actor_id ON audit_log (actor_id);
CREATE INDEX idx_audit_log_created_at ON audit_log (created_at);

CREATE TABLE api_keys (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    partner_id          UUID NOT NULL,
    key_prefix          TEXT NOT NULL,
    secret_hash         TEXT NOT NULL,
    scopes              JSONB NOT NULL DEFAULT '[]',
    rate_limit_per_min  INT NOT NULL DEFAULT 60,
    status              TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active','revoked')),
    last_used_at        TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (key_prefix)
);

CREATE TABLE reconciliation_runs (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_type            TEXT NOT NULL,
    status              TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','running','completed','failed')),
    discrepancy_count   INT NOT NULL DEFAULT 0,
    summary             JSONB NOT NULL DEFAULT '{}',
    started_at          TIMESTAMPTZ,
    finished_at         TIMESTAMPTZ,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT now()
);
