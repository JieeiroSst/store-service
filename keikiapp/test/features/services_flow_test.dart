import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';

import '../helpers.dart';

void main() {
  late List<Map<String, dynamic>> requests;

  GraphQLClient client() => GraphQLClient(
        endpoint: 'http://gw/query',
        storage: MemoryTokenStorage(validTokens()),
        httpClient: MockClient((req) async {
          final body = jsonDecode(req.body) as Map<String, dynamic>;
          requests.add(body);
          final q = body['query'] as String;
          Object data = {};
          if (q.contains('me {')) {
            data = {
              'me': {'id': 7, 'username': 'quan', 'email': 'q@keiki.vn', 'name': 'Quan', 'roles': []},
            };
          } else if (q.contains('walletSvcMyWallets')) {
            data = {'walletSvcMyWallets': []};
          } else if (q.contains('myPermissions')) {
            data = {
              'myPermissions': [],
              'myRoles': {'effectiveRoles': []},
            };
          } else if (q.contains('ticketEvents')) {
            data = {
              'ticketEvents': [
                {'id': 1, 'title': 'Đêm nhạc Hà Nội', 'city': 'Hà Nội', 'status': 'PUBLISHED'},
                {'id': 2, 'title': 'Lễ hội ẩm thực', 'city': 'Đà Nẵng', 'status': 'PUBLISHED'},
              ],
            };
          } else if (q.contains('lotteryCheck')) {
            data = {
              'lotteryCheck': {'won': true, 'prizes_won': 1, 'result_found': true, 'highest_prize': 'Giải nhất'},
            };
          }
          return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
        }),
      );

  setUp(() => requests = []);

  Future<void> pumpApp(WidgetTester tester) async {
    tester.view.physicalSize = const Size(1200, 2600);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    await tester.pumpWidget(KeikiApp(client: client(), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
  }

  Future<void> openAllFunctions(WidgetTester tester) async {
    await tester.tap(find.byIcon(Icons.more_vert).last);
    await tester.pumpAndSettle();
    await tester.tap(find.text('Tất cả chức năng'));
    await tester.pumpAndSettle();
  }

  testWidgets('browses a schema service and auto-runs a query without required args', (tester) async {
    await pumpApp(tester);

    await tester.tap(find.text('Dịch vụ'));
    await tester.pumpAndSettle();
    expect(find.text('Tất cả dịch vụ'), findsWidgets);

    await tester.enterText(find.byType(SearchBar), 'vé sự');
    await tester.pumpAndSettle();
    await tester.tap(find.text('Vé sự kiện').first);
    await tester.pumpAndSettle();
    await openAllFunctions(tester);

    await tester.tap(find.text('Danh sách sự kiện'));
    await tester.pumpAndSettle();

    expect(find.text('Đêm nhạc Hà Nội'), findsOneWidget);
    expect(find.text('Lễ hội ẩm thực'), findsOneWidget);
    final sent = requests.lastWhere((r) => (r['query'] as String).contains('ticketEvents'));
    expect(sent['query'], startsWith('query ticketEvents('));
    expect(sent['variables'], {'limit': 20});
  });

  testWidgets('validates required args, then sends typed variables', (tester) async {
    await pumpApp(tester);
    await tester.tap(find.text('Dịch vụ'));
    await tester.pumpAndSettle();
    await tester.enterText(find.byType(SearchBar), 'xổ số');
    await tester.pumpAndSettle();
    await tester.tap(find.text('Xổ số').first);
    await tester.pumpAndSettle();
    await openAllFunctions(tester);
    await tester.tap(find.textContaining('Kiểm tra'));
    await tester.pumpAndSettle();

    await tester.tap(find.text('Xem kết quả'));
    await tester.pumpAndSettle();
    expect(find.text('Bắt buộc'), findsNWidgets(3));
    expect(requests.where((r) => (r['query'] as String).contains('lotteryCheck')), isEmpty);

    await tester.enterText(find.widgetWithText(TextFormField, 'Date *'), '2026-10-20');
    await tester.enterText(find.widgetWithText(TextFormField, 'Province *'), 'Hà Nội');
    await tester.enterText(find.widgetWithText(TextFormField, 'Ticket number *'), '123456');
    await tester.tap(find.text('Xem kết quả'));
    await tester.pumpAndSettle();

    final sent = requests.lastWhere((r) => (r['query'] as String).contains('lotteryCheck'));
    expect(sent['variables'], {'date': '2026-10-20', 'province': 'Hà Nội', 'ticket_number': '123456'});
    expect(find.text('Giải nhất'), findsOneWidget);
  });
}
