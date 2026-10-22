import 'dart:io';

import 'package:keikiapp/core/storage/token_storage.dart';
import 'package:keikiapp/features/services/data/schema_catalog.dart';

class MemoryTokenStorage extends TokenStorage {
  MemoryTokenStorage([this.tokens]);

  AuthTokens? tokens;
  int clears = 0;

  @override
  Future<AuthTokens?> read() async => tokens;

  @override
  Future<void> save(AuthTokens tokens) async => this.tokens = tokens;

  @override
  Future<void> clear() async {
    clears++;
    tokens = null;
  }

  @override
  Future<String> deviceId() async => 'test-device';
}

AuthTokens validTokens({String access = 'access-1', String refresh = 'refresh-1'}) =>
    AuthTokens(accessToken: access, refreshToken: refresh, expiresAt: DateTime.now().add(const Duration(hours: 1)));

final testCatalog = SchemaCatalog.parse(File('assets/gateway_schema.json').readAsStringSync());

Future<SchemaCatalog> loadTestCatalog() => Future.value(testCatalog);
