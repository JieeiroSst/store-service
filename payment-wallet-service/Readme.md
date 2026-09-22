# Payment Wallet Service

Ví điện tử (e-wallet) nội bộ kiểu MoMo/ZaloPay/PayPal/Venmo: mỗi người dùng có một **ví** giữ số dư, có thể **nạp tiền** (top-up) từ phương thức thanh toán đã liên kết, **rút tiền** về phương thức đó, và **chuyển tiền** P2P sang ví khác — merchant cũng chỉ là một ví khác, nên thanh toán cho người bán dùng chung cơ chế chuyển tiền. Service **không** sở hữu KYC (xem `ekyc-service`) hay danh tính người dùng (xem `user_service`) — `user_id` chỉ là một chuỗi tham chiếu tới các service đó.

## 1. Yêu cầu chức năng (Functional Requirements)

- **Tạo ví**: mỗi `user_id` có đúng một ví theo một loại tiền tệ (`currency`), số dư khởi tạo bằng 0.
- **Nạp tiền (Deposit)**: cộng tiền vào ví sau khi đã xác nhận thanh toán từ phương thức bên ngoài (ngân hàng/thẻ) ở phía đối soát/webhook của PSP — service này chỉ ghi nhận kết quả cuối cùng vào ledger.
- **Rút tiền (Withdraw)**: trừ tiền khỏi ví để chuyển ra phương thức thanh toán đã liên kết; từ chối nếu số dư không đủ.
- **Chuyển tiền (Transfer)**: chuyển tiền giữa hai ví cùng loại tiền tệ, atomic hai chiều (trừ ví gửi, cộng ví nhận) — dùng cho cả P2P lẫn thanh toán merchant.
- **Lịch sử giao dịch**: mỗi ví có một ledger đầy đủ (`transactions`) giải thích được toàn bộ biến động số dư của chính nó.
- **Idempotency**: mọi lệnh nạp/rút/chuyển đều nhận một `reference_id` tùy chọn do client sinh ra; gọi lại với cùng `reference_id` trả về đúng kết quả lần đầu thay vì xử lý tiền lần nữa — bắt buộc để client có thể an toàn retry khi request bị timeout.
- **Quản lý phương thức thanh toán**: liên kết/hủy liên kết thẻ hoặc tài khoản ngân hàng, đặt một phương thức mặc định. Service chỉ lưu tham chiếu đã che/token hóa (vd. 4 số cuối, token từ PSP) — không bao giờ lưu số thẻ/tài khoản đầy đủ.
- **Đảo/hoàn tiền (Reverse/Refund)**: đảo một `Deposit`/`Withdraw` đã hoàn tất, hoặc đảo cả một `Transfer` (cả 2 chân), ghi thêm giao dịch `REVERSAL` bù trừ thay vì xoá lịch sử — đảo một `Transfer` sẽ thất bại nếu bên nhận đã tiêu hết tiền.
- **Vòng đời ví & hạn mức**: đóng băng (`FREEZE`)/mở lại (`UNFREEZE`)/đóng ví (`CLOSE`, chỉ khi số dư = 0), và đặt hạn mức giao dịch (`daily_limit`, `per_transaction_limit`) áp dụng cho Withdraw và chân đi của Transfer.
- **QR pay & Request money**: một `PaymentRequest` vừa dùng làm mã QR nhận thanh toán (ai quét cũng trả được) vừa dùng làm "yêu cầu chuyển tiền" P2P (chỉ định đúng 1 người phải trả) — trả tiền cho request thực chất là một `Transfer` bình thường.
- **Ví phụ (Pockets)**: mỗi ví có thể có nhiều "hũ" (tiết kiệm, chi tiêu...) là sub-balance của chính ví đó; nạp/rút giữa ví chính và hũ được ghi lại bằng giao dịch `POCKET_OUT`/`POCKET_IN`; đóng hũ sẽ tự trả phần còn lại về ví chính.
- **Sao kê (Statement)**: xuất lịch sử giao dịch của một ví trong khoảng thời gian dưới dạng CSV.

## 2. Yêu cầu phi chức năng (Non-functional Requirements)

