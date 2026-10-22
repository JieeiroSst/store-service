import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';

import 'helpers.dart';

GraphQLClient _client(MemoryTokenStorage storage) => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: storage,
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        Object data = {};
        if (query.contains('me {')) {
          data = {
            'me': {'id': 7, 'username': 'quan', 'email': 'q@keiki.vn', 'name': 'Quan', 'roles': []},
          };
        } else if (query.contains('walletSvcMyWallets')) {
          data = {'walletSvcMyWallets': []};
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

void main() {
  testWidgets('shows login when there is no session', (tester) async {
    await tester.pumpWidget(KeikiApp(client: _client(MemoryTokenStorage()), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();

    expect(find.widgetWithText(FilledButton, 'Đăng nhập'), findsOneWidget);
    expect(find.text('Tên đăng nhập'), findsOneWidget);
  });

  testWidgets('shows the mini-app home for a stored session', (tester) async {
    tester.view.physicalSize = const Size(1200, 4200);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(KeikiApp(client: _client(MemoryTokenStorage(validTokens())), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();

    expect(find.textContaining('Quan'), findsOneWidget);
    for (final title in ['Ví Keiki', 'Food', 'Mua sắm', 'Chat', 'Phim', 'Vé sự kiện']) {
      expect(find.text(title), findsWidgets);
    }
    expect(find.text('Chưa kích hoạt'), findsOneWidget);
    expect(find.byType(NavigationBar), findsOneWidget);
    expect(find.text('Dịch vụ'), findsOneWidget);
  });

  testWidgets('login and home fit a small phone screen', (tester) async {
    tester.view.physicalSize = const Size(640, 1136);
    tester.view.devicePixelRatio = 2;
    addTearDown(tester.view.reset);

    await tester.pumpWidget(KeikiApp(client: _client(MemoryTokenStorage()), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);

    await tester.pumpWidget(const SizedBox());
    await tester.pumpWidget(KeikiApp(client: _client(MemoryTokenStorage(validTokens())), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('Tài khoản'));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
    await tester.tap(find.text('Dịch vụ'));
    await tester.pumpAndSettle();
    expect(tester.takeException(), isNull);
  });
}
