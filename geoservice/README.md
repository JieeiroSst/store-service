# geoservice — Vietnam geospatial indexing (3 cấp hành chính + đường phố)

Service Golang theo **hexagonal architecture** (uber-go/fx), lưu dữ liệu trong
**PostgreSQL + PostGIS**. Cho tọa độ `(lng, lat)`, trả về **chuỗi hành chính đầy
đủ**: tỉnh/thành → quận/huyện → phường/xã → **đường phố gần nhất**.

Mốc dữ liệu: **63 tỉnh (cơ cấu 2008)**, **705 quận/huyện**, **10.599 phường/xã** —
seed sẵn từ Tổng cục Thống kê. **Đường phố** seed từ OpenStreetMap (Overpass API).

## Hai nguồn dữ liệu, hai vai trò

| Nguồn | Vai trò | Khóa |
| --- | --- | --- |
| **GSO** (Tổng cục Thống kê) | metadata phân cấp (tên, loại, vùng, quan hệ cha-con) — **seed sẵn** | mã GSO (tỉnh 2 số, huyện 3, xã 5) |
| **GADM** | ranh giới địa lý (polygon) — tải & gắn sau | GID_1/2/3 |
| **OpenStreetMap** | tên và đường tâm (MultiLineString) đường phố — tải qua script | OSM way ID |

## Phân cấp

| Cấp | Bảng | Số đơn vị | Ví dụ |
| --- | --- | --- | --- |
| 1 — tỉnh/thành | `provinces` | 63 | `79` Hồ Chí Minh |
| 2 — quận/huyện | `districts` | 705 | `760` Quận 1 |
| 3 — phường/xã | `wards` | 10.599 | `26740` Phường Bến Nghé |
| — đường phố | `streets` | ~tuỳ tỉnh | `Nguyễn Huệ`, `Lê Lợi` |

> **Lưu ý:** `streets` không phải cấp hành chính; dùng `province_code` làm khoá
> cha, tham chiếu `districts` tuỳ chọn. Geometry là `MultiLineString` (tuyến tính).

## Kiến trúc (hexagon / ports & adapters)

```
cmd/server         -> API server (fx)
cmd/importgeo      -> CLI gắn geometry GADM vào dòng đã seed
config             -> viper, env GEO_*
internal/
  core/
    domain         -> Province, District, Ward, Street, Point, LocationResult, GeoFeature
    port           -> GeoRepository (driven), LocatorService (driving)
    service        -> validate bbox VN, locate chain, fallback nearest
  adapter/
    handler        -> HTTP (echo)
    repository     -> PostGIS qua pgx (point-in-polygon 3 cấp + nearest street)
    importer       -> parse GADM GeoJSON, gắn polygon theo tên
    migration      -> chạy SQL seed (embed)
  app              -> fx wiring
scripts/
  gen_seed.py            -> sinh SQL seed từ JSON GSO (tỉnh/huyện/xã)
  regenerate_seed.sh     -> tải GSO + chạy gen_seed.py
  download_gadm.sh       -> tải ranh giới GADM levels 1-3
  fetch_hcm_streets.py   -> sinh 0015_seed_hcm_streets.sql (chỉ TP.HCM)
  fetch_all_streets.py   -> sinh SQL đường phố cho toàn bộ 63 tỉnh/thành
  sample_VNM_*.json      -> mẫu GADM 3 cấp (HCMC) để test
```

## Migrations

| File | Nội dung |
| --- | --- |
| `0001_init.sql` | Schema provinces / districts / wards (PostGIS) |
| `0002–0004` | Seed 63 tỉnh / 705 huyện / 10.599 xã (GSO) |
| `0005–0012` | Gắn geometry GADM L1→L2→L3; vá các trường hợp đặc biệt |
| `0013` | Vá geometry Phường Võ Thị Sáu (Quận 3, TP.HCM) |
| `0014` | Schema bảng `streets` |
| `0015` | Seed đường phố TP.HCM (~19.000 tên, từ OSM) |
| `0016` | `UNIQUE(province_code, name)` trên bảng `streets` |
| `0017` | *(tuỳ chọn)* Seed đường phố 62 tỉnh còn lại — sinh bằng script |

