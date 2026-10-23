import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter/material.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/features/admin/console/admin_console_page.dart';

import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/features/admin/console/admin_op.dart';
import 'package:keikiapp/features/admin/console/gen/admin_consoles.g.dart';
import 'package:keikiapp/features/apps/kit/kit.dart';
import 'package:keikiapp/features/services/data/humanize.dart';
import 'package:keikiapp/features/services/data/schema_catalog.dart';
import 'package:keikiapp/gen/gateway_api.g.dart';

import 'helpers.dart';

class _Recorder implements GatewayExecutor {
  String? document;
  Map<String, dynamic>? variables;

  Future<Map<String, dynamic>> _record(String doc, Map<String, dynamic>? vars) async {
    document = doc;
    variables = vars;
    return {};
  }

  @override
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables}) => _record(document, variables);

  @override
  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables}) => _record(document, variables);

  @override
  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables}) => _record(document, variables);
}

String _sampleValue(KField f) {
  switch (f.kind) {
    case FieldKind.number:
      return '1';
    case FieldKind.toggle:
      return 'true';
    case FieldKind.choice:
      return f.options.firstWhere((o) => o.$1.isNotEmpty).$1;
    case FieldKind.email:
      return 'admin@keiki.vn';
    case FieldKind.phone:
      return '0900000000';
    default:
      final hint = f.hint ?? '';
      if (hint.startsWith('JSON, ví dụ ')) return hint.substring('JSON, ví dụ '.length);
      return '1';
  }
}

final _requests = <String>[];

GraphQLClient _client() => GraphQLClient(
      endpoint: 'http://gw/query',
      storage: MemoryTokenStorage(validTokens()),
      httpClient: MockClient((req) async {
        final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
        _requests.add(query);
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
        } else if (query.contains('myPermissions')) {
          data = {
            'myPermissions': [
              {'object': '/graphql/*', 'action': '.*'},
            ],
            'myRoles': {
              'effectiveRoles': ['admin'],
            },
          };
        } else if (query.contains('crmCampaigns')) {
          data = {
            'crmCampaigns': [
              {'id': 1, 'campaign_name': 'Ra mắt Keikibook', 'campaign_sponsor': 'Keiki'},
              {'id': 2, 'campaign_name': 'Tết 2027', 'campaign_sponsor': 'Keiki'},
            ],
          };
        } else if (query.contains('crmCreateCampaign')) {
          data = {
            'crmCreateCampaign': {'id': 3, 'campaign_name': 'Mới'},
          };
        }
        return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
      }),
    );

Future<void> _open(WidgetTester tester, Widget page) async {
  tester.view.physicalSize = const Size(720, 1280);
  tester.view.devicePixelRatio = 2;
  addTearDown(tester.view.reset);
  await tester.pumpWidget(KeikiApp(client: _client(), catalogLoader: loadTestCatalog));
  await tester.pumpAndSettle();
  Navigator.of(tester.element(find.byType(Scaffold).first)).push(MaterialPageRoute<void>(builder: (_) => page));
  await tester.pumpAndSettle();
}

