import 'dart:async';

import 'package:flutter/foundation.dart';

import '../../features/notifications/data/notification_repository.dart';

class PushMessage {
  const PushMessage({this.title, this.body, this.data = const {}});

  final String? title;
  final String? body;
  final Map<String, dynamic> data;

  String? get route {
    final explicit = data['route'];
    if (explicit is String && explicit.startsWith('/')) return explicit;
    final orderId = '${data['order_id'] ?? ''}';
    if (orderId.startsWith('KS')) return '/shop/purchases';
    final reference = '${data['reference_id'] ?? ''}';
    if (data['type'] == 'order_update' && int.tryParse(reference) != null) return '/food/orders/$reference';
    return null;
  }
}

abstract interface class PushTransport {
  Future<void> init();
  Future<bool> requestPermission();
  Future<String?> token();
  Stream<String> get tokenRefresh;
  Stream<PushMessage> get foreground;
  Stream<PushMessage> get opened;
  Future<PushMessage?> initialMessage();
  String get platform;
}

class PushService {
  PushService(this._repository, this._transport);

  final NotificationRepository _repository;
  final PushTransport? _transport;
  final _foreground = StreamController<PushMessage>.broadcast();
  final _navigate = StreamController<String>.broadcast();
  final _subs = <StreamSubscription<dynamic>>[];
  bool _ready = false;
  bool _loggedIn = false;
  String? _registered;

  bool get enabled => _transport != null;

  Stream<PushMessage> get foreground => _foreground.stream;

  Stream<String> get navigate => _navigate.stream;

  Future<void> start() async {
    final t = _transport;
    if (t == null || _ready) return;
    try {
      await t.init();
      _ready = true;
    } catch (e) {
      debugPrint('push: init failed: $e');
      return;
    }
    _subs
      ..add(t.foreground.listen(_foreground.add))
      ..add(t.opened.listen(_open))
      ..add(t.tokenRefresh.listen((token) {
        if (_loggedIn) _register(token);
      }));
    final initial = await t.initialMessage();
    if (initial != null) _open(initial);
  }

  Future<void> onLogin() async {
    _loggedIn = true;
    final t = _transport;
    if (t == null) return;
    await start();
    if (!_ready || !await t.requestPermission()) return;
    final token = await t.token();
    if (token != null) await _register(token);
  }

  Future<void> onLogout() async {
    _loggedIn = false;
    if (_registered == null) return;
    _registered = null;
    try {
      await _repository.unregisterThisDevice();
    } catch (e) {
      debugPrint('push: unregister failed: $e');
    }
  }

  Future<void> _register(String token) async {
    if (token == _registered) return;
    try {
      await _repository.registerDevice(pushToken: token, deviceType: _transport!.platform);
      _registered = token;
    } catch (e) {
      debugPrint('push: register failed: $e');
    }
  }

  void _open(PushMessage m) {
    final route = m.route;
    if (route != null) _navigate.add(route);
  }

  Future<void> dispose() async {
    for (final s in _subs) {
      await s.cancel();
    }
    await _foreground.close();
    await _navigate.close();
  }
}
