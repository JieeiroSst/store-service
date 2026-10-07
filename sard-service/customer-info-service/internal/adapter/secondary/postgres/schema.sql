CREATE TABLE IF NOT EXISTS customers (
    id                text PRIMARY KEY,
    user_id           bigint      NOT NULL UNIQUE,
    username          text        NOT NULL DEFAULT '',
    full_name         text        NOT NULL DEFAULT '',
    email             text        NOT NULL DEFAULT '',
    phone             text        NOT NULL DEFAULT '',
    address           text        NOT NULL DEFAULT '',
    gender            text        NOT NULL DEFAULT '',
    status            text        NOT NULL,
    kyc_status        text        NOT NULL,
    kyc_document_hash text        NOT NULL DEFAULT '',
    kyc               jsonb       NOT NULL DEFAULT '{}'::jsonb,
    created_at        timestamptz NOT NULL,
    updated_at        timestamptz NOT NULL,
    synced_at         timestamptz NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS customers_verified_document_idx
    ON customers (kyc_document_hash)
    WHERE kyc_status = 'VERIFIED' AND kyc_document_hash <> '';
