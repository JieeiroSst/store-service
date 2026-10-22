# Keikibook — super app

Keikibook (thư mục `keikiapp`) là ứng dụng Flutter gom các dịch vụ của store-service vào một app duy nhất. App chỉ nói chuyện với **gateway-api** (GraphQL, `POST /query`). Phần lớn giao diện được dựng tự động từ schema của gateway, nên mỗi khi gateway có thêm chức năng, app cũng có ngay sau một lần sinh lại code.

## Tính năng

| Phần | Nội dung |
| --- | --- |
| Trang chủ | số dư ví (ẩn/hiện được), nạp/chuyển/nhận tiền nhanh, dịch vụ nổi bật, ưu đãi, dịch vụ dùng gần đây |
| Tất cả dịch vụ | 95 service và 1855 chức năng của gateway; form nhập và phần hiển thị kết quả dựng tự động từ schema |
| 54 mini-app dịch vụ | giao diện viết tay trên SDK sinh ra cho mọi service người dùng cuối: phim, vé sự kiện, vòng quay, xổ số, thị trường dự đoán, video, livestream, sách; homestay, khách sạn, thuê xe, bãi đỗ xe; ví thanh toán, voucher, giới thiệu bạn bè, điểm thưởng, xác thực 3-D Secure; Threads, bài viết, phòng chat; eKYC, phòng gym, tìm kiếm, tải tệp, địa giới hành chính, phí ship, nhà hàng, máy bán hàng… |
| Ví Keiki | số dư, nạp/rút/chuyển tiền, mã PIN, tài khoản ngân hàng, thẻ ảo, yêu cầu thanh toán, QR nhận tiền, điểm thưởng |
| Food | nhà hàng → thực đơn → giỏ → đặt đơn, trả bằng thẻ ví, theo dõi và huỷ đơn |
| Mua sắm | sản phẩm có giá và tồn kho, giỏ hàng lưu trên server, thanh toán bằng ví + PIN, đơn mua, yêu thích, coupon |
| Chat | danh bạ, nhóm, nhắn tin 1-1 và nhóm, trợ lý Keiki AI |
| Tài khoản | hồ sơ, sổ địa chỉ, eKYC, quyền của tôi, cài đặt thông báo |

## Bắt đầu nhanh

Yêu cầu: Flutter **3.47.7** (Dart 3.13) và một gateway-api đang chạy.

```bash
flutter pub get
flutter run --dart-define=API_URL=http://localhost:8080
```

Nếu không truyền `API_URL`, Android emulator gọi `http://10.0.2.2:8080`, các nền tảng khác gọi `http://localhost:8080`. Android chỉ cho phép HTTP không mã hoá tới `10.0.2.2`, `localhost` và `127.0.0.1`; iOS chỉ cho phép trong mạng local.

### Cấu hình (`--dart-define`)

| Khoá | Mặc định | Ý nghĩa |
| --- | --- | --- |
| `API_URL` | xem trên | gốc URL của gateway-api, không gồm `/query` |
| `DEFAULT_CURRENCY` | `VND` | loại tiền của ví chính |
| `WALLET_ISSUER_BANK_ID` | `7a1e0000-0000-4000-8000-000000000001` | ngân hàng phát hành thẻ ảo (Keiki Issuer, tạo bởi migration `003` của wallet-service) |
| `FIREBASE_API_KEY`, `FIREBASE_APP_ID`, `FIREBASE_SENDER_ID`, `FIREBASE_PROJECT_ID` | trống | bật push notification; thiếu một trong bốn khoá thì push tự tắt |
| `FIREBASE_IOS_BUNDLE_ID` | trống | bundle id của app iOS trong Firebase |

## Build

| Nền tảng | Lệnh | Kết quả |
| --- | --- | --- |
| Android | `flutter build apk --release` | `build/app/outputs/flutter-apk/app-release.apk` |
| iOS | `flutter build ios --release` (cần Xcode) | `build/ios/iphoneos/Runner.app` |
| Web | `flutter build web --release` | `build/web/` |

**Android** cần JDK 21 và Android SDK (platform 36, build-tools 36, NDK 28.2). Không bắt buộc phải có Android Studio; có thể dùng command-line tools rồi trỏ Flutter tới đó:

```bash
flutter config --android-sdk ~/Library/Android/sdk --jdk-dir <đường dẫn JDK 21>
flutter doctor
```

