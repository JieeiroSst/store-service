CREATE TABLE IF NOT EXISTS `address` (
    address_id SERIAL PRIMARY KEY,
    line1 VARCHAR (50) NOT NULL, 
    line2 VARCHAR (50), 
    line1 VARCHAR (50) NOT NULL, 
    city VARCHAR (50) NOT NULL, 
    state VARCHAR (50) NOT NULL, 
    postal_code VARCHAR (50) NOT NULL, 
    country VARCHAR (50) NOT NULL
)

CREATE TABLE IF NOT EXISTS `customers` (
    customer_id SERIAL PRIMARY KEY,
    name VARCHAR (50) NOT NULL, 
    company VARCHAR (50) NOT NULL, 
    email VARCHAR (50) NOT NULL, 
    phone_number VARCHAR (50) NOT NULL, 
)

CREATE TABLE IF NOT EXISTS  `invoices` (
    invoice_id SERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL,
    invoice_date BIGINT NOT NULL,
    due_date BIGINT NOT NULL,
    amount DOUBLE PRECISION,
    tax VARCHAR (50) NOT NULL, 
    status VARCHAR (50) NOT NULL
)

CREATE TABLE IF NOT EXISTS `plans` (
    plan_id SERIAL PRIMARY KEY,
    name VARCHAR (50) NOT NULL,
    description VARCHAR (50) NOT NULL,
    price DOUBLE PRECISION,
    billing_cycle VARCHAR (50) NOT NULL
)

CREATE TABLE IF NOT EXISTS `subscriptions` (
    subscription_id SERIAL PRIMARY KEY,
    customer_id  BIGINT NOT NULL,
    plan_id BIGINT NOT NULL,
    start_date BIGINT NOT NULL,
    end_date BIGINT NOT NULL,
    status VARCHAR (50) NOT NULL
)

CREATE TABLE IF NOT EXISTS `` (
    transaction_id SERIAL PRIMARY KEY,
    invoice_id BIGINT NOT NULL,
    payment_method VARCHAR (50) NOT NULL,
    transaction_date BIGINT NOT NULL,
    amount DOUBLE PRECISION,
    status VARCHAR (50) NOT NULL
)