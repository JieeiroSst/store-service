-- coupon-service owns the coupon domain: coupon definitions, per-coupon
-- restrictions (category/product/user-group), redemption usage records and
-- user-targeted coupon assignments. See Readme.md for the full domain model.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema, which runs each statement
-- separately by splitting the file on every semicolon character. It's
-- plain idempotent DDL (IF NOT EXISTS everywhere) - no PL/pgSQL blocks, and
-- no extra semicolon characters inside comments or string literals, since
-- either would break that naive split - so re-running it on every deploy
-- is safe.

CREATE TABLE IF NOT EXISTS coupons (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL, -- 'percentage', 'fixed_amount', 'buy_x_get_y'
    discount_value DECIMAL(10,2) NOT NULL,
    minimum_purchase DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_discount_amount DECIMAL(10,2), -- NULL for unlimited
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    max_uses INTEGER, -- NULL for unlimited
    current_uses INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT valid_dates CHECK (end_date > start_date),
    CONSTRAINT valid_discount CHECK (discount_value > 0),
    CONSTRAINT valid_minimum CHECK (minimum_purchase >= 0)
);

-- Category/product/user-group scoping for a coupon. is_exclude=false is an
-- inclusion entry ("only applies to this entity"), is_exclude=true is an
-- exclusion entry ("never applies to this entity").
CREATE TABLE IF NOT EXISTS coupon_restrictions (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    restriction_type VARCHAR(20) NOT NULL, -- 'category', 'product', 'user_group'
    restricted_entity_id BIGINT NOT NULL,
    is_exclude BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(coupon_id, restriction_type, restricted_entity_id)
);

-- One row per successful redemption. The UNIQUE(order_id, coupon_id)
-- constraint is the DB-level guard against double-applying a coupon to the
-- same order (see application.couponService.ApplyCoupon).
CREATE TABLE IF NOT EXISTS coupon_usage (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(order_id, coupon_id)
);

-- Targeted coupon assignments. A coupon with no rows here at all is public,
-- one with rows can only be redeemed by users who have an unused row.
CREATE TABLE IF NOT EXISTS user_coupons (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    is_used BOOLEAN NOT NULL DEFAULT false,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP,

    UNIQUE(user_id, coupon_id)
);

CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_dates ON coupons(start_date, end_date);
CREATE INDEX IF NOT EXISTS idx_coupon_restrictions_coupon ON coupon_restrictions(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_user ON coupon_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_coupon_usage_order ON coupon_usage(order_id);
CREATE INDEX IF NOT EXISTS idx_user_coupons_user ON user_coupons(user_id);
