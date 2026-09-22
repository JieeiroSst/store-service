-- ekyc-service owns citizen-ID (CCCD) and facial-biometric verification
-- data. It does NOT own user identity/auth - accounts and credentials
-- belong to user-service, which this service calls over HTTP to validate
-- that a user_id exists before accepting a submission for it (see
-- internal/adapter/secondary/userclient and internal/domain/port.UserClient).
--
-- user_id is stored as plain TEXT, not UUID: user-service's own primary
-- key is an integer, but every other service in this cluster already
-- treats it as an opaque string over the wire (see
-- internal/adapter/secondary/userclient/http_client.go) - there is no
-- cross-service foreign key, since each service owns its own database.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema. It's plain idempotent DDL
-- (IF NOT EXISTS everywhere), so re-running it on every deploy is safe.
--
-- Static files (card photos, chip portraits, face-scan crops) live in
-- MinIO, not here (see internal/adapter/secondary/storage and
-- port.ObjectStorage) - the *_key columns below hold only their object
-- keys.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS citizen_identities (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    -- 'card_ocr' (OCR over the printed card) or 'nfc_chip' (decoded from
    -- the chip's own DG1/DG2, verified against EF.SOD) - see
    -- internal/domain/model.IdentitySource.
    source TEXT NOT NULL DEFAULT 'card_ocr',
    document_number TEXT,
    surname TEXT,
    given_names TEXT,
    nationality TEXT,
    date_of_birth TEXT,
    sex TEXT,
    date_of_expiry TEXT,
    mrz_line1 TEXT,
    mrz_line2 TEXT,
    mrz_line3 TEXT,
    checksum_valid BOOLEAN NOT NULL DEFAULT FALSE,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0,
    front_image_key TEXT,
    back_image_key TEXT,
    -- True once a chip read's per-data-group hashes and EF.SOD signature
    -- both checked out (see nfcreader.SODResult) - not a full CSCA
    -- trust-chain validation, see that type's doc comment.
    nfc_verified BOOLEAN NOT NULL DEFAULT FALSE,
    chip_face_image_key TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_citizen_identities_user_id ON citizen_identities (user_id);
CREATE INDEX IF NOT EXISTS idx_citizen_identities_document_number ON citizen_identities (document_number);

CREATE TABLE IF NOT EXISTS face_biometrics (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    template BYTEA NOT NULL,
    face_image_key TEXT,
    quality_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_face_biometrics_user_id ON face_biometrics (user_id);

CREATE TABLE IF NOT EXISTS ekyc_verifications (
    id UUID PRIMARY KEY,
    user_id TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    match_score DOUBLE PRECISION NOT NULL DEFAULT 0,
    verified_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_ekyc_verifications_user_id ON ekyc_verifications (user_id);
