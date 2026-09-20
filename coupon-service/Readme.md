# Coupon Service

Hệ thống quản lý và áp dụng mã giảm giá (coupon) cho đơn hàng: nhiều loại giảm giá, giới hạn số lần dùng, ràng buộc theo danh mục/sản phẩm, và coupon phát riêng cho từng khách hàng.

## 1. Yêu cầu chức năng (Functional Requirements)

- Nhiều **loại coupon** (`type`): `percentage` (giảm theo %), `fixed_amount` (giảm số tiền cố định), `buy_x_get_y` (mua X tặng Y — quy về một giá trị giảm cố định tương đương, xem mục 4).
- **Hiệu lực theo thời gian**: `start_date` → `end_date`.
- **Giới hạn số lần dùng**: tổng số lần (`max_uses`/`current_uses`, `NULL` = không giới hạn) và theo từng đơn hàng (một coupon chỉ áp dụng được một lần cho một `order_id`).
- **Điều kiện đơn hàng tối thiểu** (`minimum_purchase`) và **mức giảm tối đa** (`max_discount_amount`, `NULL` = không giới hạn).
- **Ràng buộc theo danh mục/sản phẩm** (`coupon_restrictions`): danh sách bao gồm (chỉ áp dụng cho các danh mục/sản phẩm này) hoặc loại trừ (không bao giờ áp dụng).
- **Coupon phát riêng cho người dùng** (`user_coupons`): một coupon không có dòng nào trong bảng này là coupon công khai; nếu có, chỉ những user được gán (và chưa dùng) mới được áp dụng.
- **Xác thực (validate) coupon** mà không ghi nhận giao dịch — dùng để hiển thị mức giảm trước khi khách xác nhận đơn.
- **Áp dụng (apply) coupon**: xác thực + ghi nhận lượt dùng (`coupon_usage`) + tăng `current_uses` + đánh dấu `user_coupons.is_used` (nếu là coupon riêng) — toàn bộ có bảo vệ chống race condition và chống áp dụng trùng trên cùng một đơn hàng.
- Lịch sử redemption (`coupon_usage`) phục vụ tra cứu, đối soát theo coupon hoặc theo user.

## 2. Yêu cầu phi chức năng (Non-functional Requirements)

- **Concurrency**: hai request áp dụng cùng một coupon đang gần chạm `max_uses` không được cùng vượt giới hạn.
- **Idempotency**: áp dụng lại coupon cho cùng một `order_id` phải bị từ chối, không được trừ thêm lượt dùng.
- **Consistency**: `current_uses`, `coupon_usage` và `user_coupons.is_used` phải nhất quán ngay sau một lần `apply` thành công.
- **Extensibility**: thêm loại coupon mới, loại ràng buộc mới mà không phá vỡ logic hiện có.

## 3. Domain Model (Thực thể chính)

| Entity | Mô tả |
|---|---|
| `Coupon` | Định nghĩa coupon: mã, loại, giá trị giảm, điều kiện, hiệu lực, số lần dùng |
| `CouponRestriction` | Ràng buộc coupon theo `category`/`product`/`user_group`, bao gồm hoặc loại trừ |
| `CouponUsage` | Một lần redemption cụ thể: coupon nào, user nào, order nào, giảm bao nhiêu |
| `UserCoupon` | Gán một coupon cho một user cụ thể (coupon riêng), theo dõi đã dùng hay chưa |

### Loại coupon (CouponType)

```
percentage, fixed_amount, buy_x_get_y
```

### Loại ràng buộc (RestrictionType)

```
category, product, user_group   // user_group hiện chưa có nguồn dữ liệu nhóm user, xem mục 8
```

## 4. Kiến trúc tổng quan

