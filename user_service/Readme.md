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
    inbound/ssohttp/handler.go               driving adapter - OIDC login/callback routes (absorbed sso-service)
    outbound/
      oidcprovider/provider.go               driven adapter - OIDC discovery, code exchange, id_token verify
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

#### `GET /api/v1/sso/login` → `GET /api/v1/sso/callback` - Đăng nhập SSO (OIDC)

Gộp từ `sso-service` cũ. Chỉ bật khi `sso.enabled = true` trong Consul
(`issuer`, `client_id`, `client_secret`, `redirect_url`, `scopes`;
`SSO_CLIENT_SECRET` env ghi đè `client_secret`).

1. `GET /api/v1/sso/login` tạo `state` ngẫu nhiên, lưu cookie `sso_state`
   (HttpOnly, 10 phút) rồi `302` sang trang đăng nhập của IdP.
2. IdP redirect về `GET /api/v1/sso/callback?code=...&state=...`. Service so
   `state` với cookie, đổi `code` lấy `id_token`, verify chữ ký/issuer/audience.
3. Tìm user theo `email` (IdP phải trả `email_verified = true`); chưa có thì tạo
   user mới (username = `preferred_username`, trùng thì dùng email; mật khẩu
   ngẫu nhiên) và gán role `user`. User bị khoá → `AUTH_ACCOUNT_LOCKED` (403).
4. Response `200` giống `POST /api/v1/login`:
   `{ "session_token": "...", "refresh_token": "...", "expiry_time": 900 }`.

Lỗi SSO (state sai, code không hợp lệ, email chưa verify, IdP từ chối):
`AUTH_SSO_FAILED` (401).

#### Đăng ký / đăng nhập mạng xã hội (`/api/v1/auth/{provider}/*`)

`provider` ∈ `google`, `facebook`, `twitter`, `microsoft`. Mỗi provider bật riêng
qua `social_login.<provider>.enabled` trong Consul (`client_id`, `client_secret`,
`redirect_url`, `scopes`, `extra_audiences`; Microsoft thêm `tenant`). Secret có
thể ghi đè bằng env `GOOGLE_CLIENT_SECRET`, `FACEBOOK_CLIENT_SECRET`,
`TWITTER_CLIENT_SECRET`, `MICROSOFT_CLIENT_SECRET`. `redirect_url` phải được khai
báo ở console của provider.

| Provider  | Giao thức                    | PKCE | `POST .../token` (mobile/SPA)            |
|-----------|------------------------------|------|------------------------------------------|
| google    | OIDC `accounts.google.com`   | có   | `{"id_token"}` - aud = client_id hoặc `extra_audiences` |
| microsoft | OIDC Entra ID v2.0           | có   | `{"id_token"}` - như trên                 |
| facebook  | OAuth2 + Graph API v21.0     | không| `{"access_token"}` - kiểm `debug_token` app_id |
| twitter   | OAuth2 (X API v2)            | bắt buộc | không hỗ trợ → `NOT_IMPLEMENTED` (501) |

Luồng redirect: `GET /api/v1/auth/{provider}/login` → `302` sang provider (cookie
`sso_state` lưu `state` + PKCE verifier, path `/api/v1/auth/{provider}`) →
`GET /api/v1/auth/{provider}/callback?code&state` → `200` giống `POST /api/v1/login`.

Chỉ **lần đầu** user đồng ý (consent) ở màn hình của provider, thông tin được lưu
một lần vào bảng `social_logins` (unique `(provider, subject)`, SSO Keycloak dùng `provider = sso`):

| Cột | Ý nghĩa |
|-----|---------|
| `user_id` | user nội bộ được liên kết |
| `provider`, `subject` | định danh ở provider (`sub` / Facebook id / X id) |
| `email`, `email_verified` | email provider trả về và đã xác minh hay chưa |
| `name`, `username`, `avatar_url` | hồ sơ công khai (Google `picture`, Facebook `picture`, X `profile_image_url`) |
| `scopes` | quyền user đã cấp, cách nhau bằng dấu cách |
| `profile` | `jsonb` - claims ID token (OIDC) hoặc response `/me` (Facebook, X), tối đa 64KB; không chứa token |
| `consent_ip`, `consented_at` | IP và thời điểm user đồng ý lần đầu |

