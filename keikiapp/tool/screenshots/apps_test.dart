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
import 'fixtures/fixtures.dart';

final _field = RegExp(r'^\s*(?:query|mutation)\s+\w+[^{]*\{\s*(\w+)');

GraphQLClient _client() => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final q = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        Object? data;
        if (q.contains('me {')) {
          data = {'me': {'id': 7, 'username': 'quan', 'email': 'quan@keiki.vn', 'name': 'Lưu Mạnh Quân', 'roles': []}};
        } else {
          final name = _field.firstMatch(q)?.group(1);
          data = {?name: fixtures[name]};
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json; charset=utf-8'});
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

  Future<void> shoot(WidgetTester tester, String name, String route, {String? tap, String? tab}) async {
    tester.view.physicalSize = const Size(1170, 2532);
    tester.view.devicePixelRatio = 3;
    addTearDown(tester.view.reset);
    debugDisableShadows = false;
    await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog));
    await tester.pumpAndSettle();
    GoRouter.of(tester.element(find.byType(Navigator).first)).push(route);
    await tester.pumpAndSettle();
    if (tab != null) {
      await tester.tap(find.text(tab).last);
      await tester.pumpAndSettle();
    }
    if (tap != null) {
      await tester.tap(find.text(tap).first);
      await tester.pumpAndSettle();
    }
    expect(tester.takeException(), isNull);
    try {
      await expectLater(find.byType(MaterialApp), matchesGoldenFile('screens/apps/$name.png'));
    } finally {
      debugDisableShadows = true;
    }
  }

  for (final s in shots) {
    testWidgets(s.name, (t) => shoot(t, s.name, s.route, tap: s.tap, tab: s.tab));
  }

  for (final p in pageShots.entries) {
    testWidgets(p.key, (tester) async {
      tester.view.physicalSize = const Size(1170, 2532);
      tester.view.devicePixelRatio = 3;
      addTearDown(tester.view.reset);
      debugDisableShadows = false;
      await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog));
      await tester.pumpAndSettle();
      Navigator.of(tester.element(find.byType(Scaffold).first)).push(MaterialPageRoute<void>(builder: (_) => p.value()));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
      try {
        await expectLater(find.byType(MaterialApp), matchesGoldenFile('screens/apps/${p.key}.png'));
      } finally {
        debugDisableShadows = true;
      }
    });
  }
}
