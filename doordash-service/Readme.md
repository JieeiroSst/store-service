# Doordash Service

Dịch vụ điều phối đơn hàng giao đồ ăn kiểu DoorDash: đặt đơn, theo dõi hành trình giao hàng (tracking), và gán/điều phối tài xế. Service **không** sở hữu dữ liệu khách hàng/tài xế, thực đơn nhà hàng, thanh toán hay thông báo — những dữ liệu đó thuộc về các service khác trong hệ thống và được gọi qua HTTP mỗi khi cần (xem mục 4).

## 1. Yêu cầu chức năng (Functional Requirements)

- **Tạo đơn hàng**: xác thực nhà hàng đang hoạt động, định giá từng món theo đúng dữ liệu thực đơn lấy từ restaurant service tại thời điểm đặt (không tin giá client gửi lên), giữ tiền (authorize) qua payment service cho tổng đơn, rồi mới lưu đơn.
- **Vòng đời trạng thái đơn** (`OrderStatus`): `created → confirmed → preparing → ready_for_pickup → picked_up → delivered`, có thể `cancelled` ở bất kỳ bước nào trước khi giao xong. Mọi bước chuyển trạng thái đều được kiểm tra hợp lệ (không cho nhảy cóc, không sửa đơn đã ở trạng thái cuối).
- **Tracking**: mỗi lần đổi trạng thái tự động ghi một checkpoint (`order_tracking`); có thể ghi thêm checkpoint thủ công (vd. vị trí GPS tài xế) không làm đổi trạng thái đơn.
- **Gán tài xế**: đề nghị (`driver_assignments`, trạng thái `pending`) tới một tài xế đang hoạt động; tài xế `accept` (lúc này đơn mới thực sự được gán `driver_id`) hoặc `reject` (kèm lý do, đơn vẫn chưa có tài xế); `complete` khi đã giao xong chặng của tài xế đó.
- **Thông báo khách hàng**: mỗi lần tạo đơn, đổi trạng thái, hoặc huỷ đơn đều gửi thông báo qua notification service — best-effort, lỗi gửi thông báo không làm hỏng thao tác nghiệp vụ chính.

## 2. Yêu cầu phi chức năng (Non-functional Requirements)

- **Không tin dữ liệu từ client**: giá món ăn luôn lấy lại từ restaurant service khi tạo đơn, không dùng giá do client truyền lên — tránh việc client tự sửa giá.
- **Toàn vẹn trạng thái**: chuyển trạng thái đơn/assignment đi qua đúng state machine (`model.OrderStatus.CanTransitionTo`, `AssignmentStatus`), không cho phép cập nhật tuỳ ý.
- **Chịu lỗi của service phụ thuộc**: lỗi từ notification service không rollback đơn hàng đã tạo/cập nhật; lỗi từ payment service (authorize thất bại) thì chặn việc tạo đơn ngay từ đầu, trước khi ghi gì vào DB.
- **Tách biệt dữ liệu (bounded context)**: doordash-service chỉ lưu bảng thuộc domain đặt đơn/giao hàng của chính nó — xem `database.sql`.

## 3. Domain Model (Thực thể chính)

| Entity | Mô tả |
|---|---|
| `Order` | Đơn hàng: khách, nhà hàng, tài xế (nếu có), địa chỉ giao, trạng thái, các khoản phí, tổng tiền, trạng thái thanh toán |
| `OrderItem` | Một dòng món trong đơn — snapshot tên/giá tại thời điểm đặt, không tham chiếu sống tới thực đơn |
| `OrderTracking` | Một checkpoint hành trình đơn: trạng thái, toạ độ (tuỳ chọn), thời điểm, ghi chú |
| `DriverAssignment` | Một lần đề nghị chặng giao hàng cho một tài xế: `pending → accepted/rejected`, rồi `accepted → completed` |

Các thực thể **không** thuộc doordash-service (chỉ là DTO đọc, lấy qua HTTP, không lưu DB): `Customer`, `Driver` (từ user_service), `Restaurant`, `MenuItem` (từ restaurant service), `PaymentAuthorization` (từ payment service) — xem `internal/domain/model/external.go`.

## 4. Kiến trúc tổng quan

