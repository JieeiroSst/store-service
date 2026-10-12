# ejabberd-service

Máy chủ XMPP được port từ ejabberd (Erlang) sang Go, giữ nguyên logic lõi IM:
đăng nhập (STARTTLS, SASL PLAIN / SCRAM-SHA-1/256/512), session, định tuyến
message/presence/iq, roster + subscription, tin nhắn offline, ACL/access rules/shaper,
đăng ký in-band và HTTP API quản trị (`mod_http_api`).

Kiến trúc hexagonal, dependency injection bằng [uber-go/fx](https://github.com/uber-go/fx).

## Mục lục

- [Chạy nhanh](#chạy-nhanh)
- [Cổng và endpoint](#cổng-và-endpoint)
- [Cấu hình](#cấu-hình)
- [Đăng nhập qua user_service](#đăng-nhập-qua-user_service)
- [Kết nối bằng client XMPP](#kết-nối-bằng-client-xmpp)
- [Tích hợp mobile (chat real-time)](#tích-hợp-mobile-chat-real-time)
- [HTTP API](#http-api)
  - [Xác thực và phân quyền](#xác-thực-và-phân-quyền)
  - [Định dạng response](#định-dạng-response)
  - [Danh sách lệnh](#danh-sách-lệnh)
- [Health check](#health-check)
- [Triển khai Kubernetes](#triển-khai-kubernetes)
- [Phát triển](#phát-triển)
- [Giới hạn so với bản Erlang](#giới-hạn-so-với-bản-erlang)

## Chạy nhanh

```bash
# Chạy local với config mặc định (lưu trữ in-memory, mất dữ liệu khi tắt)
make run

# Hoặc bằng Docker
make docker
docker run --rm -p 5222:5222 -p 5223:5223 -p 5280:5280 ejabberd-service
```

Kiểm tra server đã lên:

```bash
curl -s http://127.0.0.1:5280/livez
# {"status":"ok"}

curl -s -4 -X POST http://127.0.0.1:5280/api/status -d '{}'
# "The node ejabberd@localhost is started with status: started\nejabberd dev is running in that node"
```

> Dùng `127.0.0.1` (hoặc `curl -4`). Quyền `public commands` mặc định chỉ khớp
> `127.0.0.1/8`; nếu curl đi qua `::1` bạn sẽ nhận lỗi 403.

## Cổng và endpoint

| Cổng | Listener | Mô tả |
|------|----------|-------|
| 5222 | `ejabberd_c2s` | XMPP client, bắt buộc STARTTLS (`starttls_required: true`) |
| 5223 | `ejabberd_c2s` | XMPP client qua TLS trực tiếp (`tls: true`) |
| 5280 | `ejabberd_http` | `/api/<lệnh>` (HTTP API), `/health`, `/livez` |

Nếu không khai báo `certfiles`, server tự sinh chứng chỉ self-signed cho các `hosts`
khi khởi động (client cần chấp nhận chứng chỉ này khi test).

## Cấu hình

Server đọc file YAML cùng định dạng với `ejabberd.yml` của bản Erlang. Thứ tự tìm file:

1. Biến môi trường `EJABBERD_CONFIG`
2. `/etc/ejabberd/ejabberd.yml`
3. `./ejabberd.yml`

Config mặc định nằm ở [conf/ejabberd.yml](conf/ejabberd.yml). Các option hỗ trợ:
`hosts`, `loglevel`, `language`, `certfiles`, `listen`, `acl`, `access_rules`, `shaper`,
`shaper_rules`, `api_permissions`, `auth_method` (`internal` | `sql` | `user_service`), `user_service`,
`default_db` (`mnesia` = in-memory | `sql`), `auth_password_format` (`plain` | `scram`),
`auth_scram_hash` (`sha` | `sha256` | `sha512`), `resource_conflict`,
`negotiation_timeout`, `registration_timeout`, `sql_*`, `modules`.

Module hỗ trợ: `mod_roster`, `mod_offline`, `mod_ping`, `mod_disco`, `mod_version`,
`mod_register`, `mod_last`, `mod_http_api`, `mod_http_upload` (lưu file ở upload_service). Option/module khác được bỏ qua và ghi cảnh báo
vào log lúc khởi động.

### Biến môi trường (ghi đè file config)

| Biến | Ý nghĩa |
|------|---------|
| `EJABBERD_CONFIG` | Đường dẫn file config |
| `EJABBERD_HOSTS` | Danh sách host, phân tách bằng dấu phẩy (`a.com,b.com`) |
| `EJABBERD_LOGLEVEL` | `debug` \| `info` \| `warning` \| `error` \| `none` |
| `EJABBERD_DEFAULT_DB` | `mnesia` (in-memory) hoặc `sql` |
| `EJABBERD_AUTH_METHOD` | `internal` (in-memory) hoặc `sql` |
| `EJABBERD_CERTFILES` | Danh sách file PEM (cert + key), phân tách bằng dấu phẩy |
| `SQL_SERVER`, `SQL_PORT`, `SQL_DATABASE`, `SQL_USERNAME`, `SQL_PASSWORD` | Kết nối PostgreSQL |
| `USER_SERVICE_URL` | URL HTTP của user_service (ghi đè `user_service.url`) |
| `USER_SERVICE_INTERNAL_TOKEN` | Token gọi `/internal/v1/users/{id}` (chỉ cần khi `jid_field: user_id`) |
| `UPLOAD_SERVICE_URL`, `UPLOAD_SERVICE_KEY` | URL và `SERVICE_API_KEY` của upload_service (nơi lưu file đính kèm) |
| `HTTP_UPLOAD_PUT_URL`, `HTTP_UPLOAD_GET_URL` | URL công khai để client upload/tải file (ghi đè `put_url` / `get_url`) |
| `HTTP_UPLOAD_SECRET` | Khoá ký URL upload (ghi đè `external_secret`) |
| `EJABBERD_ADMIN_USERNAME`, `EJABBERD_ADMIN_PASSWORD`, `EJABBERD_ADMIN_EMAIL` | Tài khoản admin tạo sẵn trong user_service lúc khởi động (thêm vào hoặc ghi đè mục cùng username trong `user_service.admins`) |

### Dùng PostgreSQL

```yaml
default_db: sql
auth_method: sql
auth_password_format: scram
sql_type: pgsql
sql_server: localhost
sql_port: 5432
sql_database: ejabberd
sql_username: ejabberd
sql_password: secret
```

Bảng được tạo tự động khi khởi động, đúng schema `sql/pg.new.sql` của ejabberd Erlang, nên
có thể trỏ thẳng vào database ejabberd cũ:

- Luôn tạo: `rosterusers`, `rostergroups`, `roster_version`, `spool` (dữ liệu chat).
- Chỉ tạo khi `auth_method: sql`: `users` (username, mật khẩu/SCRAM). Với
  `auth_method: user_service`, tài khoản và mật khẩu do user_service quản lý nên
  ejabberd-service không tạo và không dùng bảng này.

### Tạo tài khoản admin

`conf/ejabberd.yml` không định nghĩa ACL `admin`, nên mặc định không ai gọi được các lệnh
quản trị. Thêm vào config:

```yaml
acl:
  admin:
    user:
      - admin@localhost
```

Sau đó tạo tài khoản `admin@localhost` (đăng ký in-band từ máy local, xem bên dưới) và gọi
API bằng Basic auth `admin@localhost:<mật khẩu>`.

Khi dùng `auth_method: user_service`, tài khoản admin phải nằm trong user_service; xem
[Khởi tạo tài khoản admin](#khởi-tạo-tài-khoản-admin), ejabberd-service tự tạo lúc khởi động.

## Đăng nhập qua user_service

Thay vì tự lưu mật khẩu, ejabberd-service có thể xác thực người dùng bằng `user_service`
(tương đương `auth_method: external` / `jwt` của ejabberd Erlang). Người dùng chỉ có một
tài khoản, là tài khoản của app.

```yaml
auth_method: user_service
user_service:
  url: http://user-api-svc        # hoặc biến USER_SERVICE_URL
  jid_field: username             # username | user_id
  allow_password_login: true      # cho phép đăng nhập XMPP bằng mật khẩu user_service
  timeout: 5                      # giây
  cache_ttl: 300                  # giây, cache kết quả "user có tồn tại không"
  # internal_token: ...           # chỉ cần khi jid_field: user_id
  admins:                         # tạo sẵn trong user_service lúc khởi động
    - username: admin
      password: Admin12345        # nên để trống ở đây và đặt EJABBERD_ADMIN_PASSWORD
      email: ejabberd.admin@store.com
default_db: sql                   # roster và tin offline vẫn lưu ở ejabberd-service
```

### Cách xác thực

Client gửi SASL `PLAIN` (trên TLS) với JID và "mật khẩu". Mật khẩu có thể là:

| Mật khẩu XMPP | ejabberd-service làm gì | Dùng cho |
|---------------|------------------------|----------|
| `sessionToken` (JWT) từ `POST /api/v1/login` | Gọi `POST /api/v1/validate`, rồi kiểm tra token thuộc đúng người dùng của JID (claim `username`, hoặc `userId` khi `jid_field: user_id`) | **Mobile** (khuyến nghị) |
| Mật khẩu thật của user_service | Gọi `POST /api/v1/login`, thành công thì gọi `/api/v1/logout` ngay để không để lại session | Client desktop (Gajim, Dino…). Tắt bằng `allow_password_login: false` |

Không lấy được token của người khác để đăng nhập thay họ: JID phải khớp chủ của token.

| `jid_field` | JID | Ví dụ |
|-------------|-----|-------|
| `username` (mặc định) | `<username>@<domain>` | user `alice` → `alice@store.example.com` |
| `user_id` | `<id>@<domain>` | user id 42 → `42@store.example.com` (chỉ đăng nhập bằng token) |

Với `jid_field: username`, username trong user_service phải là localpart XMPP hợp lệ: chữ
thường, không có khoảng trắng hay các ký tự `@ / : " & ' < >`. Username có chữ hoa sẽ được
hạ thành chữ thường trong JID, nhưng user_service tìm user theo đúng chữ hoa/thường, nên
không khớp được. Nếu username có thể là email hoặc chứa ký tự đặc biệt, dùng
`jid_field: user_id`.

### Những gì thay đổi khi bật user_service

- Chỉ có cơ chế `PLAIN`. SCRAM bị tắt vì ejabberd-service không giữ mật khẩu, giống store
  type `external` của ejabberd. Bảng `users` không được tạo.
- Tạo, đổi mật khẩu và xoá tài khoản làm ở user_service (`POST /user/sign-up`…). Đăng ký
  in-band và các lệnh API `register`, `change_password`, `unregister`, `registered_users`
  trả lỗi `Accounts are managed by user_service`.
- `check_account` và việc lưu tin offline hỏi user_service xem người nhận có tồn tại
  (`GET /user?username=`, hoặc `/internal/v1/users/{id}`). Kết quả được cache
  `cache_ttl` (không tồn tại: 1 phút).
- Nếu user_service không phản hồi: đăng nhập thất bại; tin nhắn vẫn được lưu offline (coi
  như người nhận tồn tại) để không mất tin.
- Token chỉ được kiểm tra **lúc kết nối**. Kết nối đang mở không bị ngắt khi token hết hạn
  (mặc định 15 phút), nhưng lần kết nối lại cần token còn hạn: app gọi
  `POST /api/v1/refresh` để lấy token mới trước khi reconnect.
- Tài khoản trong ACL `admin` (Basic auth cho HTTP API) cũng phải tồn tại trong
  user_service; Basic auth chấp nhận cả mật khẩu lẫn token. Dùng `user_service.admins`
  để tự tạo, xem bên dưới.

### Khởi tạo tài khoản admin

Mỗi lần khởi động, với mỗi mục trong `user_service.admins` (và tài khoản từ
`EJABBERD_ADMIN_USERNAME` / `EJABBERD_ADMIN_PASSWORD` / `EJABBERD_ADMIN_EMAIL`),
ejabberd-service chạy nền:

1. Gọi `POST /user/sign-up` của user_service.
   - `200`: tài khoản vừa được tạo.
   - `409` (username đã có): giữ nguyên tài khoản, **không** đổi mật khẩu.
   - `400` (mật khẩu yếu, email sai…): ghi log lỗi và bỏ qua, không thử lại.
   - user_service chưa sẵn sàng (lỗi mạng, `5xx`): thử lại với backoff 2 s → 30 s cho tới
     khi được. Server vẫn phục vụ XMPP bình thường trong lúc chờ.
2. Thử đăng nhập bằng mật khẩu đã cấu hình. Nếu tài khoản đã có sẵn với mật khẩu khác, log
   cảnh báo: Basic auth bằng mật khẩu trong config sẽ thất bại cho tới khi đổi mật khẩu bên
   user_service.
3. Cảnh báo nếu `<username>@<host đầu tiên>` không nằm trong ACL `admin`, vì khi đó tài
   khoản không có quyền admin trên HTTP API.

Ràng buộc của user_service khi tạo tài khoản:

- Mật khẩu phải chứa một chữ hoa theo sau là chữ/số (regex `([A-Z])\w+`), ví dụ
  `Admin12345`.
- Email bắt buộc và phải khớp
  `^[a-z][a-z0-9_.]{5,32}@[a-z0-9]{2,}(\.[a-z0-9]{2,4}){1,2}$`. Ví dụ `ejabberd.admin@store.com`
  hợp lệ; `admin@store.local` không hợp lệ (phần trước `@` quá ngắn và `.local` dài hơn
  4 ký tự).
- Sign-up gán role mặc định qua authorize-service, nên authorize-service cũng phải chạy.

Ví dụ local:

```bash
EJABBERD_AUTH_METHOD=user_service USER_SERVICE_URL=http://127.0.0.1:1235 \
EJABBERD_ADMIN_USERNAME=admin EJABBERD_ADMIN_PASSWORD=Admin12345 \
EJABBERD_ADMIN_EMAIL=ejabberd.admin@store.com make run
# log: "bootstrap admin: created in user_service" username=admin login_ok=true

curl -s -u admin@localhost:Admin12345 -X POST http://127.0.0.1:5280/api/connected_users_number -d '{}'
```

(`conf/ejabberd.yml` chưa khai báo ACL `admin`; thêm `admin: {user: admin@localhost}` vào
`acl` như mục [Tạo tài khoản admin](#tạo-tài-khoản-admin) để tài khoản có quyền admin.)

## Kết nối bằng client XMPP

Dùng client bất kỳ (Gajim, Conversations, Psi, Dino…) với:

- JID: `user@localhost`
- Server: `127.0.0.1`, cổng `5222` (STARTTLS) hoặc `5223` (TLS)
- Chấp nhận chứng chỉ self-signed nếu chưa cấu hình `certfiles`

**Đăng ký in-band** (XEP-0077) chỉ được phép từ mạng `trusted_network`
(mặc định `127.0.0.0/8`), theo `mod_register: ip_access`. Mỗi IP chỉ đăng ký được một tài
khoản trong `registration_timeout` (mặc định 600 giây; đặt `infinity` để tắt).

Các tính năng lõi đã hỗ trợ:

- Roster (RFC 6121), roster versioning (XEP-0237)
- Presence: broadcast, probe, directed presence, subscription
- Message: định tuyến theo priority; lưu offline khi người nhận không online, gửi lại kèm
  `<delay>` "Offline Storage" khi đăng nhập
- `max_user_sessions` và `max_user_offline_messages` theo `shaper_rules`
- Ping (XEP-0199), Service Discovery (XEP-0030), Software Version (XEP-0092),
  Last Activity (XEP-0012, uptime server)

## Tích hợp mobile (chat real-time)

### Nguyên tắc: chat real-time không đi qua HTTP API

Mobile **không gọi HTTP API để chat**. App mở **một kết nối XMPP (TCP) lâu dài** tới
ejabberd-service và giữ nó suốt phiên dùng app. Mọi thứ real-time đều đi trên kết nối đó:
gửi/nhận tin, trạng thái online, lời mời kết bạn, roster push. Server chủ động đẩy stanza
xuống ngay khi có, nên app không cần polling.

HTTP API (`/api/...`) chỉ dành cho **backend của bạn** (cần quyền admin), không nhúng vào
app mobile:

```
                 (1) POST /api/v1/login
   ┌──────────┐ ───────────────────────────►  ┌──────────────┐
   │  Mobile  │ ◄─────────── sessionToken ─── │ user_service │
   └────┬─────┘                               └──────▲───────┘
        │ (2) XMPP TCP 5223/5222, giữ mở suốt phiên  │ (3) POST /api/v1/validate
        │     JID alice@domain, mật khẩu = token     │
        ▼                                            │
   ┌──────────────────┐ ─────────────────────────────┘
   │ ejabberd-service │
   └────────▲─────────┘
            │ HTTP /api (admin): add_rosteritem, send_message, get_offline_count…
   ┌────────┴─────────┐
   │  Backend của bạn │
   └──────────────────┘
```

| Việc | Ai làm | Bằng gì |
|------|--------|---------|
| Đăng ký / đổi mật khẩu / xoá tài khoản | Mobile → user_service | `POST /user/sign-up`… (không cần tạo riêng tài khoản chat) |
| Lấy "mật khẩu" XMPP | Mobile → user_service | `POST /api/v1/login` → `sessionToken` |
| Tạo sẵn quan hệ bạn bè hai chiều | Backend | `add_rosteritem` (gọi 2 lần, mỗi phía một lần, `subs: both`) |
| Gửi thông báo hệ thống vào khung chat | Backend | `send_message` |
| Kiểm tra user có đang online | Backend | `user_resources`, `connected_users` |
| Đếm tin chưa nhận để hiển thị badge | Backend | `get_offline_count` |
| Đăng nhập, gửi/nhận tin, online/offline, kết bạn | **Mobile** | Kết nối XMPP |

Chart k8s bật sẵn [đăng nhập qua user_service](#đăng-nhập-qua-user_service): app đăng
nhập user_service như bình thường, rồi kết nối XMPP với JID `<username>@<domain>` và
**`sessionToken` làm mật khẩu**. Không có tài khoản chat riêng phải tạo hay đồng bộ.

```
POST /api/v1/login {"username":"alice","password":"..."}   →  {"sessionToken":"eyJ...","refreshToken":"...","expiryTime":"900"}
XMPP: JID alice@store.example.com, password = eyJ...        →  <success/>
```

(Nếu tắt user_service thì backend phải tạo tài khoản chat bằng `POST /api/register` và gửi
mật khẩu chat cho app.)

### Thông số kết nối cho app

| Tham số | Giá trị |
|---------|---------|
| Host | Địa chỉ LoadBalancer của `ejabberd-service-svc` (staging/prod). Ở dev, service là `ClusterIP`: dùng `kubectl port-forward svc/ejabberd-service-svc 5222 5223 -n store-dev` |
| Cổng | `5223` (TLS trực tiếp, nên dùng) hoặc `5222` (bắt buộc STARTTLS) |
| XMPP domain | Giá trị `hosts`, ví dụ `dev.store.local`, `store.example.com` |
| JID | `<username>@<domain>`, ví dụ `alice@store.example.com` |
| Xác thực | SASL `PLAIN` trên TLS, mật khẩu = `sessionToken` của user_service. Chỉ khi tắt user_service mới có `SCRAM-SHA-256` |
| Resource | Mỗi thiết bị một giá trị cố định, ví dụ `android-<deviceId>`, `ios-<deviceId>` |
| Chứng chỉ | Prod dùng cert thật (`certfiles`). Dev là self-signed: chỉ ở bản debug mới cho phép bỏ qua kiểm tra cert |

Thư viện XMPP gợi ý (phải hỗ trợ **TCP**, vì server không có WebSocket/BOSH):

| Nền tảng | Thư viện |
|----------|----------|
| Android (Kotlin/Java) | [Smack](https://github.com/igniterealtime/Smack) (`smack-tcp`, `smack-android`, `smack-im`, `smack-extensions`) |
| iOS (Swift/ObjC) | [XMPPFramework](https://github.com/robbiehanson/XMPPFramework) hoặc [Tigase Martin](https://github.com/tigase/Martin) |
| Flutter / React Native | Thư viện XMPP có transport TCP, hoặc native module bọc Smack/XMPPFramework |

### Luồng một phiên chat

Các thư viện trên tự lo bước 1–3. Phần dưới đây là stanza thực tế trên dây, để debug hoặc
khi tự viết client.

**1. Mở kết nối, TLS, xác thực**

```xml
<!-- app → server (cổng 5223 đã là TLS) -->
<stream:stream to='store.example.com' xmlns='jabber:client'
  xmlns:stream='http://etherx.jabber.org/streams' version='1.0'>

<!-- server → app: danh sách cơ chế xác thực -->
<stream:features>
  <mechanisms xmlns='urn:ietf:params:xml:ns:xmpp-sasl'>
    <mechanism>SCRAM-SHA-256</mechanism><mechanism>PLAIN</mechanism> ...
  </mechanisms>
</stream:features>

<!-- app xác thực, nhận <success/>, rồi mở lại stream.
     Với user_service: PLAIN, nội dung = base64("\0" + "alice" + "\0" + sessionToken) -->
<auth xmlns='urn:ietf:params:xml:ns:xmpp-sasl' mechanism='PLAIN'>AGFsaWNlAGV5SmhiR2NpT2k...</auth>
<success xmlns='urn:ietf:params:xml:ns:xmpp-sasl'/>
```

**2. Bind resource cho thiết bị**

```xml
<iq type='set' id='bind1'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><resource>android-8f3a</resource></bind>
</iq>

<iq type='result' id='bind1'>
  <bind xmlns='urn:ietf:params:xml:ns:xmpp-bind'><jid>alice@store.example.com/android-8f3a</jid></bind>
</iq>
```

Nếu cùng tài khoản đăng nhập lại với **cùng resource**, phiên cũ bị ngắt với stream error
`conflict` "Replaced by new connection". Vì vậy mỗi thiết bị cần một resource riêng. Mỗi
tài khoản có tối đa `max_user_sessions` (mặc định 10) phiên cùng lúc.

**3. Lấy danh bạ (roster)**

```xml
<iq type='get' id='roster1'><query xmlns='jabber:iq:roster' ver='abc123'/></iq>
```

Gửi lại `ver` của lần trước: nếu danh bạ không đổi, server trả `<iq type='result'/>` rỗng
và app dùng bản cache.

**4. Báo online. Bước này bắt buộc để nhận tin offline**

```xml
<presence><priority>5</priority></presence>
```

Sau presence đầu tiên (priority ≥ 0), server sẽ:

- Gửi các tin nhắn đến lúc app offline, mỗi tin có
  `<delay xmlns='urn:xmpp:delay' stamp='...'>Offline Storage</delay>`. App dùng `stamp`
  làm thời gian gửi thật.
- Gửi lại các lời mời kết bạn đang chờ.
- Báo cho bạn bè biết user online, và gửi trạng thái online hiện tại của bạn bè xuống app.

Nếu app chưa gửi presence, server coi thiết bị là "chưa sẵn sàng": tin gửi tới JID trần
(`alice@domain`) **không** được đẩy xuống thiết bị đó mà vào hàng offline.

**5. Gửi và nhận tin**

```xml
<!-- gửi -->
<message to='bob@store.example.com' type='chat' id='m-001'>
  <body>Xin chào</body>
</message>

<!-- nhận (server đẩy xuống bất cứ lúc nào) -->
<message from='bob@store.example.com/ios-77c1' to='alice@store.example.com/android-8f3a'
  type='chat' id='m-900'>
  <body>Chào bạn</body>
</message>
```

- Nên gửi tới **JID trần** (`bob@domain`). Server giao tới mọi thiết bị online có priority
  cao nhất của Bob; nếu Bob không online thì lưu offline.
- Luôn đặt `id` duy nhất cho mỗi tin để chống trùng khi gửi lại.
- Server chuyển tiếp nguyên vẹn các phần tử con của `<message>`. Vì vậy app có thể dùng
  "đang gõ" (XEP-0085, `<composing xmlns='http://jabber.org/protocol/chatstates'/>`) và báo
  đã nhận (XEP-0184, `<request xmlns='urn:xmpp:receipts'/>` / `<received .../>`). Tin chỉ
  chứa trạng thái "đang gõ" không được lưu offline.
- Gửi tới tài khoản không tồn tại: nhận lại `<message type='error'>` với
  `service-unavailable`. Hàng offline của người nhận đầy (`max_user_offline_messages`,
  mặc định 100): nhận lỗi `resource-constraint`.

**6. Kết bạn từ app**

```xml
<!-- Alice gửi lời mời -->
<presence to='bob@store.example.com' type='subscribe'/>

<!-- Bob (online, hoặc khi online lại) nhận: -->
<presence from='alice@store.example.com' type='subscribe'/>

<!-- Bob đồng ý, và mời ngược lại để thành hai chiều -->
<presence to='alice@store.example.com' type='subscribed'/>
<presence to='alice@store.example.com' type='subscribe'/>
```

Mỗi thay đổi danh bạ được server đẩy xuống mọi thiết bị của user dưới dạng roster push
(`<iq type='set'><query xmlns='jabber:iq:roster'>...`). App phải trả lời
`<iq type='result' id='...'/>` và cập nhật danh bạ local.

**7. Thoát**

```xml
<presence type='unavailable'/>
</stream:stream>
```

### Gửi file, ảnh (XEP-0363)

File đính kèm **không** đi trong kết nối XMPP. App upload file qua HTTP rồi gửi tin nhắn
chứa link, theo XEP-0363 HTTP File Upload (`mod_http_upload`). ejabberd-service không tự
lưu file: mọi file được chuyển vào **upload_service** (MinIO + MongoDB), với
`receiver_id = "ejabberd:<user>/<secret>"`. Chart `upload-service` khoá prefix
`ejabberd:` (`SERVICE_ONLY_RECEIVER_PREFIXES`) nên chỉ ejabberd-service, dùng service key,
mới đọc được các file này qua upload_service; người nhận tải file qua link do
ejabberd-service cấp.

```
App A ──(1) IQ request slot──► upload.<domain>      (XMPP)
      ◄─────── put URL + get URL ────
App A ──(2) HTTP PUT file ──► ejabberd-service /upload ──► upload_service POST /api/v1/upload
App A ──(3) <message> chứa get URL ──► App B          (XMPP, real-time hoặc offline)
App B ──(4) HTTP GET get URL ──► ejabberd-service /upload ──► upload_service .../content
```

**1. Xin slot** (gửi tới `upload.<domain>`, ví dụ `upload.store.example.com`; app có thể
tìm thấy service này qua `disco#items` của server):

```xml
<iq type='get' id='up1' to='upload.store.example.com'>
  <request xmlns='urn:xmpp:http:upload:0' filename='anh.jpg' size='204800' content-type='image/jpeg'/>
</iq>

<iq type='result' id='up1' from='upload.store.example.com'>
  <slot xmlns='urn:xmpp:http:upload:0'>
    <put url='https://chat.store.example.com/upload/2f1c…/Xk3…/anh.jpg?s=204800&amp;e=1791798622&amp;v=d8d0…'/>
    <get url='https://chat.store.example.com/upload/2f1c…/Xk3…/anh.jpg'/>
  </slot>
</iq>
```

Lỗi khi xin slot:

| Lỗi | Khi nào |
|-----|---------|
| `not-acceptable` + `<file-too-large><max-file-size>20971520</max-file-size>` | `size` lớn hơn `max_size` (chart: 20 MB, bằng giới hạn của upload_service) |
| `bad-request` | thiếu `filename` hoặc `size` không hợp lệ |
| `forbidden` | tài khoản không khớp `access` (mặc định `local`: chỉ user của server) |

**2. Upload:** `PUT` đúng URL `put` (giữ nguyên query string), body là nội dung file,
header `Content-Length` bằng đúng `size` đã khai báo. Slot hết hạn sau 5 giờ và chỉ dùng
được một lần.

| HTTP | Ý nghĩa |
|------|---------|
| `201` | Đã lưu vào upload_service |
| `403` | URL sai, hết hạn hoặc slot đã được dùng |
| `413` | `Content-Length` khác `size` của slot, hoặc file vượt giới hạn của upload_service |
| `415` | upload_service không nhận loại file này. Loại file được nhận diện **theo nội dung**, mặc định chỉ PDF, JPEG, PNG, WebP, GIF (`UPLOAD_ALLOWED_TYPES` của upload-service) |
| `503` | upload_service không phản hồi; app thử lại |

**3. Gửi link** trong tin nhắn bình thường. Thêm `jabber:x:oob` để client hiển thị ảnh
hoặc file thay vì một link trần:

```xml
<message to='bob@store.example.com' type='chat' id='m-42'>
  <body>https://chat.store.example.com/upload/2f1c…/Xk3…/anh.jpg</body>
  <x xmlns='jabber:x:oob'><url>https://chat.store.example.com/upload/2f1c…/Xk3…/anh.jpg</url></x>
</message>
```

Tin này đi như mọi tin khác: online thì đến ngay, offline thì được lưu và giao khi người
nhận online lại.

**4. Tải file:** `GET` (hoặc `HEAD`) URL `get`, không cần đăng nhập. URL chứa chuỗi ngẫu
nhiên 40 ký tự nên không đoán được; ai có link thì tải được, giống ejabberd. Ảnh, video,
audio trả về `Content-Disposition: inline`, các loại khác `attachment`.

Android với Smack (`smack-experimental`) đã có sẵn XEP-0363:

```kotlin
val upload = HttpFileUploadManager.getInstanceFor(conn)
val url: URL = upload.uploadFile(File(path))           // bước 1 + 2
val msg = MessageBuilder.buildMessage()
    .to(JidCreate.entityBareFrom("bob@store.example.com"))
    .ofType(Message.Type.chat)
    .setBody(url.toString())
    .addExtension(StandardExtensionElement.builder("x", "jabber:x:oob")
        .addElement("url", url.toString()).build())
    .build()
conn.sendStanza(msg)                                    // bước 3
```

Cấu hình phía server:

```yaml
upload_service:
  url: http://upload-service-svc          # UPLOAD_SERVICE_URL
  service_key: ...                        # UPLOAD_SERVICE_KEY, = SERVICE_API_KEY của upload_service
listen:
  - port: 5280
    module: ejabberd_http
    request_handlers:
      /api: mod_http_api
      /upload: mod_http_upload
modules:
  mod_http_upload:
    put_url: https://chat.store.example.com/upload   # URL công khai, qua ingress
    max_size: 20971520
    external_secret: ...                  # HTTP_UPLOAD_SECRET; trống thì slot mất hiệu lực khi restart
    # host: upload.@HOST@   access: local   get_url: (mặc định = put_url)
```

### Giữ kết nối sống

Mạng di động hay rớt mà TCP không báo ngay. Cần làm cả hai phía:

**Phía app:** gửi ping mỗi 60 giây (nhiều thư viện có sẵn, ví dụ `PingManager` của Smack).
Không nhận được kết quả trong khoảng 30 giây thì coi như mất kết nối và kết nối lại.

```xml
<iq type='get' id='ping1' to='store.example.com'><ping xmlns='urn:xmpp:ping'/></iq>
<!-- server trả --> <iq type='result' id='ping1' from='store.example.com'/>
```

**Phía server:** bật ping chủ động để dọn phiên "chết" (nếu không, server vẫn tưởng thiết bị
online và giao tin vào một kết nối đã chết):

```yaml
modules:
  mod_ping:
    send_pings: true
    ping_interval: 60       # giây không có dữ liệu thì server ping
    ping_ack_timeout: 32    # không trả lời trong 32 giây...
    timeout_action: kill    # ...thì đóng phiên, tin mới sẽ vào hàng offline
```

App phải trả lời ping của server (`<iq type='result'>`); các thư viện XMPP làm việc này tự
động.

**Kết nối lại:** khi mất kết nối, làm lại bước 1 → 4 (Smack: bật `ReconnectionManager`).
Token user_service thường đã hết hạn sau 15 phút, nên trước khi reconnect app gọi
`POST /api/v1/refresh` lấy `sessionToken` mới và dùng nó làm mật khẩu.
Tin đến trong lúc mất kết nối sẽ được giao khi app gửi presence ở bước 4.

### Khi app chạy nền

Server **chưa có push notification** (XEP-0357 / `mod_push`) và **chưa có stream
management** (XEP-0198). Hệ quả:

- Khi iOS/Android đóng socket lúc app chạy nền, user trở thành offline. Tin mới vào hàng
  offline và chỉ tới khi app mở lại và kết nối.
- Tin gửi vào đúng lúc kết nối vừa chết (trước khi ping phát hiện) có thể bị mất. Bật
  `send_pings` + `timeout_action: kill` để thu hẹp khoảng này. App nên dùng receipt
  (XEP-0184) để đánh dấu tin chưa được xác nhận và gửi lại.
- Muốn thông báo đẩy, backend cần tự gửi qua FCM/APNs. Ví dụ: backend gửi tin bằng
  `send_message`, trước đó gọi `user_resources` để biết người nhận có online không, nếu
  không thì đẩy FCM/APNs. Dùng `get_offline_count` để lấy số badge.

### Ví dụ Android (Smack 4.4)

```kotlin
// sessionToken lấy từ POST /api/v1/login (hoặc /api/v1/refresh) của user_service
val config = XMPPTCPConnectionConfiguration.builder()
    .setXmppDomain("store.example.com")
    .setHost("xmpp.store.example.com")
    .setPort(5222)
    .setSecurityMode(ConnectionConfiguration.SecurityMode.required)
    .setUsernameAndPassword(username, sessionToken)
    .setResource("android-$deviceId")
    .build()

val conn = XMPPTCPConnection(config)
ReconnectionManager.getInstanceFor(conn).enableAutomaticReconnection()
PingManager.getInstanceFor(conn).pingInterval = 60

val chats = ChatManager.getInstanceFor(conn)
chats.addIncomingListener { from, message, _ ->
    showMessage(from.asBareJid().toString(), message.body)
}

conn.connect().login()

chats.chatWith(JidCreate.entityBareFrom("bob@store.example.com")).send("Xin chào")
```

`login()` mặc định gửi presence đầu tiên, nên app sẽ nhận ngay các tin offline qua listener.

Smack tự chọn SASL `PLAIN` khi server chỉ cung cấp `PLAIN`. Khi tự reconnect,
`ReconnectionManager` dùng lại mật khẩu cũ, nên nếu token đã hết hạn thì cần tạo lại
connection với token mới (lắng nghe `ConnectionListener.connectionClosedOnError`, gọi
`/api/v1/refresh`, rồi build lại `XMPPTCPConnectionConfiguration`).

## HTTP API

```
POST http://<host>:5280/api/<lệnh>
Content-Type: application/json

{ "tham_số": "giá trị", ... }
```

Có thể dùng `GET /api/<lệnh>?tham_số=giá_trị`. Lệnh không cần tham số vẫn phải gửi body
`{}` khi dùng POST.

### Xác thực và phân quyền

1. **Không gửi `Authorization`**: request chỉ được xét theo IP nguồn.
2. **Basic auth** `user@host:password`: server kiểm tra mật khẩu. Sai mật khẩu trả 401.

Quyền được quyết định bởi `api_permissions`, giống ejabberd Erlang. Một request được phép
khi có ít nhất một mục mà:

- `from` (nếu có) chứa `mod_http_api`,
- `who` khớp caller (`all`, tên ACL, `ip: ...`, `user: ...`, hoặc `access: {...}`),
- `what` chứa tên lệnh hoặc `"*"`, và không có `"!<lệnh>"`.

Ví dụ trong `conf/ejabberd.yml`:

```yaml
api_permissions:
  "admin access":        # admin dùng Basic auth: mọi lệnh trừ stop/start
    who:
      access:
        allow:
          - acl: admin
    what: ["*", "!stop", "!start"]
  "public commands":     # gọi từ localhost không cần đăng nhập
    who:
      ip: 127.0.0.1/8
    what: [status, connected_users_number]
```

Lưu ý: trong `access` dạng `allow: [acl: a, acl: b]`, các điều kiện được **AND** với nhau
(giống `acl:match_acls` của Erlang). Muốn OR thì tách thành nhiều mục `allow`.

### Định dạng response

| Loại kết quả | Thành công | Thất bại |
|--------------|-----------|----------|
| rescode | `200` body `0` | `200` body `1` |
| restuple / string | `200` body `"chuỗi"` | xem lỗi bên dưới |
| integer | `200` body `5` | |
| list | `200` body `["a","b"]` hoặc mảng object | |

Lỗi trả về JSON:

```json
{"status": "error", "code": 32, "message": "AccessRules: Account does not have the right to perform the operation."}
```

| HTTP | code | Khi nào |
|------|------|---------|
| 400 | — | Body rỗng hoặc không phải JSON (`"400 Bad Request"`), thiếu tham số (`"Missing argument: user"`) |
| 400 | 10 | Tham số không hợp lệ (JID sai, host không phục vụ…) |
| 401 | 10 | Basic auth sai |
| 403 | 32 | Không có quyền theo `api_permissions` |
| 404 | 44 | Lệnh không tồn tại |
| 409 | 10090 | `register` khi tài khoản đã có |

### Danh sách lệnh

Các ví dụ dưới đây gọi với admin:

```bash
API=http://127.0.0.1:5280/api
ADMIN=admin@localhost:adminpw
```

#### Hệ thống

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `status` | — | string |
| `registered_vhosts` | — | danh sách host |
| `stats` | `name`: `registeredusers` \| `onlineusers` \| `onlineusersnode` \| `uptimeseconds` | integer |

```bash
curl -s -u $ADMIN -X POST $API/stats -d '{"name":"registeredusers"}'
# 3
```

#### Tài khoản

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `register` | `user`, `host`, `password` | string, 409 nếu đã tồn tại |
| `unregister` | `user`, `host` | `""` (xoá roster, offline, ngắt session) |
| `check_account` | `user`, `host` | rescode |
| `check_password` | `user`, `host`, `password` | rescode |
| `change_password` | `user`, `host`, `newpass` | rescode |
| `registered_users` | `host` | danh sách username |

```bash
curl -s -u $ADMIN -X POST $API/register \
  -d '{"user":"alice","host":"localhost","password":"secret"}'
# "User alice@localhost successfully registered"

curl -s -u $ADMIN -X POST $API/check_account -d '{"user":"alice","host":"localhost"}'
# 0

curl -s -u $ADMIN -X POST $API/change_password \
  -d '{"user":"alice","host":"localhost","newpass":"newsecret"}'
# 0

curl -s -u $ADMIN -X POST $API/registered_users -d '{"host":"localhost"}'
# ["admin","alice"]
```

#### Session

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `connected_users` | — | danh sách `user@host/resource` |
| `connected_users_number` | — | integer |
| `connected_users_vhost` | `host` | danh sách `user@host/resource` |
| `user_resources` | `user`, `host` | danh sách resource |
| `num_resources` | `user`, `host` | integer |
| `kick_user` | `user`, `host` | số session bị ngắt |

```bash
curl -s -u $ADMIN -X POST $API/connected_users -d '{}'
# ["alice@localhost/phone"]

curl -s -u $ADMIN -X POST $API/kick_user -d '{"user":"alice","host":"localhost"}'
# 1
```

Session bị kick nhận stream error `policy-violation` "Kicked by administrator".

#### Gửi tin

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `send_message` | `type` (`chat` \| `normal` \| `headline` \| `groupchat`), `from`, `to`, `subject`, `body` | rescode |
| `send_stanza` | `from`, `to`, `stanza` (XML) | rescode |

```bash
curl -s -u $ADMIN -X POST $API/send_message -d '{
  "type": "chat",
  "from": "admin@localhost",
  "to": "alice@localhost",
  "subject": "",
  "body": "Xin chào"
}'
# 0

curl -s -u $ADMIN -X POST $API/send_stanza -d '{
  "from": "admin@localhost",
  "to": "alice@localhost",
  "stanza": "<message type=\"headline\"><body>Thông báo</body></message>"
}'
# 0
```

Tin nhắn đi qua router như tin từ client: người nhận offline thì `chat`/`normal` được lưu
offline, `headline` bị bỏ.

#### Roster

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `get_roster` | `user`, `host` | mảng `{jid, nick, subscription, ask, groups}` |
| `add_rosteritem` | `localuser`, `localhost`, `user`, `host`, `nick`, `groups` (mảng) hoặc `group` (chuỗi), `subs` (`none` \| `to` \| `from` \| `both`) | rescode |
| `delete_rosteritem` | `localuser`, `localhost`, `user`, `host` | rescode |

```bash
curl -s -u $ADMIN -X POST $API/add_rosteritem -d '{
  "localuser": "alice", "localhost": "localhost",
  "user": "bob", "host": "localhost",
  "nick": "Bob", "groups": ["Friends"], "subs": "both"
}'
# 0

curl -s -u $ADMIN -X POST $API/get_roster -d '{"user":"alice","host":"localhost"}'
# [{"jid":"bob@localhost","nick":"Bob","subscription":"both","ask":"none","groups":["Friends"]}]
```

Thay đổi roster được push ngay tới các session đang online của `localuser`. Lệnh này chỉ
sửa roster của `localuser`; muốn hai chiều thì gọi thêm một lần cho `bob`.

#### Tin nhắn offline

| Lệnh | Tham số | Kết quả |
|------|---------|---------|
| `get_offline_count` | `user`, `host` | integer |
| `delete_expired_messages` | — | rescode, xoá tin có `jabber:x:expire` đã hết hạn |

```bash
curl -s -u $ADMIN -X POST $API/get_offline_count -d '{"user":"alice","host":"localhost"}'
# 2
```

## Health check

| Endpoint | Ý nghĩa |
|----------|---------|
| `GET /livez` | Tiến trình còn sống (dùng cho liveness/startup probe) |
| `GET /health` | Ping PostgreSQL khi dùng SQL; trả `503` nếu DB không phản hồi (readiness probe) |

Hai endpoint này có trên mọi listener `ejabberd_http` và không cần xác thực.

## Triển khai Kubernetes

Helm chart nằm ở [chart/ejabberd-service](../chart/ejabberd-service):

- Config được render từ `chart/ejabberd-service/files/ejabberd.yml` vào ConfigMap `ejabberd-service-config`;
  sửa qua `ejabberdService.xmpp.*` trong values (`hosts`, `admins`, `apiNetworks`, `certfiles`…).
- Database `ejabberd_service` và role `ejabberd_service_svc` được tạo bởi `chart/postgres`
  (`serviceDatabases`); credentials lấy từ Secret `ejabberd-service-db-credentials`.
- `apiNetworks` cho phép các pod trong dải IP đó gọi `apiNetworkCommands` mà không cần
  đăng nhập (dev mở `10.0.0.0/8`).
- `ejabberdService.userService` bật đăng nhập qua user_service (`http://user-api-svc` cùng
  namespace); NetworkPolicy cho phép đi tới pod `user-api-deployment` cổng 1235. Đặt
  `userService.enabled: false` để quay lại tài khoản riêng của ejabberd (bảng `users`).
- `ejabberdService.httpUpload` bật gửi file: file lưu ở `upload-service-svc`, service key
  đọc từ Secret `upload-service-secret` (key `serviceApiKey`) của chart upload-service.
  Ingress (dev `chat.dev.store.local`, staging/prod `chat.…`) chỉ mở đường dẫn `/upload`,
  giới hạn body 25 MB. Chart upload-service thêm `ejabberd:` vào
  `serviceOnlyReceiverPrefixes`. Khoá ký URL: `ejabberd-service-secret` /
  `HTTP_UPLOAD_SECRET` (prod: `store/prod/ejabberd-service/http-upload`, property `secret`).
- `ejabberdService.userService.admin` khai báo tài khoản admin được tạo lúc khởi động
  (`admin`, khớp `xmpp.admins`). Mật khẩu lấy từ Secret:
  - local/dev/staging: `ejabberd-service-secret` (chart tạo, key `EJABBERD_ADMIN_PASSWORD`,
    giá trị mặc định trong `values.yaml`; nên đổi).
  - prod: secret manager, key `store/prod/ejabberd-service/admin` property `password`,
    đồng bộ vào `ejabberd-service-db-credentials`. Tạo key này trước lần sync đầu.

Gọi API từ trong cluster:

```bash
curl -s -X POST http://ejabberd-service-svc:5280/api/send_message \
  -d '{"type":"chat","from":"bob@dev.store.local","to":"alice@dev.store.local","subject":"","body":"Xin chào"}'
```

Session nằm trong bộ nhớ tiến trình (không có clustering như Erlang), vì vậy chart chạy
**1 replica**.

CI/CD:

- `.github/workflows/go.yml` (job `ejabberd_service`): build, vet, test (có Postgres service).
- `.github/workflows/ejabberd-service-cicd.yml`: chạy tay, build và push `ghcr.io/jieeirosst/ejabberd-service`.
- `deploy/ci/services.yaml`: CI trong cluster build mỗi lần push lên master và rollout
  vào `store-dev`.

## Phát triển

```bash
make build   # go build ./...
make vet     # go vet ./...
make test    # go test -race ./...
```

Test tích hợp Postgres chỉ chạy khi có biến môi trường:

```bash
EJABBERD_TEST_SQL_SERVER=localhost EJABBERD_TEST_SQL_PORT=5432 \
EJABBERD_TEST_SQL_USERNAME=ejabberd EJABBERD_TEST_SQL_PASSWORD=ejabberd \
EJABBERD_TEST_SQL_DATABASE=ejabberd make test
```

Cấu trúc thư mục:

```
cmd/ejabberd-service/         entrypoint (fx.New)
internal/domain/              JID, XML stanza, roster state machine, offline, ACL, shaper, SCRAM/SASL
internal/port/                interface repository, router, session sink
internal/application/         session manager, router, auth, roster, offline, c2s session, module IQ, admin
internal/adapter/primary/     xmpp (c2s listener) và http (mod_http_api)
internal/adapter/secondary/   memory (thay mnesia) và postgres
internal/config/              đọc ejabberd.yml
internal/infrastructure/      wiring fx, TLS, HTTP server, test end-to-end
```

## Giới hạn so với bản Erlang

Chưa port: MUC, PubSub/PEP, MAM, privacy/blocking, carbons, stream management, s2s
(gửi tới domain khác trả `remote-server-not-found`), BOSH/WebSocket, MQTT, STUN/TURN,
web admin, OAuth, CAPTCHA, clustering, và `ejabberdctl`.
