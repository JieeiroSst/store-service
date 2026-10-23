import 'dart:async';
import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:go_router/go_router.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:keikiapp/app/app.dart';
import 'package:keikiapp/core/config/env.dart';
import 'package:keikiapp/core/network/graphql_client.dart';
import 'package:keikiapp/core/network/graphql_exception.dart';
import 'package:keikiapp/core/tracking/attribution_cubit.dart';
import 'package:keikiapp/core/tracking/link_source.dart';
import 'package:keikiapp/core/tracking/link_touch.dart';
import 'package:keikiapp/features/tracking/link_tracking_page.dart';
import 'package:keikiapp/gen/gateway_api.g.dart';

import 'helpers.dart';

class FakeLinkSource implements LinkSource {
  final controller = StreamController<Map<dynamic, dynamic>>.broadcast();
  final identities = <String>[];
  final handled = <String>[];
  int logouts = 0;
  bool inited = false;

  @override
  bool get enabled => true;

  @override
  String get label => 'Branch test • keikibook.app.link';

  @override
  Future<void> init() async => inited = true;

  @override
  Stream<Map<dynamic, dynamic>> sessions() => controller.stream;

  @override
  Future<Map<dynamic, dynamic>> latestParams() async => {'+clicked_branch_link': false};

  @override
  Future<Map<dynamic, dynamic>> firstParams() async => {'+is_first_session': true};

  @override
  void identify(String userId) => identities.add(userId);

  @override
  void logout() => logouts++;

  @override
  void handle(String url) => handled.add(url);

  @override
  void validate() {}

  @override
  Stream<String> get logs => const Stream.empty();
}

class _Executor implements GatewayExecutor {
  _Executor(this.respond);

  Object? Function(String op, Map<String, dynamic> vars) respond;
  final calls = <(String, Map<String, dynamic>)>[];

  Future<Map<String, dynamic>> _run(String doc, Map<String, dynamic>? vars) async {
    final op = RegExp(r'\{\s*(\w+)').firstMatch(doc)!.group(1)!;
    calls.add((op, vars ?? const {}));
    return {op: respond(op, vars ?? const {})};
  }

  @override
  Future<Map<String, dynamic>> query(String document, {Map<String, dynamic>? variables}) => _run(document, variables);

  @override
  Future<Map<String, dynamic>> mutation(String document, {Map<String, dynamic>? variables}) => _run(document, variables);

  @override
  Future<Map<String, dynamic>> public(String document, {Map<String, dynamic>? variables}) => _run(document, variables);
}

const branchParams = {
  '+clicked_branch_link': true,
  '+is_first_session': false,
  '~referring_link': 'https://keikibook.app.link/aBc123',
  '~channel': 'zalo',
  '~campaign': 'sale_tet_2027',
  '~feature': 'referral',
  'ref_code': 'abc12345',
  'click_id': 'click-1',
  'utm_medium': 'referral',
  r'$deeplink_path': 'services/ticket',
};

