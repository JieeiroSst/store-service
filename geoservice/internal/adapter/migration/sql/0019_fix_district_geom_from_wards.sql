-- 0019_fix_district_geom_from_wards.sql
-- Recompute geometry for 14 districts where GADM boundary doesn't match
-- current ward boundaries (admin splits/merges post-GADM).
-- Strategy: union of all ward geometries that belong to each district.

UPDATE districts d
SET
  geom     = sub.geom,
  area_km2 = ST_Area(sub.geom::geography) / 1e6
FROM (
  SELECT
    district_code,
    ST_Multi(ST_MakeValid(ST_Union(geom))) AS geom
  FROM wards
  WHERE district_code IN (
    '045',  -- Hà Quảng, Cao Bằng
    '047',  -- Trùng Khánh, Cao Bằng
    '094',  -- Điện Biên Phủ, Điện Biên
    '133',  -- Nghĩa Lộ, Yên Bái
    '148',  -- Hòa Bình, Hoà Bình
    '156',  -- Mai Châu, Hoà Bình
    '193',  -- Hạ Long, Quảng Ninh
    '343',  -- Kiến Xương, Thái Bình
    '474',  -- Huế, Thừa Thiên Huế
    '525',  -- Trà Bồng, Quảng Ngãi
    '769',  -- Thủ Đức, Hồ Chí Minh
    '817',  -- Cai Lậy, Tiền Giang
    '866',  -- Cao Lãnh, Đồng Tháp
    '868'   -- Hồng Ngự, Đồng Tháp
  )
  AND geom IS NOT NULL
  GROUP BY district_code
) sub
WHERE d.code = sub.district_code;
