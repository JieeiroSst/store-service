# book-service

Service quản lý sách (Play Framework 2.8, Scala 2.13): thông tin sách, thể loại, ảnh bìa và nội dung từng chương. Mọi danh sách đều phân trang bằng cursor. Metadata nằm trong **PostgreSQL**; khi chạy local và chạy test thì dùng H2 ở chế độ PostgreSQL. Nội dung chương và ảnh bìa lưu trong **MinIO** (S3).

## API

Đọc (public):

| Method | Path | Mô tả |
|---|---|---|
| GET | `/api/v1/books?cursor=&limit=&q=&author=&year=&category=` | Danh sách sách, mới nhất trước. `q` tìm trong tên sách hoặc tác giả (không phân biệt hoa thường), các filter kết hợp bằng AND |
| GET | `/api/v1/books/:id` | Thông tin sách, kèm `categories`, `coverUrl`, `chapterCount` |
| GET | `/api/v1/books/:id/chapters?cursor=&limit=` | Nội dung sách theo thứ tự đọc, đọc từ MinIO |
| GET | `/api/v1/books/:id/chapters/:number` | Một chương |
| GET | `/api/v1/books/:id/cover` | Ảnh bìa, proxy từ MinIO |
| GET | `/api/v1/categories` | Danh sách thể loại |
| GET | `/api/v1/categories/:slug` | Một thể loại |
| GET | `/api/v1/categories/:slug/books?cursor=&limit=&q=` | Sách thuộc thể loại (cursor) |
| GET | `/health`, `/health/ready` | Liveness; readiness (kiểm tra PostgreSQL và bucket MinIO) |

Ghi (cần header `X-Api-Key` khi có cấu hình `BOOK_ADMIN_API_KEY`):

| Method | Path | Body | Kết quả |
|---|---|---|---|
| POST | `/api/v1/books` | `{"title","author","isbn"?,"description"?,"publishedYear"?,"categories"?:["slug"]}` | `201` + `Location`; `409` nếu trùng ISBN; `400` nếu dữ liệu không hợp lệ hoặc có thể loại không tồn tại |
| PUT | `/api/v1/books/:id` | như POST (thay toàn bộ, kể cả thể loại) | `200` / `404` / `409` |
| DELETE | `/api/v1/books/:id` | | `204`, xoá luôn chương và ảnh bìa trên MinIO |
| PUT | `/api/v1/books/:id/chapters/:number` | `{"title","content"}` (tối đa 1 MiB text) | `201` khi tạo mới, `200` khi thay thế |
| DELETE | `/api/v1/books/:id/chapters/:number` | | `204` |
| PUT | `/api/v1/books/:id/cover` | ảnh dạng raw body, `Content-Type: image/jpeg\|png\|webp\|gif`, tối đa 5 MB | `200` |
| DELETE | `/api/v1/books/:id/cover` | | `204` |
| POST | `/api/v1/categories` | `{"slug","name"}` | `201` / `409` |
| PUT | `/api/v1/categories/:slug` | `{"name"}` | `200` |
| DELETE | `/api/v1/categories/:slug` | | `204`; sách vẫn giữ, chỉ bỏ liên kết |

```sh
curl -X POST localhost:9000/api/v1/books -H 'X-Api-Key: ...' -H 'Content-Type: application/json' \
  -d '{"title":"Dune","author":"Frank Herbert","isbn":"978-0441013593","categories":["science-fiction"]}'
curl -X PUT localhost:9000/api/v1/books/13/chapters/1 -H 'X-Api-Key: ...' -H 'Content-Type: application/json' \
  -d '{"title":"Book One","content":"A beginning is a very delicate time."}'
curl -X PUT localhost:9000/api/v1/books/13/cover -H 'X-Api-Key: ...' -H 'Content-Type: image/jpeg' --data-binary @cover.jpg
```

### Phân trang cursor

```json
{
  "data": [ ... ],
  "pagination": { "limit": 20, "nextCursor": "djE6OA", "hasMore": true }
}
```

