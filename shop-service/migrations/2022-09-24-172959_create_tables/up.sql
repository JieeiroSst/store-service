-- Your SQL goes here
create table products (
    id              TEXT PRIMARY KEY,
    product_name    TEXT NOT NULL,
    description     TEXT NOT NULL,
    price           INT NOT NULL,
    media_id        TEXT NOT NULL,
    destroy         BOOLEAN DEFAULT TRUE,
    created_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table carts (
    id              TEXT PRIMARY KEY,
    total           INT NOT NULL,
    user_id         TEXT NOT NULL,
    destroy         BOOLEAN DEFAULT TRUE,
    created_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table cart_items (
    id              TEXT PRIMARY KEY,
    cart_id         TEXT NOT NULL,
    total           INT NOT NULL,
    amount          INT NOT NULL,
    destroy         BOOLEAN DEFAULT TRUE,
    created_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

create table medias(
    id              TEXT PRIMARY KEY,
    name            TEXT NOT NULL,
    url             TEXT NOT NULL,
    description     TEXT NOT NULL,
    destroy         BOOLEAN DEFAULT TRUE,
    created_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at      Timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);