Bản release hiện vẫn ký bằng debug key và dùng `applicationId` `com.example.keikiapp`. Trước khi đưa lên Play Store cần đổi `applicationId` trong `android/app/build.gradle.kts` và cấu hình keystore ký release.

**iOS** cần Xcode đầy đủ (từ App Store), deployment target là iOS 15. Nếu máy không có Xcode, dùng CI ở mục dưới.

### CI

`.github/workflows/keikiapp-build.yml` là workflow chạy tay (Actions → Run workflow), gồm ba job:

1. `test`: `flutter analyze` và `flutter test`.
2. `android`: build APK release, đính kèm `keikiapp-android-apk`.
3. `ios`: build trên runner `macos-15` với `--no-codesign`, đính kèm `keikiapp-ios-unsigned` (`Runner.app.zip`, chưa ký nên chưa cài được lên máy thật).

Workflow đọc cấu hình từ repo: variable `KEIKI_API_URL` (hoặc input `api_url`), `KEIKI_IOS_BUNDLE_ID`, và các secret `KEIKI_FIREBASE_API_KEY`, `KEIKI_FIREBASE_APP_ID`, `KEIKI_FIREBASE_SENDER_ID`, `KEIKI_FIREBASE_PROJECT_ID`.

## Kiến trúc

```
lib/
  app/          KeikiApp, router (go_router, 3 tab), theme, danh sách mini-app
  core/
    network/    GraphQL client: refresh token, multipart upload, X-Device-Id
    push/       PushService + FirebasePushTransport
    storage/    token và device id trong flutter_secure_storage
    widgets/    design system (ui.dart), StatusView, empty/error state
  gen/          SDK sinh từ schema — không sửa tay
  features/
    services/   mini-app "Tất cả dịch vụ": catalog, quyền, form theo kiểu, hiển thị kết quả, tên tiếng Việt
    wallet/ food/ shopping/ chat/        mini-app viết tay: data/ bloc/ view/ + <name>_mini_app.dart
    apps/       54 mini-app dịch vụ dùng thẳng SDK `lib/gen/api`
      kit/      bộ khung chung: ServiceApp (header + tab), DetailPage, AsyncList, SearchableList,
                showFormSheet (text/số/ngày/chọn/tệp), runAction, KCard, EntityTile, PosterCard, StatusPill…
      entertainment/ travel/ finance/ community/ utility/
      apps.dart serviceApps: service id → mini-app
    auth/ home/ profile/ address/ notifications/
tool/
  gen_gateway.dart   sinh lib/gen/** và assets/gateway_schema.json
```

- **State:** Bloc/Cubit. `AuthBloc`, `WalletCubit` và `ServicesCubit` sống ở cấp app. Mỗi mini-app bọc route của nó trong một `ShellRoute` riêng; cubit của mini-app được tạo khi vào và huỷ khi rời mini-app đó.
- **Đăng nhập:** access token và refresh token lưu trong `flutter_secure_storage`. Client tự refresh khi token sắp hết hạn hoặc khi gateway trả `UNAUTHENTICATED`/401. Lời gọi refresh gửi không kèm bearer, vì gateway từ chối token hết hạn ngay ở middleware. Refresh thất bại thì app xoá phiên và quay về màn đăng nhập.
- **Quyền:** chức năng có `@hasPermission` chỉ hiện khi `myPermissions` cho phép. Logic so khớp giống authorize-service: `keyMatch` cho object, `regexMatch` cho action. Server vẫn kiểm tra lại mọi lời gọi.
- **Idempotency:** mỗi lần nạp/rút/chuyển/thanh toán gửi một `idempotency_key` mới.

### Vai trò: khách hàng và admin

App có hai vai trò, xác định trong `lib/app/role.dart` (`RoleCubit`):

- **Admin** khi `me.roles` hoặc `myRoles.effectiveRoles` (authorize-service) có `admin` hoặc `super_admin`.
- **Khách hàng** với mọi tài khoản còn lại.

| | Khách hàng | Admin |
| --- | --- | --- |
| Thanh điều hướng | Trang chủ, Dịch vụ, Tài khoản | thêm tab **Quản trị** |
| Mini-app | tính năng tiêu dùng | thêm nút mở khu quản lý (Ban tổ chức, Chủ nhà, Quầy giao xe, Voucher doanh nghiệp, Đối tác Bonuslink, Chuyên mục bài viết, đo lường deep link, phân xử thị trường) |
| Tab Dịch vụ / màn động | ẩn mục Quản trị và mọi chức năng `@hasPermission` | đầy đủ theo quyền được cấp |
| App chỉ cho admin | — | `manage` (SSO Keycloak), chuyến giao tài xế |

