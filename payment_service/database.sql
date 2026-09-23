-- payment-service owns payment records: which provider (paypal, payoneer,
-- stripe or wise - see internal/domain/model.Provider) handled each one, the
-- provider's own external id, amount/currency and status. It does not own
-- customer identity or order data - those live in their own services and
-- are referenced only via the caller-supplied payer_email/description on
-- each payment.
--
-- transactions is the lifecycle log appended to on every gateway call a
-- payment goes through (create, refund, inbound webhook) - see
-- internal/application.paymentService.recordTransaction. The payments row
-- only ever holds current state (including the running refunded_amount for
-- partial refunds), so this table is what lets a payment's full history be
-- replayed. idempotency_key is unique (NULLs excluded) so a retried create
-- request can be detected and replayed instead of double-charging.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema, which runs each statement
-- separately by splitting the file on every semicolon character. Indexes
-- are declared inline in the CREATE TABLE statement (MySQL has no
-- CREATE INDEX IF NOT EXISTS) so the whole file stays idempotent and safe
-- to re-run on every deploy.

CREATE TABLE IF NOT EXISTS payments (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    provider VARCHAR(20) NOT NULL,
    external_id VARCHAR(128) NULL,
    amount BIGINT NOT NULL,
    currency CHAR(3) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    payer_email VARCHAR(255) NULL,
    description VARCHAR(255) NULL,
    failure_reason TEXT NULL,
    idempotency_key VARCHAR(128) NULL,
    refunded_amount BIGINT NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    KEY idx_payments_provider (provider),
    KEY idx_payments_status (status),
    KEY idx_payments_external_id (external_id),
    UNIQUE KEY idx_payments_idempotency_key (idempotency_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE IF NOT EXISTS transactions (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    payment_id BIGINT UNSIGNED NOT NULL,
    type VARCHAR(20) NOT NULL,
    status VARCHAR(20) NOT NULL,
    amount BIGINT NOT NULL,
    external_id VARCHAR(128) NULL,
    raw_status VARCHAR(64) NULL,
    error_message TEXT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    KEY idx_transactions_payment (payment_id),
    CONSTRAINT fk_transactions_payment FOREIGN KEY (payment_id) REFERENCES payments (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
