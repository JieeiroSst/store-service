import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/app/deep_links.dart';
import 'package:keikiapp/core/links/share_link.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/tracking/attribution_cubit.dart';
import 'package:keikiapp/features/apps/entertainment/movie_app.dart';
import 'package:keikiapp/features/apps/entertainment/ticket_app.dart';
import 'package:keikiapp/features/apps/travel/mobility_apps.dart';
import 'package:keikiapp/features/apps/travel/stays_apps.dart';
import 'package:keikiapp/features/food/view/food_home_page.dart';

import 'helpers.dart';

GraphQLClient _client() => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        Object data = <String, Object?>{};
        if (query.contains('me {')) {
          data = {
            'me': {
              'id': 7,
              'username': 'quan',
              'name': 'Quan',
              'roles': [
                {'id': 1, 'name': 'admin'},
              ],
            },
          };
        } else if (query.contains('movieVideo(')) {
          final id = ((jsonDecode(req.body) as Map)['variables'] as Map)['id'];
          data = {
            'movieVideo': id == 'mat-biec' ? {'id': 'mat-biec', 'title': 'Mắt Biếc', 'description': 'Phim của Victor Vũ', 'status': 'ready'} : null,
          };
        } else if (query.contains('carRentalGetVehicle')) {
          data = {
            'carRentalGetVehicle': {'vehicleId': 'veh-01', 'make': 'VinFast', 'model': 'VF 8', 'dailyRate': 1200000},
          };
        } else if (query.contains('foodRestaurants')) {
          data = {
            'foodRestaurants': [
              {'id': 2, 'name': 'Phở Thìn'},
              {'id': 3, 'name': 'Bún chả Hương Liên', 'address': '24 Lê Văn Hưu'},
            ],
          };
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

Future<GoRouter> _pump(WidgetTester tester) async {
  tester.view.physicalSize = const Size(1170, 2532);
  tester.view.devicePixelRatio = 3;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog, attributionStore: MemoryAttributionStore()));
  await tester.pumpAndSettle();
  return GoRouter.of(tester.element(find.byType(Navigator).first));
}

void main() {
  test('catalog patterns are unique and examples are absolute paths', () {
    expect(deepLinkCatalog.map((d) => d.pattern).toSet(), hasLength(deepLinkCatalog.length));
    for (final d in deepLinkCatalog) {
      expect(d.example, startsWith('/'), reason: d.title);
    }
  });

  test('web and app links keep the path and query', () {
    expect(webLinkFor('/hotel/4?guests=2'), 'https://keikibook.vn/hotel/4?guests=2');
    expect(appLinkFor('/event/9'), 'keikibook://app/event/9');
    expect(deepLinkCatalog.firstWhere((d) => d.pattern == '/event/:id').appUrl, 'keikibook://app/event/9');
  });

  group('share links', () {
    tearDown(() => shortLinkBuilder = (path, web, {title, description, imageUrl, channel = 'app'}) async => null);

    test('falls back to the website link without Branch', () async {
      shortLinkBuilder = (path, web, {title, description, imageUrl, channel = 'app'}) async => null;
      expect(await buildShareLink('/event/9', title: 'Hoà nhạc'), 'https://keikibook.vn/event/9');
    });

    test('uses the Branch short link with the deep link path', () async {
      String? sentPath;
      String? sentWeb;
      shortLinkBuilder = (path, web, {title, description, imageUrl, channel = 'app'}) async {
        sentPath = path;
        sentWeb = web;
        return 'https://keikibook.app.link/xyz';
      };
      expect(await buildShareLink('/post/abc'), 'https://keikibook.app.link/xyz');
      expect(sentPath, '/post/abc');
      expect(sentWeb, 'https://keikibook.vn/post/abc');
    });

    test('Branch failure still shares the website link', () async {
      shortLinkBuilder = (path, web, {title, description, imageUrl, channel = 'app'}) async => throw Exception('offline');
      expect(await buildShareLink('/video/a'), 'https://keikibook.vn/video/a');
    });

    testWidgets('share button shows the link', (tester) async {
      shortLinkBuilder = (path, web, {title, description, imageUrl, channel = 'app'}) async => null;
      await tester.pumpWidget(const MaterialApp(home: Scaffold(body: Center(child: ShareLinkButton(path: '/event/9', title: 'Hoà nhạc')))));
      await tester.tap(find.byIcon(Icons.share_outlined));
      await tester.pumpAndSettle();
      expect(find.text('https://keikibook.vn/event/9'), findsOneWidget);
      expect(find.text('Mở trong app: keikibook://app/event/9'), findsOneWidget);
      expect(find.text('Sao chép'), findsOneWidget);
    });
  });

  testWidgets('every catalog link opens a real screen', (tester) async {
    final router = await _pump(tester);
    for (final d in deepLinkCatalog.where((d) => d.pattern != '/scan')) {
      router.go(d.example);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull, reason: d.example);
      expect(find.byType(LinkNotFoundPage), findsNothing, reason: d.example);
    }
  });

  testWidgets('scan link opens the scanner', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
    final router = await _pump(tester);
    router.go('/scan');
    await tester.pumpAndSettle();
    expect(find.text('Nhập mã'), findsOneWidget);
    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets('broken links show a friendly page', (tester) async {
    final router = await _pump(tester);
    for (final bad in ['/event/abc', '/shop/products/x', '/food/orders/x', '/chat/direct/x', '/hotel/x', '/khong-co-trang-nay']) {
      router.go(bad);
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull, reason: bad);
      expect(find.byType(LinkNotFoundPage), findsOneWidget, reason: bad);
    }
    await tester.tap(find.text('Về trang chủ'));
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/');
  });

  testWidgets('external app and website links open content', (tester) async {
    final router = await _pump(tester);
    router.go('keikibook://app/event/9');
    await tester.pumpAndSettle();
    expect(find.byType(TicketEventPage), findsOneWidget);
    router.go('https://keikibook.vn/hotel/4?check_in=2026-12-24&check_out=2026-12-26&guests=3');
    await tester.pumpAndSettle();
    final hotel = tester.widget<HotelLinkPage>(find.byType(HotelLinkPage));
    expect(hotel.id, 4);
    expect(hotel.query, {'check_in': '2026-12-24', 'check_out': '2026-12-26', 'guests': '3'});
    router.go('keikibook://app/r/ABC12345');
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/');
  });

  testWidgets('restaurant link opens that restaurant even from a cold start', (tester) async {
    final router = await _pump(tester);
    router.go('/food/restaurants/3');
    await tester.pumpAndSettle();
    expect(find.byType(RestaurantPage), findsOneWidget);
    expect(find.text('Bún chả Hương Liên'), findsWidgets);
    router.go('/food/restaurants/99');
    await tester.pumpAndSettle();
    expect(find.text('Không tìm thấy nhà hàng'), findsOneWidget);
  });

  testWidgets('movie and rental car links open their detail screens', (tester) async {
    final router = await _pump(tester);
    router.go('keikibook://app/movie/mat-biec');
    await tester.pumpAndSettle();
    expect(find.byType(MovieDetailPage), findsOneWidget);
    expect(find.text('Mắt Biếc'), findsWidgets);
    router.go('/movie/khong-co');
    await tester.pumpAndSettle();
    expect(find.text('Phim không tồn tại hoặc chưa sẵn sàng'), findsOneWidget);
    router.go('https://keikibook.vn/vehicle/veh-01');
    await tester.pumpAndSettle();
    expect(find.byType(VehicleLinkPage), findsOneWidget);
    expect(find.text('VinFast VF 8'), findsWidgets);
    expect(tester.takeException(), isNull);
  });
}
