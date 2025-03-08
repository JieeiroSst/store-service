CREATE TABLE subscriptions (
    subscription_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    plan_id UUID NOT NULL REFERENCES subscription_plans(plan_id),
    status VARCHAR(20) NOT NULL, -- active, cancelled, expired, trial
    start_date DATE NOT NULL,
    end_date DATE,
    auto_renewal BOOLEAN DEFAULT TRUE,
    trial_end_date DATE,
    next_billing_date DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Payment Methods - Lưu phương thức thanh toán
CREATE TABLE payment_methods (
    payment_method_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    provider VARCHAR(50) NOT NULL, -- visa, mastercard, paypal, etc.
    token_id VARCHAR(255), -- Token ID từ payment gateway
    last_four_digits VARCHAR(4),
    expiry_date VARCHAR(7), -- MM/YYYY
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Transactions - Lưu lịch sử thanh toán
CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY,
    subscription_id UUID NOT NULL REFERENCES subscriptions(subscription_id),
    payment_method_id UUID REFERENCES payment_methods(payment_method_id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'USD',
    status VARCHAR(20) NOT NULL, -- successful, failed, pending, refunded
    gateway_transaction_id VARCHAR(255),
    error_message TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Bảng Invoices - Lưu hóa đơn
CREATE TABLE invoices (
    invoice_id UUID PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(transaction_id),
    subscription_id UUID NOT NULL REFERENCES subscriptions(subscription_id),
    user_id UUID NOT NULL REFERENCES users(user_id),
    amount DECIMAL(10, 2) NOT NULL,
    tax_amount DECIMAL(10, 2) DEFAULT 0,
    total_amount DECIMAL(10, 2) NOT NULL,
    status VARCHAR(20) NOT NULL, -- paid, unpaid, void
    due_date DATE NOT NULL,
    paid_date DATE,
    invoice_number VARCHAR(50) UNIQUE NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);