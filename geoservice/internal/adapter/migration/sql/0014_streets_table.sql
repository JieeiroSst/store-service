-- 0014_streets_table.sql : bảng đường phố (streets) — geometry tuyến tính.
-- gid  : ID ngoài (OpenStreetMap way ID hoặc mã khác).
-- geom : MultiLineString (đường là đối tượng tuyến, không phải vùng).
-- Seed TP.HCM được nạp riêng qua script scripts/fetch_hcm_streets.py.

CREATE TABLE IF NOT EXISTS streets (
    id            SERIAL PRIMARY KEY,
    province_code TEXT NOT NULL,
    district_code TEXT,                       -- NULL nếu đường trải dài qua nhiều quận
    name          TEXT NOT NULL,
    name_full     TEXT NOT NULL DEFAULT '',   -- "Đường Nguyễn Huệ"
    name_en       TEXT NOT NULL DEFAULT '',
    slug          TEXT NOT NULL DEFAULT '',
    gid           TEXT,                       -- OSM way ID hoặc ID ngoài
    length_m      DOUBLE PRECISION,           -- độ dài tính bằng mét
    geom          geometry(MultiLineString, 4326),
    CONSTRAINT fk_street_province
        FOREIGN KEY (province_code) REFERENCES provinces(code) ON DELETE CASCADE,
    CONSTRAINT fk_street_district
        FOREIGN KEY (district_code) REFERENCES districts(code) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_streets_geom     ON streets USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_streets_province ON streets (province_code);
CREATE INDEX IF NOT EXISTS idx_streets_district ON streets (district_code);
CREATE INDEX IF NOT EXISTS idx_streets_name     ON streets (name);
CREATE INDEX IF NOT EXISTS idx_streets_slug     ON streets (slug);
