# Referral Service

Golang service theo kiến trúc **Hexagonal Architecture (Ports & Adapters)** để quản lý hệ thống mời khách hàng qua referral code — từ app iOS/Android đến App Store và Google Play.

## Kiến trúc

```
cmd/server/
└── main.go                    ← fx wiring, entrypoint

internal/
├── config/
│   └── config.go              ← đọc .env → Config struct
│
├── core/                      ← HEXAGON (không phụ thuộc framework nào)
│   ├── domain/
│   │   └── referral.go        ← entities: ReferralLink, Event, Reward, Stats
│   ├── ports/
│   │   └── ports.go           ← interfaces (primary + secondary ports)
│   └── services/
│       └── referral_service.go← business logic thuần túy
│
└── adapters/
    ├── primary/
    │   └── http/
    │       ├── handler.go     ← Gin HTTP handlers (primary adapter)
    │       └── server.go      ← gin.Engine + fx lifecycle
    └── secondary/
        └── dynamodb/
            ├── client.go      ← AWS DynamoDB client
            ├── link_repo.go   ← referral_links table
            └── repos.go       ← events, rewards, stats tables

pkg/logger/
└── logger.go                  ← zap logger provider
```

### Nguyên tắc Hexagonal Architecture

```
[HTTP Handler] → (Primary Port) → [Service] → (Secondary Port) → [DynamoDB Adapter]
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
    dynamo.Module,   // DynamoDB client + 4 repositories
    services.Module, // ports.ReferralService
    http.Module,     // *http.Handler
    http.ServerModule, // gin.Engine + graceful shutdown
)
```

## DynamoDB Tables

| Table | PK | SK | GSI |
|---|---|---|---|
| `referral_links` | `ref_code` | `created_at` | `owner_user_id-index` |
| `referral_events` | `ref_code` | `event_id` | `new_user_id-index` |
| `referral_rewards` | `owner_user_id` | `ref_code` | — |
| `user_referral_stats` | `user_id` | `"STATS"` | — |

## Cài đặt

```bash
# 1. Clone và cài dependencies
go mod download

# 2. Copy và chỉnh .env
cp .env.example .env

# 3. Khởi động local DynamoDB (Docker)
make docker-up

# 4. Chạy service
make run
```

## Biến môi trường

| Biến | Bắt buộc | Mô tả | Ví dụ |
|---|---|---|---|
| `AWS_REGION` | ✅ | AWS region | `ap-southeast-1` |
| `DEEP_LINK_BASE_URL` | ✅ | Base URL của redirect server (không có trailing slash) | `https://ref.yourapp.com` |
| `APP_STORE_URL` | ✅ | App Store link iOS | `https://apps.apple.com/app/id123456789` |
| `PLAY_STORE_URL` | ✅ | Play Store link Android (không có `&referrer`) | `https://play.google.com/store/apps/details?id=com.yourapp` |
| `APP_URL_SCHEME` | — | Deep link scheme để thử mở app đã cài | `yourapp://open` |
| `REFERRAL_TTL_DAYS` | — | Số ngày link còn hiệu lực (mặc định 30) | `30` |
| `MAX_REFERRAL_PER_DAY` | — | Giới hạn link mỗi ngày (mặc định 50) | `50` |

## Redirect Server (tự build, không dùng third-party)

Khi user chia sẻ referral link, link có dạng:

```
https://ref.yourapp.com/r/ABC123
```

Server phát hiện platform qua `User-Agent` rồi serve một trang HTML + JavaScript thực hiện logic sau:

```
User click link
      ↓
  Có APP_URL_SCHEME?
  ├─ Có → thử mở app: yourapp://open?ref=ABC123
  │         ↓ (chờ 1.5s)
  │    App đã cài → mở app ✅
  │    Chưa cài   → fallback store
  └─ Không → redirect thẳng đến store

  iOS    → https://apps.apple.com/app/id<APP_ID>
  Android→ https://play.google.com/store/apps/details?id=com.yourapp&referrer=ref_code%3DABC123
```

