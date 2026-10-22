// ignore_for_file: deprecated_member_use

import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';

import 'push_service.dart';

class FirebasePushConfig {
  static const apiKey = String.fromEnvironment('FIREBASE_API_KEY');
  static const appId = String.fromEnvironment('FIREBASE_APP_ID');
  static const senderId = String.fromEnvironment('FIREBASE_SENDER_ID');
  static const projectId = String.fromEnvironment('FIREBASE_PROJECT_ID');
  static const iosBundleId = String.fromEnvironment('FIREBASE_IOS_BUNDLE_ID');

  static bool get configured =>
      !kIsWeb && apiKey.isNotEmpty && appId.isNotEmpty && senderId.isNotEmpty && projectId.isNotEmpty;

  static FirebaseOptions get options => FirebaseOptions(
        apiKey: apiKey,
        appId: appId,
        messagingSenderId: senderId,
        projectId: projectId,
        iosBundleId: iosBundleId.isEmpty ? null : iosBundleId,
      );
}

PushMessage _toMessage(RemoteMessage m) =>
    PushMessage(title: m.notification?.title, body: m.notification?.body, data: Map<String, dynamic>.from(m.data));

class FirebasePushTransport implements PushTransport {
  FirebaseMessaging get _m => FirebaseMessaging.instance;

  @override
  Future<void> init() async {
    if (Firebase.apps.isEmpty) await Firebase.initializeApp(options: FirebasePushConfig.options);
  }

  @override
  Future<bool> requestPermission() async {
    final s = await _m.requestPermission();
    return s.authorizationStatus == AuthorizationStatus.authorized || s.authorizationStatus == AuthorizationStatus.provisional;
  }

  @override
  Future<String?> token() => _m.getToken();

  @override
  Stream<String> get tokenRefresh => _m.onTokenRefresh;

  @override
  Stream<PushMessage> get foreground => FirebaseMessaging.onMessage.map(_toMessage);

  @override
  Stream<PushMessage> get opened => FirebaseMessaging.onMessageOpenedApp.map(_toMessage);

  @override
  Future<PushMessage?> initialMessage() async {
    final m = await _m.getInitialMessage();
    return m == null ? null : _toMessage(m);
  }

  @override
  String get platform => defaultTargetPlatform == TargetPlatform.iOS ? 'ios' : 'android';
}
