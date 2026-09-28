# draw-image-service

Service ghép nhiều ảnh thành **một ảnh JPEG**, lưu ảnh vào **MinIO**, lưu thông tin ảnh vào **MySQL**. Mỗi ảnh ghép có một **UUID** riêng để lấy lại sau.

Có 3 kiểu ghép (layout):

| Layout | Mô tả |
|---|---|
| `grid` (mặc định) | Xếp ảnh thành lưới N cột, mỗi ô được cắt vừa ô vuông |
| `row` | Xếp tất cả ảnh trên một hàng ngang |
| `center` | Một ảnh chính to ở giữa, tất cả ảnh khác dán đè lên 4 góc và các cạnh để làm nổi bật ảnh chính |

---

## API

Base URL khi chạy local: `http://localhost:8080`

Mọi lỗi đều trả JSON dạng `{"error": "<mô tả>"}`.

### 1. `POST /upload`: ghép và lưu ảnh

Request `multipart/form-data`:

| Tên | Nơi | Bắt buộc | Mô tả |
|---|---|---|---|
| `files` | form, lặp lại được | có | Các ảnh cần ghép (JPEG, PNG, GIF, BMP, TIFF, WebP), kích thước bất kỳ (xem [Kích thước ảnh đầu vào](#kích-thước-ảnh-đầu-vào)) |
| `center` | form, tối đa 1 | không | Ảnh chính cho `layout=center`. Nếu bỏ trống, file `files` đầu tiên làm ảnh chính |
| `layout` | query | không | `grid` \| `row` \| `center`, mặc định `grid` |
| `columns` | query | không | Số cột cho `grid` (số nguyên dương), mặc định `DEFAULT_COLUMNS` |

Response `201 Created`, header `Location: /images/<id>`:

```json
{
  "id": "3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91",
  "url": "/images/3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91",
  "bucket": "draw-image-service",
  "object_key": "collages/3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91.jpg",
  "content_type": "image/jpeg",
  "size": 84213,
  "width": 800,
  "height": 800,
  "image_count": 2,
  "layout": "center",
  "columns": 0,
  "created_at": "2026-09-28T08:15:30.123Z"
}
```

Ví dụ:

```sh
# Lưới 2 cột (mặc định)
curl -F files=@a.jpg -F files=@b.jpg -F files=@c.jpg http://localhost:8080/upload

# Lưới 3 cột
curl -F files=@a.jpg -F files=@b.jpg -F files=@c.jpg "http://localhost:8080/upload?columns=3"

# Một hàng ngang
curl -F files=@a.jpg -F files=@b.jpg "http://localhost:8080/upload?layout=row"

# Ảnh ghế làm trung tâm, sticker Flash Sale dán lên góc
curl -F center=@chair.jpg -F files=@flash-sale.png "http://localhost:8080/upload?layout=center"
```

### 2. `GET /images/:id`: lấy ảnh đã ghép

Trả ảnh JPEG (`Content-Type: image/jpeg`). Ảnh không bao giờ thay đổi theo id, nên response có `Cache-Control: public, max-age=31536000, immutable` và `ETag`.

```sh
curl -o collage.jpg http://localhost:8080/images/3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91
```

Có thể dùng trực tiếp trong HTML: `<img src="http://localhost:8080/images/<id>">`

### 3. `GET /images/:id/info`: lấy thông tin ảnh

Trả JSON giống response của `POST /upload`.

```sh
curl http://localhost:8080/images/3f1c8a4e-2b7d-4c1e-9a5f-0d6b2e8c7a91/info
```

### 4. `GET /health`

Trả `{"status":"ok"}`. Kubernetes dùng endpoint này cho readiness/liveness probe.

### Mã lỗi

| HTTP | Khi nào |
|---|---|
| `400` | Không có ảnh; file rỗng, hỏng hoặc định dạng không hỗ trợ (ví dụ HEIC); `layout`/`columns` sai; gửi nhiều hơn 1 `center`; `id` không phải UUID |
| `404` | Không tìm thấy ảnh với `id` này |
| `413` | Body vượt `MAX_UPLOAD_MB`; số ảnh vượt `MAX_IMAGES`; ảnh vượt `MAX_PIXELS` |
| `500` | Lỗi MinIO/MySQL (chi tiết chỉ ghi vào log server) |

---

## Các layout

### `grid` / `row`

Mỗi ảnh được thu nhỏ và **cắt giữa** cho vừa ô `CELL_WIDTH × CELL_HEIGHT` (mặc định 100×100), rồi xếp từ trái sang phải, trên xuống dưới. Nền trắng.

```
grid, columns=2, 3 ảnh:        row, 3 ảnh:
┌───┬───┐                      ┌───┬───┬───┐
│ 1 │ 2 │                      │ 1 │ 2 │ 3 │
├───┼───┘                      └───┴───┴───┘
│ 3 │
└───┘
```

### `center`

Canvas `CENTER_WIDTH × CENTER_HEIGHT` (mặc định 800×800), nền trắng.

- **Ảnh chính** được phóng/thu cho vừa vùng giữa (khoảng 72% canvas), **giữ nguyên toàn bộ, không cắt**, có bóng đổ nhẹ.
- **Tất cả ảnh phụ** (không giới hạn riêng, chỉ bị giới hạn bởi `MAX_IMAGES`) được xếp quanh 4 góc và các cạnh của ảnh chính. Tâm mỗi ảnh phụ nằm đúng trên góc/cạnh nên luôn đè lên viền ảnh chính:
  1. 4 ảnh đầu vào 4 góc: trên trái → trên phải → dưới trái → dưới phải.
  2. Các ảnh còn lại chia lần lượt cho cạnh trên → dưới → trái → phải, lặp lại. Trên mỗi cạnh, các ảnh **cách đều nhau**.

```
8 ảnh phụ:                        16 ảnh phụ:
  1 ────── 5 ────── 2               1 ── 5 ── 9 ── 13 ── 2
  │                 │               │                    │
  │                 │               7                    8
  7    ẢNH CHÍNH    8               11    ẢNH CHÍNH     12
  │                 │               15                  16
  │                 │               │                    │
  3 ────── 6 ────── 4               3 ── 6 ── 10 ─ 14 ── 4
```

Kích thước khung mỗi ảnh phụ là giá trị nhỏ nhất của:
- 28% canvas (224px với canvas 800);
- ½ cạnh ảnh chính, để **phần giữa ảnh chính luôn nhìn thấy được**;
- khoảng cách giữa hai ảnh phụ kề nhau trên cùng một cạnh, để **các ảnh phụ không bao giờ đè lên nhau**.

Càng nhiều ảnh phụ thì mỗi ảnh càng nhỏ. Ví dụ với canvas 800 và ảnh chính 432×576 (ảnh ghế): 1–4 ảnh phụ có khung 216px, 8 ảnh có khung 216px, 16 ảnh có khung 108px.

Mẹo: dùng **PNG nền trong suốt** cho ảnh phụ (sticker, logo, nhãn giá). Ảnh JPEG có nền trắng sẽ hiện thành một khối trắng trên ảnh chính.

---

## Kích thước ảnh đầu vào

Ảnh gửi vào có thể ở bất kỳ kích thước và tỉ lệ nào; service tự co giãn cho vừa layout:

| Ảnh đầu vào | `grid` / `row` | `center` |
|---|---|---|
| Rất nhỏ (1×1, icon 16×16) | Phóng to cho lấp đầy ô | Phóng to cho vừa vùng ảnh chính / khung ảnh phụ |
| Rất lớn (ảnh điện thoại 48–50MP) | Thu nhỏ rồi cắt giữa vừa ô | Thu nhỏ ngay sau khi decode về tối đa kích thước canvas để tiết kiệm RAM |
| Rất dài/hẹp (1×5000, 5000×1) | Cắt giữa vừa ô | Giữ nguyên tỉ lệ, cạnh ngắn tối thiểu 1px |
| Ảnh dọc / ngang | Cắt giữa vừa ô | Giữ nguyên toàn bộ ảnh, không cắt |
| PNG/WebP nền trong suốt | Phần trong suốt thành nền trắng | Phần trong suốt để lộ ảnh bên dưới |
| Ảnh chụp có EXIF xoay | Tự xoay đúng chiều | Tự xoay đúng chiều |
| Grayscale, paletted, 16-bit, CMYK | Tự chuyển sang RGB | Tự chuyển sang RGB |

Giới hạn (đều trả lỗi rõ ràng thay vì làm sập service):

- Mỗi ảnh tối đa `MAX_PIXELS` (mặc định 50 triệu pixel ≈ 8165×6124); vượt thì `413`.
- Tổng request tối đa `MAX_UPLOAD_MB` (mặc định 100MB); vượt thì `413`.
- File rỗng, hỏng, header báo kích thước 0 hoặc định dạng không hỗ trợ (HEIC, AVIF, SVG…) thì `400`. Hãy chuyển HEIC sang JPEG trước khi gửi.

---

## Cấu hình (biến môi trường)

| Biến | Mặc định | Mô tả |
|---|---|---|
| `PORT_HTTP_SERVER` | `8080` | Cổng HTTP |
| `MAX_UPLOAD_MB` | `100` | Giới hạn kích thước body upload |
| `MAX_IMAGES` | `50` | Số ảnh tối đa mỗi request |
| `MAX_PIXELS` | `50000000` | Số pixel tối đa mỗi ảnh (chống decompression bomb) |
| `JPEG_QUALITY` | `95` | Chất lượng JPEG đầu ra |
| `CELL_WIDTH` / `CELL_HEIGHT` | `100` / `100` | Kích thước ô cho `grid`/`row` |
| `DEFAULT_COLUMNS` | `2` | Số cột mặc định cho `grid` |
| `CENTER_WIDTH` / `CENTER_HEIGHT` | `800` / `800` | Kích thước canvas cho `center` |
| `MINIO_ENDPOINT` | `localhost:9000` | Địa chỉ MinIO |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `minioadmin` / `minioadmin` | Thông tin đăng nhập MinIO |
| `MINIO_BUCKET` | `draw-image-service` | Bucket lưu ảnh (tự tạo khi khởi động nếu chưa có) |
| `MINIO_USE_SSL` | `false` | `true` nếu MinIO dùng HTTPS |
| `MYSQL_HOST` / `MYSQL_PORT` | `localhost` / `3306` | Địa chỉ MySQL |
| `MYSQL_USER` / `MYSQL_PASSWORD` | `root` / _(trống)_ | Thông tin đăng nhập MySQL |
| `MYSQL_DBNAME` | `draw_image_service` | Database (phải tồn tại sẵn; bảng `collages` tự tạo) |

---

## Chạy local

### 1. Khởi động MinIO và MySQL

```sh
docker run -d --name minio -p 9000:9000 -p 9001:9001 \
  -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin \
  minio/minio server /data --console-address ":9001"

docker run -d --name mysql -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=password -e MYSQL_DATABASE=draw_image_service \
  mysql:8.0
```

### 2. Chạy service

```sh
cd draw-image-service
go mod tidy
MYSQL_PASSWORD=password go run ./cmd
```

### 3. Thử

```sh
curl -F center=@chair.jpg -F files=@flash-sale.png "http://localhost:8080/upload?layout=center"
# mở http://localhost:8080/images/<id> trên trình duyệt
```

MinIO console: http://localhost:9001 (xem ảnh trong bucket `draw-image-service`, thư mục `collages/`).

### Ghép thử không cần MinIO/MySQL

`cmd/compose-local` chạy đúng các bước ghép của service trên file local và ghi ra một file JPEG:

```sh
go run ./cmd/compose-local -layout center -out center.jpg chair.jpg flash-sale.png
go run ./cmd/compose-local -layout grid -columns 3 -out grid.jpg a.jpg b.jpg c.jpg
CELL_WIDTH=400 CELL_HEIGHT=400 go run ./cmd/compose-local -layout row -out row.jpg a.jpg b.jpg
```

Với `-layout center`, file đầu tiên là ảnh chính.

### Test

```sh
go vet ./... && go test ./...
```

Test không cần MinIO hay MySQL vì chúng dùng fake cho các port.

---

## Kiến trúc

Hexagonal architecture (ports & adapters), dependency injection bằng [uber-go/fx](https://github.com/uber-go/fx).

```
draw-image-service/
├── cmd/
│   ├── main.go                     # fx.New(infrastructure.Module).Run()
│   └── compose-local/              # CLI ghép ảnh local
├── config/config.go                # đọc biến môi trường
└── internal/
    ├── domain/                     # Collage, Layout, lỗi domain
    ├── port/                       # interface: CollageService, ImageProcessor,
    │                               #            ObjectStorage, CollageRepository
    ├── application/                # CollageService: validate → ghép → MinIO → MySQL
    ├── adapter/
    │   ├── primary/http/           # gin handler + router
    │   └── secondary/
    │       ├── imaging/            # decode, thumbnail, grid, center (disintegration/imaging)
    │       ├── minio/              # ObjectStorage trên MinIO
    │       └── repository/         # CollageRepository trên MySQL (gorm)
    └── infrastructure/             # fx module, HTTP server, kết nối MySQL
```

Luồng `POST /upload`:

1. Handler đọc multipart và đưa file `center` lên đầu danh sách.
2. Application kiểm tra số ảnh và layout, decode từng ảnh (ảnh vượt `MAX_PIXELS` bị từ chối trước khi decode), rồi ghép và encode JPEG.
3. Tạo UUID, upload lên MinIO tại `collages/<uuid>.jpg`.
4. Ghi một dòng vào bảng `collages`. Nếu bước này lỗi, object vừa upload lên MinIO bị xoá để không còn file mồ côi.

Bảng `collages` (tạo tự động bằng gorm `AutoMigrate`):

| Cột | Kiểu | Ghi chú |
|---|---|---|
| `id` | `char(36)` PK | UUID |
| `bucket`, `object_key` | `varchar` | Vị trí ảnh trong MinIO |
| `content_type`, `size`, `width`, `height` | | Thông tin file ảnh |
| `image_count`, `layout`, `columns_count` | | Cách ghép |
| `created_at` | `datetime(3)` | UTC |

---

## Deploy

- **Docker**: `docker build -t draw-image-service ./draw-image-service`. Image chạy bằng user không phải root và mở cổng 8080.
- **Helm**: `chart/draw-image-service` gồm Deployment, Service `draw-image-service-svc` (port 80), Ingress (tắt mặc định) và HPA. Service kết nối `mysql-svc:80` và `minio-svc:9000`.
  - Database `draw_image_service` được tạo bởi `chart/mysql` (`serviceDatabases`).
  - Bucket `draw-image-service` được tạo bởi `chart/minio` (`buckets`).
  - Mật khẩu trong `values.yaml` là giá trị dev; môi trường thật hãy override bằng `--set` hoặc chuyển sang Secret.

  ```sh
  helm template ./chart/draw-image-service | kubectl apply -f -
  ```
- **CI/CD**:
  - `.github/workflows/go.yml` (job `draw_image_service`) chạy build, vet và test.
  - `.github/workflows/draw-image-service-cicd.yml` chạy test, build image multi-arch, push lên `ghcr.io/jieeirosst/draw-image-service`, rồi deploy bằng Helm trên runner self-hosted.
