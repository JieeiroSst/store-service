# vending-machine-service

Service quản lý máy bán hàng tự động: máy, danh mục/sản phẩm, tồn kho theo slot, và luồng mua hàng
**session → giữ hàng (reservation) → thanh toán → nhả hàng (dispense)**. Dữ liệu lưu trong **PostgreSQL**.

## Kiến trúc (hexagonal + [fx](https://github.com/uber-go/fx))

```
cmd/main.go                         fx.New(infrastructure.Module)
config/                             đọc env
internal/
  domain/                           entity + business rule thuần (không phụ thuộc DB/HTTP)
  port/
    inbound.go                      use case mà adapter primary gọi vào
    outbound.go                     repository, Transactor, PaymentGateway mà application cần
  application/                      hiện thực use case, chỉ phụ thuộc port
  adapter/
    primary/http/                   REST API (gin)
    primary/worker/                 job định kỳ trả hàng của reservation/session hết hạn
    secondary/postgres/             repository + migration (embed, tự chạy khi khởi động)
    secondary/payment/              PaymentGateway: ví (payment-wallet-service) + máy POS giả lập
    secondary/coupon/               CouponProvider gọi coupon-service
    secondary/notification/         Notifier gửi Slack qua notification-service
    secondary/httpx/                HTTP client JSON dùng chung (timeout, retry, X-Requested-By)
  infrastructure/                   pool Postgres, HTTP server, wiring fx
```

Chiều phụ thuộc luôn hướng vào trong: `adapter → port ← application → domain`. Muốn đổi DB hay cổng
thanh toán chỉ cần viết adapter mới hiện thực port và đổi `fx.Provide` trong `module.go` tương ứng.

## Tích hợp với service khác

| Service | Dùng để | Endpoint gọi tới | Khi để trống URL |
|---|---|---|---|
| payment-wallet-service (`WALLET_SERVICE_URL`) | thanh toán `payment_method: "wallet"` | `POST /api/v1/wallets/:id/withdraw` (`reference_id` = payment ID, retry an toàn), `POST /api/v1/transactions/:id/reverse` khi hoàn tiền | thanh toán ví bị từ chối (402) |
| coupon-service (`COUPON_SERVICE_URL`) | giảm giá bằng `coupon_code` | `POST /api/v1/coupons/validate` trước khi trừ tiền, `POST /api/v1/coupons/apply` với `order_id` = `order_no` sau khi có order | mã coupon bị từ chối (422) |
| notification-service (`NOTIFICATION_SERVICE_URL`) | cảnh báo Slack cho vận hành | `POST /api/v1/notifications/slack` | cảnh báo chỉ ghi log |

Cảnh báo gửi qua **outbox**: mọi thay đổi ghi event vào bảng `events` trong cùng transaction, worker đọc
event chưa gửi (`FOR UPDATE SKIP LOCKED`) mỗi `ALERT_INTERVAL` và gửi các loại cần cảnh báo:
máy đổi trạng thái, slot sắp hết / hết hàng, máy nhả hàng lỗi, hoàn tiền lỗi, áp coupon lỗi.
Notification-service lỗi thì event được gửi lại ở lần sau, không ảnh hưởng giao dịch mua.

Thanh toán `card`, `cash`, `mobile` vẫn qua `payment.Simulated` (duyệt mọi giao dịch). Cần thay bằng
adapter của máy POS thật trước khi dùng production.

Tiền gửi sang coupon-service đổi từ cent sang đơn vị tiền (`150` → `1.5`); sang ví gửi nguyên cent.

## Luồng mua hàng

1. `POST /api/v1/sessions` mở phiên trên máy (máy phải `active`), hết hạn sau `SESSION_TTL`.
2. `POST /api/v1/sessions/:id/reservations` giữ 1 sản phẩm ở slot, trừ tồn kho ngay (khóa dòng
   `FOR UPDATE` nên nhiều khách cùng mua không bán vượt số lượng). Hết hạn sau `RESERVATION_TTL`.