Các lần đăng nhập sau chỉ đọc bản ghi này để tìm user, **không** ghi/cập nhật lại
(thông tin là ảnh chụp tại thời điểm consent; lịch sử đăng nhập nằm ở
`user_sessions`). Access/refresh token của provider **không** được lưu.

Liên kết tài khoản:

1. Đã có liên kết `(provider, subject)` → đăng nhập vào user đó (kể cả khi email ở
   provider đã đổi).
2. Provider trả email **đã xác minh** → gắn vào user có cùng email, chưa có thì
   tạo user mới (username = username provider → email → `{provider}_{subject}`).
3. Không có email đã xác minh (X không cấp `users.email`, tài khoản công ty
   Microsoft không có `xms_edov`) → luôn tạo user mới, không lưu email, **không
   bao giờ** gắn vào tài khoản có sẵn (chống chiếm tài khoản kiểu nOAuth).

Microsoft `tenant`: `common` / `organizations` / `consumers` (đa tenant, kiểm
`iss` khớp `tid`) hoặc GUID tenant cụ thể.

### Phiên đăng nhập (session lifecycle)

Mỗi lần đăng nhập (mật khẩu, SSO, mạng xã hội) tạo một phiên trong bảng
`user_sessions` (`id`, `provider`, `ip`, `user_agent`, `created_at`,
`last_active_at`, `expires_at`, `revoked_at`, `revoke_reason`); token vẫn nằm ở
Redis để validate nhanh.

- **Refresh**: giữ nguyên `id` phiên, xoay access/refresh token, cập nhật
  `last_active_at` và gia hạn `expires_at` (sliding, `refresh_token_exp_hours`).
  Refresh token cũ dùng lại → `AUTH_REFRESH_TOKEN_INVALID`. Phiên đã thu hồi
  không refresh được.
- **Giới hạn**: `token.max_sessions_per_user` (0 = không giới hạn); vượt quá thì
  phiên ít hoạt động nhất bị thu hồi (`evicted`).
- **Kết thúc**: logout (`logout`), user tự thu hồi (`revoked_by_user`), khoá tài
  khoản thu hồi mọi phiên (`account_locked`); user bị khoá không đăng nhập bằng
  mật khẩu được nữa (`AUTH_ACCOUNT_LOCKED`).
- IP/User-Agent lấy từ `X-Forwarded-For` (phần tử đầu) / `User-Agent`
  (`grpcgateway-user-agent` với gRPC).

Tất cả cần `Authorization: Bearer <session_token>`:

| Method | Path | Mô tả |
|--------|------|-------|
| `GET` | `/api/v1/sessions` | Danh sách phiên còn hiệu lực, `current: true` là phiên đang gọi |
| `DELETE` | `/api/v1/sessions/{id}` | Thu hồi một phiên của chính mình → `204`; không phải của mình → `SESSION_NOT_FOUND` (404) |
| `DELETE` | `/api/v1/sessions` | Đăng xuất mọi thiết bị khác; `?include_current=true` thu hồi cả phiên hiện tại → `{ "revoked": n }` |

```json
{
  "sessions": [
    {
      "id": "pX3...", "provider": "google", "ip": "203.0.113.5",
      "user_agent": "Mozilla/5.0 ...", "created_at": "2026-10-14T09:00:00Z",
      "last_active_at": "2026-10-14T09:30:00Z", "expires_at": "2026-10-21T09:30:00Z",
      "current": true
    }
  ]
}
```

Token phát hành trước bản này (chưa có `id` phiên) vẫn dùng được; lần refresh đầu
tiên sẽ chuyển thành phiên có theo dõi.

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

#### `GET /user` - Tìm user

Query: `username`, `email`, `page` (mặc định 1), `limit` (mặc định 10, tối đa 100), đều optional.
Không có `username`/`email` thì trả danh sách user theo trang.

Response `200`:

```json
{ "users": [ { "id": 1234567, "username": "quanluu", "email": "quanluu@gmail.com", "...": "..." } ], "total": 1 }
```

`password` luôn rỗng.

### Internal (service-to-service, không dành cho mobile)

#### `GET /internal/v1/users/{id}` - Lấy user theo id

