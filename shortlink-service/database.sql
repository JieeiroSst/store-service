-- Links Table
CREATE TABLE links (
    id VARCHAR(255) PRIMARY KEY,
    original_url TEXT NOT NULL,
    shortlink TEXT NOT NULL,
    short_code VARCHAR(50) NOT NULL UNIQUE,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    expired_at TIMESTAMP WITH TIME ZONE,
    total_clicks INTEGER DEFAULT 0,
    status INTEGER DEFAULT 1, -- 1: active, 0: inactive
    
    -- Optional additional constraints
    CONSTRAINT valid_short_code CHECK (length(short_code) > 0),
    CONSTRAINT valid_original_url CHECK (original_url ~ '^https?://'),
    CONSTRAINT future_expiration CHECK (expired_at IS NULL OR expired_at > created_at)
);

-- Link Clicks Table
CREATE TABLE link_clicks (
    id VARCHAR(255) PRIMARY KEY,
    link_id VARCHAR(255) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE NOT NULL,
    ip_address INET,
    country VARCHAR(100),
    browser VARCHAR(255),
    device_type VARCHAR(100),
    
    -- Foreign key constraint to links table
    CONSTRAINT fk_link 
        FOREIGN KEY (link_id) 
        REFERENCES links(id) 
        ON DELETE CASCADE
);

-- Indexes for performance
CREATE INDEX idx_links_user_id ON links(user_id);
CREATE INDEX idx_links_short_code ON links(short_code);
CREATE INDEX idx_link_clicks_link_id ON link_clicks(link_id);
CREATE INDEX idx_link_clicks_country ON link_clicks(country);
CREATE INDEX idx_link_clicks_browser ON link_clicks(browser);
CREATE INDEX idx_link_clicks_device_type ON link_clicks(device_type);