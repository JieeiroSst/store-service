import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';

import '../../test/helpers.dart';

Object _data(String q) {
  if (q.contains('me {')) {
    return {
      'me': {'id': 7, 'username': 'quan', 'email': 'quan@keiki.vn', 'name': 'Lưu Mạnh Quân', 'phone': '0901234567', 'address': 'Quận 1, TP.HCM', 'roles': []},
    };
  }
  if (q.contains('walletSvcMyWallets')) {
    return {
      'walletSvcMyWallets': [
        {'id': 'w-1', 'currency': 'VND', 'balance': '2450000', 'status': 'ACTIVE', 'tier': 'VERIFIED', 'has_pin': true},
      ],
    };
  }
  if (q.contains('walletSvcMyBalance')) {
    return {
      'walletSvcMyBalance': {'balance': '2450000', 'held': '0', 'available': '2450000', 'currency': 'VND'},
    };
  }
  if (q.contains('walletSvcMyTransactions')) {
    return {
      'walletSvcMyTransactions': {
        'items': [
          {'id': 't1', 'type': 'TOPUP', 'status': 'COMPLETED', 'amount': '500000', 'currency': 'VND', 'description': 'Nạp tiền từ Vietcombank', 'wallet_id': 'w-1', 'created_at': '2026-10-20T08:30:00Z'},
          {'id': 't2', 'type': 'PAYMENT', 'status': 'COMPLETED', 'amount': '125000', 'currency': 'VND', 'description': 'Keiki Food • Phở Thìn', 'wallet_id': 'w-1', 'created_at': '2026-10-19T12:10:00Z'},
          {'id': 't3', 'type': 'TRANSFER_OUT', 'status': 'COMPLETED', 'amount': '300000', 'currency': 'VND', 'description': 'Chuyển cho Minh Anh', 'wallet_id': 'w-1', 'created_at': '2026-10-18T19:45:00Z'},
        ],
        'total': 3,
      },
    };
  }
  if (q.contains('walletSvcMyBankAccounts')) return {'walletSvcMyBankAccounts': []};
  if (q.contains('walletSvcMyPaymentRequests')) return {'walletSvcMyPaymentRequests': {'items': []}};
  if (q.contains('pointMyBalance')) return {'pointMyBalance': {'balance': 1280, 'total_earned': 4200}};
  if (q.contains('myPermissions')) return {'myPermissions': [], 'myRoles': {'effectiveRoles': []}};
  if (q.contains('foodRestaurants')) {
    return {
      'foodRestaurants': [
        {'id': 1, 'name': 'Phở Thìn Lò Đúc', 'description': 'Phở bò tái lăn gia truyền từ 1979', 'address': '13 Lò Đúc, Hai Bà Trưng'},
        {'id': 2, 'name': 'Bún Chả Hương Liên', 'description': 'Bún chả Obama nổi tiếng', 'address': '24 Lê Văn Hưu'},
      ],
    };
  }
  if (q.contains('catalogueCategories')) {
    return {
      'catalogueCategories': [
        {'id': 1, 'name': 'Thời trang', 'is_public': true},
        {'id': 2, 'name': 'Điện tử', 'is_public': true},
        {'id': 3, 'name': 'Sách', 'is_public': true},
      ],
    };
  }
  if (q.contains('catalogueProducts')) {
    return {
      'catalogueProducts': [
        for (final (i, t) in ['Áo thun Keiki Basic', 'Tai nghe không dây Air', 'Sổ tay da Keikibook', 'Bình giữ nhiệt 500ml'].indexed)
          {'id': i + 1, 'title': t, 'rating': 4.5 + i * 0.1, 'price': 159000.0 + i * 70000, 'price_currency': 'VND', 'in_stock': true, 'num_available': 12, 'images': []},
      ],
    };
  }
  if (q.contains('basketBaskets')) return {'basketBaskets': []};
  if (q.contains('basketCreateBasket')) return {'basketCreateBasket': {'id': 1, 'status': 'Open'}};
  if (q.contains('basketBasketLines')) return {'basketBasketLines': []};
  if (q.contains('wishlistDefault')) return {'wishlistDefault': null};
  if (q.contains('myDoordashOrders')) return {'myDoordashOrders': []};
  return {};
}

GraphQLClient _client({bool loggedIn = true}) => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(loggedIn ? validTokens() : null),
      httpClient: MockClient((req) async {
        final q = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        return http.Response(jsonEncode({'data': _data(q)}), 200, headers: {'content-type': 'application/json; charset=utf-8'});
      }),
    );

Future<void> _loadFonts() async {
  final family = FontLoader('BeVietnamPro');
  for (final w in ['Regular', 'Medium', 'SemiBold', 'Bold', 'ExtraBold']) {
    family.addFont(rootBundle.load('assets/fonts/BeVietnamPro-$w.ttf'));
  }
  await family.load();
  final icons = FontLoader('MaterialIcons');
  final sdk = Platform.environment['FLUTTER_ROOT'] ?? '${Platform.environment['HOME']}/flutter';
  icons.addFont(Future.value(ByteData.sublistView(File('$sdk/bin/cache/artifacts/material_fonts/MaterialIcons-Regular.otf').readAsBytesSync())));
  await icons.load();
}

void main() {
  setUpAll(_loadFonts);

  Future<void> shoot(WidgetTester tester, String name, {bool loggedIn = true, String? route}) async {
    tester.view.physicalSize = const Size(1170, 2532);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(KeikiApp(client: _client(loggedIn: loggedIn), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
    if (route != null) {
      GoRouter.of(tester.element(find.byType(Navigator).first)).push(route);
      await tester.pumpAndSettle();
    }
    await expectLater(find.byType(MaterialApp), matchesGoldenFile('screens/$name.png'));
  }

  testWidgets('login', (t) => shoot(t, 'login', loggedIn: false));
  testWidgets('home', (t) => shoot(t, 'home'));
  testWidgets('services', (t) => shoot(t, 'services', route: '/services'));
  testWidgets('wallet', (t) => shoot(t, 'wallet', route: '/wallet'));
  testWidgets('food', (t) => shoot(t, 'food', route: '/food'));
  testWidgets('shop', (t) => shoot(t, 'shop', route: '/shop'));
  testWidgets('profile', (t) => shoot(t, 'profile', route: '/profile'));
}
