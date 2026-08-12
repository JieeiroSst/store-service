CREATE TABLE feature_usage_metrics (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_flag_id UUID NOT NULL REFERENCES feature_flags(id) ON DELETE CASCADE,
    environment_id  UUID NOT NULL REFERENCES environments(id) ON DELETE CASCADE,
    app_name        VARCHAR(255) NOT NULL,
    yes_count       BIGINT NOT NULL DEFAULT 0,
    no_count        BIGINT NOT NULL DEFAULT 0,
    window_start    TIMESTAMPTZ NOT NULL,
    window_stop     TIMESTAMPTZ NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (feature_flag_id, environment_id, app_name, window_start)
);

CREATE INDEX idx_feature_usage_metrics_flag_id ON feature_usage_metrics(feature_flag_id);