- **Toàn vẹn số dư**: số dư không bao giờ âm (ràng buộc ở cả tầng ứng dụng và `CHECK (balance >= 0)` trong DB); mọi thay đổi số dư đi kèm đúng một dòng ledger giải thích nó.
- **Concurrency an toàn**: hai request nạp/rút/chuyển đồng thời trên cùng một ví không được ghi đè số dư của nhau (row lock, xem mục 8); hai chuyển khoản ngược chiều giữa cùng một cặp ví không được deadlock.
- **Idempotent theo thiết kế**: xử lý trùng lặp một `reference_id` là hành vi được định nghĩa rõ (trả lại kết quả cũ), không phải lỗi.
- **Không dùng số thực cho tiền**: số dư/số tiền là số nguyên ở đơn vị nhỏ nhất của tiền tệ (minor units, vd. cent với USD) để tránh sai số làm tròn dấu phẩy động — cùng cách tiếp cận Stripe/PayPal dùng.

## 3. Domain Model (Thực thể chính)

| Entity | Mô tả |
|---|---|
| `Wallet` | Ví của một user: số dư (minor units), loại tiền tệ, trạng thái (`ACTIVE`/`FROZEN`/`CLOSED`), hạn mức (`daily_limit`/`per_transaction_limit`) |
| `Transaction` | Một dòng ledger trên **một** ví: `DEPOSIT`, `WITHDRAW`, `TRANSFER_IN`, `TRANSFER_OUT`, `REVERSAL`, `POCKET_OUT`, `POCKET_IN` |
| `Transfer` | Một lần chuyển tiền giữa hai ví, tham chiếu tới đúng 2 `Transaction` (out + in) mà nó tạo ra |
| `PaymentMethod` | Phương thức thanh toán đã liên kết: `BANK_ACCOUNT`/`CARD`, có thể đặt mặc định |
| `Pocket` | Sub-balance của một `Wallet` (hũ tiết kiệm/chi tiêu) |
| `PaymentRequest` | Yêu cầu được trả tiền: QR pay (`payer_wallet_id` = NULL) hoặc request money (chỉ định `payer_wallet_id`) |

## 4. Kiến trúc tổng quan

