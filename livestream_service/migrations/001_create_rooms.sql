CREATE TABLE IF NOT EXISTS rooms (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_user_id   TEXT NOT NULL,
    slug            TEXT NOT NULL UNIQUE,
    title           TEXT NOT NULL,
    description     TEXT NOT NULL DEFAULT '',
    stream_key      TEXT NOT NULL UNIQUE,
    status          TEXT NOT NULL DEFAULT 'offline',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms (status);
CREATE INDEX IF NOT EXISTS idx_rooms_owner_user_id ON rooms (owner_user_id);
