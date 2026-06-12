# Referral Service

Golang service theo kiến trúc **Hexagonal Architecture (Ports & Adapters)** để quản lý hệ thống mời khách hàng qua referral code — từ app iOS/Android đến App Store và Google Play.

## Kiến trúc

```text
cmd/server/
└── main.go                    ← fx wiring, entrypoint

internal/
├── config/
│   └── config.go              ← đọc .env → Config struct
│
├── core/                      ← HEXAGON (không phụ thuộc framework nào)
│   ├── domain/
│   │   ├── referral.go        ← entities: ReferralLink, Event, Reward, Stats, Attribution
│   │   └── errors.go          ← domain errors: ErrNotFound, ErrSelfReferral, ErrAlreadyReferred…
│   ├── ports/
│   │   └── ports.go           ← interfaces (primary + secondary ports)
│   └── services/
│       └── referral_service.go← business logic thuần túy
│
└── adapters/
    ├── primary/
    │   └── http/
    │       ├── handler.go     ← Gin HTTP handlers (primary adapter)
    │       ├── response.go    ← envelope + error codes
    │       └── server.go      ← gin.Engine + fx lifecycle
    └── secondary/
        └── mysql/
            ├── client.go          ← MySQL (sqlx) client + fx Module
            ├── link_repo.go       ← referral_links table
            ├── attribution_repo.go← referral_attributions table
            └── repos.go           ← events, rewards, stats, reward_programs tables

pkg/logger/
└── logger.go                  ← zap logger provider
```

### Nguyên tắc Hexagonal Architecture

```text
[HTTP Handler] → (Primary Port) → [Service] → (Secondary Port) → [MySQL Adapter]
     ↑                               ↑                                    ↑
 primary adapter              business logic only                  secondary adapter
 (calls the port)             (no framework deps)                  (implements the port)
```

- **Core** (`domain` + `ports` + `services`) không import bất kỳ framework nào
- **Adapters** implement các interface trong `ports`
- **fx** wire tất cả lại ở `main.go`

## Dependency Injection với uber/fx

```go
// main.go — toàn bộ wiring ở một chỗ
app := fx.New(
    config.Module,   // *config.Config
    logger.Module,   // *zap.Logger
    mysql.Module,    // MySQL (sqlx) client + 6 repositories
    services.Module, // ports.ReferralService
    http.Module,     // *http.Handler
    http.ServerModule, // gin.Engine + graceful shutdown
)
```

## MySQL Schema

| Table | Primary key | Index | Notes |
| --- | --- | --- | --- |
| `referral_links` | `ref_code` | `(owner_user_id, created_at)` | `ref_code` is globally unique (service retries on collision) |
| `referral_events` | `event_id` | `(ref_code, occurred_at)`, `new_user_id` | `event_id` is a UUID |
| `referral_rewards` | `(owner_user_id, ref_code)` | — | — |
| `user_referral_stats` | `user_id` | — | counters updated via `INSERT ... ON DUPLICATE KEY UPDATE` |
| `reward_programs` | `program_id` | `status` | `tiers` stored as a `JSON` column |
| `referral_attributions` | `(owner_user_id, new_user_id)` | `new_user_id` | index enables "who referred this user?" lookups |

> **`reward_programs`**: only one row should have `status = 'active'` at a time — enforced at the application level (see `SetRewardProgramStatus`), not via a DB constraint.
>
> **`referral_attributions`**: ghi nhận quan hệ owner ↔ new user sau khi new user xác nhận cài đặt app. Index `new_user_id` cho phép tra cứu ngược "ai đã mời user này?".

## MySQL Setup — Migrations