Service được cài đặt theo **Hexagonal Architecture** (Ports & Adapters), dùng [`uber-go/fx`](https://github.com/uber-go/fx) để dependency injection — không có `main.go` nào tự tay khởi tạo/nối dây từng struct, toàn bộ được khai báo (declarative) qua các `fx.Module`, cùng khuôn mẫu với `parking-lot-service`, `doordash-service`, `coupon-service` trong repo này.

```
cmd/main.go                        → fx.New(infrastructure.Module).Run()

internal/
├── domain/
│   ├── model/                     → Wallet, Transaction, Transfer, PaymentMethod, Pocket, PaymentRequest
│   └── port/
│       ├── driving.go             → WalletUsecase, TransactionUsecase, PaymentMethodUsecase,
│       │                             PocketUsecase, PaymentRequestUsecase (port vào)
│       └── driven.go              → WalletRepository, TransactionRepository, TransferRepository,
│                                     PaymentMethodRepository, PocketRepository, PaymentRequestRepository (port ra)
├── application/                   → cài đặt driving port: validate, idempotency check, orchestrate
├── adapter/
│   ├── primary/http/              → adapter vào: Gin handler/router, gọi driving port
│   └── secondary/repository/      → adapter ra: GORM/Postgres, cài đặt driven port
└── infrastructure/                → wiring: config, logger, database, http server, fx.Module tổng
```

- **domain** là lõi, không import Gin/GORM/fx — chỉ chứa entity và interface (port).
- **application** chỉ phụ thuộc `domain/port`, không biết HTTP hay SQL cụ thể là gì — có thể test bằng fake repository thuần Go (`internal/application/fakes_test.go`), không cần Postgres thật.
- **`WalletRepository`/`PocketRepository` là nơi duy nhất ghi tiền**: `Deposit`/`Withdraw`/`Transfer`/`Reverse`/`ReverseTransfer` (trên ví) và `MoveToPocket`/`MoveFromPocket`/`Close` (trên hũ) mỗi hàm thực hiện trọn vẹn một giao dịch DB (khóa dòng, kiểm tra, cập nhật số dư, ghi ledger) bên trong adapter — `application` không tự điều phối transaction đa bước, nó chỉ gọi đúng một hàm port và nhận kết quả cuối cùng.
- **`PaymentRequestUsecase` tái dùng `TransactionUsecase`**: trả tiền cho một `PaymentRequest` (`Pay`) không tự viết logic chuyển tiền — nó gọi thẳng `TransactionUsecase.Transfer`, nên mọi ràng buộc của Transfer (idempotency, khớp tiền tệ, hạn mức, khóa 2 ví) tự động áp dụng, không bị lặp code.
- **fx** khởi tạo theo đúng thứ tự phụ thuộc (`config → database → repository → application → http handler → http server`), xem `internal/infrastructure/module.go`.

## 5. Luồng nghiệp vụ chính

### Nạp tiền (`TransactionUsecase.Deposit`, `internal/application/transaction.go`)

1. `amount` phải dương, ngược lại `ErrInvalidAmount`.
2. Nếu có `reference_id`, tra `TransactionRepository.GetByReferenceID` trên đúng ví đó — nếu đã tồn tại, trả lại transaction cũ (idempotent replay), **không** cộng tiền lần nữa.
3. Gọi `WalletRepository.Deposit`: khóa dòng ví (`SELECT ... FOR UPDATE`), kiểm tra ví đang `ACTIVE`, cộng số dư, ghi dòng `Transaction` — toàn bộ trong một DB transaction.

### Rút tiền (`Withdraw`)

Tương tự Deposit, nhưng `WalletRepository.Withdraw` kiểm tra thêm `balance >= amount` trong cùng transaction đã khóa dòng — không có khoảng hở giữa lúc đọc số dư và lúc trừ tiền để hai request rút đồng thời cùng vượt qua kiểm tra.

### Chuyển tiền (`Transfer`)

1. Từ chối nếu `amount <= 0`, hai ví trùng nhau (`ErrSameWallet`), hoặc khác loại tiền tệ (`ErrCurrencyMismatch`).
2. Nếu có `reference_id`, tra `TransferRepository.GetByReferenceID` (idempotency toàn cục, không theo từng ví) — trùng thì trả lại transfer cũ.
3. Gọi `WalletRepository.Transfer`: khóa **cả hai** ví theo một thứ tự cố định (so sánh `wallet_id`, luôn khóa ví "nhỏ hơn" trước bất kể ai là sender/receiver) để hai lệnh chuyển ngược chiều giữa cùng một cặp ví không bao giờ deadlock chờ nhau; kiểm tra cả hai ví `ACTIVE` và số dư đủ; trừ ví gửi, cộng ví nhận, ghi 2 dòng `Transaction` (`TRANSFER_OUT`/`TRANSFER_IN`) + 1 dòng `Transfer` — tất cả trong một DB transaction.
4. Nếu `sender.DailyLimit`/`PerTransactionLimit` > 0, `amount` bị kiểm tra trước khi gọi `WalletRepository.Transfer` (mục "Hạn mức giao dịch" bên dưới).

### Đảo/hoàn tiền (`Reverse`, `ReverseTransfer`)

- `Reverse(transactionID, reason)`: chỉ áp dụng cho `Transaction` đang `COMPLETED` và có `Type` là `DEPOSIT`/`WITHDRAW` (không đảo trực tiếp một chân của `Transfer` — dùng `ReverseTransfer` cho việc đó). `WalletRepository.Reverse` khóa ví, đổi chiều số dư (đảo Deposit thì trừ lại, đảo Withdraw thì cộng lại), rồi `UPDATE transactions SET status='REVERSED' WHERE ... AND status='COMPLETED'` — điều kiện `AND status='COMPLETED'` trong cùng câu UPDATE nghĩa là hai lệnh đảo đồng thời trên cùng giao dịch chỉ một cái thắng, cái còn lại nhận `ErrTransactionNotReversible`.
- `ReverseTransfer(transferID, reason)`: khóa cả hai ví theo cùng thứ tự cố định như `Transfer`, hoàn tiền cho sender, trừ lại của receiver — **thất bại với `ErrInsufficientBalance` nếu receiver đã tiêu hết tiền nhận được**, đúng như hành vi hoàn tiền thật (không thể "rút ngược" tiền đã tiêu). Đánh dấu `REVERSED` cho cả `Transfer` lẫn 2 `Transaction` gốc.
- Cả hai đều ghi thêm một dòng `Transaction` mới loại `REVERSAL` thay vì xoá/sửa lịch sử cũ — ledger luôn chỉ được nối thêm (append-only).

### Vòng đời ví (`WalletUsecase.FreezeWallet`/`UnfreezeWallet`/`CloseWallet`)

Chuyển trạng thái đi qua kiểm tra hợp lệ đơn giản (`ACTIVE → FROZEN → ACTIVE`, không cho freeze một ví đã `FROZEN`/`CLOSED`) trước khi gọi `WalletRepository.UpdateStatus`. `CloseWallet` gọi `WalletRepository.Close`, khóa dòng ví và từ chối với `ErrWalletNotEmpty` nếu `balance != 0` — một ví có tiền không bao giờ bị đóng nhầm.

### Hạn mức giao dịch (`checkLimits`, `internal/application/transaction.go`)

Áp dụng cho `Withdraw` và chân đi (sender) của `Transfer`, **không** áp dụng cho `Deposit` (hạn mức nạp tiền là việc của PSP, không phải ledger này): nếu `PerTransactionLimit > 0` và `amount` vượt quá → `ErrPerTransactionLimitExceeded`; nếu `DailyLimit > 0`, cộng dồn `TransactionRepository.SumOutgoingSince` (tổng `WITHDRAW`+`TRANSFER_OUT` đã `COMPLETED` từ đầu ngày) với `amount` — vượt `DailyLimit` thì `ErrDailyLimitExceeded`.

### QR pay & Request money (`PaymentRequestUsecase`)

Một `PaymentRequest` phục vụ cả 2 luồng bằng cùng một cơ chế:

- **QR pay**: `payer_wallet_id = NULL` lúc tạo → `GET /payment-requests/:id/qr` trả về một payload (`paywallet://pay?request_id=...`) để render QR; **bất kỳ ví nào** gọi `Pay` với đúng `id` đều thanh toán được.
- **Request money (P2P)**: chỉ định `payer_wallet_id` lúc tạo → chỉ ví đó mới `Pay` được, ví khác nhận `ErrPaymentRequestPayerMismatch`.

`Pay` kiểm tra request đang `PENDING` và chưa hết hạn (`expires_at`, tự chuyển `EXPIRED` nếu đã quá hạn), rồi gọi thẳng `TransactionUsecase.Transfer` (payer → requester) — do đó thừa hưởng toàn bộ idempotency/hạn mức/khóa ví của Transfer. `UpdateStatus` trong `PaymentRequestRepository` chỉ áp dụng khi request còn `PENDING` (`... WHERE status = 'PENDING'`), nên `Pay` và `Cancel` chạy đồng thời không thể cùng thắng.

### Ví phụ / Pockets (`PocketUsecase`)

`DepositToPocket`/`WithdrawFromPocket` gọi `PocketRepository.MoveToPocket`/`MoveFromPocket`, khóa **cả** dòng ví lẫn dòng pocket trong một DB transaction rồi chuyển tiền qua lại, ghi một `Transaction` loại `POCKET_OUT`/`POCKET_IN` (không tạo cặp giao dịch như Transfer, vì tiền không rời khỏi ví — chỉ đổi chỗ nội bộ). `ClosePocket` cộng thẳng số dư còn lại của pocket về ví rồi xoá pocket, atomic — không có API nào đóng một pocket không rỗng và mất tiền.

## 6. API

| Method | Endpoint | Mô tả |
|---|---|---|
| POST | `/api/v1/wallets` | Tạo ví (`{"user_id", "currency"}`) |
| GET | `/api/v1/wallets/:id` | Chi tiết ví theo `wallet_id` |
| GET | `/api/v1/wallets/user/:userId` | Chi tiết ví theo `user_id` |
| POST | `/api/v1/wallets/:id/deposit` | Nạp tiền (`{"amount", "reference_id", "description"}`) |
| POST | `/api/v1/wallets/:id/withdraw` | Rút tiền (cùng payload) |
| GET | `/api/v1/wallets/:id/transactions?limit=&offset=` | Lịch sử giao dịch của ví |
| GET | `/api/v1/wallets/:id/statement?from=&to=` | Xuất sao kê CSV (`YYYY-MM-DD`, mặc định 30 ngày gần nhất) |
| POST | `/api/v1/wallets/:id/freeze` | Đóng băng ví (`{"reason"}`) |
| POST | `/api/v1/wallets/:id/unfreeze` | Mở lại ví đang bị đóng băng |
| POST | `/api/v1/wallets/:id/close` | Đóng ví vĩnh viễn (chỉ khi số dư = 0) |
| PATCH | `/api/v1/wallets/:id/limits` | Đặt hạn mức (`{"daily_limit", "per_transaction_limit"}`, 0 = không giới hạn) |
| POST | `/api/v1/wallets/:id/pockets` | Tạo ví phụ/pocket (`{"name"}`) |
| GET | `/api/v1/wallets/:id/pockets` | Danh sách pocket của ví |
| POST | `/api/v1/pockets/:pocketId/deposit` | Chuyển tiền từ ví chính vào pocket (`{"amount"}`) |
| POST | `/api/v1/pockets/:pocketId/withdraw` | Chuyển tiền từ pocket về ví chính (`{"amount"}`) |
| DELETE | `/api/v1/pockets/:pocketId` | Đóng pocket, tự trả số dư còn lại về ví |
| POST | `/api/v1/transfers` | Chuyển tiền (`{"sender_wallet_id", "receiver_wallet_id", "amount", "reference_id", "description"}`) |
| GET | `/api/v1/transfers/:id` | Chi tiết một lần chuyển tiền |
| POST | `/api/v1/transfers/:id/reverse` | Đảo một lần chuyển tiền (`{"reason"}`) |
| GET | `/api/v1/transactions/:id` | Chi tiết một giao dịch |
| POST | `/api/v1/transactions/:id/reverse` | Đảo một giao dịch nạp/rút (`{"reason"}`) |
| POST | `/api/v1/payment-requests` | Tạo yêu cầu thanh toán/QR pay (`{"requester_wallet_id", "payer_wallet_id"?, "amount", "description", "expires_in_seconds"?}`) |
| GET | `/api/v1/payment-requests/:id` | Chi tiết yêu cầu thanh toán |
| GET | `/api/v1/payment-requests/:id/qr` | Payload QR để quét thanh toán |
| GET | `/api/v1/payment-requests?wallet_id=` | Danh sách yêu cầu liên quan tới một ví (bên tạo hoặc bên trả) |
| POST | `/api/v1/payment-requests/:id/pay` | Trả một yêu cầu thanh toán (`{"payer_wallet_id"}`) |
| POST | `/api/v1/payment-requests/:id/cancel` | Hủy yêu cầu thanh toán |
| POST | `/api/v1/payment-methods` | Liên kết phương thức thanh toán (`{"user_id", "type", "provider", "account_number"}`) |
| GET | `/api/v1/payment-methods?user_id=` | Danh sách phương thức của user |
| DELETE | `/api/v1/payment-methods/:id` | Hủy liên kết (vô hiệu hóa mềm) |
| PATCH | `/api/v1/payment-methods/:id/default` | Đặt làm phương thức mặc định |
| GET | `/health` | Health check |

`amount` luôn là số nguyên ở đơn vị nhỏ nhất của tiền tệ (vd. `500` = $5.00 với USD).

## 7. Database Schema

Schema thật (được `internal/infrastructure/database.applySchema` tự áp dụng khi service khởi động) nằm ở [`database.sql`](database.sql), dùng `UUID` làm khóa chính (sinh bởi `google/uuid` ở tầng application) và `BIGINT` cho mọi cột tiền (minor units, không dùng `DECIMAL`/float). Không có seed data — service dùng được ngay qua API mà không cần dữ liệu khởi tạo.

Idempotency được ép cả ở tầng DB, không chỉ ở tầng application: `UNIQUE INDEX ... WHERE reference_id IS NOT NULL` trên `(wallet_id, reference_id)` của `transactions` và trên `reference_id` của `transfers`.

## 8. Vấn đề đồng thời (Concurrency)

`WalletRepository.Deposit`/`Withdraw`/`Transfer` (`internal/adapter/secondary/repository/wallet.go`) mở một DB transaction và khóa dòng ví bằng `SELECT ... FOR UPDATE` trước khi đọc/sửa số dư — khác với `parking-lot-service` (dùng `SKIP LOCKED` vì có thể bỏ qua một chỗ đỗ và thử chỗ khác), tiền không thể "bỏ qua": request thứ hai phải **chờ** thay vì được phép đọc số dư cũ. `Transfer` khóa hai ví theo thứ tự cố định (so sánh `wallet_id`) bất kể vai trò sender/receiver, để hai lệnh chuyển tiền ngược chiều giữa cùng một cặp ví luôn khóa theo cùng một thứ tự và không thể deadlock lẫn nhau.

## 9. Hướng mở rộng

- Tích hợp PSP thật (Stripe/VNPay/Momo Business API) để xác nhận Deposit/Withdraw qua webhook thay vì ghi nhận trực tiếp qua API.
- Đa tiền tệ trên cùng một user (hiện tại 1 ví = 1 user = 1 currency).
- Gắn hạn mức mặc định theo trạng thái KYC (gọi `ekyc-service`) thay vì đặt tay qua `PATCH /wallets/:id/limits`, giống cách `doordash-service` gọi service khác qua HTTP thay vì join chéo database.
- Sinh ảnh QR thật (PNG) từ payload ở `GET /payment-requests/:id/qr` thay vì chỉ trả chuỗi payload.
- Thông báo realtime khi số dư thay đổi hoặc payment request được trả (tích hợp `notification_service`).

## 10. Chạy & Triển khai

### Chạy local

```bash
export POSTGRES_HOST=localhost POSTGRES_PORT=5432 POSTGRES_USER=postgres \
       POSTGRES_PASSWORD=postgres POSTGRES_DBNAME=payment_wallet
make run               # go run ./cmd, mặc định lắng nghe :8088
make test              # go test ./...
```

### Docker

`Dockerfile` build multi-stage (Go 1.22 alpine → alpine runtime), copy kèm `database.sql` để service tự áp dụng schema khi khởi động:

```bash
docker build -t payment-wallet-service .
docker run -p 8088:8088 --env-file .env payment-wallet-service
```

### CI/CD

- [`/.github/workflows/payment-wallet-service-cicd.yml`](../.github/workflows/payment-wallet-service-cicd.yml): build + test → build & push image đa kiến trúc lên `ghcr.io/jieeirosst/payment-wallet-service` → `helm template | kubectl apply` khi push vào `master` (chỉ chạy khi đổi trong `payment-wallet-service/`, `chart/payment-wallet-service/` hoặc chính file workflow).
- [`/.github/workflows/go.yml`](../.github/workflows/go.yml): job `payment_wallet_service` chạy `go build`/`go test` trên mọi push/PR vào `master`, cùng chỗ với các service Go khác trong repo.

### Kubernetes (Helm)

Chart nằm ở [`/chart/payment-wallet-service`](../chart/payment-wallet-service) (Deployment + Service + Secret + HPA + Ingress tùy chọn), theo đúng khuôn mẫu các service khác trong repo (`chart/parking-lot-service`, `chart/doordash-service`, ...) và đã được khai báo làm dependency (`paymentWalletService`) trong chart tổng ở [`/chart/Chart.yaml`](../chart/Chart.yaml).

```bash
helm template ./chart/payment-wallet-service \
  --set paymentWalletService.image.tag=<tag> \
  | kubectl apply -n default -f -
```
