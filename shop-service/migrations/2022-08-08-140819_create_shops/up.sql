-- Your SQL goes here
create table products (
    id SERIAL PRIMARY KEY,
    product_name TEXT NOT NULL,
    description TEXT NOT NULL,
    price INT NOT NULL,
    media_id INT NOT NULL,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);