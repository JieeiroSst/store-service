ALTER TABLE orders ADD COLUMN IF NOT EXISTS order_no BIGINT GENERATED ALWAYS AS IDENTITY;
CREATE UNIQUE INDEX IF NOT EXISTS idx_orders_order_no ON orders(order_no);

ALTER TABLE payments ADD COLUMN IF NOT EXISTS discount_cents INTEGER NOT NULL DEFAULT 0;
ALTER TABLE payments ADD COLUMN IF NOT EXISTS coupon_code VARCHAR(50) NOT NULL DEFAULT '';

ALTER TABLE events ADD COLUMN IF NOT EXISTS dispatched_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_events_undispatched ON events(occurred_at) WHERE dispatched_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_orders_fulfilled ON orders(fulfilled_at) WHERE status = 'completed';