void main() {
  group('LinkTouch', () {
    test('reads a clicked Branch link', () {
      final t = LinkTouch.fromBranch(branchParams)!;
      expect(t.clickedBranchLink, isTrue);
      expect(t.refCode, 'ABC12345');
      expect(t.clickId, 'click-1');
      expect(t.utm, {'utm_medium': 'referral', 'utm_source': 'zalo', 'utm_campaign': 'sale_tet_2027'});
      expect(t.route, '/services/ticket');
      expect(t.url, 'https://keikibook.app.link/aBc123');
    });

    test('splits ref_code with an embedded click id', () {
      final t = LinkTouch.fromBranch({'+clicked_branch_link': 'true', 'ref_code': 'XYZ:click-9'})!;
      expect(t.refCode, 'XYZ');
      expect(t.clickId, 'click-9');
    });

    test('non Branch links and empty sessions', () {
      expect(LinkTouch.fromBranch({'+clicked_branch_link': false}), isNull);
      final t = LinkTouch.fromBranch({'+clicked_branch_link': false, '+non_branch_link': 'keikibook://app/wallet?utm_source=sms'})!;
      expect(t.clickedBranchLink, isFalse);
      expect(t.route, '/wallet?utm_source=sms');
      expect(t.utm, {'utm_source': 'sms'});
    });

    test('reads referral QR and web links', () {
      final qr = LinkTouch.fromUri(Uri.parse('keikibook://open?ref=abc12'))!;
      expect(qr.refCode, 'ABC12');
      expect(qr.route, isNull);
      final web = LinkTouch.fromUri(Uri.parse('https://keikibook.vn/r/ZZ99?utm_source=fb'))!;
      expect(web.refCode, 'ZZ99');
      expect(web.channel, 'fb');
      expect(LinkTouch.fromUri(Uri.parse('https://example.com/a')), isNull);
    });

    test('json round trip', () {
      final t = LinkTouch.fromBranch(branchParams)!;
      final back = LinkTouch.fromJson(jsonDecode(jsonEncode(t.toJson())) as Map<String, dynamic>);
      expect(back.toJson(), t.toJson());
    });
  });

  group('AttributionCubit', () {
    late FakeLinkSource source;
    late _Executor executor;
    late MemoryAttributionStore store;
    late List<String> routes;

    AttributionCubit make() => AttributionCubit(source: source, api: GatewayApi(executor), deviceId: () async => 'device-1', store: store)..onNavigate = routes.add;

    setUp(() {
      source = FakeLinkSource();
      store = MemoryAttributionStore();
      routes = [];
      executor = _Executor((op, _) => {'attributed': true, 'owner_user_id': '42', 'reward_type': 'points'});
    });

    test('Branch link before login waits, navigates, then activates after login', () async {
      final cubit = make();
      await cubit.start();
      expect(source.inited, isTrue);
      source.controller.add(branchParams);
      await pumpEventQueue();
      expect(routes, ['/services/ticket']);
      expect(cubit.state.pending?.refCode, 'ABC12345');
      expect(executor.calls, isEmpty);
      expect(store.value, contains('ABC12345'));

      await cubit.onAuthenticated('7');
      expect(source.identities, ['7']);
      final (op, vars) = executor.calls.single;
      expect(op, 'referralActivate');
      expect(vars, containsPair('ref_code', 'ABC12345'));
      expect(vars, containsPair('click_id', 'click-1'));
      expect(vars, containsPair('device_id', 'device-1'));
      expect(vars, containsPair('utm_source', 'zalo'));
      expect(vars, containsPair('utm_campaign', 'sale_tet_2027'));
      expect(cubit.state.pending, isNull);
      expect(cubit.state.lastResult?.attributed, isTrue);
      expect(cubit.state.lastResult?.ownerUserId, '42');
      await cubit.close();
    });

    test('install from a Branch link uses confirm-install', () async {
      final cubit = make();
      await cubit.start();
      await cubit.onAuthenticated('7');
      source.controller.add({...branchParams, '+is_first_session': true});
      await pumpEventQueue();
      expect(executor.calls.single.$1, 'referralConfirmInstall');
      await cubit.close();
    });

    test('duplicate session events are recorded once', () async {
      final cubit = make();
      await cubit.start();
      source.controller
        ..add(branchParams)
        ..add(branchParams);
      await pumpEventQueue();
      expect(cubit.state.history, hasLength(1));
      expect(routes, hasLength(1));
      await cubit.close();
    });

    test('permanent errors drop the pending link, network errors keep it', () async {
      executor.respond = (_, _) => throw const GraphQLException('confirm install: self referral not allowed');
      final cubit = make();
      await cubit.start();
      await cubit.record(LinkTouch.fromBranch(branchParams)!);
      await cubit.onAuthenticated('42');
      expect(cubit.state.pending, isNull);
      expect(cubit.state.lastResult?.error, contains('self referral'));

      executor.respond = (_, _) => throw Exception('SocketException');
      await cubit.record(LinkTouch.fromUri(Uri.parse('keikibook://open?ref=NEW1'))!);
      expect(cubit.state.pending?.refCode, 'NEW1');
      executor.respond = (_, _) => {'attributed': true};
      await cubit.retry();
      expect(cubit.state.pending, isNull);
      expect(cubit.state.lastResult?.refCode, 'NEW1');
      await cubit.close();
    });

    test('pending link survives an app restart', () async {
      final first = make();
      await first.start();
      source.controller.add(branchParams);
      await pumpEventQueue();
      await first.close();

      source = FakeLinkSource();
      final second = make();
      await second.start();
      expect(second.state.pending?.refCode, 'ABC12345');
      expect(second.state.history, hasLength(1));
      await second.onAuthenticated('7');
      expect(executor.calls.single.$1, 'referralActivate');
      await second.close();
    });

    test('manual code reports server errors to the form', () async {
      executor.respond = (_, _) => throw const GraphQLException('referral link not found');
      final cubit = make();
      await cubit.start();
      await cubit.onAuthenticated('7');
      await expectLater(cubit.applyManual('nope'), throwsA(isA<GraphQLException>()));
      expect(cubit.state.pending, isNull);
      await cubit.close();
    });

    test('disabled source still tracks deep links', () async {
      final cubit = AttributionCubit(source: const DisabledLinkSource(), api: GatewayApi(executor), deviceId: () async => 'd', store: store);
      await cubit.start();
      expect(cubit.state.enabled, isFalse);
      cubit.observeUri(Uri.parse('keikibook://open?ref=QR1&utm_source=poster'), source: LinkTouchSource.scan);
      await pumpEventQueue();
      expect(cubit.state.pending?.refCode, 'QR1');
      expect(cubit.state.pending?.source, LinkTouchSource.scan);
      await cubit.close();
    });
  });

  test('Branch domains', () {
    expect(Env.branchLinkDomain, 'keikibook.app.link');
    expect(Env.branchAlternateDomain, 'keikibook-alternate.app.link');
    expect(Env.isBranchHost('keikibook-alternate.app.link'), isTrue);
    expect(Env.isBranchHost('keikibook.test-app.link'), isTrue);
    expect(Env.isBranchHost('keikibook.vn'), isFalse);
  });

  test('link inspection classifies links', () {
    expect(LinkInspection.of('https://keikibook.app.link/abc')!.kind, LinkKind.branch);
    expect(LinkInspection.of('keikibook://app/wallet')!.route, '/wallet');
    expect(LinkInspection.of('keiki://pay?wallet_id=w')!.kind, LinkKind.payment);
    expect(LinkInspection.of('https://ref.keikibook.vn/r/ABC')!.kind, LinkKind.referralPage);
    expect(LinkInspection.of('https://keikibook.vn/services/movie')!.kind, LinkKind.website);
    expect(LinkInspection.of('not a link'), isNull);
  });

  group('in the app', () {
    late FakeLinkSource source;

    GraphQLClient client() => GraphQLClient(
          endpoint: 'http://gw/query',
          storage: MemoryTokenStorage(validTokens()),
          httpClient: MockClient((req) async {
            final query = (jsonDecode(req.body) as Map<String, dynamic>)['query'] as String;
            Object data = <String, Object?>{};
            if (query.contains('me {')) {
              data = {
                'me': {'id': 7, 'username': 'quan', 'name': 'Quan', 'roles': []},
              };
            } else if (query.contains('referralActivate')) {
              data = {
                'referralActivate': {'attributed': true, 'owner_user_id': '42'},
              };
            } else if (query.contains('referralLink(')) {
              data = {
                'referralLink': {'ref_code': 'ABC12345', 'owner_user_id': '42', 'status': 'active', 'channel': 'zalo'},
              };
            } else if (query.contains('referralStatus')) {
              data = {
                'referralStatus': {'ref_code': 'ABC12345', 'status': 'active', 'invitation_status': 'pending'},
              };
            }
            return http.Response(jsonEncode({'data': data}), 200, headers: {'content-type': 'application/json'});
          }),
        );

    Future<void> pump(WidgetTester tester) async {
      source = FakeLinkSource();
      tester.view.physicalSize = const Size(1170, 2532);
      tester.view.devicePixelRatio = 3;
      addTearDown(tester.view.reset);
      await tester.pumpWidget(KeikiApp(client: client(), catalogLoader: loadTestCatalog, linkSource: source, attributionStore: MemoryAttributionStore()));
      await tester.pumpAndSettle();
    }

    testWidgets('Branch link opens its screen and is attributed to the signed in user', (tester) async {
      await pump(tester);
      expect(source.identities, ['7']);
      source.controller.add(branchParams);
      await tester.pumpAndSettle();
      final router = GoRouter.of(tester.element(find.byType(Navigator).first));
      expect(router.routerDelegate.currentConfiguration.last.matchedLocation, '/services/ticket');
      final cubit = BlocProvider.of<AttributionCubit>(tester.element(find.byType(Navigator).first));
      expect(cubit.state.lastResult?.attributed, isTrue);
      expect(cubit.state.lastResult?.ownerUserId, '42');
    });

    testWidgets('Branch domains opened by the OS do not hit an unknown route', (tester) async {
      await pump(tester);
      final router = GoRouter.of(tester.element(find.byType(Navigator).first));
      router.go('https://keikibook.app.link/aBc123');
      await tester.pumpAndSettle();
      expect(router.routerDelegate.currentConfiguration.uri.path, '/');
      expect(tester.takeException(), isNull);
    });

    testWidgets('tracking page checks a referral link against the server', (tester) async {
      await pump(tester);
      tester.view.physicalSize = const Size(1170, 9000);
      Navigator.of(tester.element(find.byType(Scaffold).first)).push(MaterialPageRoute<void>(builder: (_) => const LinkTrackingPage(initialLink: 'https://ref.keikibook.vn/r/ABC12345?utm_source=zalo')));
      await tester.pumpAndSettle();
      expect(find.text('Link mời qua referral-service (/r/…)'), findsOneWidget);
      expect(find.text('ABC12345'), findsOneWidget);
      expect(find.text('zalo'), findsWidgets);
      expect(find.text('active'), findsOneWidget);
      expect(find.text('pending'), findsOneWidget);
      await tester.tap(find.text('Mở như người dùng bấm'));
      await tester.pumpAndSettle();
      final cubit = BlocProvider.of<AttributionCubit>(tester.element(find.byType(Scaffold).first));
      expect(cubit.state.history.first.refCode, 'ABC12345');
      expect(tester.takeException(), isNull);
    });
  });
}
