#!/usr/bin/env bash
set -uo pipefail

BASE_URL="${BASE_URL:-http://localhost:1235}"
API="$BASE_URL/api/v1"
REQUESTED_BY="${REQUESTED_BY:-curl-sample}"
USER_ID="${USER_ID:-42}"
EMAIL="${EMAIL:-khach.hang@example.com}"
PHONE="${PHONE:-0912345678}"
DEVICE_ID="${DEVICE_ID:-install-uuid-stored-by-the-app}"

title() { printf '\n\033[1;34m== %s ==\033[0m\n' "$1"; }

call() {
  local method="$1" path="$2" body="${3:-}"
  printf '\033[2m%s %s\033[0m\n' "$method" "$path"
  if [ -n "$body" ]; then
    curl -sS -X "$method" "$API$path" \
      -H 'Content-Type: application/json' \
      -H "X-Requested-By: $REQUESTED_BY" \
      -d "$body" -w '\n[HTTP %{http_code}]\n'
  else
    curl -sS -X "$method" "$API$path" \
      -H "X-Requested-By: $REQUESTED_BY" \
      -w '\n[HTTP %{http_code}]\n'
  fi
}

health() {
  title "Health"
  curl -sS "$BASE_URL/health" -w '\n[HTTP %{http_code}]\n'
}

devices() {
  title "Devices of a user (tokens stay in notification-service, only token_preview is returned)"
  call GET "/devices?user_id=$USER_ID"

  title "Logout on one device (by device_id, no token needed)"
  call POST /devices/unregister "{\"user_id\": $USER_ID, \"device_id\": \"$DEVICE_ID\"}"

  title "Logout on every device of the user"
  call POST /devices/unregister "{\"user_id\": $USER_ID}"
}

contacts() {
  title "Set user contact (email / phone used to find the user's devices)"
  call PUT "/users/$USER_ID/contact" "{\"email\": \"$EMAIL\", \"phone\": \"+84 912 345 678\"}"

  title "Get user contact"
  call GET "/users/$USER_ID/contact"
}

push() {
  title "Push by user_id"
  call POST /notifications/push "{
    \"user_id\": $USER_ID,
    \"title\": \"Đơn hàng DH-1024\",
    \"message\": \"Đơn hàng của bạn đang được giao\",
    \"data\": {\"screen\": \"order_detail\", \"order_id\": \"1024\"}
  }"

  title "Push by email"
  call POST /notifications/push "{\"email\": \"$EMAIL\", \"title\": \"Khuyến mãi\", \"message\": \"Giảm 20% hôm nay\"}"

  title "Push by phone"
  call POST /notifications/push "{\"phone\": \"$PHONE\", \"title\": \"Mã OTP\", \"message\": \"Mã của bạn là 123456\"}"
}

email() {
  title "Email from template"
  call POST /notifications/email "{
    \"recipient\": \"$EMAIL\",
    \"template_type\": \"welcome\",
    \"template_data\": {\"Name\": \"Quan\", \"Email\": \"$EMAIL\"}
  }"

  title "Raw HTML email"
  call POST /notifications/email "{
    \"recipient\": \"$EMAIL\",
    \"subject\": \"Hóa đơn #INV-2026-001\",
    \"html\": \"<h2>Cảm ơn bạn đã mua hàng</h2><p>Tổng tiền: <b>1.250.000đ</b></p>\"
  }"

  title "Raw text email with raw_data (rendered as a key/value table)"
  call POST /notifications/email "{
    \"recipient\": \"$EMAIL\",
    \"subject\": \"Báo cáo doanh thu ngày 28/09\",
    \"text\": \"Chào anh/chị,\\nDưới đây là số liệu trong ngày.\",
    \"raw_data\": {
      \"revenue\": 125000000,
      \"orders\": 342,
      \"refunds\": 3,
      \"top_products\": [\"SKU-001\", \"SKU-017\"],
      \"by_channel\": {\"web\": 210, \"app\": 132}
    }
  }"

  title "Template email plus raw_data"
  call POST /notifications/email "{
    \"recipient\": \"$EMAIL\",
    \"template_type\": \"welcome\",
    \"template_data\": {\"Name\": \"Quan\"},
    \"raw_data\": {\"referral_code\": \"QUAN2026\"}
  }"
}

