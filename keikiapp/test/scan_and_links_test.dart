import 'dart:convert';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/scan/scan.dart';
import 'package:keikiapp/features/address/view/ghn_area_picker.dart';
import 'package:keikiapp/features/wallet/view/wallet_forms.dart';
import 'package:keikiapp/gen/gateway_api.g.dart';

import 'helpers.dart';

final _queries = <String>[];

GraphQLClient _client({bool loggedIn = true, bool hasWallet = true}) => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(loggedIn ? validTokens() : null),
      httpClient: MockClient((req) async {
        final body = jsonDecode(req.body) as Map<String, dynamic>;
        final query = body['query'] as String;
        _queries.add(query);
        Object data = <String, Object?>{};
        if (query.contains('me {')) {
          data = {
            'me': {'id': 7, 'username': 'quan', 'name': 'Quan', 'roles': []},
          };
        } else if (query.contains('walletSvcMyWallets')) {
          data = {
            'walletSvcMyWallets': [
              if (hasWallet) {'id': 'w-me', 'balance': '100000', 'currency': 'VND', 'status': 'active', 'tier': 'basic', 'has_pin': true},
            ],
          };
        } else if (query.contains('shipProvinces')) {
          data = {
            'shipProvinces': [
              {'id': 201, 'name': 'Hà Nội'},
              {'id': 202, 'name': 'Hồ Chí Minh'},
            ],
          };
        } else if (query.contains('shipDistricts')) {
          data = {
            'shipDistricts': [
              {'id': 1442, 'province_id': 202, 'name': 'Quận 1'},
            ],
          };
        } else if (query.contains('shipWards')) {
          data = {
            'shipWards': [
              {'code': '20109', 'district_id': 1442, 'name': 'Phường Bến Nghé'},
            ],
          };
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

Future<GoRouter> _pumpApp(WidgetTester tester, {bool loggedIn = true, bool hasWallet = true}) async {
  tester.view.physicalSize = const Size(1170, 2532);
  tester.view.devicePixelRatio = 3;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(KeikiApp(client: _client(loggedIn: loggedIn, hasWallet: hasWallet), catalogLoader: loadTestCatalog));
  await tester.pumpAndSettle();
  return GoRouter.of(tester.element(find.byType(Navigator).first));
}

class _QuoteExecutor implements GatewayExecutor {
  _QuoteExecutor(this.onVars);

  final void Function(Map<String, dynamic>) onVars;

  Future<Map<String, dynamic>> _run(Map<String, dynamic>? vars) async {
    onVars(vars ?? const {});
    return {
      'shipQuote': {
        'options': [
          {'fee': 28000},
        ],
        'recommended': {
          'CHEAPEST': {'fee': 28000},
        },
      },
    };
  }

  @override
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables}) => _run(variables);

  @override
  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables}) => _run(variables);

  @override
  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables}) => _run(variables);
}