Schema được quản lý bằng [`golang-migrate`](https://github.com/golang-migrate/migrate), file SQL nằm ở `migrations/`.

- **Local (Docker Compose):** `make docker-up` tự khởi động `mysql` rồi chạy migration một lần qua service `migrate` (image `migrate/migrate`) — không cần thao tác gì thêm.
- **Staging/Prod:** cài CLI rồi áp migration thủ công, nhắm vào `MYSQL_*` env vars:

  ```bash
  go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

  make migrate-up    # áp toàn bộ migration còn lại
  make migrate-down  # rollback migration gần nhất
  ```

  Có thể override `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_DATABASE` khi gọi `make`, ví dụ:

  ```bash
  make migrate-up MYSQL_HOST=prod-db.internal MYSQL_USER=app MYSQL_PASSWORD=*** MYSQL_DATABASE=referral_service
  ```

---

## Cài đặt

```bash
# 1. Clone và cài dependencies
go mod download

# 2. Copy và chỉnh .env
cp .env.example .env

# 3. Khởi động local MySQL + áp migrations (Docker)
make docker-up

# 4. Chạy service
make run
```

## Biến môi trường

| Biến | Bắt buộc | Mô tả | Ví dụ |
| --- | --- | --- | --- |
| `APP_URL_SCHEME` | ✅ | Deep link scheme của app — nội dung được encode vào QR code (`{scheme}?ref={ref_code}`) | `yourapp://open` |
| `APP_STORE_URL` | ✅ | App Store link iOS | `https://apps.apple.com/app/id123456789` |
| `PLAY_STORE_URL` | ✅ | Play Store link Android | `https://play.google.com/store/apps/details?id=com.yourapp` |
| `APP_PUBLIC_URL` | — | URL công khai của service (không có trailing slash). Mặc định `http://localhost:{APP_PORT}` | `https://ref.yourapp.com` |
| `REFERRAL_TTL_DAYS` | — | Số ngày link còn hiệu lực (mặc định 30) | `30` |
| `MAX_REFERRAL_PER_DAY` | — | Giới hạn link mỗi ngày (mặc định 50) | `50` |
| `MYSQL_HOST` | — | MySQL host (mặc định `localhost`) | `prod-db.internal` |
| `MYSQL_PORT` | — | MySQL port (mặc định `3306`) | `3306` |
| `MYSQL_DATABASE` | — | Tên database (mặc định `referral_service`) | `referral_service` |
| `MYSQL_MAX_OPEN_CONNS` | — | Số connection tối đa trong pool (mặc định 20) | `20` |
| `MYSQL_MAX_IDLE_CONNS` | — | Số connection idle giữ trong pool (mặc định 10) | `10` |
| `MYSQL_CONN_MAX_LIFETIME` | — | Thời gian sống tối đa của một connection (mặc định `5m`) | `5m` |
| `COM_MYSQL_USERNAME` | — | MySQL user (mặc định `root`) | `root` |
| `COM_MYSQL_PASSWORD` | — | MySQL password (mặc định `root`) | `root` |

## Luồng chia sẻ qua QR Code

Thay vì serve HTML redirect, service sinh ảnh QR code trực tiếp. Mobile app lấy PNG về và hiển thị cho user chia sẻ.

```text
[Mobile] GET /api/v1/referral/qr/{ref_code}
              ↓
      Server validate link
              ↓
  Encode: yourapp://open?ref=ABC123
              ↓
     Trả về ảnh PNG (image/png)
              ↓
   Mobile render QR → user chia sẻ màn hình
              ↓
   Người nhận scan QR → mở thẳng app (nếu đã cài)
```

> Nội dung QR là deep link scheme (`APP_URL_SCHEME?ref={ref_code}`), không qua browser hay HTML trung gian.

## API Endpoints

### Response envelope

Mọi API response đều dùng cùng một cấu trúc envelope:

```jsonc
// Success — single resource
{ "data": { ... } }

// Success — danh sách + pagination
{ "data": [...], "meta": { "count": 2, "next_cursor": "..." } }

// Error
{ "error": { "code": "NOT_FOUND", "message": "link not found" } }
```

**Error codes:**

| Code | HTTP | Ý nghĩa |
| --- | --- | --- |
| `VALIDATION_ERROR` | 400 | Thiếu field bắt buộc hoặc sai kiểu |
| `INVALID_FIELD` | 400 | Giá trị enum không hợp lệ hoặc `APP_URL_SCHEME` chưa cấu hình |
| `SELF_REFERRAL` | 400 | User tự mời bản thân |
| `NOT_FOUND` | 404 | Resource không tồn tại |
| `LINK_NOT_ACTIVE` | 422 | Link đã dùng hoặc hết hạn |
| `INTERNAL_ERROR` | 500 | Lỗi server (message luôn là "internal server error") |

---

### Sinh referral link mới

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/generate \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{
    "owner_user_id": "user-123",
    "channel": "copy",
    "platform": "ios"
  }' | jq .
```

```json
{
  "data": {
    "ref_code":  "ABC12",
    "deep_link": "https://ref.yourapp.com/r/ABC12",
    "expires_at": 1751299200000
  }
}
```

### Lấy QR code chia sẻ

```text
GET /api/v1/referral/qr/:ref_code?size=256
```

| Query param | Bắt buộc | Mô tả |
| --- | --- | --- |
| `size` | — | Kích thước ảnh PNG (px), từ 64 đến 1024. Mặc định 256 |

```bash
# Lấy về file PNG
curl -s "http://localhost:8080/api/v1/referral/qr/ABC12?size=300" \
  -o qr_ABC12.png
```

- **Content-Type:** `image/png`
- **Cache-Control:** `public, max-age=3600`
- QR encode nội dung: `yourapp://open?ref=ABC12`
- Trả về `422 LINK_NOT_ACTIVE` nếu link đã dùng hoặc hết hạn
- Trả về `404 NOT_FOUND` nếu `ref_code` không tồn tại

### Lấy thông tin link

```bash
curl -s http://localhost:8080/api/v1/referral/link/ABC12 \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' | jq .
```

```json
{
  "data": {
    "ref_code": "ABC12",
    "owner_user_id": "user-123",
    "channel": "copy",
    "status": "active",
    "expires_at": 1751299200000,
    "deep_link": "https://ref.yourapp.com/r/ABC12",
    "platform": "ios"
  }
}
```

### Danh sách link của user (có pagination)

```bash
curl -s "http://localhost:8080/api/v1/referral/user/user-123/links?limit=20&cursor=" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' | jq .
```

```json
{
  "data": [ { "ref_code": "ABC12", "..." : "..." } ],
  "meta": { "count": 1, "next_cursor": "eyJ..." }
}
```

### Track sự kiện thủ công

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/event \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{
    "ref_code":   "ABC12",
    "event_type": "link_clicked",
    "platform":   "ios",
    "device_id":  "device-xyz"
  }' | jq .
```

```json
{ "data": { "status": "tracked" } }
```

### Kích hoạt referral từ app (gọi sau install hoặc lần đầu mở)

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/activate \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{
    "ref_code":  "ABC12",
    "user_id":   "user-456",
    "platform":  "ios",
    "device_id": "device-xyz"
  }' | jq .
```

```json
{
  "data": {
    "attributed":    true,
    "owner_user_id": "user-123",
    "reward_type":   "credit"
  }
}
```

- Link được đánh dấu `used`, không thể dùng lại
- Tự động tạo reward record cho `owner_user_id`
- Nếu `attributed: false` → link đã dùng, hết hạn, hoặc tự mời bản thân

### Kiểm tra trạng thái referral

```bash
curl -s "http://localhost:8080/api/v1/referral/status?ref_code=ABC12" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' | jq .
```

```json
{
  "data": {
    "ref_code":          "ABC12",
    "status":            "active",
    "invitation_status": "clicked",
    "owner_user_id":     "user-123",
    "activated_at":      1717200000000,
    "platform":          "android",
    "new_user_id":       "user-456",
    "reward_status":     "pending",
    "reward_value":      50000
  }
}
```

**`invitation_status`** theo dõi hành trình mời khách:

| Giá trị | Ý nghĩa |
| --- | --- |
| `pending` | Link đã tạo, chưa có ai click |
| `clicked` | Khách đã mở link, chưa cài app |
| `installed` | App đã được cài, reward đang chờ xử lý |
| `rewarded` | Reward đã thanh toán (`reward_status = paid`) |
| `expired` | Link hết hạn trước khi được dùng |

### Xác nhận cài đặt (endpoint cũ — vẫn hoạt động)

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/confirm-install \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{
    "ref_code":    "ABC12",
    "new_user_id": "user-456",
    "platform":    "ios",
    "device_id":   "device-xyz"
  }' | jq .
```

```json
{
  "data": { "attributed": true, "owner_user_id": "user-123", "reward_type": "credit" }
}
```

### Thống kê của referrer

```bash
curl -s http://localhost:8080/api/v1/referral/user/user-123/stats \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' | jq .
```

```json
{
  "data": {
    "user_id":          "user-123",
    "total_invited":    5,
    "total_installed":  3,
    "total_rewarded":   3,
    "total_reward_amt": 150000,
    "last_active_at":   1717210000000
  }
}
```

### Quản lý chương trình thưởng

#### Tạo chương trình mới

```bash
curl -s -X POST http://localhost:8080/api/v1/programs \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{
    "name": "Campaign Q3",
    "activate": true,
    "tiers": [
      { "min_count": 1,  "max_count": 9,  "reward_value": 50000  },
      { "min_count": 10, "max_count": 19, "reward_value": 100000 },
      { "min_count": 20, "max_count": -1, "reward_value": 150000 }
    ]
  }' | jq .
