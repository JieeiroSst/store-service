CREATE TABLE IF NOT EXISTS users (
    user_id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    phone_number VARCHAR(20),
    address TEXT,
    user_type VARCHAR(30) NOT NULL CHECK (user_type IN ('customer', 'staff', 'admin')),
    driving_license VARCHAR(50),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);


CREATE TABLE IF NOT EXISTS vehicle_categories (
    category_id UUID PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    description TEXT,
    vehicle_type VARCHAR(30) NOT NULL CHECK (vehicle_type IN ('car', 'motorcycle', 'bicycle'))
);

CREATE TABLE IF NOT EXISTS locations (
    location_id UUID PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address TEXT NOT NULL,
    city VARCHAR(100) NOT NULL,
    state VARCHAR(100),
    country VARCHAR(100) NOT NULL,
    zip_code VARCHAR(20),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    contact_phone VARCHAR(20),
    opening_hours JSONB
);

CREATE TABLE IF NOT EXISTS vehicles (
    vehicle_id UUID PRIMARY KEY,
    category_id UUID REFERENCES vehicle_categories(category_id),
    registration_number VARCHAR(20) UNIQUE,
    make VARCHAR(50) NOT NULL,
    model VARCHAR(50) NOT NULL,
    year INT NOT NULL,
    color VARCHAR(30),
    mileage FLOAT,
    status VARCHAR(30) NOT NULL CHECK (status IN ('available', 'rented', 'maintenance', 'retired')),
    hourly_rate DECIMAL(10, 2) NOT NULL,
    daily_rate DECIMAL(10, 2) NOT NULL,
    location_id UUID REFERENCES locations(location_id),
    features JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reservations (
    reservation_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    vehicle_id UUID REFERENCES vehicles(vehicle_id),
    pickup_location_id UUID REFERENCES locations(location_id),
    return_location_id UUID REFERENCES locations(location_id),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP NOT NULL,
    status VARCHAR(30) NOT NULL CHECK (status IN ('pending', 'confirmed', 'cancelled', 'completed')),
    cancellation_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS rentals (
    rental_id UUID PRIMARY KEY,
    reservation_id UUID REFERENCES reservations(reservation_id),
    vehicle_id UUID REFERENCES vehicles(vehicle_id),
    user_id UUID REFERENCES users(user_id),
    pickup_time TIMESTAMP NOT NULL,
    actual_return_time TIMESTAMP,
    pickup_location_id UUID REFERENCES locations(location_id),
    return_location_id UUID REFERENCES locations(location_id),
    pickup_mileage FLOAT,
    return_mileage FLOAT,
    status VARCHAR(30) NOT NULL CHECK (status IN ('active', 'completed', 'overdue')),
    base_fee DECIMAL(10, 2) NOT NULL,
    additional_fees DECIMAL(10, 2) DEFAULT 0,
    notes TEXT,
    payment_status VARCHAR(30) NOT NULL CHECK (payment_status IN ('pending', 'completed', 'failed', 'refunded')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS maintenance_records (
    record_id UUID PRIMARY KEY,
    vehicle_id UUID REFERENCES vehicles(vehicle_id),
    maintenance_type VARCHAR(100) NOT NULL,
    description TEXT,
    cost DECIMAL(10, 2),
    performed_by VARCHAR(100),
    maintenance_date DATE NOT NULL,
    next_maintenance_date DATE,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS payments (
    payment_id UUID PRIMARY KEY,
    rental_id UUID REFERENCES rentals(rental_id),
    user_id UUID REFERENCES users(user_id),
    amount DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(30) NOT NULL CHECK (payment_method IN ('credit_card', 'debit_card', 'cash', 'online')),
    transaction_id VARCHAR(100),
    payment_status VARCHAR(30) NOT NULL CHECK (payment_status IN ('pending', 'completed', 'failed', 'refunded')),
    payment_date TIMESTAMP NOT NULL,
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS reviews (
    review_id UUID PRIMARY KEY,
    rental_id UUID REFERENCES rentals(rental_id),
    user_id UUID REFERENCES users(user_id),
    vehicle_id UUID REFERENCES vehicles(vehicle_id),
    rating INT CHECK (rating BETWEEN 1 AND 5) NOT NULL,
    comment TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_documents (
    document_id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(user_id),
    document_type VARCHAR(30) NOT NULL CHECK (document_type IN ('driving_license', 'id_proof', 'passport', 'other')),
    document_number VARCHAR(100) NOT NULL,
    expiry_date DATE,
    document_url VARCHAR(255),
    verification_status VARCHAR(30) NOT NULL CHECK (verification_status IN ('pending', 'verified', 'rejected')),
    verified_by UUID REFERENCES users(user_id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS car_details (
    car_detail_id UUID PRIMARY KEY,
    vehicle_id UUID REFERENCES vehicles(vehicle_id) UNIQUE,
    fuel_type VARCHAR(30) NOT NULL CHECK (fuel_type IN ('petrol', 'diesel', 'electric', 'hybrid')),
    transmission VARCHAR(30) NOT NULL CHECK (transmission IN ('manual', 'automatic')),
    seating_capacity INT NOT NULL,
    trunk_capacity FLOAT,
    air_conditioning BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS motorcycle_details (
    motorcycle_detail_id UUID PRIMARY KEY,
    vehicle_id UUID REFERENCES vehicles(vehicle_id) UNIQUE,
    engine_capacity INT NOT NULL,
    motorcycle_type VARCHAR(30) NOT NULL CHECK (motorcycle_type IN ('cruiser', 'sport', 'touring', 'standard', 'off_road')),
    helmet_included BOOLEAN DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS bicycle_details (
    bicycle_detail_id UUID PRIMARY KEY,
    vehicle_id UUID REFERENCES vehicles(vehicle_id) UNIQUE,
    bicycle_type VARCHAR(30) NOT NULL CHECK (bicycle_type IN ('mountain', 'road', 'hybrid', 'city', 'electric')),
    frame_size VARCHAR(20) NOT NULL,
    gear_count INT,
    has_basket BOOLEAN DEFAULT FALSE
);