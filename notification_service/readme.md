# notification-service

## Dùng thử nhanh: `curl.sh`

[curl.sh](curl.sh) có request mẫu cho mọi API, chạy từng nhóm hoặc tất cả:

```
./curl.sh                      # xem danh sách nhóm
./curl.sh email slack          # gửi thử email và Slack
BASE_URL=http://localhost:1235 REQUESTED_BY=order-service ./curl.sh push
./curl.sh all
```

Nhóm: `health devices contacts push email slack notifications campaigns audit all`. Biến môi trường: `BASE_URL`, `REQUESTED_BY` (gửi thành header `X-Requested-By`, được ghi vào audit), `USER_ID`, `EMAIL`, `PHONE`, `DEVICE_ID`. Không request nào cần FCM token: token do notification-service lưu và quản lý; việc app giao token cho backend (`POST /devices`) thuộc phía app mobile, xem mục "Thiết bị và FCM token".

## API

### Send email

```
POST /api/v1/notifications/email
```

Lưu notification (status `pending`), publish lên RabbitMQ; consumer sẽ render template rồi gửi qua Resend API.

Request:

```json
{
  "user_id": 1,
  "recipient": "someone@example.com",
  "template_type": "welcome",
  "template_data": {
    "Name": "Quan",
    "Email": "someone@example.com"
  },
  "priority": 0
}
```

Gửi nội dung thô thay vì template:

```json
{
  "recipient": "someone@example.com",
  "subject": "Báo cáo doanh thu ngày 28/09",
  "text": "Chào anh/chị,\nDưới đây là số liệu trong ngày.",
  "raw_data": {"revenue": 125000000, "orders": 342, "by_channel": {"web": 210, "app": 132}}
}
```

- Chọn một trong hai: `template_type` + `template_data`, **hoặc** `subject` + `html` / `text` (`text` được hiển thị nguyên định dạng, có escape HTML).
- `raw_data`: JSON object bất kỳ (tối đa 64KB), dùng được với cả hai kiểu; được thêm vào cuối email dưới dạng bảng key/value (key sắp xếp theo alphabet, giá trị được escape HTML, object/mảng hiển thị dạng JSON).
- `recipient`: bắt buộc, phải là email hợp lệ.
- `template_type`: tên file (không có `.html`) trong [internal/adapter/secondary/template/templates](internal/adapter/secondary/template/templates). Hiện có: `welcome`, `otp`, `reset_password`.
- `template_data`: `map[string]string`, dùng để thay các placeholder `{{.Key}}` trong template.

Response `202 Accepted`: notification vừa tạo (status `pending`).

Muốn thêm template mới: thả file `.html` mới vào `internal/adapter/secondary/template/templates/`, định nghĩa 2 block `{{define "subject"}}...{{end}}` và `{{define "html"}}...{{end}}`, không cần sửa code Go.

### Send Slack message

```
POST /api/v1/notifications/slack
```

Cùng cơ chế với email: lưu notification, publish RabbitMQ, consumer render template rồi gửi qua Slack webhook.

Request:

```json
{
  "user_id": 1,
  "template_type": "order_status",
  "template_data": {
    "OrderID": "1024",
    "Status": "shipped",
    "CustomerName": "Quan"
  },
  "priority": 0
}
```

Gửi nội dung thô thay vì template:

```json
{
  "title": "Cảnh báo thanh toán lỗi",
  "text": "Tỉ lệ lỗi vượt ngưỡng trong *5 phút* qua",
  "raw_data": {"service": "payment-service", "error_rate": 0.12, "p95_ms": 1840}
}
```

- Chọn một trong hai: `template_type` + `template_data`, **hoặc** `title` + `text` (cú pháp Slack mrkdwn). Có thể chỉ gửi `raw_data`.
- `raw_data`: JSON object bất kỳ (tối đa 64KB), được thêm vào cuối tin nhắn dưới dạng khối code JSON.
- `template_type`: tên file (không có `.txt`) trong [internal/adapter/secondary/slacktemplate/templates](internal/adapter/secondary/slacktemplate/templates). Hiện có: `alert`, `order_status`, `daily_report`.
- `template_data`: `map[string]string`, dùng để thay các placeholder `{{.Key}}` trong template.

