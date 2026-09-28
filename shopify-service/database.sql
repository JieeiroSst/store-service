CREATE TABLE IF NOT EXISTS products (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shopify_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL,
    handle VARCHAR(255) NOT NULL DEFAULT '',
    status VARCHAR(32) NOT NULL DEFAULT '',
    vendor VARCHAR(255) NOT NULL DEFAULT '',
    product_type VARCHAR(255) NOT NULL DEFAULT '',
    tags TEXT,
    shopify_created_at DATETIME(3) NULL,
    shopify_updated_at DATETIME(3) NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_products_shopify_id (shopify_id)
);

CREATE TABLE IF NOT EXISTS product_variants (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    product_id BIGINT NOT NULL,
    shopify_id VARCHAR(255) NOT NULL,
    title VARCHAR(255) NOT NULL DEFAULT '',
    sku VARCHAR(255) NOT NULL DEFAULT '',
    price DECIMAL(20, 4) NOT NULL DEFAULT 0,
    inventory_quantity INT NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_product_variants_shopify_id (shopify_id),
    KEY idx_product_variants_product_id (product_id)
);

CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    shopify_id VARCHAR(255) NOT NULL,
    name VARCHAR(64) NOT NULL DEFAULT '',
    email VARCHAR(255) NOT NULL DEFAULT '',
    financial_status VARCHAR(64) NOT NULL DEFAULT '',
    fulfillment_status VARCHAR(64) NOT NULL DEFAULT '',
    currency VARCHAR(8) NOT NULL DEFAULT '',
    total_price DECIMAL(20, 4) NOT NULL DEFAULT 0,
    cancelled_at DATETIME(3) NULL,
    processed_at DATETIME(3) NULL,
    shopify_created_at DATETIME(3) NULL,
    shopify_updated_at DATETIME(3) NULL,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_orders_shopify_id (shopify_id),
    KEY idx_orders_processed_at (processed_at)
);

CREATE TABLE IF NOT EXISTS order_line_items (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    order_id BIGINT NOT NULL,
    shopify_id VARCHAR(255) NOT NULL,
    variant_shopify_id VARCHAR(255) NOT NULL DEFAULT '',
    title VARCHAR(255) NOT NULL DEFAULT '',
    sku VARCHAR(255) NOT NULL DEFAULT '',
    quantity INT NOT NULL DEFAULT 0,
    unit_price DECIMAL(20, 4) NOT NULL DEFAULT 0,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    UNIQUE KEY uk_order_line_items_shopify_id (shopify_id),
    KEY idx_order_line_items_order_id (order_id)
);

CREATE TABLE IF NOT EXISTS webhook_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    webhook_id VARCHAR(255) NOT NULL,
    topic VARCHAR(128) NOT NULL,
    shop_domain VARCHAR(255) NOT NULL DEFAULT '',
    processed_at DATETIME(3) NOT NULL,
    UNIQUE KEY uk_webhook_events_webhook_id (webhook_id)
);