void main() {
  group('scanned codes', () {
    test('wallet receive QR opens a prefilled transfer', () {
      expect(routeForCode('keiki://pay?wallet_id=w1&amount=50000'), '/wallet/transfer?wallet_id=w1&amount=50000');
    });

    test('app links and website links map to routes', () {
      expect(routeForCode('keikibook://app/services/ticket'), '/services/ticket');
      expect(routeForCode('keikibook://wallet/receive'), '/wallet/receive');
      expect(routeForCode('https://keikibook.vn/services/movie?id=3'), '/services/movie?id=3');
      expect(routeForCode('https://example.com/a'), isNull);
      expect(routeForCode('TICKET-123'), isNull);
    });

    test('code values are extracted for form fields', () {
      expect(codeValue('keiki://pay?wallet_id=w9'), 'w9');
      expect(codeValue('https://keikibook.vn/r/ABC'), 'ABC');
      expect(codeValue('  VC-001 '), 'VC-001');
    });
  });

  group('GHN area', () {
    test('code round trip', () {
      const a = GhnArea(provinceId: 202, districtId: 1442, wardCode: '20109');
      expect(a.code, 'ghn:202:1442:20109');
      expect(GhnArea.parse(a.code), a);
      expect(GhnArea.parse('70000'), isNull);
      expect(const GhnArea(provinceId: 1).code, '');
    });

    test('fee is read from different quote shapes', () {
      expect(ghnFee(32000), 32000);
      expect(ghnFee({'total': 25000}), 25000);
      expect(ghnFee({'data': {'service_fee': 18000}}), 18000);
      expect(ghnFee([{'carrier': 'GHN', 'total': 30000}, {'carrier': 'GHTK', 'fee': 27000}]), 27000);
      expect(ghnFee({'quotes': []}), isNull);
      expect(
        ghnFee({
          'options': [
            {'service_name': 'Chuẩn', 'fee': 32000},
            {'service_name': 'Tiết kiệm', 'fee': 28000},
          ],
          'recommended': {
            'FASTEST': {'fee': 41000},
            'CHEAPEST': {'fee': 28000},
          },
        }),
        28000,
      );
      expect(
        ghnFee({
          'options': [
            {'fee': 35000},
            {'fee': 30000},
          ],
          'recommended': {},
        }),
        30000,
      );
    });

    test('quote request matches the shipping-service contract', () async {
      Map<String, dynamic>? sent;
      final api = GatewayApi(_QuoteExecutor((vars) => sent = vars));
      final fee = await quoteGhn(
        api,
        const GhnArea(provinceId: 202, districtId: 1442, wardCode: '20109'),
        name: 'Lưu Mạnh Quân',
        phone: '0912345678',
        street: '24 Lê Văn Hưu',
        items: const [GhnItem('Áo thun', 2, price: 150000), GhnItem('Mũ', 1, price: 90000)],
        insurance: 9000000,
      );
      expect(fee, 28000);
      final input = sent!['input'] as Map;
      expect(input['recipient'], containsPair('name', 'Lưu Mạnh Quân'));
      expect(input['recipient'], containsPair('phone', '0912345678'));
      expect(input['recipient'], containsPair('street', '24 Lê Văn Hưu'));
      expect(input['recipient'], containsPair('ward_code', '20109'));
      final parcel = input['parcel'] as Map;
      expect(parcel['weight_gram'], 1500);
      expect(parcel['height_cm'], 25);
      expect((parcel['items'] as List).first, containsPair('name', 'Áo thun'));
      expect((parcel['items'] as List).first, containsPair('weight_gram', 500));
      expect(input['insurance_value'], 5000000);
    });

    testWidgets('picker walks province, district and ward', (tester) async {
      GhnArea? area;
      await tester.pumpWidget(MaterialApp(
        home: Scaffold(
          body: Builder(
            builder: (context) => RepositoryProvider<GatewayApi>.value(value: GatewayApi(_client()), child: GhnAreaPicker(onChanged: (a) => area = a)),
          ),
        ),
      ));
      await tester.tap(find.text('Tỉnh/Thành phố'));
      await tester.pumpAndSettle();
      await tester.enterText(find.byType(TextField).last, 'ho chi');
      await tester.pumpAndSettle();
      expect(find.text('Hà Nội'), findsNothing);
      await tester.tap(find.text('Hồ Chí Minh'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Quận/Huyện'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Quận 1'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Phường/Xã'));
      await tester.pumpAndSettle();
      await tester.tap(find.text('Phường Bến Nghé'));
      await tester.pumpAndSettle();
      expect(area?.code, 'ghn:202:1442:20109');
      expect(area?.provinceName, 'Hồ Chí Minh');
    });
  });

  testWidgets('scan page falls back to manual entry without a camera', (tester) async {
    debugDefaultTargetPlatformOverride = TargetPlatform.macOS;
    addTearDown(() => debugDefaultTargetPlatformOverride = null);
    String? result;
    await tester.pumpWidget(MaterialApp(
      home: Builder(
        builder: (context) => TextButton(onPressed: () async => result = await scanCode(context), child: const Text('scan')),
      ),
    ));
    await tester.tap(find.text('scan'));
    await tester.pumpAndSettle();
    expect(find.textContaining('nhập mã'), findsOneWidget);
    await tester.tap(find.text('Nhập mã'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(TextField), 'keiki://pay?wallet_id=w1');
    await tester.tap(find.text('Xong'));
    await tester.pumpAndSettle();
    expect(result, 'keiki://pay?wallet_id=w1');
    debugDefaultTargetPlatformOverride = null;
  });

  testWidgets('payment QR deep link opens transfer with the wallet prefilled', (tester) async {
    final router = await _pumpApp(tester);
    router.go('keiki://pay?wallet_id=w-friend&amount=50000');
    await tester.pumpAndSettle();
    expect(find.byType(TransferPage), findsOneWidget);
    expect(find.text('w-friend'), findsOneWidget);
    expect(find.text('50000'), findsOneWidget);
  });

  testWidgets('deep link target survives the login screen', (tester) async {
    final router = await _pumpApp(tester, loggedIn: false);
    router.go('/wallet/receive');
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/login');
    expect(router.routerDelegate.currentConfiguration.uri.queryParameters['from'], '/wallet/receive');
  });

  testWidgets('payment link without a wallet asks to activate instead of crashing', (tester) async {
    final router = await _pumpApp(tester, hasWallet: false);
    router.go('keiki://pay?wallet_id=w-friend');
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    expect(find.text('Bạn chưa kích hoạt Ví Keiki'), findsOneWidget);
    await tester.tap(find.text('Kích hoạt ví'));
    await tester.pumpAndSettle();
    expect(router.routerDelegate.currentConfiguration.uri.path, '/wallet');
  });
}
