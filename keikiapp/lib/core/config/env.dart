import 'dart:io' show Platform;

import 'package:flutter/foundation.dart';

class Env {
  static const _apiUrl = String.fromEnvironment('API_URL');

  static String get graphqlEndpoint {
    if (_apiUrl.isNotEmpty) return '$_apiUrl/query';
    if (!kIsWeb && Platform.isAndroid) return 'http://10.0.2.2:8080/query';
    return 'http://localhost:8080/query';
  }

  static const defaultCurrency = String.fromEnvironment('DEFAULT_CURRENCY', defaultValue: 'VND');

  static const walletIssuerBankId = String.fromEnvironment('WALLET_ISSUER_BANK_ID', defaultValue: '7a1e0000-0000-4000-8000-000000000001');

  static const _webHost = String.fromEnvironment('WEB_HOST');

  static String get webHost => _webHost.isEmpty ? 'keikibook.vn' : _webHost;

  static const branchKey = String.fromEnvironment('BRANCH_KEY');

  static const _branchLinkDomain = String.fromEnvironment('BRANCH_LINK_DOMAIN');

  static String get branchLinkDomain => _branchLinkDomain.isEmpty ? 'keikibook.app.link' : _branchLinkDomain;

  static String get branchAlternateDomain {
    final m = RegExp(r'^([^.]+)\.((test-)?app\.link)$').firstMatch(branchLinkDomain);
    return m == null ? branchLinkDomain : '${m[1]}-alternate.${m[2]}';
  }

  static bool isBranchHost(String host) => host == branchLinkDomain || host == branchAlternateDomain || host.endsWith('.app.link') || host.endsWith('.test-app.link');
}
