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

App gọi 1860/1861 operation GraphQL:

- **166 public + 509 user:** màn viết tay trong các mini-app. Op duy nhất không dùng là `carRentalRegisterUser`, vì car-rental-service luôn trả Unimplemented (tài khoản nằm ở user-service).
- **1185 admin:** console quản trị sinh tự động, một console cho mỗi service (68 console). Mở từ tab Quản trị → "Dịch vụ quản trị".

Kiểm tra lại độ phủ bằng cách so tên root field trong `assets/gateway_schema.json` với mã trong `lib/` (trừ `lib/gen`). Test `test/admin_console_test.dart` cũng kiểm tra mọi op admin đều có console.

### Quét mã, deep link và địa chỉ giao hàng

- **Quét QR** (`lib/core/scan/scan.dart`, dùng `mobile_scanner`):
  - mở từ nút quét ở trang chủ, Ví Keiki và trong các ô nhập mã có `KField(..., scannable: true)` (mã ví, yêu cầu thanh toán, voucher, vé gửi xe, mã đặt xe…);
  - ban tổ chức soát vé liên tục bằng camera, mỗi lần quét tự check-in;
  - đọc được mã trong ảnh có sẵn; thiết bị không có camera thì nhập tay.
- **Điều hướng theo nội dung mã:**
  - `keiki://pay?wallet_id=…&amount=…` (QR ở màn Nhận tiền) mở Chuyển tiền đã điền sẵn;
  - `keikibook://app/<đường dẫn>` và `https://keikibook.vn/<đường dẫn>` mở màn tương ứng;
  - nội dung khác hiện ra để sao chép.
- **Deep link từ ngoài app:**
  - Android khai báo `keikibook://`, `keiki://pay` và App Link `https://keikibook.vn` (cần `assetlinks.json` trên domain, xem console "Deep link & đo lường");
  - iOS khai báo hai scheme trên;
  - nếu chưa đăng nhập, đích đến được giữ qua `?from=` và mở ngay sau khi đăng nhập.
- **Ví:** các route `/wallet/*` được bọc `WalletGate`, nên link mở lúc ví đang tải hoặc khi người dùng chưa có ví sẽ hiện màn chờ hoặc hướng dẫn kích hoạt thay vì lỗi.
- **Ảnh:** trường tệp trong form cho chọn Chụp ảnh, Thư viện ảnh hoặc Tệp khác (`image_picker`), dùng cho eKYC và các form tải ảnh.
- **Địa chỉ giao hàng:**
  - chọn Tỉnh → Quận → Phường từ danh mục GHN (`shipProvinces/Districts/Wards`, có tìm kiếm không dấu);
  - mã GHN được lưu vào trường `postcode` của địa chỉ dưới dạng `ghn:<tỉnh>:<quận>:<mã phường>`;
  - **khách trả phí ship:** giỏ hàng gọi `basketShippingQuote`. basket-service tính phí bằng đúng logic của checkout (giá hàng từ catalogue, địa chỉ đọc bằng token người mua, báo giá GHN qua shipping-service) rồi trả về tạm tính, phí ship và tổng;
  - thanh toán gửi kèm `shipping_fee` mà khách đã thấy:
    - server tính lại phí; nếu phí thật cao hơn thì trả 409 và không trừ tiền, app hiện phí mới để khách xác nhận lại;
    - nếu phí thật thấp hơn hoặc bằng thì trừ theo phí thật;
    - ví bị trừ tiền hàng + phí ship, đơn lưu phí ở `shipping_incl_tax`;
  - địa chỉ cũ (tạo trước khi có bộ chọn GHN) được đánh dấu "Cần bổ sung Tỉnh/Quận/Phường" trong danh sách địa chỉ:
    - khi chọn, app tự đoán khu vực từ chữ đã lưu (so khớp không dấu, bỏ tiền tố "TP.", "Quận", "P."…);
    - người dùng kiểm tra lại rồi Lưu; app gọi `updateMyUserAddress` để cập nhật `line2`/`state`/`postcode`, các trường khác giữ nguyên;
    - phần nào đoán không chắc (ví dụ tên trùng một phần) thì để trống cho người dùng tự chọn;
  - GHN không giao tới được thì không cho thanh toán, kèm thông báo cụ thể;
  - tuyến được báo giá (kho + dịch vụ GHN) được khoá đến lúc tạo vận đơn; nếu tuyến đó không còn thì order-processing-service chuyển sang tuyến rẻ nhất;
  - khi đơn được thanh toán, server tự tạo vận đơn GHN cho địa chỉ này;
  - mini-app Vận chuyển cũng báo giá theo đúng phường đã chọn.
