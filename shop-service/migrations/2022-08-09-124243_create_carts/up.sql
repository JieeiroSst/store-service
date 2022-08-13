-- Your SQL goes here
create table carts (
    id SERIAL PRIMARY KEY,
    total INT NOT NULL,
    user_id INT NOT NULL,
    destroy boolean default true,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);