Tab Quản trị gom số mục chờ duyệt (sự kiện, khách sạn, homestay), lối vào các khu quản lý và danh sách dịch vụ có chức năng quản trị. Nút "Xem như khách" cho admin xem thử giao diện khách hàng; quay lại bằng mục "Quay lại chế độ quản trị" trong tab Tài khoản. Ẩn giao diện không thay cho phân quyền: gateway và từng service vẫn kiểm tra quyền ở mọi lời gọi.

### Độ phủ API

Màn viết tay gọi 674/675 operation GraphQL không phải admin (165 public, 509 user). Op duy nhất không dùng là `carRentalRegisterUser`: car-rental-service luôn trả Unimplemented vì tài khoản nằm ở user-service. 1185 op admin vẫn dùng màn động "Tất cả chức năng". Kiểm tra lại độ phủ bằng cách so tên root field trong `assets/gateway_schema.json` với mã trong `lib/` (trừ `lib/gen`).

Ngoài phần cho khách, một số mini-app có khu quản lý mở từ nút trên thanh tiêu đề (server vẫn kiểm tra quyền từng lời gọi):

| Mini-app | Khu quản lý |
| --- | --- |
| Vé sự kiện | Ban tổ chức: tạo/sửa/nhân bản/huỷ sự kiện, suất diễn, loại vé, sinh và nhập sơ đồ ghế, áp sơ đồ địa điểm, khuyến mãi, nhân viên, vé mời, đơn và khách tham dự, địa điểm (mẫu, xem trước, chia sẻ), soát vé, kiểm duyệt |
| Khách sạn / Homestay | Chủ nhà: đăng ký, loại phòng, phòng, lịch phòng và giá, khuyến mãi, dịch vụ thêm, báo cáo, nhận/trả phòng; homestay: giá theo kỳ, lịch trống, đặt chỗ và hợp đồng của khách, kiểm duyệt |
| Thị trường dự đoán | Tạo sự kiện, phân xử, thưởng tạo lập thị trường, cấp vốn |
| Voucher | Doanh nghiệp: mua số lượng lớn, ngân sách, phát hàng loạt, đối tác, phát hành, nhập kho mã, thu hồi |
| Thuê xe | Quầy giao xe: giao và nhận lại xe |

### Thêm một mini-app

- **Mini-app dịch vụ (khuyên dùng):** viết một widget dùng `ServiceApp(serviceId: …, tabs: [...])` và các khối trong `lib/features/apps/kit/kit.dart`, gọi API qua `context.api.<service>` (SDK sinh ra), rồi thêm vào `serviceApps` trong `lib/features/apps/apps.dart`. Route `/services/<id>` sẽ mở mini-app này; menu ⋮ → "Tất cả chức năng" (`/services/<id>?all=1`) vẫn mở màn động với mọi chức năng của service. Màu chủ đạo đặt trong `_accents` của kit.
- **Từ schema có sẵn:** thêm `MiniApp.service('<service_id>')` vào `featuredServices` trong `lib/app/mini_apps.dart`. Tên, icon và nhóm lấy từ `lib/features/services/data/service_meta.dart`.
- **Viết tay:** tạo `lib/features/<name>/` gồm `data/`, `bloc/`, `view/`, rồi khai báo `final <name>MiniApp = MiniApp(...)` kèm `routes`. Đăng ký repository trong `lib/app/app.dart` và thêm mini-app vào `miniApps`.

## Thương hiệu và giao diện

- **Logo:** chữ K ghép từ gáy sách và hai trang sách mở, trên nền gradient chàm `#4338CA` → tím `#7C3AED` → hồng `#DB2777`. File nguồn nằm ở `assets/brand/`:
  - `keikibook_icon.svg`: bo góc, dùng cho web và trong app;
  - `keikibook_icon_square.svg`: tràn viền, dùng cho iOS;
  - `keikibook_mark.svg`: foreground cho adaptive icon Android;
  - `keikibook_monochrome.svg`: themed icon Android 13+;
  - `keikibook_background.svg`.
