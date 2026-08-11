CREATE TABLE IF NOT EXISTS vod_recordings (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    stream_id        UUID NOT NULL REFERENCES streams (id) ON DELETE CASCADE,
    room_id          UUID NOT NULL REFERENCES rooms (id) ON DELETE CASCADE,
    object_key       TEXT NOT NULL,
    duration_seconds INTEGER NOT NULL DEFAULT 0,
    size_bytes       BIGINT NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_vod_recordings_room_id ON vod_recordings (room_id);