3. `POST /api/v1/reservations/:id/checkout` thanh toán, có thể kèm `coupon_code` và `wallet_id`.
   Cổng thanh toán được gọi **ngoài** transaction DB; unique index trên `payments(reservation_id)` chặn
   trừ tiền 2 lần. Nếu reservation hết hạn trong lúc đang trừ tiền, tiền được hoàn lại. Coupon giảm hết
   giá thì không gọi cổng thanh toán.
4. `POST /api/v1/orders/:id/dispense` máy báo kết quả nhả hàng. `{"dispensed": false}` sẽ hoàn tiền;
   gọi lại được nếu lần hoàn tiền trước lỗi.
5. `POST /api/v1/orders/:id/refund` vận hành hoàn tiền đơn đã hoàn tất (khách khiếu nại).

Reservation/session bỏ dở được worker trả lại kho sau mỗi `SWEEP_INTERVAL`.

## Thanh toán chưa rõ kết quả

Nếu ví timeout / trả 5xx sau 3 lần thử (cùng `reference_id` nên không trừ trùng), service tra
`GET /wallets/:id/transactions/by-reference/:paymentId` của payment-wallet-service. Vẫn chưa chắc thì
payment **giữ `pending`** (không đánh dấu thất bại) và checkout trả `202`:

```json
{"status": "pending", "payment_id": "…", "status_url": "/api/v1/payments/…"}
```

Trong lúc đó unique index chặn mọi checkout lại trên reservation (409), nên khách không thể bị trừ tiền
lần hai. Worker đối soát (`vending.reconcileInterval`) lấy các payment `pending` cũ hơn
`vending.reconcileAfter`, hỏi lại nhà cung cấp:

- có giao dịch → tạo order (hoặc hoàn tiền nếu reservation đã hết hạn);
- không có → đánh dấu `failed`, khách có thể trả lại;
- không hỏi được → để nguyên, thử lại lần sau.

Nếu một lần trừ tiền đến muộn sau khi payment đã bị đánh dấu `failed`, tiền được hoàn lại.
`reconcileAfter` bắt buộc lớn hơn 8 × `upstream.timeout` để không đối soát một checkout còn đang chạy.

## Phân trang

Các API danh sách máy, tồn kho, slot sắp hết, danh mục, sản phẩm dùng cursor:
`?limit=20&cursor=<next_cursor của trang trước>` (`limit` 1–100, mặc định 20).

```json
{"items": [...], "next_cursor": "WyIyMDI2…", "is_last_page": false, "limit": 20}
```

`is_last_page: true` ở trang cuối, khi đó không có `next_cursor`. Cursor là keyset trên khóa sắp xếp
nên không bị trùng/sót khi dữ liệu thay đổi giữa hai lần gọi; cursor không hợp lệ trả 400.

## API

Mọi lỗi trả JSON `{"error": "<mô tả>"}`: 400 dữ liệu sai, 404 không tìm thấy, 409 xung đột trạng thái
(hết hàng, session hết hạn, giá vừa đổi, ...), 402 thanh toán lỗi, 422 coupon không hợp lệ,
502 service phụ thuộc không phản hồi.

