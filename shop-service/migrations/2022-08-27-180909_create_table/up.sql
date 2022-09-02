-- Your SQL goes here
create table products (
    id SERIAL PRIMARY KEY,
    product_name TEXT NOT NULL,
    description TEXT NOT NULL,
    price INT NOT NULL,
    media_id INT NOT NULL,
    destroy boolean default true,
    created_at     TIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIME DEFAULT CURRENT_TIMESTAMP
);

create table carts (
    id SERIAL PRIMARY KEY,
    total INT NOT NULL,
    user_id INT NOT NULL,
    destroy boolean default true,
    created_at     TIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIME DEFAULT CURRENT_TIMESTAMP
);

create table cart_items (
    id SERIAL PRIMARY KEY,
    cart_id int not null,
    total int not null,
    amount int not null,
    destroy boolean default true,
    created_at     TIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIME DEFAULT CURRENT_TIMESTAMP
);

create table medias(
    id SERIAL PRIMARY KEY,
    name text not null,
    url text not null,
    description text not null,
    destroy boolean default true,
    created_at     TIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIME DEFAULT CURRENT_TIMESTAMP
);