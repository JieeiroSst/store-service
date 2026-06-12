#!/usr/bin/env bash
# Regenerate the GSO administrative seed migrations (provinces/districts/wards)
# from the upstream open dataset (Tổng cục Thống kê, đóng gói bởi madnh/hanhchinhvn).
#
# Chạy lại khi muốn cập nhật danh sách đơn vị hành chính. Yêu cầu: curl, python3.
# Output: internal/adapter/migration/sql/000{2,3,4}_seed_*.sql
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
SQLDIR="$ROOT/internal/adapter/migration/sql"
TMP="$(mktemp -d)"
BASE="https://raw.githubusercontent.com/madnh/hanhchinhvn/master/dist"

echo ">> Downloading GSO admin dataset..."
curl -fsSL -o "$TMP/tinh_tp.json"    "$BASE/tinh_tp.json"
curl -fsSL -o "$TMP/quan_huyen.json" "$BASE/quan_huyen.json"
curl -fsSL -o "$TMP/xa_phuong.json"  "$BASE/xa_phuong.json"

echo ">> Generating seed SQL into $SQLDIR ..."
python3 "$ROOT/scripts/gen_seed.py" "$TMP" "$SQLDIR"

rm -rf "$TMP"
echo ">> Done. Migrations 0002/0003/0004 regenerated."
