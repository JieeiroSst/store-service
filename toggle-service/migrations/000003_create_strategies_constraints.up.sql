CREATE TABLE activation_strategies (
    id                          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    feature_flag_environment_id UUID NOT NULL REFERENCES feature_flag_environments(id) ON DELETE CASCADE,
    strategy_type               VARCHAR(50) NOT NULL,
    parameters                  JSONB NOT NULL DEFAULT '{}'::jsonb,
    sort_order                  INTEGER NOT NULL DEFAULT 0,
    created_at                  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at                  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE constraints (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    strategy_id      UUID NOT NULL REFERENCES activation_strategies(id) ON DELETE CASCADE,
    context_field    VARCHAR(255) NOT NULL,
    operator         VARCHAR(50) NOT NULL,
    values           JSONB NOT NULL DEFAULT '[]'::jsonb,
    case_insensitive BOOLEAN NOT NULL DEFAULT false
);

CREATE INDEX idx_activation_strategies_flag_env ON activation_strategies(feature_flag_environment_id);
CREATE INDEX idx_constraints_strategy_id ON constraints(strategy_id);
