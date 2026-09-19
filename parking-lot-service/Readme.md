# Parking Lot Service

Hệ thống quản lý bãi đỗ xe, hỗ trợ nhiều loại phương tiện (xe máy, ô tô, xe tải, xe khách...), nhiều tầng/khu vực đỗ, tính phí theo thời gian và phát hành vé ra/vào.

## 1. Yêu cầu chức năng (Functional Requirements)

- Bãi xe có nhiều **tầng/khu vực** (Floor/Zone), mỗi khu vực có nhiều **chỗ đỗ** (Spot) với loại khác nhau.
- Hỗ trợ nhiều **loại phương tiện**: Motorcycle (xe máy), Car (ô tô), Truck (xe tải), Bus (xe khách), Bicycle (xe đạp), Electric Vehicle (xe điện).
- Mỗi loại phương tiện chỉ được đỗ vào loại chỗ tương ứng hoặc chỗ lớn hơn (ví dụ: ô tô không đỗ được vào chỗ dành cho xe máy, nhưng xe máy có thể đỗ tạm vào chỗ ô tô nếu cấu hình cho phép).
- Ghi nhận xe **vào bãi** (check-in): tìm chỗ trống phù hợp, gán chỗ, phát hành vé (ticket), gán cổng vào (gate).
- Ghi nhận xe **ra bãi** (check-out): tính thời gian đỗ, tính phí, xử lý thanh toán, giải phóng chỗ.
- Tính phí gửi xe theo **bảng giá** (Rate) riêng cho từng loại phương tiện (giá theo giờ, giá tối đa/ngày, giá theo lượt).
- Hỗ trợ chỗ đỗ **ưu tiên/đặc biệt**: chỗ cho người khuyết tật, chỗ có trạm sạc điện, chỗ VIP.
- Hỗ trợ **đặt chỗ trước** (reservation) — tùy chọn mở rộng.
- Tra cứu **tình trạng chỗ trống** theo loại xe, theo tầng.
- Lịch sử gửi/lấy xe (Parking History) phục vụ tra cứu, đối soát, báo cáo.
- Quản lý nhiều **bãi xe** (multi-lot) nếu triển khai cho chuỗi/nhiều địa điểm.

## 2. Yêu cầu phi chức năng (Non-functional Requirements)

- **Concurrency**: nhiều xe vào/ra cùng lúc không được gán trùng một chỗ đỗ (cần lock/transaction ở tầng DB hoặc distributed lock).
- **Consistency**: trạng thái chỗ đỗ (available/occupied) phải nhất quán ngay sau mỗi giao dịch check-in/check-out.
- **Availability**: service cần chịu tải tốt ở giờ cao điểm (ra/vào đồng loạt).
- **Extensibility**: dễ dàng thêm loại phương tiện mới, loại chỗ mới, chính sách giá mới mà không phá vỡ logic hiện có.
- **Auditability**: mọi giao dịch vào/ra/thanh toán đều phải log lại để đối soát.

## 3. Domain Model (Thực thể chính)

| Entity | Mô tả |
|---|---|
| `ParkingLot` | Một bãi đỗ xe (địa chỉ, tổng số tầng, tổng số chỗ) |
| `Floor` | Một tầng/khu vực trong bãi, thuộc một `ParkingLot` |
| `ParkingSpot` | Một chỗ đỗ cụ thể: loại chỗ, trạng thái trống/đầy, thuộc `Floor` |
| `Vehicle` | Phương tiện: biển số, loại xe |
| `Gate` | Cổng ra/vào của bãi (Entry/Exit) |
| `Ticket` (ParkingHistory) | Vé gửi xe: xe nào, chỗ nào, vào lúc nào, ra lúc nào, cổng nào |
| `Rate` | Bảng giá theo loại phương tiện (giá/giờ, giá tối đa) |
| `Payment` | Giao dịch thanh toán gắn với một `Ticket` |

