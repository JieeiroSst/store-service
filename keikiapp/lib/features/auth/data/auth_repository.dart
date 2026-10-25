import '../../../core/network/graphql_client.dart';
import '../../../core/network/graphql_exception.dart';
import '../../../gen/api/user.g.dart';
import '../../../gen/models.g.dart' show MfaStatus, TotpSetup;
import 'user.dart';

class MfaRequiredException implements Exception {
  const MfaRequiredException(this.mfaToken, {this.method = 'totp'});

  final String mfaToken;
  final String method;

  @override
  String toString() => 'Vui lòng nhập mã xác thực 2 lớp';
}

class AuthRepository {
  AuthRepository(this._client) : _api = UserApi(_client);

  final GraphQLClient _client;
  final UserApi _api;

  static const _userFields = 'id username email name phone address sex roles { id name }';

  Stream<void> get sessionExpired => _client.sessionExpired;

  Future<bool> hasSession() async => await _client.storage.read() != null;

  Future<void> login(String username, String password) async {
    final Map<String, dynamic> data;
    try {
      data = await _client.public(r'''
        mutation Login($input: LoginInput!) {
          login(input: $input) { accessToken refreshToken expiresIn }
        }
      ''', variables: {
        'input': {'username': username, 'password': password},
      });
    } on GraphQLException catch (e) {
      final token = e.details['mfa_token'];
      if (e.reason == 'AUTH_MFA_REQUIRED' && token != null && token.isNotEmpty) {
        throw MfaRequiredException(token, method: e.details['mfa_method'] ?? 'totp');
      }
      rethrow;
    }
    await _client.storage.save(GraphQLClient.tokensFromJson(data['login'] as Map<String, dynamic>));
  }

  Future<void> verifyTotpLogin(String mfaToken, String code) async {
    final token = await _api.verifyTotpLogin(mfaToken: mfaToken, code: code);
    if (token == null) throw const GraphQLException('Xác thực thất bại');
    await _client.storage.save(GraphQLClient.tokensFromJson({
      'accessToken': token.accessToken,
      'refreshToken': token.refreshToken,
      'expiresIn': token.expiresIn,
    }));
  }

  Future<MfaStatus> mfaStatus() async => await _api.mfaStatus() ?? const MfaStatus(enabled: false, recoveryCodesRemaining: 0);

  Future<TotpSetup> setupTotp() async {
    final setup = await _api.setupTotp();
    if (setup == null) throw const GraphQLException('Không tạo được khoá bí mật');
    return setup;
  }

  Future<List<String>> enableTotp(String code) async => (await _api.enableTotp(code: code))?.recoveryCodes ?? const [];

  Future<void> disableTotp(String code) => _api.disableTotp(code: code);

  Future<List<String>> regenerateRecoveryCodes(String code) async =>
      (await _api.regenerateMfaRecoveryCodes(code: code))?.recoveryCodes ?? const [];

  Future<void> signUp({
    required String username,
    required String password,
    required String email,
    required String name,
    String? phone,
  }) async {
    await _client.public(r'''
      mutation SignUp($input: SignUpInput!) { signUp(input: $input) { message } }
    ''', variables: {
      'input': {
        'username': username,
        'password': password,
        'email': email,
        'name': name,
        if (phone != null && phone.isNotEmpty) 'phone': phone,
      },
    });
  }

  Future<User?> me() async {
    final data = await _client.query('query Me { me { $_userFields } }');
    final me = data['me'];
    return me == null ? null : User.fromJson(me as Map<String, dynamic>);
  }

  Future<User?> updateProfile({String? name, String? email, String? phone, String? address, String? sex}) async {
    final data = await _client.mutation('''
      mutation UpdateProfile(\$input: UpdateProfileInput!) {
        updateProfile(input: \$input) { message user { $_userFields } }
      }
    ''', variables: {
      'input': {
        'name': ?name,
        'email': ?email,
        'phone': ?phone,
        'address': ?address,
        'sex': ?sex,
      },
    });
    final user = (data['updateProfile'] as Map<String, dynamic>)['user'];
    return user == null ? null : User.fromJson(user as Map<String, dynamic>);
  }

  Future<void> logout() async {
    try {
      await _client.mutation('mutation Logout { logout }');
    } finally {
      await _client.storage.clear();
    }
  }
}
