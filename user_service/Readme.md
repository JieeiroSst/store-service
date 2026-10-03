# user_service

User accounts, roles, and the login/session lifecycle for the platform:
sign-up, profile management, role-based access, and a JWT access + rotating
refresh token flow for login/refresh/logout. Exposed over gRPC with an
HTTP/JSON gateway (grpc-gateway) generated from the shared `lib-gateway`
proto definitions.

This service absorbed what used to be two separate services -
`session-cookie-auth-service` (cookie-based session demo) and
`oauth2-service` (OAuth2 authorization-server skeleton) - since neither had
a real, working token store and nothing else in the platform depended on
OAuth2 client/authorization-code flows. What was reusable from both (JWT
access tokens, refresh-token rotation) now lives here as one Redis-backed
implementation; both source services have been removed.

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
main.go                                      entrypoint - cmd.Execute()
cmd/cmd.go                                   cobra "api" command -> fx.New(di.Module).Run()
config/config.go                             env/Consul-based configuration, incl. TokenPolicy
di/
  module.go                                  fx providers: config, db, cache, adapters, services, handler
  server.go                                  fx.Lifecycle: starts/stops the gRPC server + HTTP gateway

internal/
  domain/                                    plain structs, no framework deps
    user.go, role.go, roleitem.go            User, Role, RoleItem entities
    token.go                                 AccessClaims, Session, TokenPair
    errors.go                                sentinel errors + cache key format

  port/
    input/                                   driving ports - what the gRPC adapter calls INTO the app
      auth.go, user.go, role.go, roleitem.go
    output/                                  driven ports - what the app calls OUT to infra
      user_repository.go, role_repository.go, roleitem_repository.go
      token_generator.go, token_store.go, hasher.go

  application/                               business logic, implements the driving ports
    auth/auth_service.go                     login/logout/refresh/validate - see below
    user/user_service.go                     sign-up, profile, cache-through FindUser
    role/role_service.go, roleitem/roleitem_service.go

  adapter/
    inbound/grpcadapter/handler.go           driving adapter - implements the generated UserServiceServer
    outbound/
      pg/                                    driven adapter - GORM/Postgres repositories
      sessionstore/token_store.go            driven adapter - Redis-backed session store
      jwttoken/token_generator.go            driven adapter - JWT access tokens + opaque refresh tokens
      pwhash/hasher.go                       driven adapter - bcrypt
```

Dependencies only ever point inward (adapter -> application -> domain/port).
The application layer never imports gorm, redis, or jwt directly - only the
`port` interfaces - so any of those can be swapped without touching the
login logic itself.

## Login/session lifecycle

1. **Login** (`internal/application/auth`) checks the password (bcrypt) and
   issues an access token (JWT, short TTL) and a refresh token (opaque
   random value, long TTL), persisted together as one session in Redis.
2. **While the access token is valid**, `ValidateSession`/`Authentication`
   check it directly: JWT signature + expiry, plus a store lookup so a
   logout can revoke it before it naturally expires.
3. **Once the access token has expired**, the caller calls `RefreshToken`
   with the refresh token. If it's still valid, a brand new access+refresh
   pair is issued and persisted *before* the old pair is invalidated, so a
   transient store error never leaves the caller with zero valid sessions.
4. **If the refresh token itself is invalid or expired**, `RefreshToken`
   returns `Unauthenticated` - there's nothing left to rotate, so the caller
   has to `Login` again to get a fresh pair.
5. **Logout** deletes the session (both the access- and refresh-token
   entries) from the store - real revocation, not just a client-side
   forget.

TTLs are controlled by `config.TokenPolicy` (`Token.AccessTokenExpMinutes` /
`Token.RefreshTokenExpHours` in Consul/`.env`, defaulting to 15m/168h if
unset). Keep the access TTL well below the refresh TTL - the whole
refresh-then-relogin ordering above depends on the access token expiring
first.

## Features

Beyond the login flow, the service exposes (see
`internal/adapter/inbound/grpcadapter/handler.go` for the full RPC set):

- **Accounts** - `SignUp`, `UpdateProfile`, `FindUser` (Redis cache-through).
- **Roles** - `CreateRole`, `UpdateRole`, `DeleteRole`, `GetRole`, `ListRoles`.
- **Role assignment** - `AddRoleItem`, `UpdateItemRole`, `RemoveRoleItem`.

`LockAccount` and `Authentication` exist in the application/port layer but
aren't currently wired to a gRPC method.

## Running locally

```
# .env points at Consul (HostConsul/KeyConsul/ServiceConsul); adjust as needed
# seed Consul's KV with consul.json (or point HostConsul at your own KV)
go run . api
```

Postgres tables (`users`, `roles`, `user_roles`) are created automatically
via `AutoMigrate` on startup (see `di.newDB` in `di/module.go`).

- gRPC: `:1236` (`Server.PortGrpcServer`)
- HTTP/JSON gateway: `:1235` (`Server.PortHttpServer`)

### Docker

```
docker build -t user-service .
docker run -p 1235:1235 -p 1236:1236 user-service
```

## API cho Mobile (HTTP/JSON)

Mobile gọi qua HTTP/JSON gateway (grpc-gateway). Route và message được sinh
từ `lib-gateway/user-service/service.proto`.

**Base URL**

| Môi trường | Base URL |
|---|---|
| Local | `http://localhost:1235` |
| K8s (ingress, hiện đang tắt trong `chart/user/values.yaml`) | `http://user-api.local` |

