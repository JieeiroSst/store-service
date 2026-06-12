#!/usr/bin/env python3
"""Tải đường phố toàn bộ 63 tỉnh/thành Việt Nam từ OpenStreetMap (Overpass API).

Usage:
    # Tất cả 63 tỉnh (chạy ~20-40 phút do rate-limit):
    python3 scripts/fetch_all_streets.py \
        > internal/adapter/migration/sql/0015_seed_all_streets.sql

    # Chỉ một số tỉnh (GSO code cách nhau bởi dấu phẩy):
    python3 scripts/fetch_all_streets.py --only 01,79,31 \
        > internal/adapter/migration/sql/0015_seed_all_streets.sql

    # Bỏ qua các tỉnh đã có (nối tiếp session bị gián đoạn):
    python3 scripts/fetch_all_streets.py --skip 01,79 \
        >> internal/adapter/migration/sql/0015_seed_all_streets.sql

Yêu cầu: Python 3.8+, không cần thư viện ngoài.

Nguồn dữ liệu:
  - data/gadm41_VNM_1.json  : bounding box từng tỉnh
  - internal/adapter/migration/sql/0002_seed_provinces.sql : GSO codes + tên
"""

import argparse
import json
import math
import re
import sys
import time
import unicodedata
import urllib.error
import urllib.parse
import urllib.request
from pathlib import Path

# ---------------------------------------------------------------------------
# Cấu hình
# ---------------------------------------------------------------------------
REPO_ROOT = Path(__file__).parent.parent
GADM_L1   = REPO_ROOT / "data" / "gadm41_VNM_1.json"
SEED_SQL  = REPO_ROOT / "internal" / "adapter" / "migration" / "sql" / "0002_seed_provinces.sql"

OVERPASS_URL = "https://overpass-api.de/api/interpreter"

HIGHWAY_FILTER = (
    "trunk|trunk_link|primary|primary_link|secondary|secondary_link"
    "|tertiary|tertiary_link|residential|living_street|unclassified"
)

BETWEEN_REQUESTS_S = 10   # giây chờ giữa các request (tránh bị 429)
OVERPASS_TIMEOUT   = 120  # giây timeout cho Overpass

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def normalize(s: str) -> str:
    s = s.lower().replace("đ", "d")
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if unicodedata.category(c) != "Mn")
    return re.sub(r"[^a-z0-9]", "", s)


def slugify(name: str) -> str:
    name = name.lower().replace("đ", "d")
    name = unicodedata.normalize("NFD", name)
    name = "".join(c for c in name if unicodedata.category(c) != "Mn")
    return re.sub(r"[^a-z0-9]+", "-", name).strip("-")


def esc(s: str) -> str:
    return (s or "").replace("'", "''")


def haversine_m(lon1, lat1, lon2, lat2) -> float:
    dlat = math.radians(lat2 - lat1)
    dlon = math.radians(lon2 - lon1)
    a = (math.sin(dlat / 2) ** 2
         + math.cos(math.radians(lat1)) * math.cos(math.radians(lat2))
         * math.sin(dlon / 2) ** 2)
    return 6_371_000 * 2 * math.asin(math.sqrt(a))


def compute_length_m(segments: list) -> float:
    total = 0.0
    for seg in segments:
        for i in range(len(seg) - 1):
            total += haversine_m(*seg[i], *seg[i + 1])
    return total


# ---------------------------------------------------------------------------
# Load province mapping (GADM bbox + GSO code)
# ---------------------------------------------------------------------------

def load_province_map() -> dict:
    """Return {gso_code: {name, bbox(s,w,n,e)}}."""
    # 1. parse GADM bboxes
    gadm_data = json.loads(GADM_L1.read_text(encoding="utf-8"))
    gadm: dict[str, tuple] = {}
    for f in gadm_data["features"]:
        p = f["properties"]
        g = f["geometry"]
        pts = [pt for poly in g["coordinates"] for ring in poly for pt in ring]
        lons = [c[0] for c in pts]
        lats = [c[1] for c in pts]
        key = normalize(p["NAME_1"])
        gadm[key] = (
            round(min(lats), 4),
            round(min(lons), 4),
            round(max(lats), 4),
            round(max(lons), 4),
        )

    # 2. parse seed SQL for GSO codes + names
    seed_text = SEED_SQL.read_text(encoding="utf-8")
    provinces: dict[str, dict] = {}
    for m in re.finditer(r"'(\d{2})','([^']+)'", seed_text):
        code, name = m.group(1), m.group(2)
        key = normalize(name)
        bbox = gadm.get(key)
        if bbox is None:
            # partial match
            for k, v in gadm.items():
                if key in k or k in key:
                    bbox = v
                    break
        if bbox is None:
            print(f"-- WARNING: no bbox for province {code} '{name}'", file=sys.stderr)
            continue
        provinces[code] = {"name": name, "bbox": bbox}

    return provinces


# ---------------------------------------------------------------------------
# Overpass fetch
# ---------------------------------------------------------------------------

