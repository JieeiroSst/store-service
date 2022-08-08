-- Your SQL goes here
create table medias(
    id SERIAL PRIMARY KEY,
    name text not null,
    url text not null,
    description text not null,
    created_at     DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at     DATETIME DEFAULT CURRENT_TIMESTAMP
);