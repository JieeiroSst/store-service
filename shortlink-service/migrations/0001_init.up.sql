-- Initial schema for LinkForty Core (Go port).
-- Flattened from the upstream TypeScript project's imperative
-- CREATE TABLE / ALTER TABLE bootstrap (src/lib/database.ts) into the
-- final shape, rather than replaying every historical migration step.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- Organizations: minimal by design. Only what the redirect fallback chain
-- and the owner-restriction gate read. Multi-tenant scoping is optional;
-- deployments that never populate organization_id simply get NULL joins.
CREATE TABLE organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255),
    settings JSONB NOT NULL DEFAULT '{}',
    suspended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE link_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    settings JSONB NOT NULL DEFAULT '{}',
    is_default BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE links (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    organization_id UUID REFERENCES organizations(id) ON DELETE SET NULL,
    template_id UUID REFERENCES link_templates(id) ON DELETE SET NULL,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    title VARCHAR(255),
    description TEXT,
    -- App store URLs (renamed from ios_url/android_url upstream for clarity)
    ios_app_store_url TEXT,
    android_app_store_url TEXT,
    web_fallback_url TEXT,
    -- App deep linking configuration
    app_scheme VARCHAR(255),
    ios_universal_link TEXT,
    android_app_link TEXT,
    deep_link_path TEXT,
    deep_link_parameters JSONB NOT NULL DEFAULT '{}',
    -- UTM + targeting
    utm_parameters JSONB NOT NULL DEFAULT '{}',
    targeting_rules JSONB NOT NULL DEFAULT '{}',
    -- Open Graph
    og_title VARCHAR(255),
    og_description TEXT,
    og_image_url TEXT,
    og_type VARCHAR(50) DEFAULT 'website',
    -- Attribution
    attribution_window_hours INTEGER NOT NULL DEFAULT 168,
    -- Safety / lifecycle
    is_active BOOLEAN NOT NULL DEFAULT true,
    expires_at TIMESTAMPTZ,
    warn_at TIMESTAMPTZ,
    disabled_at TIMESTAMPTZ,
    disabled_reason TEXT,
    append_click_id BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE click_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    device_type VARCHAR(20),
    platform VARCHAR(20),
    country_code CHAR(2),
    country_name VARCHAR(100),
    region VARCHAR(100),
    city VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    timezone VARCHAR(100),
    utm_source VARCHAR(255),
    utm_medium VARCHAR(255),
    utm_campaign VARCHAR(255),
    referrer TEXT,
    is_bot BOOLEAN NOT NULL DEFAULT false,
    bot_reason VARCHAR(16)
);

CREATE TABLE device_fingerprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    click_id UUID NOT NULL REFERENCES click_events(id) ON DELETE CASCADE,
    fingerprint_hash VARCHAR(64) NOT NULL,
    ip_address INET,
    user_agent TEXT,
    timezone VARCHAR(100),
    language VARCHAR(10),
    screen_width INTEGER,
    screen_height INTEGER,
    platform VARCHAR(50),
    platform_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE install_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID REFERENCES links(id) ON DELETE SET NULL,
    click_id UUID REFERENCES click_events(id) ON DELETE SET NULL,
    fingerprint_hash VARCHAR(64) NOT NULL,
    confidence_score DECIMAL(5, 2),
    attribution_method VARCHAR(20),
    matched_factors TEXT[],
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    first_open_at TIMESTAMPTZ,
    deep_link_retrieved BOOLEAN NOT NULL DEFAULT false,
    deep_link_data JSONB NOT NULL DEFAULT '{}',
    attribution_window_hours INTEGER NOT NULL DEFAULT 168,
    ip_address INET,
    user_agent TEXT,
    timezone VARCHAR(100),
    language VARCHAR(10),
    screen_width INTEGER,
    screen_height INTEGER,
    platform VARCHAR(50),
    platform_version VARCHAR(50),
    device_id VARCHAR(255),
    sdk_name VARCHAR(50),
    sdk_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE in_app_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    install_id UUID NOT NULL REFERENCES install_events(id) ON DELETE CASCADE,
    event_name VARCHAR(255) NOT NULL,
    event_data JSONB NOT NULL DEFAULT '{}',
    event_timestamp TIMESTAMPTZ NOT NULL,
    attributed_link_id UUID REFERENCES links(id) ON DELETE SET NULL,
    attributed_click_id UUID,
    attributed_at TIMESTAMPTZ,
    session_id UUID,
    sdk_name VARCHAR(50),
    sdk_version VARCHAR(50),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE webhooks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    name VARCHAR(255) NOT NULL,
    url TEXT NOT NULL,
    secret VARCHAR(255) NOT NULL,
    events TEXT[] NOT NULL DEFAULT '{}',
    is_active BOOLEAN NOT NULL DEFAULT true,
    retry_count INTEGER NOT NULL DEFAULT 3,
    timeout_ms INTEGER NOT NULL DEFAULT 10000,
    headers JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_links_short_code ON links(short_code);
CREATE INDEX idx_links_user_id ON links(user_id);
CREATE INDEX idx_links_created_at ON links(created_at DESC);
CREATE INDEX idx_links_template_id ON links(template_id);
CREATE INDEX idx_links_organization_id ON links(organization_id);

CREATE INDEX idx_clicks_link_id ON click_events(link_id);
CREATE INDEX idx_clicks_timestamp ON click_events(clicked_at DESC);
CREATE INDEX idx_clicks_link_date ON click_events(link_id, clicked_at DESC);
CREATE INDEX idx_clicks_human_link_date ON click_events(link_id, clicked_at DESC) WHERE is_bot = false;

CREATE INDEX idx_fingerprints_hash ON device_fingerprints(fingerprint_hash);
CREATE INDEX idx_fingerprints_click_id ON device_fingerprints(click_id);

CREATE INDEX idx_installs_fingerprint ON install_events(fingerprint_hash);
CREATE INDEX idx_installs_link_id ON install_events(link_id);
CREATE INDEX idx_installs_timestamp ON install_events(installed_at DESC);
CREATE INDEX idx_installs_link_date ON install_events(link_id, installed_at DESC);

CREATE UNIQUE INDEX idx_link_templates_slug ON link_templates(slug);
CREATE INDEX idx_link_templates_user_id ON link_templates(user_id);

CREATE INDEX idx_webhooks_user_id ON webhooks(user_id);
CREATE INDEX idx_webhooks_active ON webhooks(is_active) WHERE is_active = true;

CREATE INDEX idx_in_app_events_install_id ON in_app_events(install_id);
CREATE INDEX idx_in_app_events_name ON in_app_events(event_name);
CREATE INDEX idx_in_app_events_timestamp ON in_app_events(event_timestamp DESC);
CREATE INDEX idx_in_app_events_attributed_link ON in_app_events(attributed_link_id, event_timestamp DESC);
CREATE INDEX idx_in_app_events_session ON in_app_events(session_id) WHERE session_id IS NOT NULL;