- **Kích thước APK:** Android dùng bản ML Kit tải qua Google Play Services (`dev.steenbakker.mobile_scanner.useUnbundled=true`), nên APK nhỏ hơn bản nhúng model khoảng 15 MB.

### Deep link

Mọi màn có thể chia sẻ đều có đường dẫn riêng. Danh mục (53 link) nằm ở `lib/app/deep_links.dart` (`deepLinkCatalog`), và các route nội dung cũng đăng ký ở cùng file đó. Mỗi đường dẫn mở được bằng ba cách:

- `keikibook://app<đường dẫn>`, ví dụ `keikibook://app/event/9`;
- `https://keikibook.vn<đường dẫn>` (App Link Android; tên miền đổi được bằng `--dart-define=WEB_HOST=...`);
- link Branch có `$deeplink_path=<đường dẫn>`, kể cả lần mở đầu sau khi cài app.

| Nhóm | Đường dẫn | Màn | Ví dụ |
| --- | --- | --- | --- |
| Màn chính | `/` | Trang chủ | `/` |
| Màn chính | `/services` | Tất cả dịch vụ | `/services` |
| Màn chính | `/services/:service` | Mở một mini-app | `/services/ticket` |
| Màn chính | `/services/:service?all=1` | Tất cả chức năng của dịch vụ | `/services/movie?all=1` |
| Màn chính | `/profile` | Tài khoản | `/profile` |
| Màn chính | `/profile/edit` | Sửa hồ sơ | `/profile/edit` |
| Màn chính | `/notifications` | Thông báo | `/notifications` |
| Màn chính | `/scan` | Quét mã QR | `/scan` |
| Ví | `/wallet` | Ví Keiki | `/wallet` |
| Ví | `/wallet/transfer?wallet_id=&amount=` | Chuyển tiền (điền sẵn) | `/wallet/transfer?wallet_id=w-123&amount=50000` |
| Ví | `/wallet/receive` | Nhận tiền (QR) | `/wallet/receive` |
| Ví | `/wallet/topup` | Nạp tiền | `/wallet/topup` |
| Ví | `/wallet/withdraw` | Rút tiền | `/wallet/withdraw` |
| Ví | `/wallet/transactions` | Lịch sử giao dịch | `/wallet/transactions` |
| Ví | `/wallet/banks` | Tài khoản ngân hàng | `/wallet/banks` |
| Ví | `/wallet/pin` | Mã PIN ví | `/wallet/pin` |
| Mua sắm & đồ ăn | `/shop` | Mua sắm | `/shop` |
| Mua sắm & đồ ăn | `/shop/products/:id` | Sản phẩm | `/shop/products/12` |
| Mua sắm & đồ ăn | `/shop/cart` | Giỏ hàng | `/shop/cart` |
| Mua sắm & đồ ăn | `/shop/purchases` | Đơn mua | `/shop/purchases` |
| Mua sắm & đồ ăn | `/shop/wishlist` | Yêu thích | `/shop/wishlist` |
| Mua sắm & đồ ăn | `/shop/coupons` | Mã giảm giá | `/shop/coupons` |
| Mua sắm & đồ ăn | `/food` | Đồ ăn | `/food` |
| Mua sắm & đồ ăn | `/food/restaurants/:id` | Nhà hàng | `/food/restaurants/3` |
| Mua sắm & đồ ăn | `/food/orders/:id` | Đơn đồ ăn | `/food/orders/41` |
| Kết nối | `/chat` | Trò chuyện | `/chat` |
| Kết nối | `/chat/assistant` | Trợ lý AI | `/chat/assistant` |
| Kết nối | `/chat/direct/:userId` | Nhắn tin với một người | `/chat/direct/7` |
| Kết nối | `/chat/group/:id` | Nhóm chat | `/chat/group/5` |
| Kết nối | `/room/:id` | Phòng chat | `/room/2` |
| Kết nối | `/thread/:id` | Bài đăng Threads | `/thread/66f0a1` |
| Kết nối | `/post/:id` | Bài viết | `/post/3b9c` |
| Giải trí | `/event/:id` | Sự kiện bán vé | `/event/9` |
| Giải trí | `/ticket-order/:id` | Đơn vé của tôi | `/ticket-order/120` |
| Giải trí | `/prediction/:id` | Sự kiện dự đoán | `/prediction/aff-cup-2026` |
| Giải trí | `/market/:id` | Thị trường dự đoán | `/market/1` |
| Giải trí | `/book/:id` | Sách đọc online | `/book/15` |
| Giải trí | `/store-book/:id` | Sách trong nhà sách | `/store-book/8` |
| Giải trí | `/movie/:id` | Phim | `/movie/mat-biec` |
| Giải trí | `/video/:id` | Video | `/video/a1b2` |
| Giải trí | `/live/:id` | Phòng livestream | `/live/room-1` |
| Giải trí | `/lucky/:game` | Trò chơi may mắn | `/lucky/wheel` |
| Du lịch | `/vehicle/:id` | Xe cho thuê | `/vehicle/veh-01` |
| Du lịch | `/hotel/:id?check_in=&check_out=&guests=` | Khách sạn | `/hotel/4?check_in=2026-12-24&check_out=2026-12-26&guests=2` |
| Du lịch | `/homestay/:id?check_in=&check_out=&guests=` | Homestay | `/homestay/11?guests=4` |
| Du lịch | `/reservation/:id` | Đặt phòng khách sạn | `/reservation/77` |
| Du lịch | `/homestay-booking/:id` | Đặt homestay | `/homestay-booking/31` |
| Du lịch | `/lease/:id` | Hợp đồng thuê | `/lease/5` |
| Tiện ích | `/qr/:id` | Mã QR của tôi | `/qr/qr_123` |
| Tiện ích | `/r/:code` | Link mời bạn bè | `/r/ABC12345` |
| Quản trị | `/admin` | Quản trị (chỉ admin) | `/admin` |
| Quản trị | `/admin/console/:service` | Console quản trị dịch vụ (chỉ admin) | `/admin/console/crm` |
| Quản trị | `/admin/tracking` | Kiểm tra link & tracking (chỉ admin) | `/admin/tracking` |

