CREATE TABLE IF NOT EXISTS machines (
    machine_id VARCHAR(36) PRIMARY KEY,
    location VARCHAR(255) NOT NULL,
    model VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, inactive, maintenance, out_of_service
    last_maintenance_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_machines_location ON machines(location);
CREATE INDEX IF NOT EXISTS idx_machines_status ON machines(status);

CREATE TABLE IF NOT EXISTS categories (
    category_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS products (
    product_id VARCHAR(36) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price_cents INTEGER NOT NULL CHECK (price_cents >= 0),
    category_id VARCHAR(36) NOT NULL,
    image_url VARCHAR(255) NOT NULL DEFAULT '',
    barcode VARCHAR(100) NOT NULL DEFAULT '',
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (category_id) REFERENCES categories(category_id)
);

CREATE INDEX IF NOT EXISTS idx_products_category ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_active ON products(is_active);
CREATE INDEX IF NOT EXISTS idx_products_barcode ON products(barcode);

CREATE TABLE IF NOT EXISTS product_attributes (
    attribute_id VARCHAR(36) PRIMARY KEY,
    product_id VARCHAR(36) NOT NULL,
    key VARCHAR(100) NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE,
    UNIQUE (product_id, key)
);

CREATE INDEX IF NOT EXISTS idx_product_attributes_product ON product_attributes(product_id);

CREATE TABLE IF NOT EXISTS inventory (
    inventory_id VARCHAR(36) PRIMARY KEY,
    machine_id VARCHAR(36) NOT NULL,
    product_id VARCHAR(36) NOT NULL,
    slot_identifier VARCHAR(20) NOT NULL,  -- Physical location in machine (e.g., "A1", "B3")
    quantity INTEGER NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    max_capacity INTEGER NOT NULL CHECK (max_capacity > 0),
    low_threshold INTEGER NOT NULL,  -- Alert threshold for restocking
    last_restocked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES machines(machine_id),
    FOREIGN KEY (product_id) REFERENCES products(product_id),
    UNIQUE (machine_id, slot_identifier),
    CHECK (quantity <= max_capacity)
);

CREATE INDEX IF NOT EXISTS idx_inventory_machine ON inventory(machine_id);
CREATE INDEX IF NOT EXISTS idx_inventory_product ON inventory(product_id);
CREATE INDEX IF NOT EXISTS idx_inventory_low ON inventory(quantity) WHERE quantity <= low_threshold;

CREATE TABLE IF NOT EXISTS sessions (
    session_id VARCHAR(36) PRIMARY KEY,
    machine_id VARCHAR(36) NOT NULL,
    started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',  -- active, completed, expired, cancelled
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES machines(machine_id)
);

CREATE INDEX IF NOT EXISTS idx_sessions_machine ON sessions(machine_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_expires ON sessions(expires_at);

CREATE TABLE IF NOT EXISTS reservations (
    reservation_id VARCHAR(36) PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    inventory_id VARCHAR(36) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, confirmed, cancelled, expired
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id),
    FOREIGN KEY (inventory_id) REFERENCES inventory(inventory_id)
);

CREATE INDEX IF NOT EXISTS idx_reservations_session ON reservations(session_id);
CREATE INDEX IF NOT EXISTS idx_reservations_inventory ON reservations(inventory_id);
CREATE INDEX IF NOT EXISTS idx_reservations_status ON reservations(status);
CREATE INDEX IF NOT EXISTS idx_reservations_expires ON reservations(expires_at);

CREATE TABLE IF NOT EXISTS payments (
    payment_id VARCHAR(36) PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    reservation_id VARCHAR(36) NOT NULL,
    amount_cents INTEGER NOT NULL,
    currency VARCHAR(3) NOT NULL DEFAULT 'USD',
    payment_method VARCHAR(50) NOT NULL,  -- card, cash, mobile, etc.
    payment_status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, completed, failed, refunded
    transaction_id VARCHAR(100) NOT NULL DEFAULT '',  -- External payment processor ID
    payment_metadata JSONB NOT NULL DEFAULT '{}',  -- Store payment-specific data
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id),
    FOREIGN KEY (reservation_id) REFERENCES reservations(reservation_id)
);

CREATE INDEX IF NOT EXISTS idx_payments_session ON payments(session_id);
-- At most one live payment per reservation, so a double-submitted checkout cannot charge twice.
CREATE UNIQUE INDEX IF NOT EXISTS idx_payments_live_reservation ON payments(reservation_id)
    WHERE payment_status IN ('pending', 'completed');
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(payment_status);
CREATE INDEX IF NOT EXISTS idx_payments_transaction ON payments(transaction_id);

CREATE TABLE IF NOT EXISTS orders (
    order_id VARCHAR(36) PRIMARY KEY,
    session_id VARCHAR(36) NOT NULL,
    reservation_id VARCHAR(36) NOT NULL UNIQUE,
    payment_id VARCHAR(36),
    status VARCHAR(50) NOT NULL DEFAULT 'pending',  -- pending, processing, completed, failed, cancelled
    fulfilled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(session_id),
    FOREIGN KEY (reservation_id) REFERENCES reservations(reservation_id),
    FOREIGN KEY (payment_id) REFERENCES payments(payment_id)
);

CREATE INDEX IF NOT EXISTS idx_orders_session ON orders(session_id);
CREATE INDEX IF NOT EXISTS idx_orders_payment ON orders(payment_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);

CREATE TABLE IF NOT EXISTS events (
    event_id VARCHAR(36) PRIMARY KEY,
    event_type VARCHAR(100) NOT NULL,
    related_entity VARCHAR(50) NOT NULL,  -- machine, session, order, payment, etc.
    entity_id VARCHAR(36) NOT NULL,
    machine_id VARCHAR(36) NOT NULL DEFAULT '',
    data JSONB NOT NULL DEFAULT '{}',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_type ON events(event_type);
CREATE INDEX IF NOT EXISTS idx_events_entity ON events(related_entity, entity_id);
CREATE INDEX IF NOT EXISTS idx_events_machine ON events(machine_id);
CREATE INDEX IF NOT EXISTS idx_events_time ON events(occurred_at);

CREATE TABLE IF NOT EXISTS maintenance_logs (
    log_id VARCHAR(36) PRIMARY KEY,
    machine_id VARCHAR(36) NOT NULL,
    technician_id VARCHAR(36) NOT NULL DEFAULT '',
    maintenance_type VARCHAR(100) NOT NULL,  -- restock, repair, cleaning, etc.
    notes TEXT NOT NULL DEFAULT '',
    performed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (machine_id) REFERENCES machines(machine_id)
);

CREATE INDEX IF NOT EXISTS idx_maintenance_machine ON maintenance_logs(machine_id);
CREATE INDEX IF NOT EXISTS idx_maintenance_type ON maintenance_logs(maintenance_type);
CREATE INDEX IF NOT EXISTS idx_maintenance_time ON maintenance_logs(performed_at);
