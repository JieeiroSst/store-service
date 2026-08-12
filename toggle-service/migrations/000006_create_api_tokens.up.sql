CREATE TABLE api_tokens (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name           VARCHAR(255) NOT NULL,
    token_hash     VARCHAR(255) NOT NULL UNIQUE,
    token_prefix   VARCHAR(32) NOT NULL,
    type           VARCHAR(50) NOT NULL,
    project_id     UUID REFERENCES projects(id) ON DELETE CASCADE,
    environment_id UUID REFERENCES environments(id) ON DELETE CASCADE,
    expires_at     TIMESTAMPTZ,
    created_by     TEXT, -- user_service user ID (external, no local FK)
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