- **Chia sẻ:**
  - các màn nội dung có nút Chia sẻ (`ShareLinkButton`): sự kiện, đơn vé, dự đoán, thị trường, sách, video, live, bài viết, Threads, phòng chat, khách sạn, homestay, sản phẩm, nhà hàng;
  - khi Branch bật, nút tạo short link Branch (`feature=share`, `$deeplink_path`, `$fallback_url` là link web); không có Branch hoặc lỗi mạng thì dùng link web;
  - bảng chia sẻ có Sao chép và Gửi (qua `share_plus`).
- **Khách sạn/homestay:** link giữ ngày nhận/trả phòng và số khách (`check_in`, `check_out`, `guests`).
- **Link sai:** id sai định dạng, nội dung không còn hoặc đường dẫn không có trong app thì hiện màn "Link không hợp lệ" có nút về trang chủ, thay vì lỗi.
- **Chưa đăng nhập:** đích đến được giữ qua màn đăng nhập (`?from=`).
- **Màn Kiểm tra link & tracking:** liệt kê toàn bộ danh mục theo nhóm; bấm để mở thử, menu bên cạnh để sao chép link web, link app hoặc kiểm tra link mẫu.
- **Thêm deep link mới:**
  1. thêm `DeepLinkSpec` vào `deepLinkCatalog`;
  2. nếu là màn mới, thêm route vào `deepLinkRoutes` (`_content(path, serviceId, builder)` tự bọc màu của mini-app và trả màn "Link không hợp lệ" khi builder trả `null`);
  3. gắn `ShareLinkButton(path: ...)` vào màn;
  4. test `every catalog link opens a real screen` trong `test/deep_links_test.dart` sẽ mở thử mọi link trong danh mục.