Response `202 Accepted`: notification vừa tạo (status `pending`).

Muốn thêm template mới: thả file `.txt` mới vào `internal/adapter/secondary/slacktemplate/templates/`, định nghĩa 2 block `{{define "title"}}...{{end}}` và `{{define "text"}}...{{end}}` (dùng cú pháp Slack mrkdwn, xem [docs.slack.dev/messaging](https://docs.slack.dev/messaging/)), không cần sửa code Go.

## Gửi push theo user_id, email hoặc số điện thoại

```
POST /api/v1/notifications/push
{"phone": "0912 345 678", "title": "Đơn hàng DH-1", "message": "Đang giao", "data": {"screen": "order", "id": "1"}}
```

Truyền một trong `user_id`, `email`, `phone`. Service tra ra user, lấy mọi FCM token đang active của user đó (Android và iOS) và gửi bằng một lần `SendEachForMulticast`; FCM tự áp cấu hình Android hoặc APNs theo từng máy. `data` được gửi kèm để app mở đúng màn hình. Email/số điện thoại chưa gắn với user nào thì notification `failed` ngay, không retry.

Bảng liên hệ `notification_user_contact` (`user_id` ↔ email ↔ phone) được cập nhật khi app đăng ký token (gửi kèm `email`, `phone`) hoặc do service khác gọi:

```
PUT /api/v1/users/:user_id/contact   {"email": "an@shop.vn", "phone": "+84912345678"}
GET /api/v1/users/:user_id/contact
```

Mỗi email/số điện thoại chỉ thuộc một user: gán cho user mới thì user cũ mất email/số đó. Số điện thoại được chuẩn hóa (`+84 912 345 678` → `0912345678`), email chuyển về chữ thường.

## Thiết bị và FCM token

FCM token chỉ có thể do Firebase SDK trên máy sinh ra: Google không có API để server tạo token cho một thiết bị. Backend quản lý toàn bộ vòng đời còn lại:

- **Xác thực khi đăng ký:** token được kiểm tra bằng FCM dry-run (không gửi thông báo thật); FCM từ chối thì trả 400 (`push.validate_on_register`, mặc định bật; lỗi mạng hoặc Firebase chưa cấu hình thì không chặn đăng nhập).
- **Gắn với user**, chuyển chủ khi đổi người đăng nhập, thay thế khi token đổi (theo `device_id`).
- **Tự tắt token:** FCM báo unregistered khi gửi, hoặc không được làm mới quá `push.stale_token_days` ngày (mặc định 270, mốc FCM coi token không hoạt động là hết hạn). Job chạy mỗi giờ. App nên gọi lại `POST /devices` mỗi lần mở app để cập nhật `last_used_at`.

FCM token chỉ nằm trong DB của notification-service: app mobile giao token một lần qua `POST /devices`, còn mọi service khác gửi thông báo bằng `user_id`, `email` hoặc `phone` và không bao giờ cần biết token. API không trả token ra ngoài (chỉ có `token_preview` dạng `abc123…wxyz` để đối chiếu), và `PUT /devices/:id` không sửa được token.

```
POST /api/v1/devices             {"user_id": 7, "device_token": "<FCM token từ Firebase SDK>", "device_id": "<id cài app>", "device_type": "ios|android|web", "email": "...", "phone": "..."}
GET  /api/v1/devices?user_id=7
POST /api/v1/devices/unregister  {"user_id": 7, "device_id": "<id cài app>"}   (đăng xuất một máy)
POST /api/v1/devices/unregister  {"user_id": 7}                                 (đăng xuất mọi máy)
```

- Đăng ký là upsert theo token (unique theo `sha256(token)`): cùng một máy đổi người đăng nhập thì token chuyển sang user mới, user cũ không còn nhận thông báo của máy đó.
- `device_id` (ID ổn định cho một lần cài app, do app tự sinh và lưu lại): khi FCM đổi token, token cũ của cùng `device_id` tự tắt.
- `push.single_device_per_user` (mặc định `false`): `false` thì mọi máy đang đăng nhập đều nhận; `true` thì đăng nhập máy mới sẽ tắt các máy khác của user.
- `user_id = 0` bị từ chối: token phải được gửi sau khi đăng nhập.
- Token iOS dạng 64 ký tự hex bị từ chối vì đó là APNs device token, không phải FCM token.
- Token FCM báo unregistered khi gửi được tự tắt.
- Khi khởi động, các dòng cũ chưa có `token_hash` được backfill: dòng mới nhất của mỗi token được giữ, dòng trùng cũ hơn và dòng token rỗng bị tắt.

Phía app:

| | Android | iOS |
|---|---|---|
| Quyền | `POST_NOTIFICATIONS` (Android 13+) | `requestAuthorization` rồi `registerForRemoteNotifications()` |
| Lấy token | `FirebaseMessaging.getInstance().token` | `Messaging.messaging().token` (FCM token, không phải APNs token) |
| Token đổi | `onNewToken()` gọi lại `POST /devices` | `messaging(_:didReceiveRegistrationToken:)` gọi lại `POST /devices` |
| Đăng xuất | `POST /devices/unregister` rồi `FirebaseMessaging.getInstance().deleteToken()` | `POST /devices/unregister` rồi `Messaging.messaging().deleteToken` |
| Firebase Console | | upload APNs Authentication Key (.p8), nếu không iOS sẽ không nhận được |

Flutter (`firebase_messaging`): `requestPermission()`, `getToken()`, `onTokenRefresh`, `deleteToken()`.

Push gửi kèm cấu hình theo nền tảng: Android `priority: high` + âm thanh mặc định, iOS `apns-priority: 10` + `sound: default`.

## Audit nội dung gửi cho khách hàng

Mọi lần gửi (push, email, Slack, campaign, FCM topic) đều được ghi lại, kể cả lần gửi thất bại và từng lần retry. Audit chỉ ghi thêm, không có API sửa hoặc xóa.

- `notification_audit_content`: nội dung **đúng như đã gửi** (email lưu subject và HTML đã render từ template, push lưu title/body/data). Mỗi nội dung lưu một lần theo hash, nên 1 triệu người nhận của một campaign cùng trỏ tới một bản.
- `notification_audit_delivery`: một dòng cho mỗi người nhận mỗi lần thử, gồm: nguồn (`notification` hoặc `campaign` + id, batch), lần thử, kênh, `user_id`, người nhận (email / `topic:<tên>` / `slack`), thiết bị (id, `ios`/`android`, **hash** của FCM token, không lưu token gốc), trạng thái `sent` / `failed` / `invalid_token`, mã message FCM, lỗi, và `requested_by`.
- `requested_by` lấy từ header **`X-Requested-By`** mà service gọi gửi kèm (ví dụ `order-service`, `marketing`); không gửi thì là `unknown`.

```
GET /api/v1/audit/deliveries?user_id=&email=&phone=&channel=&status=&source_type=&source_id=&requested_by=&from=&to=&limit=&before_id=
GET /api/v1/audit/contents/:id
```

- Tìm theo `email`, `phone` hoặc `user_id` trả về mọi lần gửi tới user đó: push tới thiết bị của user và email tới địa chỉ liên hệ của user.
- `from`, `to` theo RFC3339. Kết quả mới nhất trước; trang sau dùng `before_id = next_before_id`. `limit` tối đa 500.
- Response gồm `items` và `contents` (các nội dung được tham chiếu, theo `content_id`).
- `audit.retention_days` (mặc định `0` = giữ vĩnh viễn): nếu đặt, job chạy mỗi giờ xóa các dòng giao cũ hơn số ngày đó; nội dung được giữ lại.
- Nếu ghi audit lỗi (DB tạm lỗi), lỗi được log nhưng không làm hỏng lần gửi, để tránh gửi lặp cho khách.

## Retry và dead-letter

Consumer queue `notifications` chạy `worker.concurrency` goroutine với prefetch `worker.prefetch`, tự kết nối lại khi RabbitMQ khởi động lại.

| Kết quả gửi | Xử lý |
|---|---|
| Thành công | status `sent`, ack |
| Lỗi vĩnh viễn: user không còn thiết bị active, template sai, kênh chưa cấu hình, Resend trả 4xx (trừ 429) | status `failed` + `last_error`, ack, không retry |
| Lỗi tạm thời: timeout, FCM unavailable/quota, Resend 429/5xx | status `retrying`, đưa vào `notifications.retry.N` (chờ 5s, 30s, 2m, 10m rồi tự quay lại queue chính) |
| Vẫn lỗi sau 5 lần | status `failed`, message chuyển sang `notifications.dlq` kèm header `x-failure-reason` |

Push gửi tới mọi thiết bị của user bằng một lần gọi `SendEachForMulticast` (Firebase Admin SDK v4). Token FCM báo unregistered được tự động đánh dấu `is_active = false`.

## Campaign: gửi hàng loạt

Gửi cùng một nội dung cho số lượng lớn người nhận (đến `campaign.max_recipients`, mặc định 1.000.000) mà không tạo từng notification.

```
POST /api/v1/campaigns
GET  /api/v1/campaigns?limit=&offset=
GET  /api/v1/campaigns/:id
POST /api/v1/campaigns/:id/cancel
```

| `channel` | `audience` | Người nhận |
|---|---|---|
| `push` | `all_devices` | mọi thiết bị đang active |
| `push` | `users` | thiết bị active của các `user_ids` |
| `push` | `topic` | FCM topic `topic` (1 lần gọi, Firebase tự phát tới app đã subscribe, không theo dõi được từng người) |
| `email` | `emails` | danh sách `emails`, nội dung là `title` + `message` (HTML) hoặc `template_type` + `template_data` |

```json
{
  "channel": "push",
  "audience": "all_devices",
  "title": "Flash sale 9.9",
  "message": "Giảm 50% toàn bộ đơn hàng",
  "data": {"screen": "promo", "promo_id": "99"}
}
```

Response `202 Accepted` ngay sau khi lưu campaign (và danh sách người nhận nếu có). `GET /campaigns/:id` trả về trạng thái và tiến độ:

```json
{
  "campaign": {"id": 12, "status": "SENDING", "batches_planned": 2000, "...": "..."},
  "progress": {"batches_queued": 120, "batches_sent": 1880, "targets": 1000000, "sent": 938512, "failed": 211, "invalid_tokens": 1277}
}
```

Trạng thái campaign: `PENDING` → `PLANNING` → `SENDING` → `COMPLETED`, hoặc `CANCELLED` / `FAILED`.

Cách chạy:

1. Planner (trong từng instance, có lease trên DB nên nhiều replica không chạy trùng) duyệt người nhận theo keyset `id > cursor LIMIT 500` và tạo batch: tối đa 500 thiết bị cho push (giới hạn của FCM) hoặc 100 email (giới hạn Resend batch). Mỗi batch lưu khoảng id `(after_id, until_id]`, cursor được lưu cùng transaction nên crash giữa chừng sẽ chạy tiếp từ chỗ dừng.
2. Mỗi batch là một message trên queue `notification.campaign.batch`, xử lý song song bởi `campaign.concurrency` worker. Worker claim batch bằng lease trên DB, nên một batch bị publish hai lần vẫn chỉ gửi một lần.
3. Push dùng `SendEachForMulticast`; token unregistered bị vô hiệu hóa. Email dùng `POST https://api.resend.com/emails/batch` với `Idempotency-Key: campaign-<id>-batch-<batch_id>`, nên retry trong 24h không gửi trùng.
4. Batch lỗi tạm thời được retry qua `notification.campaign.batch.retry.N` (10s, 1m, 5m, 15m), quá `campaign.max_attempts` thì đánh dấu `FAILED`. Batch bị mất message (publish lỗi, worker chết) được sweeper publish lại sau 10 phút.
5. `campaign.rate_per_second` (0 = không giới hạn) giới hạn số người nhận mỗi giây trên từng instance, dùng khi cần tránh vượt quota FCM/Resend.
6. Cancel: batch chưa gửi chuyển `CANCELLED`, batch đang gửi dở vẫn hoàn tất.

Giới hạn: gửi lại một batch sau khi worker chết giữa lúc gọi FCM có thể làm một số thiết bị trong batch đó nhận 2 lần (FCM không có idempotency key). Tạo campaign `users`/`emails` với 1 triệu người nhận chèn danh sách vào DB đồng bộ trong request (từng khối 5.000 dòng), nên request có thể mất vài chục giây.