void main() {
  test('every admin operation in the schema has a generated console entry', () {
    for (final s in testCatalog.services) {
      final expected = s.operations.where((o) => o.access == OperationAccess.admin).map((o) => o.name).toSet();
      final generated = (adminConsoles[s.id] ?? const <AdminOp>[]).map((o) => o.name).toSet();
      expect(generated, expected, reason: s.id);
    }
  });

  test('every generated admin operation sends its document with required variables', () async {
    final recorder = _Recorder();
    final api = GatewayApi(recorder);
    var count = 0;
    for (final entry in adminConsoles.entries) {
      for (final op in entry.value) {
        final values = {for (final f in op.fields) if (f.kind != FieldKind.file) f.key: _sampleValue(f)};
        final files = {for (final f in op.fields) if (f.kind == FieldKind.file) f.key: GqlUpload(bytes: Uint8List(1), filename: 'a.bin')};
        recorder
          ..document = null
          ..variables = null;
        await op.run(api, FormData(values, files));
        final gql = testCatalog.service(entry.key)!.operations.firstWhere((o) => o.name == op.name && o.isMutation == op.mutation);
        expect(recorder.document, contains(' ${op.name}'), reason: op.name);
        for (final a in gql.args.where((a) => a.type.nonNull)) {
          expect(recorder.variables?[a.name], isNotNull, reason: '${op.name}.${a.name}');
        }
        count++;
      }
    }
    expect(count, testCatalog.services.expand((s) => s.operations).where((o) => o.access == OperationAccess.admin).length);
  });

  test('missing required values raise a readable error', () async {
    final op = adminConsoles['arrange']!.firstWhere((o) => o.name == 'arrangeMarkDelivered');
    await expectLater(op.run(GatewayApi(_Recorder()), const FormData({})), throwsA(isA<AdminInputError>().having((e) => e.toString(), 'message', contains('id'))));
  });

  test('invalid JSON raises a readable error', () {
    expect(() => const FormData({'x': '{oops'}).json('x'), throwsA(isA<AdminInputError>()));
    expect(const FormData({'x': '{"a": 1}'}).jsonMap('x'), {'a': 1});
    expect(const FormData({'x': 'a, b ,,c'}).strs('x'), ['a', 'b', 'c']);
    expect(const FormData({}).optFlag('x'), isNull);
  });

  test('admin results are converted to plain JSON values', () {
    expect(adminPlain(const ArrangeDispatchInput(orderId: 1)), {'orderId': 1});
    expect(adminPlain([const ArrangeDispatchInput(driverId: 2)]), [
      {'driverId': 2},
    ]);
  });

  test('resource titles are translated', () {
    expect(resourceTitle('general'), 'Chung');
    expect(resourceTitle('payment method'), 'Phương thức thanh toán');
    expect(resourceTitle('weird thing'), 'Weird thing');
  });

  for (final id in adminConsoles.keys) {
    testWidgets('$id admin console renders on a small phone', (tester) async {
      await _open(tester, AdminConsolePage(serviceId: id));
      expect(tester.takeException(), isNull);
      expect(find.byType(ListTile), findsWidgets);
      await tester.drag(find.byType(CustomScrollView), const Offset(0, -3000));
      await tester.pumpAndSettle();
      expect(tester.takeException(), isNull);
    });
  }

  testWidgets('console search and filter narrow the list', (tester) async {
    await _open(tester, const AdminConsolePage(serviceId: 'crm'));
    await tester.enterText(find.byType(TextField), 'crmCreateCampaign');
    await tester.pumpAndSettle();
    expect(find.descendant(of: find.byType(ListTile), matching: find.text('crmCreateCampaign')), findsOneWidget);
    expect(find.descendant(of: find.byType(ListTile), matching: find.text('crmCampaigns')), findsNothing);
    await tester.enterText(find.byType(TextField), '');
    await tester.tap(find.widgetWithText(ChoiceChip, 'Thao tác'));
    await tester.pumpAndSettle();
    expect(find.text('crmCampaigns'), findsNothing);
  });

  testWidgets('query without required fields loads on open and shows results', (tester) async {
    _requests.clear();
    final op = adminConsoles['crm']!.firstWhere((o) => o.name == 'crmCampaigns');
    await _open(tester, AdminOpPage(serviceId: 'crm', op: op, title: 'Danh sách chiến dịch'));
    expect(_requests.any((q) => q.contains('crmCampaigns')), isTrue);
    expect(find.text('Ra mắt Keikibook'), findsOneWidget);
    expect(tester.takeException(), isNull);
  });

  testWidgets('mutation opens a form and shows the server result', (tester) async {
    final op = adminConsoles['crm']!.firstWhere((o) => o.name == 'crmCreateCampaign');
    await _open(tester, AdminOpPage(serviceId: 'crm', op: op, title: 'Tạo chiến dịch'));
    await tester.tap(find.widgetWithText(FilledButton, 'Thực hiện'));
    await tester.pumpAndSettle();
    await tester.enterText(find.widgetWithText(TextFormField, 'Campaign name'), 'Mới');
    await tester.tap(find.widgetWithText(FilledButton, 'Thực hiện').last);
    await tester.pumpAndSettle();
    expect(_requests.last, contains('crmCreateCampaign'));
    expect(find.text('Mới'), findsWidgets);
    expect(tester.takeException(), isNull);
  });
}
