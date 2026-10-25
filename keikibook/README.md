# Keikibook web

Trang giới thiệu của Keikibook — super app gom ví, đồ ăn, mua sắm, chat, gọi video và 54 mini-app trên 95 dịch vụ của store-service. Dựng bằng Vite + React 18 + TypeScript, dùng chung logo, màu và font Be Vietnam Pro với app Flutter (`../keikiapp`).

## Chạy

```bash
npm ci
npm run dev        # http://localhost:3000
npm run build      # xuất ra build/
npm run preview    # xem bản build ở cổng 3000
```

| Biến môi trường | Ý nghĩa |
| --- | --- |
| `VITE_APK_URL` | link tải APK Android. Bỏ trống thì nút tải hiện trạng thái "sắp có link tải". |
| `VITE_ADSENSE_CLIENT` | mã nhà xuất bản Google AdSense (`ca-pub-…`). Có giá trị thì build chèn script AdSense + thẻ xác minh vào `<head>` và sinh `ads.txt`; bật Auto ads trong AdSense là đủ để hiện quảng cáo. Bỏ trống thì trang không tải gì của Google. |
| `VITE_ADSENSE_SLOT` | ID một đơn vị quảng cáo hiển thị (responsive). Cần cả hai biến thì trang mới hiện 3 ô "Quảng cáo" cố định: sau dải marquee, sau phần tính năng, trước phần tải app. Ô không có quảng cáo sẽ tự ẩn. |

Docker: `docker build --build-arg VITE_APK_URL=https://… --build-arg VITE_ADSENSE_CLIENT=ca-pub-… --build-arg VITE_ADSENSE_SLOT=… -t keikibook .`, chạy nginx ở cổng 3000.

## Tài nguyên dùng chung

- Logo, favicon, `logo192/512.png`, `apple-touch-icon.png` được sinh bởi `../keikiapp/tool/brand/generate.py` từ SVG trong `../keikiapp/assets/brand/`. Đổi logo thì sửa SVG ở đó rồi chạy lại script, không sửa tay ở đây.
- `public/screens/*.webp` là ảnh chụp app thật (585×1266), sinh bằng `flutter test --update-goldens tool/screenshots/screens_test.dart` (màn chính) và `tool/screenshots/apps_test.dart` (mini-app, ảnh ra `screens/apps/`) trong `../keikiapp`, rồi chuyển sang WebP.
- `public/screens/apps/<shot>.webp` là ảnh từng mini-app (390px, `cwebp -q 78 -resize 390 0`) lấy từ `tool/screenshots/screens/apps/` của `../keikiapp`; dữ liệu mẫu cho ảnh nằm ở `tool/screenshots/fixtures/batch*.dart`.
- Toàn bộ nội dung trang (tiếng Việt và tiếng Anh) nằm ở `src/content.ts`; `ui` là chuỗi giao diện, mỗi nội dung khác là `{ vi, en }`. Danh sách mini-app theo nhóm là `groups`, mỗi app có `shot` (tên ảnh), `note` (tóm tắt) và `desc` (mô tả đầy đủ); tên lấy theo `lib/features/services/data/service_meta.dart` của app. Thêm mini-app mới trong app thì thêm một mục ở đây kèm ảnh chụp.
- Section "Vai trò" (`#vai-tro`) giới thiệu hai vai trò khách hàng và quản trị của app; nội dung ở `roles` và `consoles` trong `src/content.ts`. Ảnh `admin-home`, `ticket-organizer`, `ticket-manage-event`, `poly-market` lấy từ `tool/screenshots/apps_test.dart` (fixture `batch6.dart`, `pageShots`). Khi thêm khu quản lý mới ở `lib/features/admin/admin_home_page.dart` thì cập nhật đoạn mô tả này.
- Mục "Bắt đầu" (`#bat-dau`) và "Hỏi đáp" (`#hoi-dap`) lấy nội dung từ `steps` và `faq` trong `src/content.ts`; câu trả lời chỉ nêu điều app thật sự làm được, đổi tính năng thì sửa ở đây. Ô tìm nhanh (⌘K / Ctrl K, nút kính lúp trên nav) tìm trong các mục trang và toàn bộ `groups`, nên mini-app mới tự xuất hiện.
- `public/fonts/` là Be Vietnam Pro (SIL Open Font License, xem `OFL.txt`).
