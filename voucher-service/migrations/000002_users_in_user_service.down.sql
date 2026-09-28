-- Fails if any wallet now belongs to a (numeric) user-service user id.
ALTER TABLE wallets ALTER COLUMN owner_id TYPE UUID USING owner_id::uuid;

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
ALTER TABLE users
    ADD CONSTRAINT fk_users_corporate FOREIGN KEY (corporate_id) REFERENCES corporates (id);