Dùng bởi `sard-service/customer-info-service`. Không có trong proto
`lib-gateway`, được đăng ký thẳng vào gateway mux (`internal/adapter/inbound/internalhttp`).
Khi `internal.api_tokens` (Consul) hoặc `INTERNAL_API_TOKENS` (env, phân tách bằng dấu phẩy)
được set thì bắt buộc header `Authorization: Bearer <token>` hoặc `X-Internal-Token`.
Không set thì endpoint mở (chỉ nên dùng khi gateway không có ingress).

Response `200`:

```json
{
  "id": 1234567, "username": "quanluu", "email": "quanluu@gmail.com", "name": "Quan Luu",
  "phone": "0901234567", "address": "HCM", "sex": "male", "active": true,
  "create_time": "2026-10-03T08:15:30.123Z"
}
```

Lỗi: `USER_NOT_FOUND` (404), `INVALID_REQUEST` (400), `AUTH_TOKEN_INVALID` (401).

#### `GET /api/v1/users/{id}/profile` - Hồ sơ công khai của user

Dùng bởi `wishlists-service` để hiển thị chủ wishlist. Xác thực bằng chính access token
của user (`session_token` nhận từ `POST /api/v1/login`) qua header `Authorization: Bearer <session_token>`,
được kiểm tra như `POST /api/v1/validate`. Service gọi service có thể dùng `X-Internal-Token`
(chỉ khi `INTERNAL_API_TOKENS` được set). Không có token hợp lệ thì luôn trả `401`, kể cả khi
`INTERNAL_API_TOKENS` chưa set.

Response `200` (chỉ các trường công khai, không có email/phone/address):

```json
{ "id": 1234567, "username": "quanluu", "name": "Quan Luu", "active": true }
```

Lỗi: `USER_NOT_FOUND` (404), `INVALID_REQUEST` (400), `AUTH_TOKEN_INVALID` (401).

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

## Two-factor authentication (TOTP / Google Authenticator)

authorize-service owns the second factor (`authorize.mfa.v1.MfaService`, see
`authorize_service/api/mfa/mfa.proto`): the AES-256-GCM encrypted secret in its
`user_totps` table, QR code, recovery codes, replay protection (each 30s step is
accepted once) and the per-user failure throttle (5 / 15 min). The encryption key
is authorize-service's `MFA_ENCRYPTION_KEY` (its `.env`; Consul `mfa.encryption_key`
as fallback). user-service never stores or sees the secret after setup; it only:

- authenticates the caller (Bearer access token) and forwards its user id,
- gates every login (password, SSO, social) on `GetTOTPStatus`, and keeps the
  short-lived login challenge (`mfa_token`, Redis) until the code is verified.

If authorize-service is unreachable, logins fail with `SERVICE_UNAVAILABLE` rather
than skipping the second factor. `internal/adapter/outbound/authorizeclient/proto/mfa.proto`
is a copy of the authorize proto; regenerate with `make gen-authorize-client`.

Enrollment (Bearer access token required):

| Method | Path | Body | Result |
|---|---|---|---|
| GET  | `/api/v1/mfa/totp` | - | `{enabled, method, enabled_at, recovery_codes_remaining}` |
| POST | `/api/v1/mfa/totp/setup` | - | `{secret, otpauth_uri, qr_code (PNG data URI), ...}` |
| POST | `/api/v1/mfa/totp/enable` | `{code}` | `{recovery_codes: [10]}` - shown once |
| POST | `/api/v1/mfa/totp/recovery-codes` | `{code}` (TOTP only) | new `recovery_codes` |
| POST | `/api/v1/mfa/totp/disable` | `{code}` (TOTP or recovery) | 204 |

Login with 2FA enabled: the login call fails with `401 AUTH_MFA_REQUIRED` and
`details: {mfa_token, mfa_method: "totp", expires_in}` (over gRPC the same values are
in the `ErrorInfo` metadata). Finish with:

```
POST /api/v1/mfa/totp/verify   {"mfa_token": "...", "code": "123456"}   -> LoginResponse
```

`code` may also be a recovery code (`xxxxx-xxxxx`), which is consumed. The `mfa_token`
lives 5 minutes and is single-use:

```json
"mfa": { "challenge_ttl_seconds": 300 }
```
