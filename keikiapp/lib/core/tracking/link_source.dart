import 'dart:async';

import 'package:flutter/foundation.dart';
import 'package:flutter_branch_sdk/flutter_branch_sdk.dart';

import '../config/env.dart';

abstract interface class LinkSource {
  bool get enabled;

  String get label;

  Future<void> init();

  Stream<Map<dynamic, dynamic>> sessions();

  Future<Map<dynamic, dynamic>> latestParams();

  Future<Map<dynamic, dynamic>> firstParams();

  void identify(String userId);

  void logout();

  void handle(String url);

  void validate();

  Stream<String> get logs;
}

class DisabledLinkSource implements LinkSource {
  const DisabledLinkSource([this.reason = 'Chưa cấu hình BRANCH_KEY']);

  final String reason;

  @override
  bool get enabled => false;

  @override
  String get label => reason;

  @override
  Future<void> init() async {}

  @override
  Stream<Map<dynamic, dynamic>> sessions() => const Stream.empty();

  @override
  Future<Map<dynamic, dynamic>> latestParams() async => const {};

  @override
  Future<Map<dynamic, dynamic>> firstParams() async => const {};

  @override
  void identify(String userId) {}

  @override
  void logout() {}

  @override
  void handle(String url) {}

  @override
  void validate() {}

  @override
  Stream<String> get logs => const Stream.empty();
}

class BranchLinkSource implements LinkSource {
  const BranchLinkSource();

  static bool get supported => !kIsWeb && (defaultTargetPlatform == TargetPlatform.android || defaultTargetPlatform == TargetPlatform.iOS);

  static LinkSource fromEnv() {
    if (!supported) return const DisabledLinkSource('Branch chỉ chạy trên Android/iOS');
    if (Env.branchKey.isEmpty) return const DisabledLinkSource();
    return const BranchLinkSource();
  }

  @override
  bool get enabled => true;

  @override
  String get label => '${Env.branchKey.startsWith('key_test') ? 'Branch test' : 'Branch live'} • ${Env.branchLinkDomain}';

  @override
  Future<void> init() => FlutterBranchSdk.init(enableLogging: kDebugMode, branchAttributionLevel: BranchAttributionLevel.FULL);

  @override
  Stream<Map<dynamic, dynamic>> sessions() => FlutterBranchSdk.listSession();

  @override
  Future<Map<dynamic, dynamic>> latestParams() => FlutterBranchSdk.getLatestReferringParams();

  @override
  Future<Map<dynamic, dynamic>> firstParams() => FlutterBranchSdk.getFirstReferringParams();

  @override
  void identify(String userId) => FlutterBranchSdk.setIdentity(userId);

  @override
  void logout() => FlutterBranchSdk.logout();

  @override
  void handle(String url) => FlutterBranchSdk.handleDeepLink(url);

  @override
  void validate() => FlutterBranchSdk.validateSDKIntegration();

  @override
  Stream<String> get logs => FlutterBranchSdk.platformLogs;
}
