import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/tracking/attribution_cubit.dart';

import '../helpers.dart';

class _Server {
  double fee = 32000;
  bool feeChanged = false;
  bool oldAddress = false;
  final checkouts = <Map<String, dynamic>>[];
  final addressUpdates = <Map<String, dynamic>>[];

  Map<String, dynamic> handle(String q, Map<String, dynamic> v) {
    if (q.contains('me {')) {
      return {
        'data': {
          'me': {'id': 7, 'username': 'quan', 'name': 'Quan', 'roles': []},
        },
      };
    }
    if (q.contains('walletSvcMyWallets')) {
      return {
        'data': {
          'walletSvcMyWallets': [
            {'id': 'w-1', 'currency': 'VND', 'balance': '1000000', 'status': 'active', 'tier': 'basic', 'has_pin': true},
          ],
        },
      };
    }
    if (q.contains('query basketBaskets')) {
      return {
        'data': {
          'basketBaskets': [
            {'id': 1, 'status': 'Open', 'user_id': 7, 'currency': 'VND'},
          ],
        },
      };
    }
    if (q.contains('query basketBasketLines')) {
      return {
        'data': {
          'basketBasketLines': [
            {'id': 1, 'basket_id': 1, 'stockrecord_id': 10, 'product_id': 42, 'quantity': 2, 'price_incl_tax': 150000, 'price_currency': 'VND'},
          ],
        },
      };
    }
    if (q.contains('myUserAddresses')) {
      return {
        'data': {
          'myUserAddresses': [
            {
              'id': 4,
              'first_name': 'Lưu Mạnh',
              'last_name': 'Quân',
              'line1': '24 Lê Văn Hưu',
              'line2': oldAddress ? 'P. Ben Nghe, Q.1' : 'Phường Bến Nghé, Quận 1',
              'state': oldAddress ? 'TP. HCM' : 'Hồ Chí Minh',
              'postcode': oldAddress ? '70000' : 'ghn:202:1442:20109',
              'phone_number': '0912345678',
            },
          ],
        },
      };
    }
    if (q.contains('shipProvinces')) {
      return {
        'data': {
          'shipProvinces': [
            {'id': 201, 'name': 'Hà Nội'},
            {'id': 202, 'name': 'HCM'},
          ],
        },
      };
    }
    if (q.contains('shipDistricts')) {
      return {
        'data': {
          'shipDistricts': [
            {'id': 1442, 'province_id': 202, 'name': 'Quận 1'},
            {'id': 1443, 'province_id': 202, 'name': 'Quận 10'},
          ],
        },
      };
    }
    if (q.contains('shipWards')) {
      return {
        'data': {
          'shipWards': [
            {'code': '20109', 'district_id': 1442, 'name': 'Phường Bến Nghé'},
            {'code': '20110', 'district_id': 1442, 'name': 'Phường Bến Thành'},
          ],
        },
      };
    }
    if (q.contains('updateMyUserAddress')) {
      addressUpdates.add(v);
      return {
        'data': {
          'updateMyUserAddress': {'id': 4},
        },
      };
    }
    if (q.contains('basketShippingQuote')) {
      return {
        'data': {
          'basketShippingQuote': {'currency': 'VND', 'subtotal': 300000, 'shipping_fee': fee, 'total': 300000 + fee, 'shipping_charged': true, 'service_name': 'Chuẩn'},
        },
      };
    }
    if (q.contains('mutation basketCheckout')) {
      checkouts.add(v);
      if (feeChanged) {
        fee = 35000;
        return {
          'data': {'basketCheckout': null},
          'errors': [
            {'message': 'conflict: phí giao hàng là 35000 VND, hãy xác nhận lại trước khi thanh toán'},
          ],
        };
      }
      return {
        'data': {
          'basketCheckout': {'id': 1, 'status': 'Submitted', 'user_id': 7, 'currency': 'VND', 'total_incl_tax': 300000 + fee, 'shipping_incl_tax': fee},
        },
      };
    }
    return {'data': <String, Object?>{}};
  }
}

