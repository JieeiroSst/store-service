ALTER TABLE wallets DROP CONSTRAINT IF EXISTS wallets_user_id_key;
CREATE UNIQUE INDEX wallets_user_currency_key ON wallets(user_id, currency) WHERE status <> 'CLOSED';
CREATE INDEX IF NOT EXISTS idx_wallets_user_id ON wallets(user_id);
ALTER TABLE wallets
    ADD COLUMN tier                VARCHAR(20)  NOT NULL DEFAULT 'BASIC'
                                       CHECK (tier IN ('BASIC','VERIFIED','PREMIUM')),
    ADD COLUMN pin_hash            VARCHAR(100) NOT NULL DEFAULT '',
    ADD COLUMN pin_failed_attempts INT          NOT NULL DEFAULT 0,
    ADD COLUMN pin_locked_until    TIMESTAMPTZ;

CREATE TABLE bank_accounts (
    id                  UUID         PRIMARY KEY,
    user_id             UUID         NOT NULL,
    bank_id             UUID         NOT NULL REFERENCES banks(id),
    account_number      VARCHAR(30)  NOT NULL,
    account_number_hash VARCHAR(64)  NOT NULL,
    account_name        VARCHAR(100) NOT NULL,
    is_default          BOOLEAN      NOT NULL DEFAULT FALSE,
    status              VARCHAR(20)  NOT NULL DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE','REMOVED')),
    created_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX idx_bank_accounts_active_number ON bank_accounts(user_id, account_number_hash) WHERE status = 'ACTIVE';
CREATE INDEX idx_bank_accounts_user ON bank_accounts(user_id);

UPDATE transactions SET idempotency_key = NULL WHERE idempotency_key = '';
ALTER TABLE transactions
    ALTER COLUMN idempotency_key  DROP NOT NULL,
    ALTER COLUMN merchant_id      DROP NOT NULL,
    ALTER COLUMN acquirer_bank_id DROP NOT NULL,
    ALTER COLUMN issuer_bank_id   DROP NOT NULL,
    ALTER COLUMN card_id          DROP NOT NULL,
    ALTER COLUMN card_network     DROP NOT NULL,
    ALTER COLUMN batch_id         DROP NOT NULL,
    ALTER COLUMN batch_id         DROP DEFAULT;
UPDATE transactions SET batch_id = NULL WHERE batch_id = '00000000-0000-0000-0000-000000000000';

ALTER TABLE transactions
    ADD COLUMN counterparty_wallet_id UUID REFERENCES wallets(id),
    ADD COLUMN bank_account_id        UUID REFERENCES bank_accounts(id),
    ADD COLUMN reference_id           UUID REFERENCES transactions(id),
    ADD COLUMN refunded_amount        NUMERIC(18,4) NOT NULL DEFAULT 0 CHECK (refunded_amount >= 0),
    ADD COLUMN external_ref           VARCHAR(100),
    ADD COLUMN failure_reason         TEXT,
    ADD COLUMN metadata               JSONB;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_type_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_type_check CHECK (type IN (
    'AUTHORIZATION','CAPTURE','SETTLEMENT','REFUND','VOID',
    'TOPUP','WITHDRAWAL','TRANSFER_IN','TRANSFER_OUT','PAYMENT','ADJUSTMENT'));
ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_status_check;
ALTER TABLE transactions ADD CONSTRAINT transactions_status_check CHECK (status IN (
    'PENDING','AUTHORIZED','CAPTURED','SETTLED','DECLINED','VOIDED',
    'COMPLETED','FAILED','REFUNDED','PARTIALLY_REFUNDED'));
ALTER TABLE transactions ADD CONSTRAINT transactions_refund_within_amount CHECK (refunded_amount <= amount);

CREATE INDEX idx_transactions_wallet_created ON transactions(wallet_id, created_at DESC);
CREATE INDEX idx_transactions_reference      ON transactions(reference_id) WHERE reference_id IS NOT NULL;
CREATE INDEX idx_transactions_unbatched      ON transactions(merchant_id, created_at)
    WHERE status = 'CAPTURED' AND card_id IS NOT NULL AND batch_id IS NULL;

CREATE TABLE ledger_entries (
    seq            BIGSERIAL     PRIMARY KEY,
    id             UUID          NOT NULL UNIQUE,
    wallet_id      UUID          NOT NULL REFERENCES wallets(id),
    transaction_id UUID          NOT NULL REFERENCES transactions(id),
    direction      VARCHAR(6)    NOT NULL CHECK (direction IN ('DEBIT','CREDIT')),
    amount         NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    balance_after  NUMERIC(18,4) NOT NULL CHECK (balance_after >= 0),
    created_at     TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_ledger_wallet_created ON ledger_entries(wallet_id, created_at);
CREATE INDEX idx_ledger_transaction    ON ledger_entries(transaction_id);

CREATE TABLE payment_requests (
    id                  UUID          PRIMARY KEY,
    requester_user_id   UUID          NOT NULL,
    requester_wallet_id UUID          NOT NULL REFERENCES wallets(id),
    payer_user_id       UUID          NOT NULL,
    payer_wallet_id     UUID          REFERENCES wallets(id),
    transaction_id      UUID          REFERENCES transactions(id),
    amount              NUMERIC(18,4) NOT NULL CHECK (amount > 0),
    currency            CHAR(3)       NOT NULL,
    note                VARCHAR(255),
    status              VARCHAR(20)   NOT NULL DEFAULT 'PENDING'
                            CHECK (status IN ('PENDING','PAID','DECLINED','CANCELLED','EXPIRED')),
    expires_at          TIMESTAMPTZ   NOT NULL,
    created_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_payment_requests_payer     ON payment_requests(payer_user_id, created_at DESC);
CREATE INDEX idx_payment_requests_requester ON payment_requests(requester_user_id, created_at DESC);

UPDATE settlement_batches SET status = 'OPEN' WHERE status = 'CAPTURED';
ALTER TABLE settlement_batches ALTER COLUMN status SET DEFAULT 'OPEN';
ALTER TABLE settlement_batches ADD CONSTRAINT settlement_batches_status_check
    CHECK (status IN ('OPEN','CLEARED','SETTLED'));
