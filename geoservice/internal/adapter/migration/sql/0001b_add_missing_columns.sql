-- 0001b_add_missing_columns.sql
-- Guard migration: adds columns introduced in the updated 0001_init.sql schema
-- to databases that were created before those columns existed.
-- All statements are idempotent (ADD COLUMN IF NOT EXISTS).

ALTER TABLE provinces
    ADD COLUMN IF NOT EXISTS name_full  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS name_en    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS type       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS region     TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS slug       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gid        TEXT,
    ADD COLUMN IF NOT EXISTS area_km2   DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS geom       geometry(MultiPolygon, 4326);

ALTER TABLE districts
    ADD COLUMN IF NOT EXISTS name_full  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS name_en    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS type       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS slug       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gid        TEXT,
    ADD COLUMN IF NOT EXISTS area_km2   DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS geom       geometry(MultiPolygon, 4326);

ALTER TABLE wards
    ADD COLUMN IF NOT EXISTS name_full  TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS name_en    TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS type       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS slug       TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS gid        TEXT,
    ADD COLUMN IF NOT EXISTS area_km2   DOUBLE PRECISION,
    ADD COLUMN IF NOT EXISTS geom       geometry(MultiPolygon, 4326);

CREATE INDEX IF NOT EXISTS idx_provinces_geom ON provinces USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_districts_geom ON districts USING GIST (geom);
CREATE INDEX IF NOT EXISTS idx_wards_geom     ON wards     USING GIST (geom);

CREATE INDEX IF NOT EXISTS idx_districts_province ON districts (province_code);
CREATE INDEX IF NOT EXISTS idx_wards_district     ON wards (district_code);
CREATE INDEX IF NOT EXISTS idx_wards_province     ON wards (province_code);