Future<void> _openCart(WidgetTester tester, _Server server) async {
  tester.view.physicalSize = const Size(1170, 2532);
  tester.view.devicePixelRatio = 3;
  addTearDown(tester.view.reset);
  final client = GraphQLClient(
    endpoint: 'http://gw/query',
    storage: MemoryTokenStorage(validTokens()),
    httpClient: MockClient((req) async {
      final body = jsonDecode(req.body) as Map<String, dynamic>;
      final res = server.handle(body['query'] as String, (body['variables'] as Map<String, dynamic>?) ?? const {});
      return http.Response(jsonEncode(res), 200, headers: {'content-type': 'application/json'});
    }),
  );
  await tester.pumpWidget(KeikiApp(client: client, catalogLoader: loadTestCatalog, attributionStore: MemoryAttributionStore()));
  await tester.pumpAndSettle();
  GoRouter.of(tester.element(find.byType(Navigator).first)).push('/shop/cart');
  await tester.pumpAndSettle();
  await tester.tap(find.text('Chọn địa chỉ giao hàng').first);
  await tester.pumpAndSettle();
  await tester.tap(find.text('Lưu Mạnh Quân • 0912345678'));
  await tester.pumpAndSettle();
  if (server.oldAddress) {
    expect(find.text('Bổ sung khu vực giao hàng'), findsOneWidget);
    expect(find.text('Đã tự tìm khu vực theo địa chỉ cũ, kiểm tra lại rồi bấm Lưu.'), findsOneWidget);
    await tester.tap(find.text('Lưu'));
    await tester.pumpAndSettle();
  }
}

Future<void> _pay(WidgetTester tester) async {
  await tester.tap(find.text('Thanh toán'));
  await tester.pumpAndSettle();
  await tester.enterText(find.byType(TextField).last, '123456');
  await tester.tap(find.text('Xác nhận'));
  await tester.pumpAndSettle();
}

void main() {
  testWidgets('cart shows the server shipping fee and pays it', (tester) async {
    final server = _Server();
    await _openCart(tester, server);
    expect(find.textContaining(RegExp(r'^Phí giao hàng GHN 32\.000.*Chuẩn$')), findsOneWidget);
    expect(find.textContaining('332.000'), findsWidgets);
    await _pay(tester);
    expect(server.checkouts.single['shipping_fee'], 32000);
    expect(server.checkouts.single['shipping_address_id'], 4);
    expect(find.textContaining('gồm phí giao hàng 32.000'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('a changed shipping fee is shown again before paying', (tester) async {
    final server = _Server()..feeChanged = true;
    await _openCart(tester, server);
    await _pay(tester);
    expect(find.text('Phí giao hàng đã thay đổi'), findsOneWidget);
    expect(find.textContaining('Phí giao hàng mới là 35.000'), findsOneWidget);
    expect(find.textContaining('chưa bị trừ tiền'), findsOneWidget);
    await tester.tap(find.text('Đã hiểu'));
    await tester.pumpAndSettle();
    expect(find.textContaining(RegExp(r'^Phí giao hàng GHN 35\.000.*Chuẩn$')), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('an old address gets its GHN area filled in before quoting', (tester) async {
    final server = _Server()..oldAddress = true;
    await _openCart(tester, server);
    final input = server.addressUpdates.single['input'] as Map;
    expect(server.addressUpdates.single['id'], 4);
    expect(input['postcode'], 'ghn:202:1442:20109');
    expect(input['line2'], 'Phường Bến Nghé, Quận 1');
    expect(input['state'], 'HCM');
    expect(find.textContaining(RegExp(r'^Phí giao hàng GHN 32\.000.*Chuẩn$')), findsOneWidget);
    expect(tester.takeException(), isNull);
  });
}
