# admanagement-service

Ad management & serving service: campaigns, ads, positions and targeting
rules go in; a weighted/targeted ad-serving engine and impression/click/CTR
analytics come out.

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                                  entrypoint - fx.New(infrastructure.Module).Run()
config/                                      env-based configuration

internal/
  domain/
    model/                                   plain structs, no framework deps (AdCampaign, Ad, ...)
    port/
      driving.go                             usecase interfaces - what adapters call INTO the app
      driven.go                              repository interfaces - what the app calls OUT to infra

  application/                               business logic, implements the driving ports
    campaign.go, category.go, position.go, ad.go, position_mapping.go,
    impression.go, click.go, targeting_rule.go
    performance_summary.go                   CTR rollups (per-ad and per-campaign)
    serving.go, targeting.go                 ad-serving engine: eligibility + weighted random pick

  adapter/
    primary/http/                            driving adapter - gorilla/mux HTTP handlers + routes
    secondary/repository/                    driven adapter - GORM/Postgres repositories

  infrastructure/                            fx wiring: config, logger, *gorm.DB, HTTP server lifecycle
```

Dependencies only ever point inward (adapter -> application -> domain). The
application layer never imports gorm or net/http directly - it only knows
the `port` interfaces, so the HTTP framework or the database driver can be
swapped without touching business logic.

## Features

Beyond CRUD over the 9 tables below, the service implements:

- **Campaign lifecycle** - `PATCH /campaigns/{id}/status` enforces the
  `draft -> active -> paused/completed/cancelled` transition graph instead
  of allowing an arbitrary status overwrite.
- **Ad-serving engine** - `GET /positions/{id}/serve?country=&device=&age=&...`
  picks a winning ad for a position: filters candidates by active
  campaign/ad state, schedule window (`start_date`/`end_date`), and every
  active `ad_targeting_rules` row (`equals`/`in`/`between`/`greater`/`less`
  over country, device, gender, age, hour-of-day), then does a
  weighted-random draw using `ad_position_mappings.weight`. Every served ad
  auto-records an `ad_impressions` row.
- **Click attribution & redirect** - `GET /clicks/track?ad_id=&impression_id=`
  records an `ad_clicks` row linked back to its impression and 302-redirects
  to the ad's `target_url`.
- **Performance analytics** - `POST /ads/{id}/performance/recompute?date=`
  rolls the day's raw impressions/clicks into `ad_performance_summary`
  (upsert, CTR computed as clicks/impressions); `GET /campaigns/{id}/performance?from=&to=`
  rolls those daily summaries up across every ad in a campaign.

See `internal/adapter/primary/http/routes.go` for the full endpoint list.

## Running locally

```
cp .env.example .env   # then edit as needed
go run ./cmd
```

Postgres tables are created automatically via `AutoMigrate` on startup
(see `internal/infrastructure/database/postgres.go`); the SQL below is kept
as the reference schema.

## Reference schema

***SQL***

```
-- =============================================
-- DATABASE SCHEMA FOR AD MANAGEMENT SYSTEM
-- =============================================