def build_query(bbox: tuple) -> str:
    s, w, n, e = bbox
    return (
        f"[out:json][timeout:{OVERPASS_TIMEOUT}];\n"
        f"(\n"
        f"  way[\"highway\"~\"^({HIGHWAY_FILTER})$\"][\"name\"]({s},{w},{n},{e});\n"
        f");\n"
        f"out geom;\n"
    )


def fetch_overpass(query: str, retries: int = 3) -> list:
    payload = urllib.parse.urlencode({"data": query}).encode()
    for attempt in range(retries):
        try:
            req = urllib.request.Request(
                OVERPASS_URL,
                data=payload,
                headers={
                    "Content-Type": "application/x-www-form-urlencoded",
                    "User-Agent": "geoservice-vn-streets/1.0 (Vietnam admin geo data)",
                },
            )
            with urllib.request.urlopen(req, timeout=OVERPASS_TIMEOUT + 30) as resp:
                return json.load(resp).get("elements", [])
        except urllib.error.HTTPError as e:
            if e.code == 429 and attempt < retries - 1:
                wait = 60 * (attempt + 1)
                print(f"--   Rate limited (429), chờ {wait}s...", file=sys.stderr)
                time.sleep(wait)
                continue
            print(f"--   HTTP {e.code} lần {attempt+1}/{retries}", file=sys.stderr)
            if attempt == retries - 1:
                raise
        except Exception as ex:
            print(f"--   Lỗi lần {attempt+1}/{retries}: {ex}", file=sys.stderr)
            if attempt == retries - 1:
                raise
    return []


# ---------------------------------------------------------------------------
# Process elements
# ---------------------------------------------------------------------------

def process_elements(elements: list) -> dict:
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
        if name not in streets:
            streets[name] = {
                "name": name,
                "name_en": tags.get("name:en", ""),
                "gid": str(el["id"]),
                "segments": [coords],
            }
        else:
            streets[name]["segments"].append(coords)
    return streets


# ---------------------------------------------------------------------------
# Render SQL rows
# ---------------------------------------------------------------------------

def render_rows(province_code: str, streets: dict) -> list[str]:
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
            "  ('{prov}','{name}','{nf}','{en}','{slug}','{gid}',{length},"
            "ST_Multi(ST_MakeValid(ST_SetSRID(ST_GeomFromGeoJSON('{geom}'),4326))))".format(
                prov=province_code,
                name=esc(name),
                nf=esc(name_full),
                en=esc(name_en),
                slug=esc(slug),
                gid=esc(gid),
                length=length_m,
                geom=esc(geojson),
            )
        )
    return rows


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------

def main():
    parser = argparse.ArgumentParser(description="Fetch VN streets from OSM → SQL")
    parser.add_argument("--only", help="Chỉ fetch các GSO code này (vd: 01,79,31)")
    parser.add_argument("--skip", help="Bỏ qua các GSO code này")
    args = parser.parse_args()

    only = set(args.only.split(",")) if args.only else None
    skip = set(args.skip.split(",")) if args.skip else set()

    provinces = load_province_map()
    print(f"-- Tổng {len(provinces)} tỉnh/thành được tải.", file=sys.stderr)

    codes = sorted(provinces.keys())
    if only:
        codes = [c for c in codes if c in only]
    codes = [c for c in codes if c not in skip]
    print(f"-- Sẽ fetch {len(codes)} tỉnh/thành.", file=sys.stderr)

    # SQL header
    print("-- 0015_seed_all_streets.sql : đường phố 63 tỉnh/thành VN từ OpenStreetMap.")
    print("-- Auto-generated bởi scripts/fetch_all_streets.py — không sửa tay.")
    print()
    print("INSERT INTO streets")
    print("  (province_code, name, name_full, name_en, slug, gid, length_m, geom)")
    print("VALUES")

    first_row = True
    for i, code in enumerate(codes):
        info = provinces[code]
        name = info["name"]
        bbox = info["bbox"]
        print(f"-- [{i+1}/{len(codes)}] {code} {name} bbox={bbox}", file=sys.stderr)

        try:
            elements = fetch_overpass(build_query(bbox))
        except Exception as ex:
            print(f"--   SKIP {code} {name}: {ex}", file=sys.stderr)
            continue

        streets = process_elements(elements)
        rows = render_rows(code, streets)
        print(f"--   {len(elements)} ways → {len(streets)} tên đường → {len(rows)} rows", file=sys.stderr)

        for row in rows:
            if not first_row:
                print(",")
            print(row, end="")
            first_row = False

        # Rate limit: không request quá nhanh
        if i < len(codes) - 1:
            print(f"--   Chờ {BETWEEN_REQUESTS_S}s...", file=sys.stderr)
            time.sleep(BETWEEN_REQUESTS_S)

    print()  # newline sau row cuối
    print("ON CONFLICT DO NOTHING;")
    print(f"-- Hoàn thành {len(codes)} tỉnh/thành.", file=sys.stderr)


if __name__ == "__main__":
    main()
