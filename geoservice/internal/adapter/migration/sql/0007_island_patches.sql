-- 0007_island_patches.sql
-- Thêm polygon các đảo xa bờ bị GADM L1 bỏ sót vào geometry tỉnh tương ứng.
-- Tọa độ xấp xỉ thực địa:
--   Cồn Cỏ (~2.3 km²)  → Quảng Trị  (code 45)
--   Hòn Khoai (~4.3km²) → Cà Mau     (code 96)
--   Nam Du (~13 km²)    → Kiên Giang  (code 91)

UPDATE provinces
SET
  geom = ST_Multi(ST_MakeValid(ST_Union(
    geom,
    ST_SetSRID(ST_GeomFromGeoJSON(
      '{"type":"Polygon","coordinates":[[[107.342,17.552],[107.352,17.558],[107.362,17.553],[107.365,17.542],[107.358,17.533],[107.346,17.531],[107.338,17.540],[107.342,17.552]]]}'
    ), 4326)
  ))),
  area_km2 = area_km2 + 2.3
WHERE code = '45';

UPDATE provinces
SET
  geom = ST_Multi(ST_MakeValid(ST_Union(
    geom,
    ST_SetSRID(ST_GeomFromGeoJSON(
      '{"type":"Polygon","coordinates":[[[104.822,8.475],[104.829,8.482],[104.839,8.481],[104.845,8.472],[104.842,8.461],[104.832,8.457],[104.822,8.463],[104.822,8.475]]]}'
    ), 4326)
  ))),
  area_km2 = area_km2 + 4.3
WHERE code = '96';

UPDATE provinces
SET
  geom = ST_Multi(ST_MakeValid(ST_Union(
    geom,
    ST_SetSRID(ST_GeomFromGeoJSON(
      '{"type":"MultiPolygon","coordinates":[[[[103.858,10.025],[103.865,10.032],[103.876,10.031],[103.882,10.023],[103.879,10.014],[103.869,10.010],[103.859,10.015],[103.858,10.025]]],[[[103.855,9.998],[103.861,10.005],[103.870,10.002],[103.872,9.993],[103.864,9.988],[103.855,9.993],[103.855,9.998]]]]}'
    ), 4326)
  ))),
  area_km2 = area_km2 + 13.0
WHERE code = '91';
