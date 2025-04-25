-- Food Delivery System Database Schema

-- Users table (customers, drivers, restaurant staff, admins)
CREATE TABLE users (
    user_id UUID PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    user_type ENUM('customer', 'driver', 'restaurant_staff', 'admin') NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP
);

-- Customer profiles
CREATE TABLE customer_profiles (
    customer_id UUID PRIMARY KEY REFERENCES users(user_id),
    default_address_id UUID,
    payment_methods JSONB,
    preferences JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Driver profiles
CREATE TABLE driver_profiles (
    driver_id UUID PRIMARY KEY REFERENCES users(user_id),
    vehicle_type VARCHAR(50) NOT NULL,
    license_number VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT FALSE,
    current_location GEOGRAPHY(POINT),
    rating DECIMAL(3,2),
    account_details JSONB,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Restaurants
CREATE TABLE restaurants (
    restaurant_id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    cuisine_type VARCHAR(100),
    address_id UUID NOT NULL,
    contact_phone VARCHAR(20) NOT NULL,
    contact_email VARCHAR(255),
    operating_hours JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    rating DECIMAL(3,2),
    commission_rate DECIMAL(5,2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Restaurant staff relationship
CREATE TABLE restaurant_staff (
    staff_id UUID REFERENCES users(user_id),
    restaurant_id UUID REFERENCES restaurants(restaurant_id),
    role VARCHAR(50) NOT NULL,
    PRIMARY KEY (staff_id, restaurant_id)
);

-- Menu categories
CREATE TABLE menu_categories (
    category_id UUID PRIMARY KEY,
    restaurant_id UUID NOT NULL REFERENCES restaurants(restaurant_id),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    display_order INTEGER NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Menu items
CREATE TABLE menu_items (
    item_id UUID PRIMARY KEY,
    restaurant_id UUID NOT NULL REFERENCES restaurants(restaurant_id),
    category_id UUID REFERENCES menu_categories(category_id),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    image_url VARCHAR(512),
    preparation_time INTEGER, -- in minutes
    is_vegetarian BOOLEAN DEFAULT FALSE,
    is_vegan BOOLEAN DEFAULT FALSE,
    is_gluten_free BOOLEAN DEFAULT FALSE,
    spice_level INTEGER CHECK (spice_level BETWEEN 0 AND 5),
    customization_options JSONB,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Addresses
CREATE TABLE addresses (
    address_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    address_line1 VARCHAR(255) NOT NULL,
    address_line2 VARCHAR(255),
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100) NOT NULL,
    postal_code VARCHAR(20) NOT NULL,
    country VARCHAR(100) NOT NULL,
    latitude DECIMAL(10,8),
    longitude DECIMAL(11,8),
    is_default BOOLEAN DEFAULT FALSE,
    label VARCHAR(50), -- 'home', 'work', etc.
    delivery_instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Orders
CREATE TABLE orders (
    order_id UUID PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES users(user_id),
    restaurant_id UUID NOT NULL REFERENCES restaurants(restaurant_id),
    driver_id UUID REFERENCES users(user_id),
    delivery_address_id UUID NOT NULL REFERENCES addresses(address_id),
    order_status ENUM('created', 'confirmed', 'preparing', 'ready_for_pickup', 'picked_up', 'delivered', 'cancelled') NOT NULL,
    placed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    estimated_delivery_time TIMESTAMP,
    actual_delivery_time TIMESTAMP,
    subtotal DECIMAL(10,2) NOT NULL,
    delivery_fee DECIMAL(10,2) NOT NULL,
    service_fee DECIMAL(10,2) NOT NULL,
    tax DECIMAL(10,2) NOT NULL,
    tip DECIMAL(10,2) DEFAULT 0.00,
    total_amount DECIMAL(10,2) NOT NULL,
    payment_method_id VARCHAR(255) NOT NULL,
    payment_status ENUM('pending', 'authorized', 'captured', 'refunded', 'failed') NOT NULL,
    special_instructions TEXT,
    rating_for_restaurant DECIMAL(3,2),
    rating_for_driver DECIMAL(3,2),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Order items
CREATE TABLE order_items (
    order_item_id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(order_id),
    menu_item_id UUID NOT NULL REFERENCES menu_items(item_id),
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    unit_price DECIMAL(10,2) NOT NULL,
    customizations JSONB,
    special_instructions TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Order tracking events
CREATE TABLE order_tracking (
    tracking_id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(order_id),
    status ENUM('created', 'confirmed', 'preparing', 'ready_for_pickup', 'picked_up', 'delivered', 'cancelled') NOT NULL,
    location GEOGRAPHY(POINT),
    timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    notes TEXT
);

-- Payments
CREATE TABLE payments (
    payment_id UUID PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders(order_id),
    amount DECIMAL(10,2) NOT NULL,
    payment_method_type ENUM('credit_card', 'debit_card', 'paypal', 'apple_pay', 'google_pay') NOT NULL,
    payment_status ENUM('pending', 'authorized', 'captured', 'refunded', 'failed') NOT NULL,
    transaction_id VARCHAR(255),
    processed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    refunded_amount DECIMAL(10,2) DEFAULT 0.00,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Promotions and coupons
CREATE TABLE promotions (
    promotion_id UUID PRIMARY KEY,
    code VARCHAR(50) UNIQUE,
    description TEXT,
    discount_type ENUM('percentage', 'fixed_amount') NOT NULL,
    discount_value DECIMAL(10,2) NOT NULL,
    minimum_order_value DECIMAL(10,2),
    max_discount_amount DECIMAL(10,2),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    usage_limit INTEGER,
    current_usage_count INTEGER DEFAULT 0,
    restaurant_id UUID REFERENCES restaurants(restaurant_id), -- NULL for platform-wide promotions
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- User promotions (for tracking which users have used which promotions)
CREATE TABLE user_promotions (
    user_id UUID REFERENCES users(user_id),
    promotion_id UUID REFERENCES promotions(promotion_id),
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    order_id UUID REFERENCES orders(order_id),
    discount_amount DECIMAL(10,2) NOT NULL,
    PRIMARY KEY (user_id, promotion_id, order_id)
);

-- Customer support tickets
CREATE TABLE support_tickets (
    ticket_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    order_id UUID REFERENCES orders(order_id),
    subject VARCHAR(255) NOT NULL,
    description TEXT NOT NULL,
    status ENUM('open', 'in_progress', 'resolved', 'closed') NOT NULL DEFAULT 'open',
    priority ENUM('low', 'medium', 'high', 'urgent') NOT NULL DEFAULT 'medium',
    assigned_to UUID REFERENCES users(user_id),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP
);

-- Support ticket messages
CREATE TABLE ticket_messages (
    message_id UUID PRIMARY KEY,
    ticket_id UUID NOT NULL REFERENCES support_tickets(ticket_id),
    sender_id UUID NOT NULL REFERENCES users(user_id),
    message TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    attachment_urls JSONB
);

-- Notifications
CREATE TABLE notifications (
    notification_id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(user_id),
    title VARCHAR(255) NOT NULL,
    message TEXT NOT NULL,
    type ENUM('order_update', 'promotion', 'system', 'payment', 'support') NOT NULL,
    reference_id UUID, -- Could be order_id, promotion_id, etc.
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Driver assignments
CREATE TABLE driver_assignments (
    assignment_id UUID PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES users(user_id),
    order_id UUID NOT NULL REFERENCES orders(order_id),
    status ENUM('pending', 'accepted', 'rejected', 'completed') NOT NULL,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    completed_at TIMESTAMP,
    rejection_reason TEXT
);

-- Create indexes for performance optimization
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_driver_profiles_location ON driver_profiles USING GIST(current_location);
CREATE INDEX idx_restaurants_cuisine ON restaurants(cuisine_type);
CREATE INDEX idx_restaurants_location ON restaurants(address_id);
CREATE INDEX idx_menu_items_restaurant ON menu_items(restaurant_id);
CREATE INDEX idx_orders_customer ON orders(customer_id);
CREATE INDEX idx_orders_restaurant ON orders(restaurant_id);
CREATE INDEX idx_orders_driver ON orders(driver_id);
CREATE INDEX idx_orders_status ON orders(order_status);
CREATE INDEX idx_order_items_order ON order_items(order_id);
CREATE INDEX idx_payments_order ON payments(order_id);
CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);