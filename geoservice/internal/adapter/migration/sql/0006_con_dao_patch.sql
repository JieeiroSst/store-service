-- 0006_con_dao_patch.sql
-- Thêm polygon quần đảo Côn Đảo vào geometry tỉnh Bà Rịa - Vũng Tàu (code 77).
-- GADM L1 bỏ sót đảo xa bờ này; patch mở rộng geom bằng ST_Union.
-- Tọa độ xấp xỉ thực địa: đảo Côn Sơn + Hòn Bảy Cạnh + Hòn Cau.

UPDATE provinces
SET
  geom = ST_Multi(ST_MakeValid(ST_Union(
    geom,
    ST_SetSRID(
      ST_GeomFromGeoJSON('{"type":"MultiPolygon","coordinates":[[[[106.537,8.752],[106.549,8.769],[106.563,8.773],[106.58,8.763],[106.607,8.745],[106.631,8.718],[106.643,8.688],[106.638,8.657],[106.626,8.635],[106.605,8.624],[106.582,8.619],[106.557,8.63],[106.535,8.658],[106.524,8.694],[106.529,8.729],[106.537,8.752]]],[[[106.674,8.703],[106.68,8.71],[106.692,8.707],[106.696,8.698],[106.69,8.689],[106.678,8.69],[106.674,8.703]]],[[[106.612,8.565],[106.619,8.57],[106.626,8.565],[106.622,8.557],[106.613,8.557],[106.612,8.565]]]]}'
      ),
      4326
    )
  ))),
  area_km2 = area_km2 + 59.0   -- Côn Sơn ~51.5km² + Hòn Bảy Cạnh ~5km² + Hòn Cau ~2.5km²
WHERE code = '77';
