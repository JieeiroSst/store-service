import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:uuid/uuid.dart';

class AuthTokens {
  const AuthTokens({required this.accessToken, required this.refreshToken, required this.expiresAt});

  final String accessToken;
  final String refreshToken;
  final DateTime expiresAt;

  bool get isExpired => DateTime.now().isAfter(expiresAt.subtract(const Duration(seconds: 30)));
}

class TokenStorage {
  TokenStorage([FlutterSecureStorage? storage]) : _storage = storage ?? const FlutterSecureStorage();

  final FlutterSecureStorage _storage;

  static const _accessKey = 'access_token';
  static const _refreshKey = 'refresh_token';
  static const _expiresKey = 'expires_at';
  static const _deviceKey = 'device_id';

  AuthTokens? _cache;
  String? _deviceId;

  Future<AuthTokens?> read() async {
    if (_cache != null) return _cache;
    final access = await _storage.read(key: _accessKey);
    final refresh = await _storage.read(key: _refreshKey);
    final expires = await _storage.read(key: _expiresKey);
    if (access == null || refresh == null) return null;
    _cache = AuthTokens(
      accessToken: access,
      refreshToken: refresh,
      expiresAt: DateTime.fromMillisecondsSinceEpoch(int.tryParse(expires ?? '') ?? 0),
    );
    return _cache;
  }

  Future<void> save(AuthTokens tokens) async {
    _cache = tokens;
    await _storage.write(key: _accessKey, value: tokens.accessToken);
    await _storage.write(key: _refreshKey, value: tokens.refreshToken);
    await _storage.write(key: _expiresKey, value: tokens.expiresAt.millisecondsSinceEpoch.toString());
  }

  Future<void> clear() async {
    _cache = null;
    await _storage.delete(key: _accessKey);
    await _storage.delete(key: _refreshKey);
    await _storage.delete(key: _expiresKey);
  }

  Future<String> deviceId() async {
    if (_deviceId != null) return _deviceId!;
    var id = await _storage.read(key: _deviceKey);
    if (id == null) {
      id = const Uuid().v4();
      await _storage.write(key: _deviceKey, value: id);
    }
    return _deviceId = id;
  }
}
