-- Your SQL goes here
create table cart_items (
    id SERIAL PRIMARY KEY,
    cart_id int not null,
    total int not null,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);