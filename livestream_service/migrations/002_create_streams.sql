CREATE TABLE IF NOT EXISTS streams (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    room_id     UUID NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    node_id     TEXT NOT NULL,
    status      TEXT NOT NULL DEFAULT 'pending',
    peak_viewer INTEGER NOT NULL DEFAULT 0,
    started_at  TIMESTAMPTZ,
    ended_at    TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_streams_room_id ON streams (room_id);
-- Only one active (pending/live) stream per room at a time.
CREATE UNIQUE INDEX IF NOT EXISTS idx_streams_room_active
    ON streams (room_id)
    WHERE status IN ('pending', 'live');
