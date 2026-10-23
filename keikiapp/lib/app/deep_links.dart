import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../core/scan/scan.dart';
import '../features/admin/console/admin_console_page.dart';
import '../features/apps/commerce/commerce_more.dart';
import '../features/apps/community/community_apps.dart';
import '../features/apps/entertainment/books_apps.dart';
import '../features/apps/entertainment/lucky_app.dart';
import '../features/apps/entertainment/media_apps.dart';
import '../features/apps/entertainment/movie_app.dart';
import '../features/apps/entertainment/polymarket_app.dart';
import '../features/apps/entertainment/polymarket_more.dart';
import '../features/apps/entertainment/ticket_app.dart';
import '../features/apps/entertainment/ticket_more.dart';
import '../features/apps/kit/kit.dart';
import '../features/apps/travel/mobility_apps.dart';
import '../features/apps/travel/stays_apps.dart';
import '../features/apps/travel/stays_more.dart';
import '../features/tracking/link_tracking_page.dart';

class DeepLinkSpec {
  const DeepLinkSpec(this.pattern, this.title, this.example, {required this.group, this.service, this.admin = false});

  final String pattern;
  final String title;
  final String example;
  final String group;
  final String? service;
  final bool admin;

  String get webUrl => webLinkFor(example);

  String get appUrl => appLinkFor(example);
}