### Quy ước chung

- Header: `Content-Type: application/json`.
- **Request body**: nhận cả `snake_case` (`session_token`) lẫn `camelCase`
  (`sessionToken`). Field lạ sẽ bị bỏ qua.
- **Response body**: key luôn là **camelCase** (`sessionToken`,
  `refreshToken`, `expiryTime`...). Field rỗng vẫn được trả về với giá trị
  mặc định (`""`, `0`, `false`, `[]`, `null`).
- Field kiểu `int64` (vd. `expiryTime`) được trả về dạng **string**
  (`"900"`), cần parse sang số ở phía mobile.
- Field timestamp (`createTime`, `updateTime`) là chuỗi RFC 3339, vd.
  `"2026-10-03T08:15:30.123Z"`.
- Gateway **không** kiểm tra header `Authorization`; việc xác thực token do
  các service khác gọi `/api/v1/validate`. Mobile chỉ cần lưu token và dùng
  nó cho các service cần đăng nhập.

### Mã lỗi

Mọi lỗi (của mọi endpoint) đều trả về **cùng một format**:

```json
{
  "code": "AUTH_INVALID_CREDENTIALS",
  "message": "invalid username or password",
  "status": 401
}
```

| Field | Ý nghĩa |
|---|---|
| `code` | Mã lỗi ổn định, **mobile dựa vào field này** để chọn câu hiển thị. Mã không bao giờ bị đổi tên; lỗi mới sẽ có mã mới. |
| `message` | Mô tả tiếng Anh cho dev/log. **Không** hiển thị trực tiếp cho user. |
| `status` | HTTP status, giống status code của response. |

Bảng mã lỗi và câu gợi ý hiển thị:

| `code` | HTTP | Endpoint | Gợi ý hiển thị cho user | Mobile nên làm |
|---|---|---|---|---|
| `AUTH_INVALID_CREDENTIALS` | 401 | login | Tên đăng nhập hoặc mật khẩu không đúng. | Giữ ở màn login, xoá ô mật khẩu |
| `AUTH_REFRESH_TOKEN_INVALID` | 401 | refresh | Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại. | Xoá token, về màn login |
| `AUTH_TOKEN_INVALID` | 401 | (nội bộ) | Phiên đăng nhập đã hết hạn, vui lòng đăng nhập lại. | Thử refresh; refresh lỗi thì về màn login |
| `USER_USERNAME_REQUIRED` | 400 | sign-up | Vui lòng nhập tên đăng nhập. | Báo lỗi ở ô username |
| `USER_USERNAME_TAKEN` | 409 | sign-up | Tên đăng nhập đã được sử dụng. | Báo lỗi ở ô username |
| `USER_INVALID_EMAIL` | 400 | sign-up | Email không hợp lệ. | Báo lỗi ở ô email |
| `USER_WEAK_PASSWORD` | 400 | sign-up | Mật khẩu phải chứa một chữ in hoa, theo sau là chữ hoặc số (vd. Abc123). | Báo lỗi ở ô mật khẩu |
| `USER_NOT_FOUND` | 404 | `PUT /user/{id}`, `GET /user` | Không tìm thấy tài khoản. | |
| `ROLE_NOT_FOUND` | 404 | role, role-item | Không tìm thấy vai trò. | |
| `ROLE_NOT_DEFINED` | 400 | tạo/sửa role | Vai trò chưa được khai báo trong hệ thống phân quyền. | |
| `INVALID_REQUEST` | 400 | tất cả | Dữ liệu gửi lên không hợp lệ. | Lỗi phía app (JSON/kiểu dữ liệu sai) |
| `ROUTE_NOT_FOUND` | 404 | tất cả | Đã có lỗi xảy ra, vui lòng thử lại. | Lỗi phía app (sai URL) |
| `NOT_IMPLEMENTED` | 501 | tất cả | Tính năng đang được phát triển. | |
| `SERVICE_UNAVAILABLE` | 503 | tất cả | Hệ thống đang bận, vui lòng thử lại sau. | Cho phép bấm thử lại |
| `INTERNAL_ERROR` | 500 | tất cả | Đã có lỗi xảy ra, vui lòng thử lại. | |