```

> `max_count: -1` = không giới hạn. `activate: true` = kích hoạt ngay và tự động tắt chương trình cũ.

#### Xem chương trình đang active

```bash
curl -s http://localhost:8080/api/v1/programs/active \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' | jq .
```

```json
{
  "data": {
    "program_id": "prog-uuid",
    "name":       "Campaign Q3",
    "status":     "active",
    "tiers":      [ { "min_count": 1, "max_count": 9, "reward_value": 50000 } ],
    "created_at": 1717200000000,
    "updated_at": 1717200000000
  }
}
```

#### Bật / tắt chương trình

```bash
curl -s -X PATCH http://localhost:8080/api/v1/programs/prog-uuid/status \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -d '{ "active": true }' | jq .
```

```json
{ "data": { "status": "updated" } }
```

### Request Headers

| Header | Bắt buộc | Mô tả | Ví dụ |
| --- | --- | --- | --- |
| `Content-Type` | ✅ (POST/PATCH) | Định dạng body | `application/json` |
| `X-Device-ID` | ✅ | Định danh thiết bị | `5619d406a7a0108e` |
| `X-Device-Info` | ✅ | Thông tin thiết bị (JSON string) | `{"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}` |
| `X-Amzn-Trace-Id` | — | Trace ID để correlate log. Nếu không gửi, server tự sinh UUID | `Root=1-abc123-xxxx` |
| `X-Request-ID` | — | Request ID từ client (ghi vào log nếu có) | `req-12345` |

## Luồng hoàn chỉnh (curl từng bước)

Ví dụ dưới đây chạy trên `localhost:8080`. Thay `BASE=https://your-domain.com` nếu deploy thật.

### Bước 0 — Tạo chương trình thưởng và kích hoạt

```bash
curl -s -X POST http://localhost:8080/api/v1/programs \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-setup-001" \
  -d '{
    "name": "Campaign Q3 2026",
    "activate": true,
    "tiers": [
      { "min_count": 1,  "max_count": 9,  "reward_value": 50000  },
      { "min_count": 10, "max_count": 19, "reward_value": 100000 },
      { "min_count": 20, "max_count": -1, "reward_value": 150000 }
    ]
  }' | jq .
```

```json
{
  "data": {
    "program_id": "prog-uuid-xxxx",
    "name":       "Campaign Q3 2026",
    "status":     "active",
    "tiers": [
      { "min_count": 1,  "max_count": 9,  "reward_value": 50000  },
      { "min_count": 10, "max_count": 19, "reward_value": 100000 },
      { "min_count": 20, "max_count": -1, "reward_value": 150000 }
    ],
    "created_at": 1717200000000,
    "updated_at": 1717200000000
  }
}
```

> Chương trình này sẽ thưởng **50.000** cho lượt mời thứ 1–9, **100.000** cho lượt 10–19, **150.000** từ lượt 20 trở đi.

---

### Bước 1 — Kiểm tra chương trình đang chạy

```bash
curl -s http://localhost:8080/api/v1/programs/active \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-check-001" | jq .
```

---

### Bước 2 — Owner (user-A) tạo referral link

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/generate \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-gen-001" \
  -d '{
    "owner_user_id": "user-A",
    "channel":       "whatsapp",
    "platform":      "android"
  }' | jq .