> **Android**: `referrer` được append tự động vào `PLAY_STORE_URL`. App đọc `ref_code` qua [Play Install Referrer API](https://developer.android.com/google/play/installreferrer).
>
> **iOS**: ref_code truyền qua URL scheme khi app đã cài. Với cold install, app team tự xử lý (clipboard, device fingerprint…).

## API Endpoints

### Sinh referral link mới
```bash
POST /api/v1/referral/generate
{
  "owner_user_id": "user-123",
  "channel": "copy",        # copy | whatsapp | facebook | instagram
  "platform": "ios"         # ios | android | universal
}
# Response:
# {
#   "ref_code":  "abc-uuid",
#   "deep_link": "https://ref.yourapp.com/r/abc-uuid",  ← URL để chia sẻ
#   "expires_at": "2026-06-30T00:00:00Z"
# }
```

### Lấy thông tin link
```bash
GET /api/v1/referral/link/:ref_code
```

### Danh sách link của user (có pagination)
```bash
GET /api/v1/referral/user/:user_id/links?limit=20&cursor=...
```

### Track sự kiện thủ công
```bash
POST /api/v1/referral/event
{
  "ref_code": "abc-uuid",
  "event_type": "link_clicked",   # link_copied | link_clicked | app_installed | registered
  "platform": "ios",
  "device_id": "device-xyz"
}
```

> Endpoint `/r/:ref_code` tự động track `link_clicked` — không cần gọi thêm.

### Kích hoạt referral từ app (gọi sau install hoặc lần đầu mở)
```bash
POST /api/v1/referral/activate
{
  "ref_code":  "abc-uuid",
  "user_id":   "user-456",   ← ID user mới vừa đăng ký
  "platform":  "ios",
  "device_id": "device-xyz"
}
# Response:
# { "attributed": true, "owner_user_id": "user-123", "reward_type": "credit" }
```

- Link được đánh dấu `used`, không thể dùng lại
- Tự động tạo reward record cho `owner_user_id`
- Nếu `attributed: false` → link đã dùng, hết hạn, hoặc tự mời bản thân

### Kiểm tra trạng thái referral
```bash
GET /api/v1/referral/status?ref_code=abc-uuid
# Response:
# {
#   "ref_code":     "abc-uuid",
#   "status":       "active" | "used" | "expired",
#   "owner_user_id":"user-123",
#   "activated_at": "2026-05-30T10:00:00Z",   ← chỉ có khi đã activated
#   "platform":     "android",
#   "new_user_id":  "user-456"
# }
```

### Xác nhận cài đặt (endpoint cũ — vẫn hoạt động)
```bash
POST /api/v1/referral/confirm-install
{
  "ref_code":    "abc-uuid",
  "new_user_id": "user-456",
  "platform":    "ios",
  "device_id":   "device-xyz"
}
```

### Thống kê của referrer
```bash
GET /api/v1/referral/user/:user_id/stats
```

## Luồng hoàn chỉnh

```
1. User A nhấn "Share" trong app
   → POST /generate → sinh ref_code → trả link chia sẻ:
     https://ref.yourapp.com/r/ABC123

2. User A gửi link cho bạn bè (WhatsApp, copy...)

3. Bạn bè click link
   → GET /r/ABC123
   → Server detect platform, track link_clicked
   → Serve HTML: thử mở app qua deep link → fallback store

4a. Đã cài app → app mở, đọc ref_code từ URL scheme
    → App gọi: POST /activate { ref_code, user_id }

4b. Chưa cài → vào store cài app
    Android: đọc ref_code từ Install Referrer API
    iOS:     đọc từ clipboard hoặc cơ chế app team chọn
    → Lần đầu mở app: POST /activate { ref_code, user_id }

5. Backend lưu activation, tạo reward cho User A

6. User A kiểm tra
   → GET /status?ref_code=ABC123
   → GET /user/user-123/stats
```