**Gặp `code` lạ** (mã mới mà app bản cũ chưa biết): hiển thị theo HTTP
status, `4xx` → "Dữ liệu không hợp lệ", `5xx` → "Đã có lỗi xảy ra, vui lòng
thử lại".

### Luồng đăng nhập / token

```
SignUp ──► Login ──► lưu sessionToken + refreshToken
                         │
        gọi API với sessionToken (TTL mặc định 15 phút)
                         │ hết hạn
                         ▼
                 POST /api/v1/refresh ──► 200: thay CẢ HAI token mới
                         │
                         └─► 401: refresh token hết hạn (mặc định 7 ngày) → về màn Login
Logout ──► POST /api/v1/logout (thu hồi phiên phía server) → xoá token local
```

- Refresh token **xoay vòng**: mỗi lần refresh, token cũ (cả access lẫn
  refresh) bị huỷ. Luôn ghi đè cả hai token bằng giá trị mới; không gọi
  refresh song song nhiều request cùng lúc với cùng một refresh token.
- `expiryTime` = số **giây** sống của access token (vd. `"900"`), không
  phải timestamp. Nên tự refresh trước khi hết hạn một chút.
- Access token là JWT (HS256), payload có `sub` (user id), `username`,
  `role`, `roles`, `exp` - mobile có thể decode để lấy user id.

### Auth

#### `POST /api/v1/login` - Đăng nhập

Request:

```json
{ "username": "quanluu", "password": "Abc12345" }
```

Response `200`:

```json
{
  "sessionToken": "eyJhbGciOiJIUzI1NiIs...",
  "refreshToken": "q3V0x...base64url",
  "expiryTime": "900"
}
```

Lỗi: `AUTH_INVALID_CREDENTIALS` (401). Sai username hay sai mật khẩu đều
trả cùng mã này, để không lộ username nào đã tồn tại.

#### `POST /api/v1/refresh` - Làm mới token

Request:

```json
{ "refresh_token": "q3V0x...base64url" }
```

Response `200`:

```json
{
  "newSessionToken": "eyJhbGciOiJIUzI1NiIs...",
  "newRefreshToken": "Zk9p...base64url",
  "expiryTime": "900"
}
```

Lỗi: `AUTH_REFRESH_TOKEN_INVALID` (401) → xoá token, chuyển về màn đăng nhập.

#### `POST /api/v1/logout` - Đăng xuất

Request:

```json
{ "session_token": "eyJhbGciOiJIUzI1NiIs..." }
```

Response `200`: `{ "message": "true" }` nếu đã huỷ phiên, `{ "message": "false" }`
nếu token không tồn tại / đã hết hạn. Cả hai trường hợp mobile đều xoá token local.

#### `POST /api/v1/validate` - Kiểm tra access token

Request:

```json
{ "session_token": "eyJhbGciOiJIUzI1NiIs..." }
```

Response `200`:

```json
{ "valid": true, "userId": "1234567" }
```

Token sai/hết hạn/đã logout vẫn trả `200` với `{ "valid": false, "userId": "" }`.

### Tài khoản

#### `POST /user/sign-up` - Đăng ký

Request:

```json
{
  "username": "quanluu",
  "password": "Abc12345",
  "email": "quanluu@gmail.com",
  "name": "Quan Luu",
  "phone": "0901234567",
  "address": "HCM",
  "sex": "male"
}
```

Ràng buộc:

