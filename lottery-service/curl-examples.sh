#!/usr/bin/env bash
# Ví dụ gọi API lottery-service bằng curl.
#
# Cách dùng:
#   ./curl-examples.sh                # gọi lần lượt tất cả ví dụ vào BASE_URL mặc định
#   BASE_URL=http://localhost:8010 ./curl-examples.sh
#   ./curl-examples.sh check_won      # chỉ chạy 1 ví dụ theo tên hàm
#
# Yêu cầu service đang chạy (uvicorn local hoặc docker compose up).

set -euo pipefail

BASE_URL="${BASE_URL:-http://localhost:8000}"

_hr() { printf '\n\033[1;34m>>> %s\033[0m\n' "$1"; }

health() {
  _hr "GET /health"
  curl -s "$BASE_URL/health"; echo
}

check_won() {
  _hr "POST /api/v1/check — vé trúng giải đặc biệt"
  curl -s -X POST "$BASE_URL/api/v1/check" \
    -H "Content-Type: application/json" \
    -d '{"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "145917"}'
  echo
}

check_won_by_code() {
  _hr "POST /api/v1/check — dùng province_code thay vì tên tỉnh có dấu"
  curl -s -X POST "$BASE_URL/api/v1/check" \
    -H "Content-Type: application/json" \
    -d '{"date": "2026-08-26", "province": "dongnai", "ticket_number": "145917"}'
  echo
}

check_not_won() {
  _hr "POST /api/v1/check — vé không trúng (có kết quả ngày đó, nhưng không khớp)"
  curl -s -X POST "$BASE_URL/api/v1/check" \
    -H "Content-Type: application/json" \
    -d '{"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "000000"}'
  echo
}

check_result_not_found() {
  _hr "POST /api/v1/check — chưa có kết quả cho ngày/tỉnh này (HTTP 200, result_found=false)"
  curl -s -X POST "$BASE_URL/api/v1/check" \
    -H "Content-Type: application/json" \
    -d '{"date": "2099-01-01", "province": "Đồng Nai", "ticket_number": "145917"}'
  echo
}

check_invalid_ticket() {
  _hr "POST /api/v1/check — ticket_number không hợp lệ (HTTP 422)"
  curl -s -o /dev/null -w "HTTP %{http_code}\n" -X POST "$BASE_URL/api/v1/check" \
    -H "Content-Type: application/json" \
    -d '{"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "abc"}'
}

get_result() {
  _hr "GET /api/v1/results — lấy nguyên bản kết quả 1 tỉnh/đài trong 1 ngày"
  curl -s -G "$BASE_URL/api/v1/results" \
    --data-urlencode "date=2026-08-26" \
    --data-urlencode "province=Đồng Nai"
  echo
}

trigger_scrape() {
  # Gọi mạng thật ra nguồn xổ số cấu hình ở SCRAPER_BASE_URL -> có thể chậm
  # (retry + backoff) hoặc trả provinces_saved=0 nếu môi trường không có internet.
  _hr "POST /api/v1/scrape — cào thủ công 1 ngày (cả 3 miền)"
  curl -s -X POST "$BASE_URL/api/v1/scrape" \
    -H "Content-Type: application/json" \
    -d '{"date": "2026-08-26"}'
  echo
}

trigger_backfill() {
  # Chỉ backfill 2 ngày cho ví dụ chạy nhanh; đổi start/end để backfill khoảng
  # dài hơn. Request HTTP sẽ block tới khi backfill xong toàn bộ khoảng ngày.
  _hr "POST /api/v1/backfill — cào lịch sử theo khoảng ngày (resume-safe)"
  curl -s -X POST "$BASE_URL/api/v1/backfill" \
    -H "Content-Type: application/json" \
    -d '{"start": "2026-08-25", "end": "2026-08-26", "force": false}'
  echo
}

main() {
  health
  check_won
  check_won_by_code
  check_not_won
  check_result_not_found
  check_invalid_ticket
  get_result
  trigger_scrape
  trigger_backfill
}

if [ "$#" -gt 0 ]; then
  "$@"
else
  main
fi
