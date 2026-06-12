-- Whitelist of priority cities (tỉnh/thành phố được ưu tiên).
CREATE TABLE IF NOT EXISTS whitelist_cities (
    province_code CHAR(2)     PRIMARY KEY REFERENCES provinces(code),
    note          TEXT,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed: Hồ Chí Minh (79), Hà Nội (01), Đà Nẵng (48)
INSERT INTO whitelist_cities (province_code, note) VALUES
    ('79', 'Thành phố Hồ Chí Minh'),
    ('01', 'Thành phố Hà Nội'),
    ('48', 'Thành phố Đà Nẵng')
ON CONFLICT DO NOTHING;