| Method | Path | Mô tả |
|---|---|---|
| GET | `/health` | liveness |
| GET | `/health/ready` | readiness (ping Postgres) |
| POST | `/api/v1/machines` | `{location, model, status?}` |
| GET | `/api/v1/machines?status=` | danh sách máy |
| GET | `/api/v1/machines/:id` | chi tiết máy |
| PATCH | `/api/v1/machines/:id` | sửa `{location, model}` |
| PATCH | `/api/v1/machines/:id/status` | `{status}`: `active`, `inactive`, `maintenance`, `out_of_service` |
| POST / GET | `/api/v1/machines/:id/maintenance` | ghi / xem nhật ký bảo trì `{technician_id, maintenance_type, notes}` |
| GET | `/api/v1/machines/:id/events?limit=` | sự kiện của máy |
| GET | `/api/v1/machines/:id/sales?from=&to=` | doanh thu theo sản phẩm (RFC3339 hoặc `YYYY-MM-DD`, mặc định 30 ngày) |
| GET | `/api/v1/machines/:id/inventory` | tồn kho các slot (kèm sản phẩm) |
| PUT | `/api/v1/machines/:id/slots/:slot` | nạp sản phẩm vào slot `{product_id, quantity, max_capacity, low_threshold}` |
| GET | `/api/v1/inventory/low?machine_id=` | slot sắp hết hàng |
| POST | `/api/v1/inventory/:id/restock` | `{quantity}` (không vượt `max_capacity`) |
| POST / GET | `/api/v1/categories` | `{name, description, display_order}` |
| POST | `/api/v1/products` | `{name, price_cents, category_id, description, image_url, barcode, is_active, attributes}` |
| GET | `/api/v1/products?category_id=`, `/api/v1/products/:id` | |
| PATCH | `/api/v1/products/:id` | sửa giá, tên, ẩn/hiện sản phẩm (`is_active`) |
| POST | `/api/v1/sessions` | `{machine_id}` |
| GET | `/api/v1/sessions/:id` | |
| POST | `/api/v1/sessions/:id/end` | kết thúc phiên, trả lại hàng đang giữ |
| GET | `/api/v1/sessions/:id/orders` | các đơn trong phiên |
| POST | `/api/v1/sessions/:id/reservations` | `{slot}` |
| DELETE | `/api/v1/reservations/:id` | hủy giữ hàng |
| POST | `/api/v1/reservations/:id/checkout` | `{payment_method, wallet_id?, coupon_code?}`; `payment_method`: `card`, `cash`, `mobile`, `wallet` |
| GET | `/api/v1/payments/:id` | trạng thái payment và order (nếu đã có) |
| GET | `/api/v1/orders/:id` | |
| POST | `/api/v1/orders/:id/dispense` | `{dispensed: true/false}` |
| POST | `/api/v1/orders/:id/refund` | `{reason}` |

## Cấu hình

Thứ tự ưu tiên: **mặc định < Consul KV < biến môi trường**.

- `CONSUL_ADDR` (vd `http://consul-service`) bật đọc Consul; `CONSUL_KEY` mặc định `vending_machine_service`.
  Key chưa có thì dùng mặc định + env; Consul không phản hồi thì service dừng khởi động (thử lại 10 lần).
- Cấu trúc key xem [consul.json](consul.json) (thời lượng viết dạng chuỗi `"5m"`, `"30s"`).
- Biến môi trường tương ứng: `PORT_HTTP_SERVER`, `POSTGRES_HOST|PORT|USER|PASSWORD|DBNAME|SSLMODE|MAX_CONNS`,
  `SESSION_TTL`, `RESERVATION_TTL`, `SWEEP_INTERVAL`, `ALERT_INTERVAL`, `RECONCILE_INTERVAL`,
  `RECONCILE_AFTER`, `CURRENCY`, `SERVICE_NAME`, `UPSTREAM_TIMEOUT`, `WALLET_SERVICE_URL`,
  `COUPON_SERVICE_URL`, `NOTIFICATION_SERVICE_URL`.

Trên k8s, Job `consul-seed` của chart ghi `files/consul.json` vào Consul **chỉ khi key chưa có** (sửa trong
Consul không bị ghi đè khi upgrade); provisioner của chart postgres ghi user/password DB vào cùng key, và
deployment vẫn inject chúng từ Secret. `chart/vending-machine-service/files/consul.json` là bản sao của
`consul.json`; CI fail nếu hai file lệch nhau.

## Chạy và test

```bash
docker run -d --name pg -e POSTGRES_PASSWORD=pw -e POSTGRES_DB=vending_machine -p 5432:5432 postgres:16-alpine
POSTGRES_PASSWORD=pw go run ./cmd

go test ./...                     # unit test (test Postgres tự skip)
TEST_DATABASE_URL="postgres://postgres:pw@localhost:5432/postgres?sslmode=disable" go test -race ./...
```

## Deploy

- Image: `ghcr.io/jieeirosst/vending-machine-service`, build bởi `.github/workflows/vending-machine-service-cicd.yml`.
- Helm chart: `chart/vending-machine-service`. Database `vending_machine`, role và Secret
  `vending-machine-service-db-credentials` do provisioner của `chart/postgres` tạo (entry trong
  `postgres.serviceDatabases`), nên cần upgrade chart postgres trước lần deploy đầu tiên.