```

```json
{
  "data": {
    "ref_code":  "A1B2C",
    "deep_link": "https://ref.yourapp.com/r/A1B2C",
    "expires_at": 1719792000000
  }
}
```

---

### Bước 3 — Lấy QR code để chia sẻ

App dùng `ref_code` vừa nhận để gọi API lấy ảnh QR, sau đó hiển thị cho user-A chia sẻ màn hình.

```bash
curl -s "http://localhost:8080/api/v1/referral/qr/A1B2C?size=300" \
  -H "X-Amzn-Trace-Id: Root=trace-qr-001" \
  -o qr_A1B2C.png
```

> Ảnh PNG trả về encode nội dung `yourapp://open?ref=A1B2C`. Người nhận scan QR → app mở thẳng với ref_code.

---

### Bước 4 — Kiểm tra trạng thái: `invitation_status = clicked`

Sau khi bạn bè scan QR và app track sự kiện:

```bash
curl -s "http://localhost:8080/api/v1/referral/status?ref_code=A1B2C" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-status-001" | jq .
```

```json
{
  "data": {
    "ref_code":          "A1B2C",
    "status":            "active",
    "invitation_status": "clicked",
    "owner_user_id":     "user-A"
  }
}
```

---

### Bước 5 — Khách cài app và kích hoạt referral

App đọc được `ref_code` (từ deep link khi mở app), sau đó gọi:

