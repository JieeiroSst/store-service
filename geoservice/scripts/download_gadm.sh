#!/usr/bin/env bash
# Download GADM v4.1 administrative boundaries for Vietnam (levels 1–3) as GeoJSON.
#
#   level 1 = tỉnh / thành phố          (63 đơn vị, cơ cấu 2008)
#   level 2 = quận / huyện / thị xã
#   level 3 = phường / xã / thị trấn
#
# Two strategies:
#   (A) DIRECT JSON  — download GADM's pre-exported zipped GeoJSON. No GDAL needed.
#                      Requires: curl, unzip.
#   (B) GPKG+ogr2ogr — download the GeoPackage and export each level yourself.
#                      Requires: curl, ogr2ogr (gdal-bin). Set USE_GPKG=1.
#
# Output -> ./data/gadm41_VNM_{1,2,3}.json
set -euo pipefail

OUT_DIR="${1:-data}"
BASE="https://geodata.ucdavis.edu/gadm/gadm4.1"
mkdir -p "$OUT_DIR"
cd "$OUT_DIR"

if [[ "${USE_GPKG:-0}" == "1" ]]; then
  # ---- Strategy B: GeoPackage + ogr2ogr ----
  GPKG="gadm41_VNM.gpkg"
  [[ -f "$GPKG" ]] || curl -fSL --retry 3 -o "$GPKG" "$BASE/gpkg/gadm41_VNM.gpkg"
  for lvl in 1 2 3; do
    echo ">> Exporting level ${lvl} via ogr2ogr"
    ogr2ogr -f GeoJSON -t_srs EPSG:4326 "gadm41_VNM_${lvl}.json" "$GPKG" "ADM_ADM_${lvl}"
  done
else
  # ---- Strategy A: direct zipped GeoJSON (default) ----
  for lvl in 1 2 3; do
    zip="gadm41_VNM_${lvl}.json.zip"
    echo ">> Downloading level ${lvl}: ${zip}"
    curl -fSL --retry 3 -o "$zip" "$BASE/json/${zip}"
    unzip -o "$zip" >/dev/null
    rm -f "$zip"
  done
fi

echo ">> Done. Files in $(pwd):"
ls -lh gadm41_VNM_*.json
cat <<NEXT

Next — import into Postgres/PostGIS:
  go run ./cmd/importgeo \\
    -l1 ${OUT_DIR}/gadm41_VNM_1.json \\
    -l2 ${OUT_DIR}/gadm41_VNM_2.json \\
    -l3 ${OUT_DIR}/gadm41_VNM_3.json
NEXT
