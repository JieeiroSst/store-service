import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/app/role.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/features/apps/apps.dart';
import 'package:keikiapp/features/apps/kit/kit.dart';

import 'helpers.dart';

GraphQLClient _client({bool admin = false}) => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        final data = query.contains('me {')
            ? {
                'me': {
                  'id': 7,
                  'username': 'quan',
                  'email': 'q@keiki.vn',
                  'name': 'Quan',
                  'roles': [if (admin) {'id': 1, 'name': 'admin'}],
                },
              }
            : <String, Object?>{};
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

void main() {
  for (final id in serviceApps.keys) {
    testWidgets('$id mini-app opens on a small phone', (tester) async {
      tester.view.physicalSize = const Size(720, 1280);
      tester.view.devicePixelRatio = 2;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(KeikiApp(client: _client(admin: adminOnlyApps.contains(id)), catalogLoader: loadTestCatalog));
      await tester.pumpAndSettle();
      GoRouter.of(tester.element(find.byType(Navigator).first)).push('/services/$id');
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      expect(find.byType(ServiceApp), findsOneWidget);
      final tabs = find.descendant(of: find.byType(TabBar), matching: find.byType(Tab));
      for (var i = 1; i < tabs.evaluate().length; i++) {
        await tester.tap(tabs.at(i));
        await tester.pumpAndSettle();
        expect(tester.takeException(), isNull);
      }
      await tester.pump(const Duration(seconds: 3));
    });
  }

  testWidgets('dynamic catalogue stays reachable with ?all=1', (tester) async {
    await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
    GoRouter.of(tester.element(find.byType(Navigator).first)).push('/services/movie?all=1');
    await tester.pumpAndSettle();
    expect(find.byType(ServiceApp), findsNothing);
  });
}
