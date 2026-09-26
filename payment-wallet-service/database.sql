-- payment-wallet-service owns the wallet/ledger domain: wallets (with
-- pockets/sub-balances), the transaction ledger, wallet-to-wallet
-- transfers, payment requests (QR pay / request money), and linked payment
-- methods. See Readme.md for the system design this schema implements.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema, which runs each statement
-- separately by splitting the file on every semicolon character. It's
-- plain idempotent DDL (IF NOT EXISTS everywhere) - no PL/pgSQL blocks, and
-- no extra semicolon characters inside comments or string literals, since
-- either would break that naive split - so re-running it on every deploy is
-- safe. There is no seed data: this service is fully usable through its own
-- API (create a wallet, then deposit/withdraw/transfer).
--
-- balance/amount columns are BIGINT storing minor currency units (e.g.
-- cents for USD) rather than DECIMAL, matching internal/domain/model.Wallet
-- - integer arithmetic in both the app and the DB, no float rounding drift.

CREATE TABLE IF NOT EXISTS wallets (
    wallet_id UUID PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL UNIQUE,
    balance BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('ACTIVE', 'FROZEN', 'CLOSED')) NOT NULL DEFAULT 'ACTIVE',
    -- 0 = unlimited. Gate Withdraw and the outgoing leg of Transfer only;
    -- top-up limits are the PSP's concern, not this ledger's.
    daily_limit BIGINT NOT NULL DEFAULT 0,
    per_transaction_limit BIGINT NOT NULL DEFAULT 0,
    frozen_reason TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT positive_balance CHECK (balance >= 0)
);

CREATE TABLE IF NOT EXISTS transactions (
    transaction_id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(wallet_id),
    type VARCHAR(20) CHECK (type IN (
        'DEPOSIT', 'WITHDRAW', 'TRANSFER_IN', 'TRANSFER_OUT',
        'REVERSAL', 'POCKET_OUT', 'POCKET_IN'
    )) NOT NULL,
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')) NOT NULL DEFAULT 'PENDING',
    reference_id VARCHAR(100),
    counterparty_wallet_id UUID REFERENCES wallets(wallet_id),
    pocket_id UUID,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS transfers (
    transfer_id UUID PRIMARY KEY,
    sender_wallet_id UUID NOT NULL REFERENCES wallets(wallet_id),
    receiver_wallet_id UUID NOT NULL REFERENCES wallets(wallet_id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED', 'REVERSED')) NOT NULL DEFAULT 'PENDING',
    reference_id VARCHAR(100),
    out_transaction_id UUID NOT NULL REFERENCES transactions(transaction_id),
    in_transaction_id UUID NOT NULL REFERENCES transactions(transaction_id),
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS payment_methods (
    payment_method_id UUID PRIMARY KEY,
    user_id VARCHAR(100) NOT NULL,
    type VARCHAR(20) CHECK (type IN ('BANK_ACCOUNT', 'CARD')) NOT NULL,
    provider VARCHAR(50) NOT NULL,
    account_number VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT FALSE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

-- A pocket is a sub-balance of its wallet (a savings jar / spending
-- envelope): money moves between wallet.balance and pocket.balance via
-- POCKET_OUT/POCKET_IN transactions, but pockets are never a Deposit/
-- Withdraw/Transfer endpoint on their own - only the parent wallet is.
CREATE TABLE IF NOT EXISTS pockets (
    pocket_id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(wallet_id),
    name VARCHAR(100) NOT NULL,
    balance BIGINT NOT NULL DEFAULT 0,
    currency VARCHAR(3) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    CONSTRAINT positive_pocket_balance CHECK (balance >= 0)
);

-- Backs both QR-code merchant payment (payer_wallet_id left NULL - anyone
-- who scans the code and pays fulfills it) and P2P "request money"
-- (payer_wallet_id set - only that wallet may pay it).
CREATE TABLE IF NOT EXISTS payment_requests (
    payment_request_id UUID PRIMARY KEY,
    requester_wallet_id UUID NOT NULL REFERENCES wallets(wallet_id),
    payer_wallet_id UUID REFERENCES wallets(wallet_id),
    amount BIGINT NOT NULL CHECK (amount > 0),
    currency VARCHAR(3) NOT NULL,
    status VARCHAR(20) CHECK (status IN ('PENDING', 'PAID', 'CANCELLED', 'EXPIRED')) NOT NULL DEFAULT 'PENDING',
    description TEXT,
    transfer_id UUID REFERENCES transfers(transfer_id),
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_transactions_wallet_created ON transactions(wallet_id, created_at);
-- Idempotency keys. "No key" is stored as an empty string, not NULL, so the empty string must be
-- excluded too: otherwise a wallet could have only one reversal (or one key-less transfer) ever.
-- The v2 indexes replace the originals, which did not exclude it; dropping the old names upgrades
-- databases that already have them.
DROP INDEX IF EXISTS idx_transactions_wallet_reference;
DROP INDEX IF EXISTS idx_transfers_reference;
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_wallet_reference_v2 ON transactions(wallet_id, reference_id) WHERE reference_id IS NOT NULL AND reference_id <> '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_transfers_reference_v2 ON transfers(reference_id) WHERE reference_id IS NOT NULL AND reference_id <> '';
CREATE INDEX IF NOT EXISTS idx_payment_methods_user ON payment_methods(user_id) WHERE is_active;
CREATE INDEX IF NOT EXISTS idx_pockets_wallet ON pockets(wallet_id);
CREATE INDEX IF NOT EXISTS idx_payment_requests_requester ON payment_requests(requester_wallet_id);
CREATE INDEX IF NOT EXISTS idx_payment_requests_payer ON payment_requests(payer_wallet_id);
CREATE INDEX IF NOT EXISTS idx_payment_requests_status ON payment_requests(status);
