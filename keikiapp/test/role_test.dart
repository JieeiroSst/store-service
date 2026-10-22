import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/app/role.dart';
import 'package:keikiapp/core/network/graphql_client.dart';

import 'helpers.dart';

GraphQLClient _client({List<String> meRoles = const [], List<String> effective = const []}) => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        Object data = {};
        if (query.contains('me {')) {
          data = {
            'me': {
              'id': 7,
              'username': 'quan',
              'email': 'q@keiki.vn',
              'name': 'Quan',
              'roles': [for (final r in meRoles) {'id': 1, 'name': r}],
            },
          };
        } else if (query.contains('myPermissions')) {
          data = {'myPermissions': [], 'myRoles': {'effectiveRoles': effective}};
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

Future<void> _pump(WidgetTester tester, GraphQLClient client) async {
  tester.view.physicalSize = const Size(1170, 2532);
  tester.view.devicePixelRatio = 3;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(KeikiApp(client: client, catalogLoader: loadTestCatalog));
  await tester.pumpAndSettle();
}

void main() {
  group('RoleCubit', () {
    test('admin when any source reports an admin role', () {
      final c = RoleCubit();
      expect(c.state.role, AppRole.customer);
      c.report('me', ['customer']);
      expect(c.state.isAdmin, isFalse);
      c.report('authorize', ['super_admin']);
      expect(c.state.isAdmin, isTrue);
      c.report('authorize', const []);
      expect(c.state.isAdmin, isFalse);
    });

    test('admins can preview the customer view and come back', () {
      final c = RoleCubit()..report('me', ['admin']);
      c.previewAsCustomer(true);
      expect(c.state.role, AppRole.customer);
      expect(c.state.serverAdmin, isTrue);
      c.previewAsCustomer(false);
      expect(c.state.isAdmin, isTrue);
    });

    test('customers cannot switch themselves to admin', () {
      final c = RoleCubit()..previewAsCustomer(false);
      expect(c.state.isAdmin, isFalse);
      c.reset();
      expect(c.state, const RoleState());
    });
  });

  testWidgets('customers see three tabs and no management entry points', (tester) async {
    await _pump(tester, _client());
    expect(find.text('Quản trị'), findsNothing);
    GoRouter.of(tester.element(find.byType(Navigator).first)).push('/services/ticket');
    await tester.pumpAndSettle();
    expect(find.byTooltip('Ban tổ chức'), findsNothing);
    GoRouter.of(tester.element(find.byType(Navigator).first)).go('/admin');
    await tester.pumpAndSettle();
    expect(find.text('Khu quản lý'), findsNothing);
  });

  testWidgets('admins get the Quản trị tab and management actions', (tester) async {
    await _pump(tester, _client(meRoles: ['admin']));
    expect(find.text('Quản trị'), findsOneWidget);
    await tester.tap(find.text('Quản trị'));
    await tester.pumpAndSettle();
    expect(find.text('Khu quản lý'), findsOneWidget);
    expect(find.text('Ban tổ chức sự kiện'), findsOneWidget);
    expect(tester.takeException(), isNull);

    GoRouter.of(tester.element(find.byType(Navigator).first)).push('/services/ticket');
    await tester.pumpAndSettle();
    expect(find.byTooltip('Ban tổ chức'), findsOneWidget);
  });

  testWidgets('admin role can come from authorize-service effective roles', (tester) async {
    await _pump(tester, _client(effective: ['super_admin']));
    expect(find.text('Quản trị'), findsOneWidget);
  });

  testWidgets('preview as customer hides the admin tab', (tester) async {
    await _pump(tester, _client(meRoles: ['admin']));
    await tester.tap(find.text('Quản trị'));
    await tester.pumpAndSettle();
    await tester.tap(find.text('Xem như khách'));
    await tester.pumpAndSettle();
    expect(find.text('Quản trị'), findsNothing);
    expect(find.text('Khu quản lý'), findsNothing);
  });
}
