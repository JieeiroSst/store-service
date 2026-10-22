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
}
