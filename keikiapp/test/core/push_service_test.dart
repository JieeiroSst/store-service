import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:keikiapp/core/push/push_service.dart';
import 'package:keikiapp/features/notifications/data/notification_repository.dart';
import 'package:mocktail/mocktail.dart';

class _MockRepo extends Mock implements NotificationRepository {}

class _FakeTransport implements PushTransport {
  bool granted = true;
  String? current = 'token-1';
  PushMessage? initial;
  final refresh = StreamController<String>.broadcast();
  final fg = StreamController<PushMessage>.broadcast();
  final open = StreamController<PushMessage>.broadcast();

  @override
  Future<void> init() async {}

  @override
  Future<bool> requestPermission() async => granted;

  @override
  Future<String?> token() async => current;

  @override
  Stream<String> get tokenRefresh => refresh.stream;

  @override
  Stream<PushMessage> get foreground => fg.stream;

  @override
  Stream<PushMessage> get opened => open.stream;

  @override
  Future<PushMessage?> initialMessage() async => initial;

  @override
  String get platform => 'android';
}

void main() {
  late _MockRepo repo;
  late _FakeTransport transport;

  setUp(() {
    repo = _MockRepo();
    transport = _FakeTransport();
    when(() => repo.registerDevice(pushToken: any(named: 'pushToken'), deviceType: any(named: 'deviceType'))).thenAnswer((_) async {});
    when(() => repo.unregisterThisDevice()).thenAnswer((_) async {});
  });

  test('registers the device on login, re-registers on refresh, unregisters on logout', () async {
    final push = PushService(repo, transport);
    await push.start();
    transport.refresh.add('ignored-before-login');
    await Future<void>.delayed(Duration.zero);
    verifyNever(() => repo.registerDevice(pushToken: any(named: 'pushToken'), deviceType: any(named: 'deviceType')));

    await push.onLogin();
    await push.onLogin();
    verify(() => repo.registerDevice(pushToken: 'token-1', deviceType: 'android')).called(1);

    transport.refresh.add('token-2');
    await Future<void>.delayed(Duration.zero);
    verify(() => repo.registerDevice(pushToken: 'token-2', deviceType: 'android')).called(1);

    await push.onLogout();
    verify(() => repo.unregisterThisDevice()).called(1);
    await push.onLogout();
    verifyNoMoreInteractions(repo);
  });

  test('does not register without notification permission', () async {
    transport.granted = false;
    final push = PushService(repo, transport);
    await push.onLogin();
    verifyNever(() => repo.registerDevice(pushToken: any(named: 'pushToken'), deviceType: any(named: 'deviceType')));
  });

  test('routes opened notifications to the matching screen', () async {
    transport.initial = const PushMessage(data: {'order_id': 'KS00000012'});
    final push = PushService(repo, transport);
    final routes = <String>[];
    push.navigate.listen(routes.add);

    await push.start();
    transport.open.add(const PushMessage(data: {'type': 'order_update', 'reference_id': '42'}));
    transport.open.add(const PushMessage(data: {'route': '/wallet'}));
    transport.open.add(const PushMessage(data: {'route': 'https://evil.example'}));
    await Future<void>.delayed(Duration.zero);

    expect(routes, ['/shop/purchases', '/food/orders/42', '/wallet']);
  });

  test('is a no-op without Firebase configuration', () async {
    final push = PushService(repo, null);
    expect(push.enabled, isFalse);
    await push.start();
    await push.onLogin();
    await push.onLogout();
    verifyZeroInteractions(repo);
  });
}