- `username`: bắt buộc, duy nhất.
- `password`: phải có ít nhất một chữ **in hoa** theo sau là chữ/số/`_`
  (regex `([A-Z])\w+`, vd. `Abc12345`).
- `email`: chữ thường, phần trước `@` bắt đầu bằng chữ cái và dài 6-33 ký
  tự (`a-z0-9_.`), domain dạng `gmail.com` / `abc.com.vn`.
- `address` hiện **không** được lưu khi đăng ký (dùng `PUT /user/{id}` để cập nhật).

Response `200`:

```json
{
  "message": "success",
  "user": {
    "id": 1234567,
    "username": "quanluu",
    "password": "",
    "email": "quanluu@gmail.com",
    "name": "Quan Luu",
    "phone": "0901234567",
    "address": "",
    "sex": "male",
    "checked": true,
    "createTime": "2026-10-03T08:15:30.123Z",
    "updateTime": null,
    "roles": []
  }
}
```

Lỗi: `USER_USERNAME_REQUIRED`, `USER_INVALID_EMAIL`, `USER_WEAK_PASSWORD`
(400), `USER_USERNAME_TAKEN` (409), `SERVICE_UNAVAILABLE` (503, không gán
được role mặc định; tài khoản đã được rollback nên user có thể đăng ký lại).

Đăng ký **không** tự đăng nhập - gọi tiếp `POST /api/v1/login`.

#### `PUT /user/{id}` - Cập nhật hồ sơ

`id` = user id (lấy từ `user.id` khi đăng ký hoặc `sub` trong JWT).
Chỉ gửi các field muốn đổi; field rỗng sẽ được giữ nguyên.

Request `PUT /user/1234567`:

```json
{
  "name": "Quan Luu",
  "email": "quanluu@gmail.com",
  "phone": "0901234567",
  "address": "Ha Noi",
  "sex": "male"
}
```

Response `200`:

```json
{
  "message": "success",
  "user": { "id": 1234567, "name": "Quan Luu", "address": "Ha Noi", "...": "..." }
}
```

`user` trong response chỉ phản ánh các field vừa gửi lên (`username`,
`createTime`... sẽ rỗng), không phải bản ghi đầy đủ trong DB.

Lỗi: `USER_NOT_FOUND` (404) nếu không có user với `id` này.

#### `GET /user` - Lấy thông tin user

Query: `username`, `email`, `page`, `limit` (đều optional).

### Role (dành cho admin, mobile thường không cần)

| Method | Path | Body / Query | Ghi chú |
|---|---|---|---|
| `GET` | `/api/v1/role` | `?page=1&limit=10` | Danh sách role |
| `GET` | `/api/v1/role/{id}` | - | Chi tiết role |
| `POST` | `/api/v1/role` | `{ "name": "admin" }` | `name` phải tồn tại trong authorize-service; response hiện rỗng |
| `PUT` | `/api/v1/role` | `{ "id": 1, "name": "admin" }` | Response hiện rỗng |
| `DELETE` | `/api/v1/role` | `?id=1` | Response hiện rỗng |
| `PUT` | `/api/v1/role-item` | `{ "user_id": 1, "role_id": 2 }` | Đặt role cho user |
| `POST` | `/api/v1/role-item` | `{ "user_id": 1, "role_id": 2 }` | ⚠️ trả `501` (handler đang tên `AddRoleItem`, proto là `AddRole`) |
| `DELETE` | `/api/v1/role-item` | `?user_id=1` | ⚠️ trả `501` (handler đang tên `RemoveRoleItem`, proto là `RemoveRole`) |

`POST /user/refresh` có trong proto nhưng chưa implement (`501`) - dùng
`POST /api/v1/refresh`.

### Ví dụ cURL

```bash
BASE=http://localhost:1235

curl -X POST $BASE/user/sign-up -H 'Content-Type: application/json' \
  -d '{"username":"quanluu","password":"Abc12345","email":"quanluu@gmail.com","name":"Quan Luu","phone":"0901234567","sex":"male"}'

curl -X POST $BASE/api/v1/login -H 'Content-Type: application/json' \
  -d '{"username":"quanluu","password":"Abc12345"}'

curl -X POST $BASE/api/v1/refresh -H 'Content-Type: application/json' \
  -d '{"refresh_token":"<refreshToken>"}'

curl -X POST $BASE/api/v1/logout -H 'Content-Type: application/json' \
  -d '{"session_token":"<sessionToken>"}'
```