-- 1. Ad campaigns table - Quản lý chiến dịch quảng cáo
CREATE TABLE ad_campaigns (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    budget DECIMAL(15,2),
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    status VARCHAR(20) DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'completed', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. Ad categories table - Phân loại quảng cáo
CREATE TABLE ad_categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    parent_id INTEGER REFERENCES ad_categories(id),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 3. Ad positions table - Vị trí hiển thị quảng cáo
CREATE TABLE ad_positions (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL, -- 'header', 'sidebar', 'footer', 'popup', 'inline'
    description TEXT,
    width INTEGER,
    height INTEGER,
    max_file_size INTEGER, -- in bytes
    allowed_formats JSON, -- ['jpg', 'png', 'mp4', 'webm']
    is_active BOOLEAN DEFAULT true
);

-- 4. Ads table - Bảng chính chứa thông tin quảng cáo
CREATE TABLE ads (
    id SERIAL PRIMARY KEY,
    campaign_id INTEGER REFERENCES ad_campaigns(id) ON DELETE CASCADE,
    category_id INTEGER REFERENCES ad_categories(id),
    title VARCHAR(200) NOT NULL,
    description TEXT,
    ad_type VARCHAR(20) NOT NULL CHECK (ad_type IN ('image', 'video', 'banner', 'text', 'link')),
    content_url VARCHAR(500), -- URL của file media
    target_url VARCHAR(500), -- URL đích khi click
    file_path VARCHAR(500), -- Đường dẫn file trên server
    file_size INTEGER,
    mime_type VARCHAR(100),
    duration INTEGER, -- Thời gian hiển thị (giây) hoặc độ dài video
    priority INTEGER DEFAULT 0, -- Độ ưu tiên hiển thị
    is_active BOOLEAN DEFAULT true,
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 5. Ad position mappings - Mapping quảng cáo với vị trí
CREATE TABLE ad_position_mappings (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES ads(id) ON DELETE CASCADE,
    position_id INTEGER REFERENCES ad_positions(id) ON DELETE CASCADE,
    weight INTEGER DEFAULT 1, -- Trọng số để random hiển thị
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ad_id, position_id)
);

-- 6. Ad impressions - Theo dõi lượt hiển thị
CREATE TABLE ad_impressions (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES ads(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    referrer_url VARCHAR(500),
    page_url VARCHAR(500),
    position_id INTEGER REFERENCES ad_positions(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 7. Ad clicks - Theo dõi lượt click
CREATE TABLE ad_clicks (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES ads(id) ON DELETE CASCADE,
    impression_id INTEGER REFERENCES ad_impressions(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    session_id VARCHAR(100),
    ip_address INET,
    user_agent TEXT,
    referrer_url VARCHAR(500),
    target_url VARCHAR(500),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 8. Ad targeting rules - Quy tắc targeting
CREATE TABLE ad_targeting_rules (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES ads(id) ON DELETE CASCADE,
    rule_type VARCHAR(50) NOT NULL, -- 'country', 'age', 'gender', 'device', 'time'
    rule_operator VARCHAR(20) DEFAULT 'equals', -- 'equals', 'in', 'between', 'greater', 'less'
    rule_value JSON NOT NULL, -- Giá trị rule (có thể là array, object)
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 9. Ad performance summary - Bảng tổng hợp hiệu suất
CREATE TABLE ad_performance_summary (
    id SERIAL PRIMARY KEY,
    ad_id INTEGER REFERENCES ads(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    impressions INTEGER DEFAULT 0,
    clicks INTEGER DEFAULT 0,
    ctr DECIMAL(5,4) DEFAULT 0, -- Click-through rate
    cost DECIMAL(15,2) DEFAULT 0,
    revenue DECIMAL(15,2) DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ad_id, date)
);

-- Indexes for better performance
CREATE INDEX idx_ads_campaign_id ON ads(campaign_id);
CREATE INDEX idx_ads_type_active ON ads(ad_type, is_active);
CREATE INDEX idx_ads_start_end_date ON ads(start_date, end_date);
CREATE INDEX idx_impressions_ad_id_date ON ad_impressions(ad_id, created_at);
CREATE INDEX idx_clicks_ad_id_date ON ad_clicks(ad_id, created_at);
CREATE INDEX idx_performance_ad_date ON ad_performance_summary(ad_id, date);
CREATE INDEX idx_position_mappings_position ON ad_position_mappings(position_id, is_active);

-- Triggers for updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_campaigns_updated_at BEFORE UPDATE ON ad_campaigns
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_ads_updated_at BEFORE UPDATE ON ads
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
CREATE TRIGGER update_performance_updated_at BEFORE UPDATE ON ad_performance_summary
    FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
```