Service được cài đặt theo **Hexagonal Architecture** (Ports & Adapters), dùng [`uber-go/fx`](https://github.com/uber-go/fx) để dependency injection — không có `main.go` nào tự tay khởi tạo/nối dây từng struct, toàn bộ được khai báo (declarative) qua các `fx.Module`, cùng khuôn mẫu với `coupon-service`, `threads-service`, `parking-lot-service` trong repo này.

```
cmd/main.go                        → fx.New(infrastructure.Module).Run()

internal/
├── domain/
│   ├── model/                     → Order, OrderItem, OrderTracking, DriverAssignment (thực thể sở hữu)
│   │                                 + Customer, Driver, Restaurant, MenuItem, PaymentAuthorization (DTO đọc, không sở hữu)
│   └── port/
│       ├── driving.go             → OrderUsecase, TrackingUsecase, DriverAssignmentUsecase (port vào)
│       └── driven.go              → OrderRepository, OrderTrackingRepository, DriverAssignmentRepository (port ra tới DB)
│                                     + UserClient, RestaurantClient, PaymentClient, NotifierClient (port ra tới service khác)
├── application/                   → cài đặt driving port: định giá đơn, kiểm tra chuyển trạng thái, điều phối tài xế
├── adapter/
│   ├── primary/http/              → adapter vào: Gin handler/router, gọi driving port
│   └── secondary/
│       ├── repository/            → adapter ra: GORM/Postgres, cài đặt OrderRepository/OrderTrackingRepository/DriverAssignmentRepository
│       ├── userclient/            → adapter ra: gọi HTTP tới user_service, cài đặt UserClient
│       ├── restaurantclient/      → adapter ra: gọi HTTP tới restaurant service, cài đặt RestaurantClient
│       ├── paymentclient/         → adapter ra: gọi HTTP tới payment service, cài đặt PaymentClient
│       └── notifierclient/        → adapter ra: gọi HTTP tới notification service, cài đặt NotifierClient
└── infrastructure/                → wiring: config, logger, database, http server, fx.Module tổng
```

- **domain** là lõi, không import Gin/GORM/fx — chỉ chứa entity và interface (port).
- **application** chỉ phụ thuộc `domain/port`, không biết HTTP hay SQL cụ thể là gì, và không biết `restaurantclient`/`userclient`/... là gọi HTTP — chỉ thấy interface `RestaurantClient`, `UserClient`, `PaymentClient`, `NotifierClient`. Nhờ vậy có thể thay adapter (vd. đổi HTTP client sang gRPC client, hoặc mock trong test — xem `internal/application/fakes_test.go`) mà không đụng vào business logic.
- **fx** khởi tạo theo đúng thứ tự phụ thuộc (`config → database → repository/clients → application → http handler → http server`), xem `internal/infrastructure/module.go`.

## 5. Gọi service khác để lấy dữ liệu

doordash-service chủ động gọi 4 service khác qua HTTP để lấy dữ liệu nó không sở hữu, thay vì join chéo database (mỗi service một database riêng):

| Client (`internal/adapter/secondary/...`) | Port (`domain/port.*Client`) | Gọi tới | Dùng khi |
|---|---|---|---|
| `userclient` | `UserClient` | `user_service` (`USER_SERVICE_BASE_URL`) | Xác thực khách hàng tồn tại lúc tạo đơn (`GetCustomer`); kiểm tra tài xế đang hoạt động lúc gán đơn (`GetDriver`) |
| `restaurantclient` | `RestaurantClient` | restaurant/menu catalog service (`RESTAURANT_SERVICE_BASE_URL`) | Kiểm tra nhà hàng đang mở (`GetRestaurant`); lấy giá/tên món thật để định giá đơn, không tin giá client gửi (`GetMenuItems`, gọi song song từng món giống `threads-service`'s `userclient.GetUsers`) |
| `paymentclient` | `PaymentClient` | payment service (`PAYMENT_SERVICE_BASE_URL`) | Giữ tiền (`AuthorizePayment`) cho tổng đơn trước khi lưu đơn |
| `notifierclient` | `NotifierClient` | notification service (`NOTIFICATION_SERVICE_BASE_URL`) | Báo khách khi đơn được tạo/đổi trạng thái/huỷ — best-effort, lỗi chỉ log |

Mỗi client có timeout riêng (`*_TIMEOUT`, mặc định `5s`) và base URL riêng, cấu hình qua env hoặc Consul (`config.ExternalServiceConfig`, xem mục 8). Base URL mặc định trỏ tới tên Service Kubernetes theo quy ước đã dùng trong repo (vd. `http://user-service-svc`, giống `threads-service`) — cần override bằng giá trị thật của từng service khi triển khai.

## 6. Luồng nghiệp vụ chính

### Tạo đơn (`OrderUsecase.CreateOrder`, `internal/application/order.go`)

1. Danh sách món không rỗng, ngược lại trả `ErrEmptyOrder`.
2. Lấy nhà hàng qua `RestaurantClient.GetRestaurant`; phải `IsActive`, ngược lại `ErrRestaurantInactive`.
3. Lấy toàn bộ món qua `RestaurantClient.GetMenuItems` (một lần gọi, song song bên trong); món nào thiếu hoặc không `IsActive` → `ErrMenuItemUnavailable`.
4. Ghép từng dòng `OrderItem` từ dữ liệu thực đơn thật (tên, giá) — **không** dùng giá do client gửi.
5. Xác thực khách hàng tồn tại qua `UserClient.GetCustomer`.
6. Tính `Subtotal`/`TotalAmount` (`Order.Recalculate`).
7. Giữ tiền qua `PaymentClient.AuthorizePayment`; thất bại hoặc trạng thái khác `authorized` → `ErrPaymentAuthorizationFailed`, **chưa ghi gì vào DB**.
8. Lưu đơn (`OrderRepository.Create`, đơn + items trong một lần ghi).
9. Ghi checkpoint tracking đầu tiên (`created`) — lỗi chỉ log, không chặn việc trả đơn đã tạo.
10. Gửi thông báo khách hàng — best-effort.

### Đổi trạng thái (`OrderUsecase.UpdateOrderStatus`)

Kiểm tra `order.Status.CanTransitionTo(next)` (state machine ở `model/order.go`) trước khi cập nhật; nếu chuyển sang `delivered` thì ghi thêm `ActualDeliveryTime`. Mỗi lần đổi trạng thái đều ghi một checkpoint tracking và gửi thông báo best-effort.

### Huỷ đơn (`OrderUsecase.CancelOrder`)

Từ chối nếu đơn đã ở trạng thái cuối (`ErrOrderAlreadyTerminal`), ngược lại chuyển sang `cancelled` qua cùng state machine, ghi tracking kèm lý do huỷ.

### Gán tài xế (`DriverAssignmentUsecase`, `internal/application/assignment.go`)

- `AssignDriver`: kiểm tra tài xế `IsActive` qua `UserClient.GetDriver`, tạo `DriverAssignment` trạng thái `pending`. **Chưa** đụng tới `Order.DriverID`.
- `AcceptAssignment`: chỉ chấp nhận được assignment đang `pending` (`ErrAssignmentNotPending` nếu không); khi chấp nhận mới gọi `OrderRepository.AssignDriver` để thực sự gán tài xế vào đơn.
- `RejectAssignment`: ghi lý do từ chối, đơn vẫn chưa có tài xế — dispatch có thể tạo assignment mới cho tài xế khác.
- `CompleteAssignment`: chỉ hoàn tất được assignment đang `accepted`.

## 7. API

| Method | Endpoint | Mô tả |
|---|---|---|
| POST | `/api/v1/orders` | Tạo đơn hàng |
| GET | `/api/v1/orders?customer_id=` hoặc `?restaurant_id=` | Danh sách đơn theo khách hàng hoặc theo nhà hàng |
| GET | `/api/v1/orders/:id` | Chi tiết đơn (kèm `items`) |
| PATCH | `/api/v1/orders/:id/status` | Đổi trạng thái đơn (`{"status": "confirmed"}`) |
| POST | `/api/v1/orders/:id/cancel` | Huỷ đơn (`{"reason": "..."}`, tuỳ chọn) |
| POST | `/api/v1/orders/:id/assign-driver` | Đề nghị chặng giao cho một tài xế (`{"driver_id": "..."}`) |
| GET | `/api/v1/orders/:id/tracking` | Lịch sử checkpoint của đơn |
| POST | `/api/v1/orders/:id/tracking` | Ghi checkpoint thủ công (vd. GPS tài xế) |
| GET | `/api/v1/driver-assignments?driver_id=` | Danh sách đề nghị chặng giao của một tài xế |
| POST | `/api/v1/driver-assignments/:id/accept` | Tài xế chấp nhận |
| POST | `/api/v1/driver-assignments/:id/reject` | Tài xế từ chối (`{"reason": "..."}`) |
| POST | `/api/v1/driver-assignments/:id/complete` | Đánh dấu đã giao xong chặng |
| GET | `/health` | Health check |

Ví dụ tạo đơn:

```json
POST /api/v1/orders
{
  "customer_id": "cust-1",
  "restaurant_id": "rest-1",
  "delivery_address_id": "addr-1",
  "payment_method_id": "pm-1",
  "delivery_fee": 2,
  "service_fee": 1,
  "tax": 1.5,
  "items": [
    { "menu_item_id": "item-1", "quantity": 2 }
  ]
}
```

## 8. Cấu hình

Đọc từ biến môi trường (`config.FromEnv`), có thể override qua Consul (`HostConsul`/`KeyConsul`/`ServiceConsul` trong `.env`, xem `internal/infrastructure/config_provider.go` — không cấu hình Consul thì service tự rơi về đọc env, không lỗi):

| Biến | Mặc định | Ý nghĩa |
|---|---|---|
| `PORT_HTTP_SERVER` | `8086` | Cổng HTTP |
| `POSTGRES_HOST`/`PORT`/`USER`/`PASSWORD`/`DBNAME`/`SSLMODE` | `localhost`/`5432`/`postgres`/``/`doordash`/`disable` | Kết nối Postgres |
| `USER_SERVICE_BASE_URL` / `USER_SERVICE_TIMEOUT` | `http://user-service-svc` / `5s` | user_service |
| `RESTAURANT_SERVICE_BASE_URL` / `RESTAURANT_SERVICE_TIMEOUT` | `http://restaurant-service-svc` / `5s` | Restaurant/menu catalog service |
| `PAYMENT_SERVICE_BASE_URL` / `PAYMENT_SERVICE_TIMEOUT` | `http://payment-service-svc` / `5s` | Payment service |
| `NOTIFICATION_SERVICE_BASE_URL` / `NOTIFICATION_SERVICE_TIMEOUT` | `http://notification-service-svc` / `5s` | Notification service |

## 9. Database Schema

Schema thật (được `internal/infrastructure/database.applySchema` tự áp dụng khi service khởi động) nằm ở [`database.sql`](database.sql) — chỉ 4 bảng thuộc domain đặt đơn/giao hàng của chính service này (`orders`, `order_items`, `order_tracking`, `driver_assignments`); dữ liệu khách hàng, nhà hàng/thực đơn, thanh toán, khuyến mãi và thông báo thuộc về các service khác (mục 5).

## 10. Chạy & Triển khai

### Chạy local

```bash
export POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=postgres \
       POSTGRES_PASSWORD=postgres POSTGRES_DBNAME=doordash
export USER_SERVICE_BASE_URL=http://localhost:8081
export RESTAURANT_SERVICE_BASE_URL=http://localhost:8082
export PAYMENT_SERVICE_BASE_URL=http://localhost:8083
export NOTIFICATION_SERVICE_BASE_URL=http://localhost:8084
make run               # go run ./cmd, mặc định lắng nghe :8086
make test              # go test ./...
```

### Docker

`Dockerfile` build multi-stage (Go 1.23 alpine → alpine runtime), copy kèm `database.sql` để service tự áp dụng schema khi khởi động:

```bash
docker build -t doordash-service .
docker run -p 8086:8086 --env-file .env doordash-service
```

### CI/CD

- [`/.github/workflows/doordash-service-cicd.yml`](../.github/workflows/doordash-service-cicd.yml): build + test → build & push image đa kiến trúc lên `ghcr.io/jieeirosst/doordash-service` → `helm template | kubectl apply` khi push vào `master` (chỉ chạy khi đổi trong `doordash-service/`, `chart/doordash-service/` hoặc chính file workflow).
- [`/.github/workflows/go.yml`](../.github/workflows/go.yml): job `doordash_service` chạy `go build`/`go test` trên mọi push/PR vào `master`, cùng chỗ với các service Go khác trong repo.

### Kubernetes (Helm)

Chart nằm ở [`/chart/doordash-service`](../chart/doordash-service) (Deployment + Service + Secret + HPA + Ingress tùy chọn), theo đúng khuôn mẫu các service khác trong repo (`chart/coupon-service`, `chart/threads-service`, ...) và đã được khai báo làm dependency (`doordashService`) trong chart tổng ở [`/chart/Chart.yaml`](../chart/Chart.yaml).

```bash
helm template ./chart/doordash-service \
  --set doordashService.image.tag=<tag> \
  | kubectl apply -n default -f -
```