## Chạy nhanh

```bash
make up        # PostGIS qua docker
make run       # server :8081; tự migrate + seed khi khởi động
curl localhost:8081/api/v1/stats
```

Metadata tỉnh/huyện/xã sẵn sàng ngay sau `make run`. Chỉ cần import GADM nếu
muốn `locate` theo tọa độ.

## Gắn ranh giới địa lý (GADM)

```bash
make download   # -> data/gadm41_VNM_{1,2,3}.json
make import     # gắn polygon L1->L2->L3 vào dòng đã seed (khớp theo tên)
```

Test nhanh với mẫu HCMC (không cần tải GADM):
```bash
make import-sample
curl "localhost:8081/api/v1/locate?lng=106.705&lat=10.780"
```

## Seed đường phố từ OpenStreetMap

Dữ liệu đường phố TP.HCM được bao gồm sẵn trong migration `0015`. Để seed thêm
các tỉnh còn lại, dùng script `fetch_all_streets.py` (cần kết nối internet).

### Fetch 62 tỉnh còn lại (bỏ HCM đã có)

```bash
python3 scripts/fetch_all_streets.py --skip 79 \
  > internal/adapter/migration/sql/0017_seed_other_streets.sql
```

> Mất khoảng 20–40 phút (63 request × 10 s delay giữa các request).

### Chỉ fetch một số tỉnh

```bash
python3 scripts/fetch_all_streets.py --only 01,31,48
# 01 = Hà Nội, 31 = Hải Phòng, 48 = Đà Nẵng
```

### Nối tiếp nếu bị ngắt giữa chừng

```bash
# Append vào file hiện tại, bỏ qua các tỉnh đã xong
python3 scripts/fetch_all_streets.py --skip 01,02,04,06 \
  >> internal/adapter/migration/sql/0017_seed_other_streets.sql
```

### Tái sinh dữ liệu TP.HCM

```bash
python3 scripts/fetch_hcm_streets.py \
  > internal/adapter/migration/sql/0015_seed_hcm_streets.sql
```

## Cập nhật danh sách đơn vị hành chính

```bash
make regenerate-seed   # tải lại JSON GSO, sinh lại 0002/0003/0004_seed_*.sql
```

## API

### Xác định vị trí (chuỗi đầy đủ)

```
GET /api/v1/locate?lng=106.705&lat=10.780&tolerance=500
```

```json
{
  "found": true,
  "province": { "code": "79", "name": "Hồ Chí Minh", "region": "Nam",
                "gid": "VNM.25_1", "area_km2": 2109.03 },
  "district": { "code": "760", "name": "1", "name_full": "Quận 1",
                "type": "quan", "gid": "VNM.25.13_1", "area_km2": 7.72 },
  "ward":     { "code": "26740", "name": "Bến Nghé",
                "name_full": "Phường Bến Nghé", "gid": "VNM.25.13.2_1" },
  "street":   { "name": "Nguyễn Huệ", "name_full": "Đường Nguyễn Huệ",
                "gid": "32582320", "length_m": 680.5, "distance_m": 42.1 },
  "point": { "lng": 106.705, "lat": 10.78 }
}
```

**Hành vi:**
- `district` / `ward`: dùng `ST_Contains` trước; nếu điểm rơi vào khoảng trống
  giữa các polygon GADM (gap < 500 m), tự động fallback sang polygon gần nhất.
- `street`: đường phố gần nhất trong bán kính 100 m (chỉ khi đã seed bảng `streets`).
- `tolerance` (m): điểm ngoài mọi polygon → tỉnh/huyện/xã gần nhất.
- `200` thấy · `404` ngoài tỉnh · `422` ngoài lãnh thổ VN.

### Duyệt cây hành chính

```
GET /api/v1/provinces                      # 63 tỉnh
GET /api/v1/provinces/:code                # 1 tỉnh
GET /api/v1/provinces/:code/districts      # quận/huyện thuộc tỉnh
GET /api/v1/districts/:code/wards          # phường/xã thuộc quận/huyện
GET /api/v1/provinces/:code/streets        # đường phố thuộc tỉnh
GET /api/v1/stats                          # đếm số đơn vị mỗi cấp
GET /healthz
```

