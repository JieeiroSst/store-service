-- 0001_init.sql : PostGIS + 3-tier administrative hierarchy.
-- Khóa chính = mã GSO (Tổng cục Thống kê): tỉnh 2 chữ số, huyện 3, xã 5.
-- Geometry (từ GADM) là cột tùy chọn, liên kết sau qua gid_*.

CREATE EXTENSION IF NOT EXISTS postgis;

-- ---------- Level 1: provinces (tỉnh / thành phố) ----------
CREATE TABLE IF NOT EXISTS provinces (
    code       TEXT PRIMARY KEY,              -- mã GSO, vd "01", "79"
    name       TEXT NOT NULL,                 -- "Hà Nội"
    name_full  TEXT NOT NULL DEFAULT '',      -- "Thành phố Hà Nội"
    name_en    TEXT NOT NULL DEFAULT '',
    type       TEXT NOT NULL DEFAULT '',      -- thanh-pho / tinh
    region     TEXT NOT NULL DEFAULT '',      -- Bắc / Trung / Nam
    slug       TEXT NOT NULL DEFAULT '',
    gid        TEXT,                          -- GADM GID_1 (sau khi khớp geometry)
    area_km2   DOUBLE PRECISION,
    geom       geometry(MultiPolygon, 4326)
);

-- ---------- Level 2: districts (quận / huyện / thị xã) ----------
CREATE TABLE IF NOT EXISTS districts (
    code          TEXT PRIMARY KEY,           -- mã GSO 3 chữ số, vd "001"
    province_code TEXT NOT NULL,              -- parent_code -> provinces.code
    name          TEXT NOT NULL,
    name_full     TEXT NOT NULL DEFAULT '',
    name_en       TEXT NOT NULL DEFAULT '',
    type          TEXT NOT NULL DEFAULT '',   -- quan / huyen / thi-xa / thanh-pho
    slug          TEXT NOT NULL DEFAULT '',
    gid           TEXT,                        -- GADM GID_2
    area_km2      DOUBLE PRECISION,
    geom          geometry(MultiPolygon, 4326),
    CONSTRAINT fk_district_province
        FOREIGN KEY (province_code) REFERENCES provinces(code) ON DELETE CASCADE
);

-- ---------- Level 3: wards (phường / xã / thị trấn) ----------
CREATE TABLE IF NOT EXISTS wards (
    code          TEXT PRIMARY KEY,           -- mã GSO 5 chữ số, vd "00001"
    district_code TEXT NOT NULL,              -- parent_code -> districts.code
    province_code TEXT NOT NULL,              -- denormalized cho join nhanh
    name          TEXT NOT NULL,
    name_full     TEXT NOT NULL DEFAULT '',
    name_en       TEXT NOT NULL DEFAULT '',
    type          TEXT NOT NULL DEFAULT '',   -- phuong / xa / thi-tran
    slug          TEXT NOT NULL DEFAULT '',
    gid           TEXT,                        -- GADM GID_3
    area_km2      DOUBLE PRECISION,
    geom          geometry(MultiPolygon, 4326),
    CONSTRAINT fk_ward_district
        FOREIGN KEY (district_code) REFERENCES districts(code) ON DELETE CASCADE
);

-- Spatial indexes (chỉ hiệu lực khi geom đã được nạp).
CREATE INDEX IF NOT EXISTS idx_provinces_geom ON provinces USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_districts_geom ON districts USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_wards_geom     ON wards     USING GIST (geom);

-- B-tree trên khóa cha cho truy vấn phân cấp.
CREATE INDEX IF NOT EXISTS idx_districts_province ON districts (province_code);
CREATE INDEX IF NOT EXISTS idx_wards_district     ON wards (district_code);
CREATE INDEX IF NOT EXISTS idx_wards_province     ON wards (province_code);
