CREATE INDEX IF NOT EXISTS idx_payments_pending ON payments(created_at) WHERE payment_status = 'pending';