- **Sinh icon cho mọi nền tảng:** Android (cả adaptive và splash), iOS (không alpha), macOS, web Flutter, cùng favicon/logo của web `../keikibook`:

  ```bash
  cd tool/brand && npm install && python3 generate.py
  ```

- **Font:** Be Vietnam Pro (OFL), đóng gói trong `assets/fonts/`.
- **Design system:** `lib/app/theme.dart` (màu, typography, component theme) và `lib/core/widgets/ui.dart` (`KeikiLogo`, `HeroBackdrop`, `QuickAction`, `IconBadge`…).
- **Chụp màn hình** (đăng nhập, trang chủ, dịch vụ, ví, food, shop, tài khoản) với dữ liệu mẫu, font thật, khổ 390×844:

  ```bash
  flutter test --update-goldens tool/screenshots/screens_test.dart   # ảnh ra tool/screenshots/screens/
  flutter test --update-goldens tool/screenshots/apps_test.dart      # 41 màn của các mini-app dịch vụ, ra tool/screenshots/screens/apps/
  ```

  Dữ liệu mẫu cho mini-app nằm ở `tool/screenshots/fixtures/batch*.dart`, khoá theo tên root field GraphQL.

  Công cụ này nằm ngoài `test/` để `flutter test` không so golden. Ảnh dùng để xem lại giao diện và cho trang web `keikibook`.

## SDK sinh từ schema

`tool/gen_gateway.dart` đọc `../gateway-api/graph/*.graphqls` và sinh ra:

- `lib/gen/models.g.dart`: class cho mọi type, input và enum, có `fromJson`/`toJson`;
- `lib/gen/api/<service>.g.dart`: mỗi service một class, mỗi root field một method có kiểu. Selection mặc định lấy sâu 2 cấp, có thể ghi đè bằng tham số `selectionSet`;
- `lib/gen/gateway_api.g.dart`: điểm vào, ví dụ `GatewayApi(client).basket.basketCheckout(...)`;
- `assets/gateway_schema.json`: catalog mà mini-app "Tất cả dịch vụ" đọc lúc chạy.

Mỗi khi schema gateway thay đổi:

```bash
cd ../gateway-api && make generate
cd ../gateway-internal-api && make sync
cd ../keikiapp && dart run tool/gen_gateway.dart
```