- Trang tiếp theo: gửi lại `nextCursor` qua `?cursor=`, giữ nguyên các filter. Khi `hasMore` là `false` thì `nextCursor` là `null`.
- `limit` mặc định 20, tối đa 100. Cursor là chuỗi opaque; cursor không hợp lệ trả về `400`.
- Phân trang theo keyset (`WHERE id < ? ORDER BY id DESC LIMIT n+1`), nên tốc độ không phụ thuộc trang sâu hay nông, và sách mới thêm vào không làm lặp hoặc mất bản ghi giữa các trang.
- Nếu một chương có trong DB nhưng thiếu object trong MinIO, API trả về `503`.

## Lưu nội dung trên MinIO

- Bucket `book-service` (env `MINIO_BUCKET`). Mỗi chương là object `books/<bookId>/chapters/<n>.txt`; ảnh bìa là `books/<bookId>/cover-<timestamp>` (đổi ảnh thì đổi key, nên cache không bao giờ trả ảnh cũ). Database chỉ lưu key, content type và kích thước.
- Khi ghi, service lưu object trước rồi mới ghi row, nên row không bao giờ trỏ tới object chưa tồn tại. Khi xoá thì xoá row trước, sau đó xoá object (best-effort, có log cảnh báo nếu lỗi).
- `coverUrl`: nếu có `MINIO_PUBLIC_ENDPOINT` thì là presigned URL (hiệu lực 1 giờ) để client tải thẳng từ MinIO; nếu không thì là `/api/v1/books/:id/cover`.
- Bucket được tạo bởi job của `chart/minio`. Service cũng tự tạo bucket nếu chưa có.
- Khi khởi động, `ContentSeeder` nạp catalogue mẫu từ `conf/seed/books.json` (9 thể loại, 12 sách, 15 chương) thông qua chính `BookService`: tạo những sách chưa có (so theo ISBN), và upload lại object chương nếu MinIO bị mất dữ liệu. Nếu PostgreSQL hoặc MinIO chưa sẵn sàng thì seeder retry. Tắt bằng `BOOK_SEED_ENABLED=false` (prod đã tắt sẵn).

Biến môi trường: `MINIO_ENDPOINT`, `MINIO_PUBLIC_ENDPOINT`, `MINIO_REGION`, `MINIO_BUCKET`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `DB_DRIVER`, `DB_URL`, `DB_USERNAME`, `DB_PASSWORD`, `EVOLUTIONS_USE_LOCKS`, `BOOK_ADMIN_API_KEY`, `BOOK_SEED_ENABLED`, `APPLICATION_SECRET`.

## Chạy local

```sh
sbt test      # H2 (PostgreSQL mode) + ContentStore in-memory: không cần Postgres hay MinIO
sbt run       # H2 in-memory; cần MinIO ở http://localhost:9000 (minioadmin / minioadmin123)
```

## Triển khai

- Chart: `chart/book-service` (dùng `chart/common`), overlay `values-{dev,staging,prod}.yaml`. Argo CD tự tạo app `book-service-dev`.
- CI trong cluster (`deploy/ci`): push lên `master` có thay đổi trong `book-service/` sẽ chạy `sbt test`, build `localhost:5000/book-service:<tag>`, rồi Image Updater deploy lên `store-dev`. Build tay: `deploy/bootstrap-local.sh ci-run book-service`.
- Staging/prod: chạy Actions → **Book Service CI/CD** (image `ghcr.io/jieeirosst/book-service:<sha7>`), sau đó **Promote image** với `service=book-service`.
- Database: chart `postgres` tạo database `book_service`, role `book_service_svc` và Secret `book-service-db-credentials` trong `default`; `store-dev`/`store-staging` lấy Secret đó qua ExternalSecret. Schema (`conf/evolutions`) tự apply khi khởi động, có lock để nhiều replica không apply cùng lúc. Evolutions viết bằng cú pháp mà cả PostgreSQL và H2 đều chạy được, và Play tách câu lệnh theo dấu `;` nên không được dùng `;` trong comment.
- Secret `book-service-secret` (dev/staging do chart render, prod lấy từ secret manager) chứa MinIO credentials, `APPLICATION_SECRET` và `BOOK_ADMIN_API_KEY`. Key dev: `book-service-dev-admin-key`.
