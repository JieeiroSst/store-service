CREATE TABLE IF NOT EXISTS wishlists_wishlist (
    id           BIGSERIAL    PRIMARY KEY,
    name         VARCHAR(255) NOT NULL,
    "key"        VARCHAR(64)  NOT NULL,
    visibility   VARCHAR(20)  NOT NULL DEFAULT 'Private',
    date_created TIMESTAMPTZ  NOT NULL DEFAULT now(),
    owner_id     VARCHAR(64)  NOT NULL,
    CONSTRAINT wishlists_wishlist_key_key UNIQUE ("key"),
    CONSTRAINT wishlists_wishlist_visibility_check CHECK (visibility IN ('Private', 'Shared', 'Public'))
);

CREATE INDEX IF NOT EXISTS wishlists_wishlist_owner_id_idx ON wishlists_wishlist (owner_id, id);

CREATE TABLE IF NOT EXISTS wishlists_line (
    id          BIGSERIAL    PRIMARY KEY,
    wishlist_id BIGINT       NOT NULL REFERENCES wishlists_wishlist (id) ON DELETE CASCADE,
    product_id  BIGINT       NOT NULL,
    title       VARCHAR(255) NOT NULL,
    quantity    INTEGER      NOT NULL DEFAULT 1,
    CONSTRAINT wishlists_line_wishlist_product_key UNIQUE (wishlist_id, product_id),
    CONSTRAINT wishlists_line_quantity_check CHECK (quantity >= 1)
);

CREATE INDEX IF NOT EXISTS wishlists_line_product_id_idx ON wishlists_line (product_id);