Các field trong model đầu ra đều nullable, vì `selectionSet` tuỳ chỉnh có thể bỏ bớt field. Tham số kiểu `Upload` được gửi theo [GraphQL multipart request spec](https://github.com/jaydenseric/graphql-multipart-request-spec) qua lớp `GqlUpload`.

## Luồng nghiệp vụ chính

**Mua sắm.** Khi thêm vào giỏ, basket-service tự lấy giá và kiểm tra tồn kho từ catalogues-service; giá do client gửi lên bị bỏ qua. Khi checkout, basket-service lần lượt:

1. định giá lại giỏ;
2. giữ hàng (nếu có dòng hết hàng thì nhả lại phần đã giữ);
3. trừ tiền ví qua wallet-service với PIN và khoá idempotency theo giỏ, nên bấm lại không bị trừ tiền hai lần;
4. chuyển giỏ sang `Submitted`;
5. khởi động `PaidOrderWorkflow` trên Temporal (workflow id `paid-order-basket-<id>`).

order-processing-service ghi `order_order`/`order_line`, gửi push cho khách rồi chuyển đơn sang `Processing`. Nếu Temporal không khởi động được, cron của basket-service thử lại mỗi phút.

**Food.** Đơn được trả bằng thẻ ảo của Ví Keiki: doordash-service authorize khi đặt (kiểm tra thẻ thuộc về người đặt), capture khi đơn chuyển sang `delivered`, void khi huỷ.

**Chat.** Gửi tin qua `ollamaChatSendMessage`. Tin mới đến theo thời gian thực qua WebSocket `GET /chat/ws` của gateway: gateway yêu cầu đăng nhập rồi chuyển tiếp kèm token tới `/ws` của ollama-service, nơi chỉ gửi tin cho người gửi/người nhận hoặc thành viên hiện tại của nhóm. Khi mất kết nối, app tự kết nối lại (chờ 1, 2, 4… tối đa 30 giây), trong lúc đó hỏi lại lịch sử 15 giây/lần, và tải lại lịch sử ngay khi kết nối lại để không sót tin.

**Phát video.** Phim, Video và Media phát qua đường `GET /media/<service>/...` của gateway. Đường này chỉ cho phép GET/HEAD tới danh sách đường dẫn đã khai báo trong `gateway-api/media/media.go` (stream, HLS, thumbnail của video-service), không chuyển tiếp token hay cookie. App thử HLS (`hls/master.m3u8`) trước, lỗi thì chuyển sang `stream`. Livestream phát thẳng URL HLS đã ký do `livePlayback` trả về và gửi heartbeat 20 giây/lần khi đang xem.

**Gọi video.** App mở WebSocket tới `GET /rtc/ws?room_id=…` của gateway kèm `Authorization: Bearer`. Gateway xác thực phiên rồi tự gắn `user_id` của người đăng nhập trước khi chuyển tiếp tới webrtc-service, nên không thể mạo danh người khác. Cuộc gọi là mesh (tối đa `MAX_PEERS_PER_ROOM`), tương thích với client web có sẵn của webrtc-service.

## Push notification

App đăng ký FCM registration token bằng `registerMyDevice` khi đăng nhập, đăng ký lại khi token đổi, và gọi `unregisterMyDevice` trước khi đăng xuất. Khi app đang mở, thông báo hiện thành snackbar có nút "Xem". Bấm vào thông báo sẽ mở màn tương ứng, theo `route`, `order_id` hoặc `reference_id` trong data.

Để bật push:

1. Tạo app Android và iOS trong Firebase. Lấy `apiKey`, `appId`, `messagingSenderId`, `projectId` (cộng bundle id cho iOS), rồi truyền qua `--dart-define`. Không cần `google-services.json` hay `GoogleService-Info.plist`.
2. iOS: trong Xcode, bật capability **Push Notifications** cho target Runner, rồi upload APNs key vào Firebase. `UIBackgroundModes/remote-notification` đã có sẵn trong `Info.plist`.
3. Backend: cấu hình `firebase.credentials_file` (service account của cùng project) cho notification_service.

App gửi lên registration token chứ không phải Firebase Installation ID, vì notification_service gửi bằng Firebase Admin `SendToToken`.

## Dữ liệu backend cần có

- **wallet-service:** migration `003_keiki_merchants.sql` tạo ngân hàng Keiki Issuer/Acquirer và hai merchant `Keiki Food` (`55555555-…`), `Keiki Shop` (`66666666-…`) ở mọi môi trường.
- **food-service:** nhà hàng (`POST /api/restaurants`, cần quyền admin) và món ăn có `restaurant_id`.
- **catalogues-service:** sản phẩm cần ít nhất một stock record (giá, tồn kho) mới mua được.
- **Temporal và order-processing-service:** cần chạy để đơn mua chuyển sang `Processing`. Worker cần `ORDER_DB_DSN` trỏ tới schema MySQL `basket_service`.

## Kiểm thử

```bash
flutter analyze
flutter test
```

Bộ test gồm:

- GraphQL client: refresh token, multipart upload;
- SDK sinh ra và catalog schema;
- tên tiếng Việt và luật so khớp quyền;
- các cubit (auth, food, giỏ hàng, chat), push service;
- widget test cho trang chủ, màn hình nhỏ và luồng "Tất cả dịch vụ";
- smoke test mở lần lượt 54 mini-app dịch vụ (và mọi tab) cùng 23 trang quản lý/công cụ trên màn hình 360×640;
- test vai trò: khách hàng không thấy tab Quản trị hay nút quản lý, admin (từ `me` hoặc authorize-service) thấy đủ, chế độ xem như khách.

Widget test phải truyền `catalogLoader: loadTestCatalog` vào `KeikiApp`, vì đọc asset bằng `compute` không bao giờ hoàn thành trong môi trường fake-async của test.

## Hạn chế hiện tại

- Gọi video và chat realtime chạy trên Android/iOS/macOS; bản web chưa hỗ trợ vì trình duyệt không gửi được header `Authorization` khi mở WebSocket (chat trên web tự chuyển sang hỏi lại lịch sử 15 giây/lần).
- Chưa tự động tạo vận đơn GHN cho đơn mua, vì cần ánh xạ địa chỉ của user sang mã quận/phường của GHN.
- Bản iOS từ CI chưa được ký, chưa cài lên máy thật được.