Service được cài đặt theo **Hexagonal Architecture** (Ports & Adapters), dùng [`uber-go/fx`](https://github.com/uber-go/fx) để dependency injection — không có `main.go` nào tự tay khởi tạo/nối dây từng struct, toàn bộ được khai báo (declarative) qua các `fx.Module`. Đây cũng là kiến trúc dùng chung với `parking-lot-service` trong repo này.

```
cmd/main.go                        → fx.New(infrastructure.Module).Run()

internal/
├── domain/
│   ├── model/                     → Coupon, CouponRestriction, CouponUsage, UserCoupon (không phụ thuộc framework)
│   └── port/
│       ├── driving.go             → CouponUsecase, CouponRestrictionUsecase, CouponUsageUsecase,
│       │                             UserCouponUsecase (port vào - cổng mà adapter primary gọi)
│       └── driven.go              → CouponRepository, CouponRestrictionRepository, CouponUsageRepository,
│                                     UserCouponRepository (port ra - cổng mà application gọi)
├── application/                   → cài đặt các driving port (business logic: validate/apply coupon,
│                                     kiểm tra ràng buộc, tính giảm giá, khử coupon hết hạn/hết lượt)
├── adapter/
│   ├── primary/http/              → adapter vào: Gin handler/router, gọi driving port
│   ├── primary/cron/               → adapter vào: robfig/cron scheduler, gọi driving port theo lịch
│   └── secondary/repository/      → adapter ra: GORM/Postgres, cài đặt driven port
└── infrastructure/                → wiring: config, logger, database, http server, cron, fx.Module tổng
```

- **domain** là lõi, không import Gin/GORM/fx — chỉ chứa entity và interface (port).
- **application** chỉ phụ thuộc `domain/port`, không biết HTTP hay SQL cụ thể là gì. `couponService` (trong `application/coupon.go`) là nơi duy nhất phối hợp cả 4 repository (coupon, restriction, usage, user-coupon) để thực hiện `ValidateCoupon`/`ApplyCoupon` — logic nghiệp vụ trung tâm của cả service.
- **adapter/primary** có hai adapter độc lập cùng gọi vào driving port: HTTP (request đến) và cron (lịch chạy đến) — cả hai chỉ là "cái gì kích hoạt usecase", không chứa logic nghiệp vụ. **adapter/secondary** (repository) có thể thay thế độc lập (vd. đổi Postgres sang MySQL) mà không đụng vào `application`/`domain`.
- **fx** chịu trách nhiệm khởi tạo theo đúng thứ tự phụ thuộc (`config → database → repository → application → http handler → http server / cron scheduler`), xem `internal/infrastructure/module.go`.

> **Ghi chú migration**: bản trước dùng gRPC + grpc-gateway + Cobra CLI (`api`/`cron`/`consumer`/`subscriber`) với Redis cache và NATS consumer/subscriber, nhưng phần lớn logic nghiệp vụ (restriction, usage, user-coupon, và đặc biệt là việc *áp dụng* coupon vào đơn hàng) chưa từng được cài đặt — các RPC tương ứng chỉ `return nil, nil`; cron cũ cũng chỉ có job rỗng (`AddFunc("@every 1m", func() {})`). Bản hexagonal này bỏ gRPC/Cobra/Redis/NATS (không dùng thật, chỉ tăng bề mặt phải bảo trì) để đổi lấy một luồng nghiệp vụ đầy đủ, và cron được làm lại thành một adapter thật sự (mục 5.3) thay vì job rỗng. Có thể mở rộng lại các adapter đã bỏ (gRPC, NATS) sau này mà không đụng vào `domain`/`application` nếu cần.

## 5. Luồng nghiệp vụ chính

### Validate coupon (không ghi nhận giao dịch)

`CouponUsecase.ValidateCoupon` (`internal/application/coupon.go`) kiểm tra theo thứ tự:

1. Coupon tồn tại theo `code` và `IsActive = true`.
2. Thời điểm hiện tại nằm trong `[start_date, end_date]`.
3. `purchase_amount >= minimum_purchase`.
4. Coupon còn lượt dùng (`current_uses < max_uses`, hoặc `max_uses = NULL`).
5. Không vi phạm `coupon_restrictions` theo `category_ids`/`product_ids` của giỏ hàng (loại trừ nào khớp thì từ chối ngay; nếu có danh sách bao gồm thì giỏ hàng phải khớp ít nhất một mục).
6. Nếu coupon có bất kỳ dòng `user_coupons` nào (tức là coupon riêng), user gọi phải có một dòng `user_coupons` chưa dùng (`is_used = false`); coupon không có dòng nào là coupon công khai.
7. Trả về coupon + số tiền giảm (`Coupon.CalculateDiscount`), không ghi gì vào DB.

### Apply coupon (ghi nhận redemption)

`CouponUsecase.ApplyCoupon` gọi lại toàn bộ `ValidateCoupon` ở trên, sau đó:

1. Kiểm tra `coupon_usage` đã có dòng `(order_id, coupon_id)` chưa — có thì từ chối (`ErrOrderAlreadyUsedCoupon`, chống áp dụng trùng).
2. `CouponRepository.IncrementUsage` — `UPDATE coupons SET current_uses = current_uses + 1 WHERE id = ? AND (max_uses IS NULL OR current_uses < max_uses)`; nếu `RowsAffected = 0` nghĩa là một request khác đã dùng hết lượt trong lúc ta đang xử lý → trả `ErrUsageLimitReached` (xem mục 7 - concurrency).
3. Ghi một dòng `coupon_usage` mới (ràng buộc `UNIQUE(order_id, coupon_id)` ở DB là lớp bảo vệ thứ hai chống trùng lặp).
4. Nếu là coupon riêng, đánh dấu `user_coupons.is_used = true` cho user đó (best-effort: nếu bước này lỗi thì không rollback redemption đã ghi nhận ở bước 3, vì phần tiền giảm đã được xác nhận với khách).

### Cron: tự động khử coupon hết hạn/hết lượt (`internal/adapter/primary/cron`)

`ValidateCoupon`/`ApplyCoupon` đã tự kiểm tra `end_date` và `max_uses` ngay tại thời điểm gọi, nên một coupon hết hạn/hết lượt không bao giờ áp dụng được dù `is_active` vẫn `true`. Cron job dưới đây không thay đổi tính đúng đắn đó — nó chỉ giữ cho **cột `is_active` phản ánh đúng thực tế**, phục vụ các API/màn hình chỉ lọc theo `is_active` (vd. trang quản trị liệt kê "coupon đang hiệu lực") mà không muốn tự tính lại điều kiện ngày/lượt dùng mỗi lần hiển thị:

- Dùng [`github.com/robfig/cron`](https://github.com/robfig/cron) (v1, cùng thư viện bản gRPC cũ từng dùng), chạy theo lịch cấu hình ở `Cron.DeactivateStaleSchedule` (env `CRON_DEACTIVATE_STALE_SCHEDULE`, mặc định `@every 5m`).
- Mỗi lần chạy gọi `CouponUsecase.DeactivateStaleCoupons`, cài đặt ở `application/coupon.go`, việc này ủy quyền cho `CouponRepository.DeactivateStale` — một câu `UPDATE` duy nhất, atomic ở tầng DB:
  ```sql
  UPDATE coupons SET is_active = false
  WHERE is_active = true
    AND (end_date < now() OR (max_uses IS NOT NULL AND current_uses >= max_uses));
  ```
- `internal/adapter/primary/cron/cron.go` đăng ký job qua `fx.Lifecycle` giống hệt cách `internal/infrastructure/server/server.go` khởi động HTTP server (`OnStart` → `scheduler.Start()`, `OnStop` → `scheduler.Stop()`) — cùng một app `fx.New(infrastructure.Module)` khởi động và dừng cả HTTP server lẫn cron scheduler cùng lúc, không cần chạy một binary/`cmd` riêng cho cron.
- Nếu cấu hình lịch (`DeactivateStaleSchedule`) sai cú pháp, `cron.New` trả lỗi ngay khi khởi động (fx sẽ dừng toàn bộ ứng dụng) thay vì âm thầm không bao giờ chạy job.

## 6. API

| Method | Endpoint | Mô tả |
|---|---|---|
| POST | `/api/v1/coupons` | Tạo coupon |
| GET | `/api/v1/coupons` | Danh sách coupon |
| GET | `/api/v1/coupons/:id` | Chi tiết coupon theo id |
| GET | `/api/v1/coupons/code/:code` | Chi tiết coupon theo mã |
| PUT | `/api/v1/coupons/:id` | Cập nhật coupon |
| DELETE | `/api/v1/coupons/:id` | Xoá coupon |
| POST | `/api/v1/coupons/validate` | Xác thực coupon cho một giỏ hàng, trả về mức giảm (không ghi nhận) |
| POST | `/api/v1/coupons/apply` | Áp dụng coupon cho một đơn hàng (ghi nhận redemption) |
| POST | `/api/v1/coupons/:id/restrictions` | Thêm ràng buộc cho coupon |
| GET | `/api/v1/coupons/:id/restrictions` | Danh sách ràng buộc của coupon |
| PUT | `/api/v1/restrictions/:id` | Cập nhật một ràng buộc |
| DELETE | `/api/v1/restrictions/:id` | Xoá một ràng buộc |
| GET | `/api/v1/coupons/:id/usages` | Lịch sử redemption theo coupon |
| GET | `/api/v1/users/:user_id/coupon-usages` | Lịch sử redemption theo user |
| POST | `/api/v1/user-coupons` | Gán coupon cho một user (`user_id`, `coupon_id`) |
| GET | `/api/v1/user-coupons?user_id=` | Danh sách coupon đã gán cho một user |
| GET | `/api/v1/coupons/:id/user-coupons` | Danh sách user đã được gán một coupon |
| POST | `/api/v1/user-coupons/:id/use` | Đánh dấu thủ công một coupon-gán-user là đã dùng |
| POST | `/api/v1/user-coupons/:id/unuse` | Hoàn tác đánh dấu đã dùng |
| DELETE | `/api/v1/user-coupons/:id` | Gỡ gán coupon khỏi user |
| GET | `/health` | Health check |

Ví dụ `validate`/`apply`:

```json
POST /api/v1/coupons/validate
{
  "code": "SAVE10",
  "user_id": 1,
  "purchase_amount": 250000,
  "category_ids": [12],
  "product_ids": []
}
```

```json
POST /api/v1/coupons/apply
{
  "code": "SAVE10",
  "user_id": 1,
  "order_id": 9001,
  "purchase_amount": 250000
}
```

## 7. Database Schema

Schema thật (được `internal/infrastructure/database.applySchema` tự áp dụng khi service khởi động) nằm ở [`database.sql`](database.sql):

```sql
CREATE TABLE coupons (
    id BIGSERIAL PRIMARY KEY,
    code VARCHAR(50) UNIQUE NOT NULL,
    type VARCHAR(20) NOT NULL,
    discount_value DECIMAL(10,2) NOT NULL,
    minimum_purchase DECIMAL(10,2) NOT NULL DEFAULT 0,
    max_discount_amount DECIMAL(10,2),
    description TEXT,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    max_uses INTEGER,
    current_uses INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT valid_dates CHECK (end_date > start_date),
    CONSTRAINT valid_discount CHECK (discount_value > 0),
    CONSTRAINT valid_minimum CHECK (minimum_purchase >= 0)
);

CREATE TABLE coupon_restrictions (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    restriction_type VARCHAR(20) NOT NULL,
    restricted_entity_id BIGINT NOT NULL,
    is_exclude BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(coupon_id, restriction_type, restricted_entity_id)
);

CREATE TABLE coupon_usage (
    id BIGSERIAL PRIMARY KEY,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    user_id BIGINT NOT NULL,
    order_id BIGINT NOT NULL,
    discount_amount DECIMAL(10,2) NOT NULL,
    used_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(order_id, coupon_id)
);

CREATE TABLE user_coupons (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    coupon_id BIGINT NOT NULL REFERENCES coupons(id),
    is_used BOOLEAN NOT NULL DEFAULT false,
    assigned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at TIMESTAMP,
    UNIQUE(user_id, coupon_id)
);
```

## 8. Vấn đề đồng thời (Concurrency)

`CouponRepository.IncrementUsage` (`internal/adapter/secondary/repository/coupon.go`) dùng một câu `UPDATE ... WHERE max_uses IS NULL OR current_uses < max_uses` nguyên tử thay vì "đọc rồi ghi" ở tầng ứng dụng — hai request `apply` chạy song song khi coupon chỉ còn 1 lượt sẽ không bao giờ cùng thành công: request thua sẽ nhận `RowsAffected = 0` và bị từ chối với `ErrUsageLimitReached`. Ràng buộc `UNIQUE(order_id, coupon_id)` trên `coupon_usage` là lớp bảo vệ thứ hai ở tầng DB chống việc cùng một đơn hàng bị trừ giảm giá hai lần dù tầng ứng dụng đã tự kiểm tra trước.

## 9. Hướng mở rộng

- Ràng buộc theo `user_group` (`RestrictionType = user_group`) hiện được định nghĩa trong domain nhưng chưa có nguồn dữ liệu nhóm user để đối chiếu — cần tích hợp với `user_service` khi cần.
- `buy_x_get_y` hiện được tính như một khoản giảm cố định (`DiscountValue`); để tính đúng theo số lượng sản phẩm tặng kèm cần dữ liệu dòng đơn hàng, thứ mà `coupon-service` không sở hữu — nên để `basket-service`/`order-processing-service` tính trước rồi truyền `discount_value` tương ứng, hoặc mở rộng `ValidateCouponInput` với line items.
- API tạo hàng loạt (`bulk create`) coupon riêng cho một danh sách user (hiện phải gọi `POST /user-coupons` từng người).
- Thêm job cron thứ hai để tự động **bật lại** coupon khi `start_date` tới (hiện `is_active` chỉ tắt tự động khi hết hạn/hết lượt ở mục 5.3, việc bật lại phải làm thủ công qua `PUT /coupons/:id` vì một coupon bị tắt thủ công không nên tự động bật lại).
- Khôi phục các adapter đã bỏ (gRPC, NATS consumer/subscriber) như các module `fx` riêng nếu có nhu cầu tích hợp thật, không đụng vào `domain`/`application`.

## 10. Chạy & Triển khai

### Chạy local

```bash
export POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=postgres \
       POSTGRES_PASSWORD=postgres POSTGRES_DBNAME=coupon
export CRON_DEACTIVATE_STALE_SCHEDULE="@every 5m"  # optional, this is already the default
make run               # go run ./cmd, mặc định lắng nghe :8085 (HTTP server + cron cùng một process)
make test              # go test ./...
```

### Docker

`Dockerfile` build multi-stage (Go 1.23 alpine → alpine runtime), copy kèm `database.sql` để service tự áp dụng schema khi khởi động:

```bash
docker build -t coupon-service .
docker run -p 8085:8085 --env-file .env coupon-service
```

### CI/CD

- [`/.github/workflows/coupon-service-cicd.yml`](../.github/workflows/coupon-service-cicd.yml): build + test → build & push image đa kiến trúc lên `ghcr.io/jieeirosst/coupon-service` → `helm template | kubectl apply` khi push vào `master` (chỉ chạy khi đổi trong `coupon-service/`, `chart/coupon-service/` hoặc chính file workflow).
- [`/.github/workflows/go.yml`](../.github/workflows/go.yml): job `coupon_service` chạy `go build`/`go test` trên mọi push/PR vào `master`, cùng chỗ với các service Go khác trong repo.

### Kubernetes (Helm)

Chart nằm ở [`/chart/coupon-service`](../chart/coupon-service) (Deployment + Service + Secret + HPA + Ingress tùy chọn), theo đúng khuôn mẫu các service khác trong repo (`chart/parking-lot-service`, `chart/threads-service`, ...) và đã được khai báo làm dependency (`couponService`) trong chart tổng ở [`/chart/Chart.yaml`](../chart/Chart.yaml).

```bash
helm template ./chart/coupon-service \
  --set couponService.image.tag=<tag> \
  | kubectl apply -n default -f -
```
