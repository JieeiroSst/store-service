```
CREATE TABLE payers (
    payer_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone_number VARCHAR(20)
);

CREATE TABLE buyers (
    buyer_id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(100) UNIQUE,
    phone_number VARCHAR(20)
);

CREATE TABLE transactions (
    transaction_id SERIAL PRIMARY KEY,
    payer_id INTEGER REFERENCES payers(payer_id),
    buyer_id INTEGER REFERENCES buyers(buyer_id),
    amount DECIMAL(10,2) NOT NULL,
    transaction_date TIMESTAMP DEFAULT NOW(),
    -- Additional fields as needed, e.g., transaction_type, description, etc.
);
```