> `GET /api/v1/provinces/79/streets` trả danh sách ~19.000 đường của TP.HCM
> (sắp xếp theo tên). Nên phân trang ở tầng ứng dụng khi dùng thực tế.

### Kiểm tra whitelist thành phố ưu tiên

Các thành phố được ưu tiên hiện tại: **Hồ Chí Minh**, **Hà Nội**, **Đà Nẵng**.

#### Kiểm tra tọa độ có nằm trong whitelist

```bash
# Tọa độ trong TP.HCM → whitelisted: true
curl "localhost:8081/api/v1/whitelist/check?lng=106.705&lat=10.780"
```

```json
{
  "whitelisted": true,
  "province": { "code": "79", "name": "Hồ Chí Minh", "type": "thanh-pho",
                "region": "Nam", "gid": "VNM.25_1", "area_km2": 2109.03 },
  "district": { "code": "760", "name": "1", "name_full": "Quận 1",
                "type": "quan", "province_code": "79" },
  "ward":     { "code": "26740", "name": "Bến Nghé",
                "name_full": "Phường Bến Nghé", "type": "phuong" },
  "street":   { "name": "Nguyễn Huệ", "name_full": "Đường Nguyễn Huệ",
                "length_m": 680.5, "distance_m": 42.1 },
  "point": { "lng": 106.705, "lat": 10.78 }
}
```

```bash
# Tọa độ ngoài whitelist (Cần Thơ) → whitelisted: false
curl "localhost:8081/api/v1/whitelist/check?lng=105.788&lat=10.045"
```

```json
{
  "whitelisted": false,
  "province": { "code": "92", "name": "Cần Thơ", "type": "thanh-pho",
                "region": "Nam" },
  "district": { "code": "916", "name": "Ninh Kiều", "type": "quan" },
  "ward":     { "code": "31150", "name": "An Hòa", "type": "phuong" },
  "point": { "lng": 105.788, "lat": 10.045 }
}
```

**Query params:**
| Param | Bắt buộc | Mô tả |
| --- | --- | --- |
| `lng` | ✓ | Kinh độ (WGS84) |
| `lat` | ✓ | Vĩ độ (WGS84) |
| `tolerance` | — | Bán kính fallback (mét, mặc định 1000) — dùng khi điểm rơi ngoài polygon nhưng gần biên tỉnh |

**HTTP status:** `200` luôn trả về (kể cả `whitelisted: false`) · `422` ngoài lãnh thổ VN · `400` thiếu/sai tham số.

#### Danh sách thành phố trong whitelist

```bash
curl "localhost:8081/api/v1/whitelist/cities"
```

```json
[
  { "province_code": "01", "note": "Thành phố Hà Nội" },
  { "province_code": "48", "note": "Thành phố Đà Nẵng" },
  { "province_code": "79", "note": "Thành phố Hồ Chí Minh" }
]
```

## Cấu hình (env)

| Biến | Mặc định |
| --- | --- |
| `GEO_DATABASE_URL` | `postgres://geo:geo@localhost:5432/geo?sslmode=disable` |
| `GEO_HTTP_PORT` | `8081` |
| `GEO_LOG_LEVEL` | `info` (`debug` log chi tiết) |

## Ghi chú dữ liệu

- **GSO:** cơ cấu 63 tỉnh (mốc 2008–2024). Nếu cần cơ cấu 34 tỉnh (2025), thay
  nguồn JSON và chạy `make regenerate-seed`; schema không đổi.
- **GADM:** dùng cho mục đích học thuật — kiểm tra điều khoản nếu thương mại.
  Thay thế: geoBoundaries (ODbL), OpenStreetMap Nominatim.
- **Tên GADM** đôi khi lệch nhẹ so GSO (biến thể dấu); vài đơn vị cần gắn thủ
  công (xem log importer).
- **OpenStreetMap:** dữ liệu đường phố theo giấy phép ODbL. Script chỉ lấy các
  loại đường `trunk / primary / secondary / tertiary / residential / unclassified`
  có tag `name`. Số lượng đường mỗi tỉnh phụ thuộc mức độ đóng góp OSM tại đó.

```

https://gadm.org/maps/VNM.html

https://gps-coordinates.org
```
