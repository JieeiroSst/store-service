CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    total_points BIGINT DEFAULT 0 NOT NULL,
    experience_level INTEGER DEFAULT 1 NOT NULL
);

CREATE TABLE reward_types (
    reward_type_id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50) NOT NULL CHECK (category IN ('POINTS', 'BADGE', 'DISCOUNT', 'EXPERIENCE', 'PREMIUM_CONTENT')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE links (
    link_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url TEXT NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    created_by UUID REFERENCES users(user_id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    is_active BOOLEAN DEFAULT TRUE
);

CREATE TABLE link_rewards (
    link_reward_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id UUID NOT NULL REFERENCES links(link_id) ON DELETE CASCADE,
    reward_type_id INTEGER NOT NULL REFERENCES reward_types(reward_type_id),
    reward_value NUMERIC(10, 2) NOT NULL,  
    max_claims_per_user INTEGER DEFAULT 1,
    expires_at TIMESTAMP WITH TIME ZONE,  
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE user_link_clicks (
    click_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id),
    link_id UUID NOT NULL REFERENCES links(link_id),
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ip_address INET,
    user_agent TEXT,
    is_valid BOOLEAN DEFAULT TRUE 
);

CREATE TABLE user_rewards (
    user_reward_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id),
    link_reward_id UUID NOT NULL REFERENCES link_rewards(link_reward_id),
    click_id UUID NOT NULL REFERENCES user_link_clicks(click_id),
    reward_value NUMERIC(10, 2) NOT NULL,
    awarded_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    redeemed_at TIMESTAMP WITH TIME ZONE, 
    expires_at TIMESTAMP WITH TIME ZONE    
);

CREATE TABLE user_badges (
    user_badge_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id),
    badge_name VARCHAR(100) NOT NULL,
    description TEXT NOT NULL,
    image_url TEXT,
    earned_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE premium_content_access (
    access_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(user_id),
    content_id VARCHAR(100) NOT NULL,  
    granted_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP WITH TIME ZONE, 
    granted_through_reward_id UUID REFERENCES user_rewards(user_reward_id)
);

CREATE TABLE reward_statistics (
    stat_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    date DATE NOT NULL,
    reward_type_id INTEGER REFERENCES reward_types(reward_type_id),
    link_id UUID REFERENCES links(link_id),
    clicks_count INTEGER DEFAULT 0,
    rewards_awarded_count INTEGER DEFAULT 0,
    total_value_awarded NUMERIC(15, 2) DEFAULT 0
);

CREATE INDEX idx_user_link_clicks_user_id ON user_link_clicks(user_id);
CREATE INDEX idx_user_link_clicks_link_id ON user_link_clicks(link_id);
CREATE INDEX idx_user_rewards_user_id ON user_rewards(user_id);
CREATE INDEX idx_user_rewards_link_reward_id ON user_rewards(link_reward_id);
CREATE INDEX idx_link_rewards_link_id ON link_rewards(link_id);

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_links_updated_at
BEFORE UPDATE ON links
FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

INSERT INTO reward_types (name, description, category) VALUES
('Basic Points', 'Standard points awarded for clicking links', 'POINTS'),
('First Click Badge', 'Badge awarded for first link click', 'BADGE'),
('10% Discount', 'Discount coupon for purchases', 'DISCOUNT'),
('Level-Up XP', 'Experience points toward level progression', 'EXPERIENCE'),
('Premium Article Access', 'Access to premium articles', 'PREMIUM_CONTENT');