slack() {
  title "Slack from template"
  call POST /notifications/slack "{
    \"template_type\": \"order_status\",
    \"template_data\": {\"OrderID\": \"1024\", \"Status\": \"shipped\", \"CustomerName\": \"Quan\"}
  }"

  title "Raw Slack message (mrkdwn)"
  call POST /notifications/slack "{
    \"title\": \"Deploy thành công\",
    \"text\": \"*payment-service* v1.4.2 đã lên _production_ :rocket:\"
  }"

  title "Raw Slack message with raw_data (rendered as a JSON code block)"
  call POST /notifications/slack "{
    \"title\": \"Cảnh báo thanh toán lỗi\",
    \"text\": \"Tỉ lệ lỗi vượt ngưỡng trong 5 phút qua\",
    \"raw_data\": {\"service\": \"payment-service\", \"error_rate\": 0.12, \"p95_ms\": 1840, \"regions\": [\"hcm\", \"hn\"]}
  }"

  title "Slack with only raw_data"
  call POST /notifications/slack "{\"raw_data\": {\"job\": \"sync-orders\", \"status\": \"done\", \"rows\": 10452}}"
}

notifications() {
  title "Generic notification (type push | email | slack)"
  call POST /notifications "{\"type\": \"push\", \"user_id\": $USER_ID, \"title\": \"Xin chào\", \"message\": \"Test\"}"

  title "List notifications"
  call GET /notifications

  title "Get notification 1 (status: pending | retrying | sent | failed, last_error)"
  call GET /notifications/1
}

campaigns() {
  title "Campaign: push to every active device"
  call POST /campaigns "{
    \"channel\": \"push\",
    \"audience\": \"all_devices\",
    \"title\": \"Flash sale 10.10\",
    \"message\": \"Giảm 50% toàn bộ đơn hàng\",
    \"data\": {\"screen\": \"promo\", \"promo_id\": \"1010\"}
  }"

  title "Campaign: push to selected users"
  call POST /campaigns "{
    \"channel\": \"push\",
    \"audience\": \"users\",
    \"user_ids\": [$USER_ID, 43, 44],
    \"title\": \"Ưu đãi riêng cho bạn\",
    \"message\": \"Voucher 100k đã được thêm vào ví\"
  }"

  title "Campaign: push to an FCM topic"
  call POST /campaigns "{
    \"channel\": \"push\",
    \"audience\": \"topic\",
    \"topic\": \"all-users\",
    \"title\": \"Bảo trì hệ thống\",
    \"message\": \"Hệ thống bảo trì 0h-2h ngày 01/10\"
  }"

  title "Campaign: email to a list"
  call POST /campaigns "{
    \"channel\": \"email\",
    \"audience\": \"emails\",
    \"emails\": [\"$EMAIL\", \"khach2@example.com\"],
    \"title\": \"Bản tin tháng 10\",
    \"message\": \"<h1>Bản tin</h1><p>Nội dung...</p>\"
  }"

  title "Campaign progress"
  call GET /campaigns/1

  title "List campaigns"
  call GET "/campaigns?limit=20"

  title "Cancel campaign"
  call POST /campaigns/1/cancel
}

audit() {
  title "Audit: everything sent to a phone number (push + emails to the user's address)"
  call GET "/audit/deliveries?phone=$PHONE"

  title "Audit: by email, only failures"
  call GET "/audit/deliveries?email=$EMAIL&status=failed"

  title "Audit: by user, push only, in a time range"
  call GET "/audit/deliveries?user_id=$USER_ID&channel=push&from=2026-09-01T00:00:00%2B07:00&to=2026-10-01T00:00:00%2B07:00"

  title "Audit: everything a campaign sent, requested by marketing"
  call GET "/audit/deliveries?source_type=campaign&source_id=1&requested_by=marketing&limit=100"

  title "Audit: next page (use next_before_id from the previous response)"
  call GET "/audit/deliveries?limit=50&before_id=1000"

  title "Audit: exact content that was sent"
  call GET /audit/contents/1
}

usage() {
  cat <<EOF
Usage: ./curl.sh [section ...]

Sections: health devices contacts push email slack notifications campaigns audit all

Environment:
  BASE_URL      default http://localhost:1235
  REQUESTED_BY  sent as X-Requested-By, recorded in the audit (default curl-sample)
  USER_ID EMAIL PHONE DEVICE_ID  sample values used in the requests

FCM tokens are stored and managed by notification-service; no request here needs one.

Examples:
  ./curl.sh email slack
  BASE_URL=https://notify.example.com REQUESTED_BY=order-service ./curl.sh push
EOF
}

[ $# -eq 0 ] && { usage; exit 0; }
for section in "$@"; do
  case "$section" in
    health | devices | contacts | push | email | slack | notifications | campaigns | audit) "$section" ;;
    all) health; contacts; devices; push; email; slack; notifications; campaigns; audit ;;
    -h | --help | help) usage ;;
    *) echo "unknown section: $section" >&2; usage; exit 1 ;;
  esac
done
