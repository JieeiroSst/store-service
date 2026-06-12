#!/usr/bin/env python3
"""Tải đường phố TP.HCM từ OpenStreetMap (Overpass API) và xuất SQL seed.

Usage:
    python3 scripts/fetch_hcm_streets.py > internal/adapter/migration/sql/0015_seed_hcm_streets.sql

Yêu cầu: Python 3.8+, không cần thư viện ngoài (dùng urllib).

Cách hoạt động:
  1. Truy vấn Overpass API để lấy tất cả way có tag "highway" + "name" trong
     bounding box TP.HCM.
  2. Nhóm các đoạn đường cùng tên thành một MultiLineString.
  3. Xuất INSERT SQL vào bảng streets (migration 0014).

Kết quả: migration 0015_seed_hcm_streets.sql (chạy sau 0014).
"""

import json
import re
import sys
import time
import unicodedata
import urllib.error
import urllib.parse
import urllib.request

# ---------------------------------------------------------------------------
# Cấu hình
# ---------------------------------------------------------------------------
OVERPASS_URL = "https://overpass-api.de/api/interpreter"

# Bounding box TP.HCM (south,west,north,east)
HCM_BBOX = "10.34,106.32,11.16,107.02"

PROVINCE_CODE = "79"

# Loại đường cần lấy (bỏ footway, path… để giữ đường chính)
HIGHWAY_FILTER = (
    "trunk|trunk_link|primary|primary_link|secondary|secondary_link"
    "|tertiary|tertiary_link|residential|living_street|unclassified"
)

OVERPASS_QUERY = f"""[out:json][timeout:120];
(
  way["highway"~"^({HIGHWAY_FILTER})$"]["name"]({HCM_BBOX});
);
out geom;
"""

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def slugify(name: str) -> str:
    name = name.lower().replace("đ", "d")
    name = unicodedata.normalize("NFD", name)
    name = "".join(c for c in name if unicodedata.category(c) != "Mn")
    name = re.sub(r"[^a-z0-9]+", "-", name)
    return name.strip("-")


def esc(s: str) -> str:
    return (s or "").replace("'", "''")


def fetch_overpass(query: str, retries: int = 3) -> dict:
    payload = urllib.parse.urlencode({"data": query}).encode()
    for attempt in range(retries):
        try:
            req = urllib.request.Request(
                OVERPASS_URL,
                data=payload,
                headers={
                    "Content-Type": "application/x-www-form-urlencoded",
                    "User-Agent": "geoservice-hcm-streets/1.0 (Vietnam administrative geo data)",
                },
            )
            with urllib.request.urlopen(req, timeout=150) as resp:
                return json.load(resp)
        except urllib.error.HTTPError as e:
            if e.code == 429 and attempt < retries - 1:
                wait = 30 * (attempt + 1)
                print(f"-- Rate limited, chờ {wait}s...", file=sys.stderr)
                time.sleep(wait)
                continue
            raise
    raise RuntimeError("Overpass API không phản hồi sau nhiều lần thử.")


def build_district_index() -> dict[str, str]:
    """Trả về dict {slug_name: district_code} cho TP.HCM (province 79)."""
    return {}  # Để trống; có thể bổ sung sau nếu cần join theo quận


# ---------------------------------------------------------------------------
# Xử lý dữ liệu
# ---------------------------------------------------------------------------

def process_elements(elements: list) -> dict:
    """Nhóm các way theo tên → MultiLineString."""
    streets: dict[str, dict] = {}
    for el in elements:
        if el.get("type") != "way":
            continue
        tags = el.get("tags", {})
        name = tags.get("name", "").strip()
        if not name:
            continue
        geometry = el.get("geometry", [])
        coords = [[pt["lon"], pt["lat"]] for pt in geometry if "lat" in pt and "lon" in pt]
        if len(coords) < 2:
            continue

        osm_id = str(el["id"])
        name_en = tags.get("name:en", "")

        if name not in streets:
            streets[name] = {
                "name": name,
                "name_en": name_en,
                "gid": osm_id,      # OSM ID của đoạn đầu tiên
                "segments": [coords],
            }
        else:
            streets[name]["segments"].append(coords)

    return streets


def compute_length_m(segments: list) -> float:
    """Tính tổng chiều dài (mét) theo Haversine."""
    import math
    total = 0.0
    for seg in segments:
        for i in range(len(seg) - 1):
            lon1, lat1 = seg[i]
            lon2, lat2 = seg[i + 1]
            dlat = math.radians(lat2 - lat1)
            dlon = math.radians(lon2 - lon1)
            a = (math.sin(dlat / 2) ** 2
                 + math.cos(math.radians(lat1)) * math.cos(math.radians(lat2))
                 * math.sin(dlon / 2) ** 2)
            total += 6371000 * 2 * math.asin(math.sqrt(a))
    return total


# ---------------------------------------------------------------------------
# Xuất SQL
# ---------------------------------------------------------------------------

def render_sql(streets: dict) -> str:
    lines = [
        "-- 0015_seed_hcm_streets.sql : đường phố TP.HCM từ OpenStreetMap.",
        "-- Auto-generated bởi scripts/fetch_hcm_streets.py — không sửa tay.",
        "",
        "INSERT INTO streets",
        "  (province_code, name, name_full, name_en, slug, gid, length_m, geom)",
        "VALUES",
    ]

    rows = []
    for name, s in sorted(streets.items(), key=lambda x: x[0].lower()):
        slug = slugify(name)
        name_full = "Đường " + name
        name_en = s["name_en"] or ""
        gid = s["gid"]
        length_m = round(compute_length_m(s["segments"]), 2)
        geojson = json.dumps(
            {"type": "MultiLineString", "coordinates": s["segments"]},
            separators=(",", ":"),
            ensure_ascii=False,
        )
        rows.append(
            "  ('{prov}','{name}','{name_full}','{name_en}','{slug}','{gid}',{length},"
            "ST_Multi(ST_MakeValid(ST_SetSRID(ST_GeomFromGeoJSON('{geom}'),4326))))".format(
                prov=PROVINCE_CODE,
                name=esc(name),
                name_full=esc(name_full),
                name_en=esc(name_en),
                slug=esc(slug),
                gid=esc(gid),
                length=length_m,
                geom=esc(geojson),
            )
        )

    lines.append(",\n".join(rows))
    lines.append("ON CONFLICT DO NOTHING;")
    return "\n".join(lines) + "\n"


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    print("-- Đang truy vấn Overpass API...", file=sys.stderr)
    result = fetch_overpass(OVERPASS_QUERY)
    elements = result.get("elements", [])
    print(f"-- Nhận được {len(elements)} way từ OSM.", file=sys.stderr)

    streets = process_elements(elements)
    print(f"-- Nhóm thành {len(streets)} tên đường.", file=sys.stderr)

    sql = render_sql(streets)
    print(sql)
    print(f"-- Hoàn thành. Lưu vào 0015_seed_hcm_streets.sql", file=sys.stderr)


if __name__ == "__main__":
    main()