### Loại phương tiện (VehicleType)

```
MOTORCYCLE, CAR, TRUCK, BUS, BICYCLE, ELECTRIC_VEHICLE
```

### Loại chỗ đỗ (SpotType) — ánh xạ theo loại phương tiện

| SpotType | Phương tiện phù hợp |
|---|---|
| `MOTORCYCLE_SPOT` | Motorcycle, Bicycle |
| `COMPACT_SPOT` | Car |
| `LARGE_SPOT` | Truck, Bus |
| `HANDICAP_SPOT` | Mọi loại (ưu tiên người khuyết tật) |
| `EV_CHARGING_SPOT` | Electric Vehicle |

## 4. Kiến trúc tổng quan

Service được cài đặt theo **Hexagonal Architecture** (Ports & Adapters), dùng [`uber-go/fx`](https://github.com/uber-go/fx) để dependency injection — không có `main.go` nào tự tay khởi tạo/nối dây từng struct, toàn bộ được khai báo khai báo (declarative) qua các `fx.Module`.

```
cmd/main.go                        → fx.New(infrastructure.Module).Run()

internal/
├── domain/
│   ├── model/                     → Vehicle, ParkingSpot, Ticket, Rate, Payment (không phụ thuộc framework)
│   └── port/
│       ├── driving.go             → ParkingUsecase, RateUsecase (port vào - cổng mà adapter primary gọi)
│       └── driven.go              → VehicleRepository, ParkingSpotRepository, TicketRepository,
│                                     RateRepository, PaymentRepository (port ra - cổng mà application gọi)
├── application/                   → cài đặt các driving port (business logic: check-in, check-out, tính phí)
├── adapter/
│   ├── primary/http/              → adapter vào: Gin handler/router, gọi driving port
│   └── secondary/repository/      → adapter ra: GORM/Postgres, cài đặt driven port
└── infrastructure/                → wiring: config, logger, database, http server, fx.Module tổng
```

- **domain** là lõi, không import Gin/GORM/fx — chỉ chứa entity và interface (port).
- **application** chỉ phụ thuộc `domain/port`, không biết HTTP hay SQL cụ thể là gì.
- **adapter/primary** (HTTP) và **adapter/secondary** (repository) có thể thay thế độc lập (vd. đổi Gin sang gRPC, đổi Postgres sang MySQL) mà không đụng vào `application`/`domain`.
- **fx** chịu trách nhiệm khởi tạo theo đúng thứ tự phụ thuộc (`config → database → repository → application → http handler → http server`), xem `internal/infrastructure/module.go`.

### Bộ phân giải chỗ đỗ & tính phí

- `model.CompatibleSpotTypes(vehicleType)` trả về danh sách loại chỗ phù hợp theo thứ tự ưu tiên (vd. `Car` → `COMPACT_SPOT` rồi mới tới `HANDICAP_SPOT`).
- `adapter/secondary/repository.parkingSpotRepository.FindAndReserveAvailable` dùng transaction + `SELECT ... FOR UPDATE SKIP LOCKED` để khóa và gán chỗ atomic (xem mục 8).
- `model.Rate.Calculate` tính phí theo giờ (làm tròn lên) và giới hạn bởi `daily_max_rate` mỗi ngày.

## 5. Luồng nghiệp vụ chính

### Check-in (xe vào bãi)

1. Xe vào `Gate` (Entry), quét biển số/vé.
2. `SpotAllocator` tìm `ParkingSpot` còn trống phù hợp với `VehicleType`.
3. Gán chỗ (transaction: `UPDATE ParkingSpot SET is_available = false WHERE spot_id = ? AND is_available = true` — đảm bảo atomic, tránh double-booking).
4. Tạo `Ticket` mới (parked_time = now, gate vào, spot_id).
5. Trả vé cho khách (mã vé/QR).

### Check-out (xe ra bãi)

1. Xe đến `Gate` (Exit), quét vé.
2. Lấy `Ticket` tương ứng, tính `duration = leave_time - parked_time`.
3. `FeeCalculator` tính phí theo `Rate` của loại phương tiện.
4. Xử lý `Payment` (tiền mặt/thẻ/ví điện tử).
5. Cập nhật `Ticket.leave_time`, giải phóng `ParkingSpot` (`is_available = true`).

## 6. API đề xuất

| Method | Endpoint | Mô tả |
|---|---|---|
| GET | `/api/v1/spots/available?type=CAR` | Tra cứu số chỗ trống theo loại xe |
| POST | `/api/v1/check-in` | Xe vào bãi, trả về ticket |
| POST | `/api/v1/check-out` | Xe ra bãi, tính phí, thanh toán |
| GET | `/api/v1/tickets/{ticket_id}` | Chi tiết vé |
| GET | `/api/v1/history?plate=...` | Lịch sử gửi xe theo biển số |
| GET | `/api/v1/rates` | Bảng giá hiện hành |
| POST | `/api/v1/reservations` | Đặt chỗ trước (mở rộng) |

## 7. Database Schema

Schema thật (được `internal/infrastructure/database.applySchema` tự áp dụng khi service khởi động) nằm ở [`database.sql`](database.sql), dùng `UUID` làm khóa chính (sinh bởi `google/uuid` ở tầng application, không phải `SERIAL`/auto-increment) để nhất quán với các service khác trong repo. Tóm tắt:

```sql
CREATE TABLE parking_lots (lot_id UUID PRIMARY KEY, name VARCHAR(100) NOT NULL, address VARCHAR(255));
CREATE TABLE floors (floor_id UUID PRIMARY KEY, lot_id UUID REFERENCES parking_lots(lot_id), floor_number INT NOT NULL);
CREATE TABLE gates (gate_id UUID PRIMARY KEY, lot_id UUID REFERENCES parking_lots(lot_id), name VARCHAR(50), gate_type VARCHAR(10) CHECK (gate_type IN ('ENTRY', 'EXIT')));

CREATE TABLE parking_spots (
    spot_id UUID PRIMARY KEY,
    floor_id UUID REFERENCES floors(floor_id),
    type VARCHAR(50) CHECK (type IN ('MOTORCYCLE_SPOT', 'COMPACT_SPOT', 'LARGE_SPOT', 'HANDICAP_SPOT', 'EV_CHARGING_SPOT')),
    is_available BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE vehicles (
    license_plate VARCHAR(20) PRIMARY KEY,
    type VARCHAR(50) CHECK (type IN ('MOTORCYCLE', 'CAR', 'TRUCK', 'BUS', 'BICYCLE', 'ELECTRIC_VEHICLE'))
);

CREATE TABLE rates (rate_id UUID PRIMARY KEY, vehicle_type VARCHAR(50) NOT NULL, hourly_rate DECIMAL(10,2) NOT NULL, daily_max_rate DECIMAL(10,2), effective_from TIMESTAMP NOT NULL);

CREATE TABLE parking_history (
    history_id UUID PRIMARY KEY,
    vehicle_plate VARCHAR(20) REFERENCES vehicles(license_plate),
    spot_id UUID REFERENCES parking_spots(spot_id),
    entry_gate_id UUID REFERENCES gates(gate_id),
    exit_gate_id UUID REFERENCES gates(gate_id),
    parked_time TIMESTAMP NOT NULL,
    leave_time TIMESTAMP,
    status VARCHAR(20) CHECK (status IN ('ACTIVE', 'COMPLETED')) NOT NULL DEFAULT 'ACTIVE',
    CONSTRAINT check_leave_time CHECK (leave_time IS NULL OR leave_time > parked_time)
);

CREATE TABLE payments (payment_id UUID PRIMARY KEY, history_id UUID REFERENCES parking_history(history_id), amount DECIMAL(10,2) NOT NULL, method VARCHAR(30) CHECK (method IN ('CASH', 'CARD', 'E_WALLET')), paid_at TIMESTAMP, status VARCHAR(20) CHECK (status IN ('PENDING', 'PAID', 'FAILED')) NOT NULL DEFAULT 'PENDING');
```

`database.sql` cũng seed sẵn 1 `ParkingLot`/`Floor`/2 `Gate`, bảng `rates` cho cả 6 loại xe, và 12 `parking_spots` mẫu (5 compact, 3 motorcycle, 2 large, 1 handicap, 1 EV) để service chạy được ngay mà không cần API quản trị lot/floor/gate riêng (xem mục 9).

## 8. Vấn đề đồng thời (Concurrency)

`ParkingSpotRepository.FindAndReserveAvailable` (`internal/adapter/secondary/repository/spot.go`) mở transaction, khóa 1 hàng ứng viên bằng `SELECT ... FOR UPDATE SKIP LOCKED` rồi mới `UPDATE is_available = false`. `SKIP LOCKED` nghĩa là 2 xe check-in cùng lúc sẽ không bao giờ giành cùng 1 chỗ hay bị block chờ nhau — request thua chỉ đơn giản bỏ qua hàng đã bị khóa và thử chỗ kế tiếp (hoặc loại chỗ tương thích kế tiếp theo `model.CompatibleSpotTypes`).

## 9. Hướng mở rộng

- Đặt chỗ trước (`Reservation`) và giữ chỗ trong khoảng thời gian.
- Vé tháng/thuê bao dài hạn cho một `Vehicle`.
- Tích hợp nhận diện biển số (ANPR) tại `Gate` thay vì quét vé giấy.
- Thông báo realtime (WebSocket/SSE) về số chỗ trống theo loại xe.
- Multi-tenant cho chuỗi nhiều bãi xe (`ParkingLot` đã được thiết kế sẵn cho việc này).
- API quản trị `ParkingLot`/`Floor`/`Gate` (hiện đang seed tĩnh qua `database.sql`).

## 10. Chạy & Triển khai

### Chạy local

```bash
export POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=postgres \
       POSTGRES_PASSWORD=postgres POSTGRES_DBNAME=parking_lot
make run               # go run ./cmd, mặc định lắng nghe :8084
make test              # go test ./...
```

### Docker

`Dockerfile` build multi-stage (Go 1.22 alpine → alpine runtime), copy kèm `database.sql` để service tự áp dụng schema khi khởi động:

```bash
docker build -t parking-lot-service .
docker run -p 8084:8084 --env-file .env parking-lot-service
```

### CI/CD

- [`/.github/workflows/parking-lot-service-cicd.yml`](../.github/workflows/parking-lot-service-cicd.yml): build + test → build & push image đa kiến trúc lên `ghcr.io/jieeirosst/parking-lot-service` → `helm template | kubectl apply` khi push vào `master` (chỉ chạy khi đổi trong `parking-lot-service/`, `chart/parking-lot-service/` hoặc chính file workflow).
- [`/.github/workflows/go.yml`](../.github/workflows/go.yml): job `parking_lot_service` chạy `go build`/`go test` trên mọi push/PR vào `master`, cùng chỗ với các service Go khác trong repo.

### Kubernetes (Helm)

Chart nằm ở [`/chart/parking-lot-service`](../chart/parking-lot-service) (Deployment + Service + Secret + HPA + Ingress tùy chọn), theo đúng khuôn mẫu các service khác trong repo (`chart/threads-service`, `chart/post-service`, ...) và đã được khai báo làm dependency (`parkingLotService`) trong chart tổng ở [`/chart/Chart.yaml`](../chart/Chart.yaml).

```bash
helm template ./chart/parking-lot-service \
  --set parkingLotService.image.tag=<tag> \
  | kubectl apply -n default -f -
```
