CREATE TABLE audit_events (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type    VARCHAR(100) NOT NULL,
    entity_id      UUID NOT NULL,
    action         VARCHAR(50) NOT NULL,
    project_id     UUID REFERENCES projects(id) ON DELETE SET NULL,
    environment_id UUID REFERENCES environments(id) ON DELETE SET NULL,
    user_id        TEXT, -- user_service user ID (external, no local FK)
    before_json    JSONB,
    after_json     JSONB,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_audit_events_entity ON audit_events(entity_type, entity_id);
CREATE INDEX idx_audit_events_project_id ON audit_events(project_id);
CREATE INDEX idx_audit_events_created_at ON audit_events(created_at);