```bash
curl -s -X POST http://localhost:8080/api/v1/referral/activate \
  -H "Content-Type: application/json" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-activate-001" \
  -d '{
    "ref_code":  "A1B2C",
    "user_id":   "user-B",
    "platform":  "android",
    "device_id": "5619d406a7a0108e"
  }' | jq .
```

```json
{
  "data": {
    "attributed":    true,
    "owner_user_id": "user-A",
    "reward_type":   "credit"
  }
}
```

> `attributed: true` — link hợp lệ, không tự mời, chưa dùng. Reward **50.000** (lượt invite thứ 1 của user-A) được tạo với `status: pending`.

---

### Bước 6 — Kiểm tra trạng thái: `invitation_status = installed`

```bash
curl -s "http://localhost:8080/api/v1/referral/status?ref_code=A1B2C" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-status-002" | jq .
```

```json
{
  "data": {
    "ref_code":          "A1B2C",
    "status":            "used",
    "invitation_status": "installed",
    "owner_user_id":     "user-A",
    "activated_at":      1717210000000,
    "platform":          "android",
    "new_user_id":       "user-B",
    "reward_status":     "pending",
    "reward_value":      50000
  }
}
```

---

### Bước 7 — Kiểm tra thống kê của owner

```bash
curl -s "http://localhost:8080/api/v1/referral/user/user-A/stats" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-stats-001" | jq .
```

```json
{
  "data": {
    "user_id":          "user-A",
    "total_invited":    1,
    "total_installed":  1,
    "total_rewarded":   1,
    "total_reward_amt": 50000,
    "last_active_at":   1717210000000
  }
}
```

---

### Bước 8 — Sau khi reward được thanh toán: `invitation_status = rewarded`

Khi hệ thống backend xử lý và cập nhật `reward_status → paid`, gọi lại:

```bash
curl -s "http://localhost:8080/api/v1/referral/status?ref_code=A1B2C" \
  -H "X-Device-ID: 5619d406a7a0108e" \
  -H 'X-Device-Info: {"device_id":"5619d406a7a0108e","app_version":"1.0.0-uat+1","os":"android","os_version":"16","model":"Galaxy S24 FE"}' \
  -H "X-Amzn-Trace-Id: Root=trace-status-003" | jq .
```

```json
{
  "data": {
    "ref_code":          "A1B2C",
    "status":            "used",
    "invitation_status": "rewarded",
    "owner_user_id":     "user-A",
    "activated_at":      1717210000000,
    "platform":          "android",
    "new_user_id":       "user-B",
    "reward_status":     "paid",
    "reward_value":      50000
  }
}
```

---

### Tóm tắt luồng

```text
[Admin]   POST /programs              → tạo chương trình thưởng theo chặng
[User-A]  POST /referral/generate     → lấy ref_code
[User-A]  GET  /referral/qr/:ref_code → lấy ảnh QR code PNG (encode deep link)
                                        → hiển thị cho bạn bè scan
[User-B]  scan QR                     → mở app trực tiếp qua deep link
[User-B]  POST /referral/activate     → cài xong, kích hoạt referral
                                        → reward pending được tạo (giá trị theo tier)
[System]  reward_status → paid        → invitation_status chuyển thành "rewarded"
[User-A]  GET  /referral/status       → theo dõi từng bước qua invitation_status
[User-A]  GET  /user/:id/stats        → xem tổng số khách và tổng thưởng
```

```
HTTP handler.go:
GET /qr/:ref_code?channel=zalo — QR gắn tag kênh.
Mới GET /api/v1/referral/share/:ref_code — trả share-sheet đa kênh (Zalo/Facebook/WhatsApp/Instagram/Copy), mỗi kênh có share_url + qr_url.
Mới GET /r/:ref_code — landing route mà deep_link trước đây sinh ra nhưng chưa từng được phục vụ; ghi nhận link_clicked (thời điểm mời), rồi: Android → redirect Play Store kèm referrer=ref_code=... (đọc qua Play Install Referrer API); iOS → trang HTML copy ref_code vào clipboard trước khi sang App Store (do Apple không có install-referrer).

Lưu ý quan trọng còn lại (việc của mobile team, ngoài phạm vi repo backend này):

Android cần đọc ref_code/channel qua Play Install Referrer API ở lần mở đầu.
iOS cần đọc clipboard ở lần mở đầu, bóc prefix REF:.
Cả hai nền tảng gọi POST /confirm-install hoặc /activate kèm firebase_instance_id lấy từ Firebase Analytics SDK để đối chiếu chéo về sau.
```