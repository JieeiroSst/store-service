#!/usr/bin/env python3
"""Generate GSO admin seed SQL (provinces/districts/wards) from hanhchinhvn JSON.

Usage: gen_seed.py <input_dir> <output_sql_dir>
  input_dir  must contain tinh_tp.json, quan_huyen.json, xa_phuong.json
  output_dir receives 0002_seed_provinces.sql, 0003_seed_districts.sql,
             0004_seed_wards.sql
"""
import json
import sys
import unicodedata

# Region (Bắc/Trung/Nam) keyed by province GSO code.
REGION = {
    "01": "Bắc", "02": "Bắc", "04": "Bắc", "06": "Bắc", "08": "Bắc", "10": "Bắc",
    "11": "Bắc", "12": "Bắc", "14": "Bắc", "15": "Bắc", "17": "Bắc", "19": "Bắc",
    "20": "Bắc", "22": "Bắc", "24": "Bắc", "25": "Bắc", "26": "Bắc", "27": "Bắc",
    "30": "Bắc", "31": "Bắc", "33": "Bắc", "34": "Bắc", "35": "Bắc", "36": "Bắc", "37": "Bắc",
    "38": "Trung", "40": "Trung", "42": "Trung", "44": "Trung", "45": "Trung", "46": "Trung",
    "48": "Trung", "49": "Trung", "51": "Trung", "52": "Trung", "54": "Trung", "56": "Trung",
    "58": "Trung", "60": "Trung", "62": "Trung", "64": "Trung", "66": "Trung", "67": "Trung", "68": "Trung",
    "70": "Nam", "72": "Nam", "74": "Nam", "75": "Nam", "77": "Nam", "79": "Nam", "80": "Nam",
    "82": "Nam", "83": "Nam", "84": "Nam", "86": "Nam", "87": "Nam", "89": "Nam", "91": "Nam",
    "92": "Nam", "93": "Nam", "94": "Nam", "95": "Nam", "96": "Nam",
}


def esc(s):
    return (s or "").replace("'", "''")


def ascii_fold(s):
    s = (s or "").lower().replace("đ", "d")
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if unicodedata.category(c) != "Mn")
    return s.title()


def main(indir, outdir):
    tinh = json.load(open(f"{indir}/tinh_tp.json", encoding="utf-8"))
    huyen = json.load(open(f"{indir}/quan_huyen.json", encoding="utf-8"))
    xa = json.load(open(f"{indir}/xa_phuong.json", encoding="utf-8"))

    miss = set(tinh) - set(REGION)
    if miss:
        raise SystemExit(f"province(s) missing region mapping: {sorted(miss)}")

    # provinces
    L = ["-- 0002_seed_provinces.sql : 63 tỉnh/thành (mã GSO). Auto-generated.",
         "INSERT INTO provinces (code, name, name_full, name_en, type, region, slug) VALUES"]
    rows = []
    for code, p in sorted(tinh.items()):
        rows.append("  ('%s','%s','%s','%s','%s','%s','%s')" % (
            code, esc(p["name"]), esc(p.get("name_with_type", "")),
            esc(ascii_fold(p["name"])), esc(p.get("type", "")),
            REGION[code], esc(p.get("slug", ""))))
    L.append(",\n".join(rows))
    L.append("ON CONFLICT (code) DO UPDATE SET")
    L.append("  name=EXCLUDED.name, name_full=EXCLUDED.name_full, name_en=EXCLUDED.name_en,")
    L.append("  type=EXCLUDED.type, region=EXCLUDED.region, slug=EXCLUDED.slug;")
    open(f"{outdir}/0002_seed_provinces.sql", "w", encoding="utf-8").write("\n".join(L) + "\n")

    # districts
    L = ["-- 0003_seed_districts.sql : quận/huyện/thị xã (mã GSO). Auto-generated.",
         "INSERT INTO districts (code, province_code, name, name_full, name_en, type, slug) VALUES"]
    rows = []
    for code, d in sorted(huyen.items()):
        rows.append("  ('%s','%s','%s','%s','%s','%s','%s')" % (
            code, d["parent_code"], esc(d["name"]), esc(d.get("name_with_type", "")),
            esc(ascii_fold(d["name"])), esc(d.get("type", "")), esc(d.get("slug", ""))))
    L.append(",\n".join(rows))
    L.append("ON CONFLICT (code) DO UPDATE SET")
    L.append("  province_code=EXCLUDED.province_code, name=EXCLUDED.name,")
    L.append("  name_full=EXCLUDED.name_full, name_en=EXCLUDED.name_en,")
    L.append("  type=EXCLUDED.type, slug=EXCLUDED.slug;")
    open(f"{outdir}/0003_seed_districts.sql", "w", encoding="utf-8").write("\n".join(L) + "\n")

    # wards (resolve province via district parent)
    huyen_parent = {c: d["parent_code"] for c, d in huyen.items()}
    L = ["-- 0004_seed_wards.sql : phường/xã/thị trấn (mã GSO). Auto-generated.",
         "INSERT INTO wards (code, district_code, province_code, name, name_full, name_en, type, slug) VALUES"]
    rows = []
    orphan = 0
    for code, w in sorted(xa.items()):
        dcode = w["parent_code"]
        pcode = huyen_parent.get(dcode)
        if pcode is None:
            orphan += 1
            continue
        rows.append("  ('%s','%s','%s','%s','%s','%s','%s','%s')" % (
            code, dcode, pcode, esc(w["name"]), esc(w.get("name_with_type", "")),
            esc(ascii_fold(w["name"])), esc(w.get("type", "")), esc(w.get("slug", ""))))
    L.append(",\n".join(rows))
    L.append("ON CONFLICT (code) DO UPDATE SET")
    L.append("  district_code=EXCLUDED.district_code, province_code=EXCLUDED.province_code,")
    L.append("  name=EXCLUDED.name, name_full=EXCLUDED.name_full, name_en=EXCLUDED.name_en,")
    L.append("  type=EXCLUDED.type, slug=EXCLUDED.slug;")
    open(f"{outdir}/0004_seed_wards.sql", "w", encoding="utf-8").write("\n".join(L) + "\n")

    print(f"provinces={len(tinh)} districts={len(huyen)} wards={len(xa) - orphan} orphan={orphan}")


if __name__ == "__main__":
    if len(sys.argv) != 3:
        raise SystemExit("usage: gen_seed.py <input_dir> <output_sql_dir>")
    main(sys.argv[1], sys.argv[2])
