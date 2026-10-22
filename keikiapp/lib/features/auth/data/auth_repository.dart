import '../../../core/network/graphql_client.dart';
import 'user.dart';

class AuthRepository {
  AuthRepository(this._client);

  final GraphQLClient _client;

  static const _userFields = 'id username email name phone address sex roles { id name }';

  Stream<void> get sessionExpired => _client.sessionExpired;

  Future<bool> hasSession() async => await _client.storage.read() != null;

  Future<void> login(String username, String password) async {
    final data = await _client.public(r'''
      mutation Login($input: LoginInput!) {
        login(input: $input) { accessToken refreshToken expiresIn }
      }
    ''', variables: {
      'input': {'username': username, 'password': password},
    });
    await _client.storage.save(GraphQLClient.tokensFromJson(data['login'] as Map<String, dynamic>));
  }

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