const deepLinkCatalog = <DeepLinkSpec>[
  DeepLinkSpec('/', 'Trang chủ', '/', group: 'Màn chính'),
  DeepLinkSpec('/services', 'Tất cả dịch vụ', '/services', group: 'Màn chính'),
  DeepLinkSpec('/services/:service', 'Mở một mini-app', '/services/ticket', group: 'Màn chính'),
  DeepLinkSpec('/services/:service?all=1', 'Tất cả chức năng của dịch vụ', '/services/movie?all=1', group: 'Màn chính'),
  DeepLinkSpec('/profile', 'Tài khoản', '/profile', group: 'Màn chính'),
  DeepLinkSpec('/profile/edit', 'Sửa hồ sơ', '/profile/edit', group: 'Màn chính'),
  DeepLinkSpec('/notifications', 'Thông báo', '/notifications', group: 'Màn chính'),
  DeepLinkSpec('/scan', 'Quét mã QR', '/scan', group: 'Màn chính'),
  DeepLinkSpec('/wallet', 'Ví Keiki', '/wallet', group: 'Ví'),
  DeepLinkSpec('/wallet/transfer?wallet_id=&amount=', 'Chuyển tiền (điền sẵn)', '/wallet/transfer?wallet_id=w-123&amount=50000', group: 'Ví'),
  DeepLinkSpec('/wallet/receive', 'Nhận tiền (QR)', '/wallet/receive', group: 'Ví'),
  DeepLinkSpec('/wallet/topup', 'Nạp tiền', '/wallet/topup', group: 'Ví'),
  DeepLinkSpec('/wallet/withdraw', 'Rút tiền', '/wallet/withdraw', group: 'Ví'),
  DeepLinkSpec('/wallet/transactions', 'Lịch sử giao dịch', '/wallet/transactions', group: 'Ví'),
  DeepLinkSpec('/wallet/banks', 'Tài khoản ngân hàng', '/wallet/banks', group: 'Ví'),
  DeepLinkSpec('/wallet/pin', 'Mã PIN ví', '/wallet/pin', group: 'Ví'),
  DeepLinkSpec('/shop', 'Mua sắm', '/shop', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/shop/products/:id', 'Sản phẩm', '/shop/products/12', group: 'Mua sắm & đồ ăn', service: 'catalogues'),
  DeepLinkSpec('/shop/cart', 'Giỏ hàng', '/shop/cart', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/shop/purchases', 'Đơn mua', '/shop/purchases', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/shop/wishlist', 'Yêu thích', '/shop/wishlist', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/shop/coupons', 'Mã giảm giá', '/shop/coupons', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/food', 'Đồ ăn', '/food', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/food/restaurants/:id', 'Nhà hàng', '/food/restaurants/3', group: 'Mua sắm & đồ ăn', service: 'food'),
  DeepLinkSpec('/food/orders/:id', 'Đơn đồ ăn', '/food/orders/41', group: 'Mua sắm & đồ ăn'),
  DeepLinkSpec('/chat', 'Trò chuyện', '/chat', group: 'Kết nối'),
  DeepLinkSpec('/chat/assistant', 'Trợ lý AI', '/chat/assistant', group: 'Kết nối'),
  DeepLinkSpec('/chat/direct/:userId', 'Nhắn tin với một người', '/chat/direct/7', group: 'Kết nối'),
  DeepLinkSpec('/chat/group/:id', 'Nhóm chat', '/chat/group/5', group: 'Kết nối'),
  DeepLinkSpec('/room/:id', 'Phòng chat', '/room/2', group: 'Kết nối', service: 'room'),
  DeepLinkSpec('/thread/:id', 'Bài đăng Threads', '/thread/66f0a1', group: 'Kết nối', service: 'threads'),
  DeepLinkSpec('/post/:id', 'Bài viết', '/post/3b9c', group: 'Kết nối', service: 'post'),
  DeepLinkSpec('/event/:id', 'Sự kiện bán vé', '/event/9', group: 'Giải trí', service: 'ticket'),
  DeepLinkSpec('/ticket-order/:id', 'Đơn vé của tôi', '/ticket-order/120', group: 'Giải trí', service: 'ticket'),
  DeepLinkSpec('/prediction/:id', 'Sự kiện dự đoán', '/prediction/aff-cup-2026', group: 'Giải trí', service: 'polymarket'),
  DeepLinkSpec('/market/:id', 'Thị trường dự đoán', '/market/1', group: 'Giải trí', service: 'polymarket'),
  DeepLinkSpec('/book/:id', 'Sách đọc online', '/book/15', group: 'Giải trí', service: 'book'),
  DeepLinkSpec('/store-book/:id', 'Sách trong nhà sách', '/store-book/8', group: 'Giải trí', service: 'book_store'),
  DeepLinkSpec('/movie/:id', 'Phim', '/movie/mat-biec', group: 'Giải trí', service: 'movie'),
  DeepLinkSpec('/video/:id', 'Video', '/video/a1b2', group: 'Giải trí', service: 'video'),
  DeepLinkSpec('/live/:id', 'Phòng livestream', '/live/room-1', group: 'Giải trí', service: 'livestream'),
  DeepLinkSpec('/lucky/:game', 'Trò chơi may mắn', '/lucky/wheel', group: 'Giải trí', service: 'lucky'),
  DeepLinkSpec('/vehicle/:id', 'Xe cho thuê', '/vehicle/veh-01', group: 'Du lịch', service: 'car_rental'),
  DeepLinkSpec('/hotel/:id?check_in=&check_out=&guests=', 'Khách sạn', '/hotel/4?check_in=2026-12-24&check_out=2026-12-26&guests=2', group: 'Du lịch', service: 'reservation'),
  DeepLinkSpec('/homestay/:id?check_in=&check_out=&guests=', 'Homestay', '/homestay/11?guests=4', group: 'Du lịch', service: 'rent_house'),
  DeepLinkSpec('/reservation/:id', 'Đặt phòng khách sạn', '/reservation/77', group: 'Du lịch', service: 'reservation'),
  DeepLinkSpec('/homestay-booking/:id', 'Đặt homestay', '/homestay-booking/31', group: 'Du lịch', service: 'rent_house'),
  DeepLinkSpec('/lease/:id', 'Hợp đồng thuê', '/lease/5', group: 'Du lịch', service: 'rent_house'),
  DeepLinkSpec('/qr/:id', 'Mã QR của tôi', '/qr/qr_123', group: 'Tiện ích', service: 'qr'),
  DeepLinkSpec('/r/:code', 'Link mời bạn bè', '/r/ABC12345', group: 'Tiện ích'),
  DeepLinkSpec('/admin', 'Quản trị', '/admin', group: 'Quản trị', admin: true),
  DeepLinkSpec('/admin/console/:service', 'Console quản trị dịch vụ', '/admin/console/crm', group: 'Quản trị', admin: true),
  DeepLinkSpec('/admin/tracking', 'Kiểm tra link & tracking', '/admin/tracking', group: 'Quản trị', admin: true),
];

int? intParam(GoRouterState s, String name) => int.tryParse(s.pathParameters[name] ?? '');

class LinkNotFoundPage extends StatelessWidget {
  const LinkNotFoundPage({super.key, this.location});

  final String? location;

  @override
  Widget build(BuildContext context) => Scaffold(
        appBar: AppBar(),
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.link_off_rounded, size: 64, color: Theme.of(context).colorScheme.outline),
                const SizedBox(height: 16),
                const Text('Link không hợp lệ hoặc nội dung không còn tồn tại', textAlign: TextAlign.center),
                if (location != null) Padding(padding: const EdgeInsets.only(top: 8), child: Text(location!, textAlign: TextAlign.center, style: Theme.of(context).textTheme.bodySmall)),
                const SizedBox(height: 20),
                FilledButton(onPressed: () => context.go('/'), child: const Text('Về trang chủ')),
              ],
            ),
          ),
        ),
      );
}

