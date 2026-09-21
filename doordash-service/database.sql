-- doordash-service owns order and delivery orchestration: placed orders,
-- their line items, delivery tracking checkpoints, and driver-assignment
-- offers. It does NOT own customer/driver identity (user_service), the
-- restaurant/menu catalog (restaurant service), payment methods or
-- captures (payment service), promotions (coupon-service), or
-- notification delivery (notification service) - those are fetched over
-- HTTP at request time (see internal/domain/port.UserClient,
-- RestaurantClient, PaymentClient, NotifierClient) rather than joined
-- against this database. customer_id / restaurant_id / driver_id /
-- menu_item_id / delivery_address_id below are plain string references
-- with no cross-service foreign key, since each service in this cluster
-- owns its own database.
--
-- This file is applied automatically on startup by
-- internal/infrastructure/database.applySchema, which runs each statement
-- separately by splitting the file on every semicolon character. It's
-- plain idempotent DDL (IF NOT EXISTS everywhere) - no PL/pgSQL blocks, and
-- no extra semicolon characters inside comments or string literals, since
-- either would break that naive split - so re-running it on every deploy
-- is safe.

CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id VARCHAR(64) NOT NULL,
    restaurant_id VARCHAR(64) NOT NULL,
    driver_id VARCHAR(64),
    delivery_address_id VARCHAR(64) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    placed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estimated_delivery_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    subtotal DECIMAL(10,2) NOT NULL DEFAULT 0,
    delivery_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    service_fee DECIMAL(10,2) NOT NULL DEFAULT 0,
    tax DECIMAL(10,2) NOT NULL DEFAULT 0,
    tip DECIMAL(10,2) NOT NULL DEFAULT 0,
    total_amount DECIMAL(10,2) NOT NULL DEFAULT 0,
    payment_method_id VARCHAR(64) NOT NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    special_instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_orders_customer ON orders(customer_id);
CREATE INDEX IF NOT EXISTS idx_orders_restaurant ON orders(restaurant_id);
CREATE INDEX IF NOT EXISTS idx_orders_driver ON orders(driver_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

CREATE TABLE IF NOT EXISTS order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    menu_item_id VARCHAR(64) NOT NULL,
    -- Name/unit_price are a snapshot of the restaurant service's menu item
    -- at order time (see model.OrderItem), so the order stays accurate
    -- even if the menu changes or the item is later removed.
    name VARCHAR(255) NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    customizations JSONB,
    special_instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_order_items_order ON order_items(order_id);

CREATE TABLE IF NOT EXISTS order_tracking (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    status VARCHAR(20) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);
CREATE INDEX IF NOT EXISTS idx_order_tracking_order ON order_tracking(order_id);

CREATE TABLE IF NOT EXISTS driver_assignments (
    id BIGSERIAL PRIMARY KEY,
    driver_id VARCHAR(64) NOT NULL,
    order_id BIGINT NOT NULL REFERENCES orders(id),
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    completed_at TIMESTAMP,
    rejection_reason TEXT
);
CREATE INDEX IF NOT EXISTS idx_driver_assignments_driver ON driver_assignments(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_assignments_order ON driver_assignments(order_id);