- **Phim và xe thuê:** màn chi tiết tải theo id qua `movieVideo` (gateway gọi `GET /api/recommendations/videos/:id` mới của movie-recommendation-service, chỉ trả phim đã sẵn sàng) và `carRentalGetVehicle`.
- **Chuyển giữa hai link cùng loại** (ví dụ `/movie/a` rồi `/movie/b`): route nội dung được gắn key theo URI, nên màn tải lại đúng nội dung mới.

### Branch.io và luồng tracking link

App dùng [`flutter_branch_sdk`](https://pub.dev/packages/flutter_branch_sdk) để nhận link Branch, kể cả deferred deep link sau khi cài app. App gửi nguồn mời về referral-service theo đúng luồng mô tả trong `referral-service/README.md`, mục "Branch.io (deferred deep link)".

```text
link Branch / keikibook:// / QR mời
        │
        ▼
AttributionCubit (lib/core/tracking/)
  ├─ đọc ref_code, click_id, utm_*, ~channel, ~campaign, $deeplink_path
  ├─ mở màn $deeplink_path qua router
  ├─ lưu "đang chờ gửi" vào secure storage (giữ qua lần mở app sau)
  └─ khi có người dùng đăng nhập:
       lần mở đầu sau khi cài (+is_first_session) → referralConfirmInstall
       các trường hợp khác                        → referralActivate
       (gateway tự điền user, app gửi kèm device_id và UTM)
```

- **Lỗi khi gửi:**
  - lỗi vĩnh viễn (tự mời mình, link hết hạn hoặc không tồn tại) thì bỏ link đang chờ;
  - lỗi mạng thì giữ lại để gửi lại ở lần đăng nhập sau.
- **Nhập mã mời tay** (Công cụ giới thiệu) cũng đi qua cubit này.
- **Đăng nhập/đăng xuất:** gọi `setIdentity`/`logout` của Branch.
- **Link từ hệ điều hành:** deep link mặc định của Flutter vẫn bật để `keikibook://` và QR thanh toán chạy cả khi chưa có key Branch. Router bỏ qua domain `*.app.link` (về trang chủ), còn Branch SDK lo điều hướng và ghi nhận cho link Branch.
- **Màn "Kiểm tra link & tracking"** (tab Quản trị → Khu quản lý):
  - dán hoặc quét một link để xem app hiểu link đó thế nào: loại link, màn sẽ mở, mã mời, click_id, UTM;
  - đối chiếu mã mời với referral-service: trạng thái, người mời, cảnh báo tự mời;
  - "Mở như người dùng bấm" chạy link qua đúng luồng; link Branch được gửi cho SDK giải mã (`handleDeepLink`);
  - xem link đang chờ gửi, kết quả lần gửi gần nhất và nguồn cài đặt server đang lưu (`myReferralAttribution`);
  - xem tham số Branch lần cài đầu / gần nhất, chạy `validateSDKIntegration`, xem lịch sử 20 link gần nhất.

**Cấu hình key** (không commit key):

| Nền tảng | Cách đặt |
| --- | --- |
| Dart + Android | `flutter run/build ... --dart-define=BRANCH_KEY=key_live_xxx --dart-define=BRANCH_LINK_DOMAIN=keikibook.app.link`. Gradle đọc thẳng dart-define (hoặc gradle property / biến môi trường cùng tên) để điền `io.branch.sdk.BranchKey` và intent filter App Link cho domain chính cùng domain `-alternate`. Thêm `BRANCH_TEST_MODE=true` để dùng test key. |
| iOS | Chép `ios/Flutter/Branch.xcconfig.example` thành `ios/Flutter/Branch.xcconfig` (đã gitignore) rồi điền key và domain; `Info.plist` đọc `$(BRANCH_KEY)`, `$(BRANCH_LINK_DOMAIN)`. Vẫn phải truyền dart-define như trên để phía Dart bật SDK. |
| CI | Secret `KEIKI_BRANCH_KEY` và biến `KEIKI_BRANCH_LINK_DOMAIN` trong `.github/workflows/keikiapp-build.yml`; job iOS tự sinh `Branch.xcconfig`. |

Không có `BRANCH_KEY` thì SDK không khởi tạo: deep link thường và QR vẫn được ghi nhận, nhưng không có deferred deep link.

Phía Branch dashboard:
- khai báo domain `keikibook.app.link`;
- khai báo package `com.example.keikiapp` cùng SHA-256 của khoá ký (để App Link tự xác minh);
- trỏ `$android_url`/`$ios_url` về store.

Phía server cần đặt `BRANCH_KEY` và `BRANCH_LINK_DOMAIN` cho referral-service, để `/r/:ref_code` chuyển hướng qua Branch.

### Console quản trị

`dart run tool/gen_gateway.dart` sinh cả SDK lẫn console. Phần sinh console nằm ở `tool/gen_admin.dart` và ghi ra `lib/features/admin/console/gen/`. Mỗi op admin được sinh thành một `AdminOp` gồm:

- **Nhóm:** tài nguyên suy ra từ tên op (ví dụ `crmCreateCampaign` → nhóm "Chiến dịch", động từ `create`).
- **Form:**
  - mỗi trường của input type thành một ô nhập;
  - `Int`/`Float` → ô số, `Boolean` → công tắc, enum → danh sách chọn, `Upload` → chọn tệp;
  - danh sách giá trị đơn giản → nhập cách nhau bởi dấu phẩy;
  - kiểu lồng nhau và `Map` → ô JSON có sẵn mẫu.
- **Lời gọi:** gọi SDK có kiểu (`api.<service>.<op>(...)`). Thiếu trường bắt buộc hoặc JSON sai thì báo lỗi dễ hiểu trước khi gửi.

Phần giao diện viết tay nằm ở `lib/features/admin/console/admin_console_page.dart`:

- **Trang console:** thống kê, tìm kiếm, lọc Tra cứu/Thao tác, danh sách nhóm.
- **Trang chức năng:**
  - query không có trường bắt buộc tự tải khi mở;
  - kết quả hiện dạng thẻ hoặc JSON;
  - thao tác xoá/thu hồi/huỷ được tô đỏ.

Console chỉ hiện các op mà tài khoản có grant trong authorize-service (`myPermissions`). Server vẫn kiểm tra quyền ở từng lời gọi. Tên tiếng Việt lấy từ `lib/features/services/data/humanize.dart`; thêm từ vào `_words` nếu một nhóm còn hiện tiếng Anh.

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
  flutter test --update-goldens tool/screenshots/apps_test.dart      # 43 màn của các mini-app dịch vụ và console quản trị, ra tool/screenshots/screens/apps/
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
- test vai trò: khách hàng không thấy tab Quản trị hay nút quản lý, admin (từ `me` hoặc authorize-service) thấy đủ, chế độ xem như khách;
- console quản trị:
  - mọi op admin trong schema đều có console;
  - cả 1185 op chạy được với dữ liệu mẫu và gửi đủ biến bắt buộc;
  - 68 console hiển thị không tràn trên màn hình 360×640;
  - luồng tra cứu và thao tác có form;
- quét mã và deep link:
  - phân tích nội dung QR;
  - nhập tay khi không có camera;
  - QR thanh toán mở Chuyển tiền đã điền sẵn, chưa có ví thì hướng dẫn kích hoạt;
  - giữ đích đến qua màn đăng nhập;
- khu vực GHN: chọn Tỉnh → Quận → Phường và đọc phí từ nhiều dạng báo giá;
- deep link (`test/deep_links_test.dart`):
  - mở lần lượt mọi link trong danh mục và kiểm tra không lỗi, không rơi vào màn "Link không hợp lệ";
  - link sai định dạng hiện màn báo lỗi thân thiện;
  - `keikibook://` và `https://keikibook.vn` mở đúng nội dung;
  - link nhà hàng mở đúng nhà hàng kể cả khi app vừa khởi động;
  - tạo link chia sẻ có và không có Branch;
- tracking link (`test/link_tracking_test.dart`):
  - đọc tham số Branch và URI;
  - link đến trước khi đăng nhập được giữ lại rồi gửi activate sau khi đăng nhập;
  - lần cài đầu dùng confirm-install;
  - chống ghi trùng, phân biệt lỗi vĩnh viễn và lỗi mạng, khôi phục sau khi khởi động lại;
  - màn kiểm tra link đối chiếu với server;
  - link Branch trong app mở đúng màn và được ghi nhận.

Widget test phải truyền `catalogLoader: loadTestCatalog` vào `KeikiApp`, vì đọc asset bằng `compute` không bao giờ hoàn thành trong môi trường fake-async của test.