GoRoute _content(String path, String service, Widget? Function(GoRouterState s) page) => GoRoute(
      path: path,
      builder: (context, state) {
        final p = page(state);
        return p == null ? LinkNotFoundPage(location: state.uri.toString()) : KeyedSubtree(key: ValueKey(state.uri.toString()), child: appScope(context, serviceColor(service), p));
      },
    );

final deepLinkRoutes = <RouteBase>[
  GoRoute(path: '/scan', builder: (context, _) => ScanPage(onCode: (ctx, code) => openScanned(ctx, code))),
  GoRoute(path: '/r/:code', redirect: (_, _) => '/'),
  _content('/event/:id', 'ticket', (s) => intParam(s, 'id') == null ? null : TicketEventPage(id: intParam(s, 'id')!)),
  _content('/ticket-order/:id', 'ticket', (s) => intParam(s, 'id') == null ? null : TicketOrderDetailPage(id: intParam(s, 'id')!)),
  _content('/prediction/:id', 'polymarket', (s) => PolyEventPage(id: s.pathParameters['id']!)),
  _content('/market/:id', 'polymarket', (s) => PolyMarketPage(id: s.pathParameters['id']!)),
  _content('/book/:id', 'book', (s) => intParam(s, 'id') == null ? null : BookDetailPage(id: intParam(s, 'id')!)),
  _content('/store-book/:id', 'book_store', (s) => intParam(s, 'id') == null ? null : StoreBookLinkPage(id: intParam(s, 'id')!)),
  _content('/movie/:id', 'movie', (s) => MovieLinkPage(id: s.pathParameters['id']!)),
  _content('/vehicle/:id', 'car_rental', (s) => VehicleLinkPage(id: s.pathParameters['id']!)),
  _content('/video/:id', 'video', (s) => VideoDetailPage(id: s.pathParameters['id']!)),
  _content('/live/:id', 'livestream', (s) => LiveRoomPage(id: s.pathParameters['id']!)),
  _content('/lucky/:game', 'lucky', (s) => LuckyGamePage(game: s.pathParameters['game']!)),
  _content('/post/:id', 'post', (s) => PostDetailPage(id: s.pathParameters['id']!)),
  _content('/thread/:id', 'threads', (s) => ThreadLinkPage(id: s.pathParameters['id']!)),
  _content('/room/:id', 'room', (s) => intParam(s, 'id') == null ? null : RoomLinkPage(id: intParam(s, 'id')!)),
  _content('/hotel/:id', 'reservation', (s) => intParam(s, 'id') == null ? null : HotelLinkPage(id: intParam(s, 'id')!, query: s.uri.queryParameters)),
  _content('/homestay/:id', 'rent_house', (s) => intParam(s, 'id') == null ? null : HomestayLinkPage(id: intParam(s, 'id')!, query: s.uri.queryParameters)),
  _content('/reservation/:id', 'reservation', (s) => intParam(s, 'id') == null ? null : ResvDetailPage(id: intParam(s, 'id')!)),
  _content('/homestay-booking/:id', 'rent_house', (s) => intParam(s, 'id') == null ? null : RentBookingPage(id: intParam(s, 'id')!)),
  _content('/lease/:id', 'rent_house', (s) => intParam(s, 'id') == null ? null : RentLeasePage(id: intParam(s, 'id')!)),
  _content('/qr/:id', 'qr', (s) => QrLinkPage(id: s.pathParameters['id']!)),
  _content('/admin/console/:service', 'manage', (s) => AdminConsolePage(serviceId: s.pathParameters['service']!)),
  _content('/admin/tracking', 'manage', (_) => const LinkTrackingPage()),
];
