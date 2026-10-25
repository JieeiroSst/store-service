// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class UserApi {
  const UserApi(this._executor);

  final GatewayExecutor _executor;

  Future<User?> me({String? selectionSet}) async {
    final data = await _executor.query('query me { me ${selectionSet ?? '{ id username email name phone address sex active createTime updateTime roles { id name } }'} }'
    );
    return data['me'] == null ? null : User.fromJson((data['me'] as Map).cast<String, dynamic>());
  }

  Future<UserList?> users({FindUsersInput? input, String? selectionSet}) async {
    final data = await _executor.query('query users(\$input: FindUsersInput) { users(input: \$input) ${selectionSet ?? '{ users { id username email name phone address sex active createTime updateTime roles { id name } } total }'} }', variables: {
      if (input != null) 'input': input.toJson(),
    }
    );
    return data['users'] == null ? null : UserList.fromJson((data['users'] as Map).cast<String, dynamic>());
  }

  Future<UserRole?> userRole({required int id, String? selectionSet}) async {
    final data = await _executor.query('query userRole(\$id: Int!) { userRole(id: \$id) ${selectionSet ?? '{ id name users { id username email name phone address sex active createTime updateTime } }'} }', variables: {
      'id': id,
    }
    );
    return data['userRole'] == null ? null : UserRole.fromJson((data['userRole'] as Map).cast<String, dynamic>());
  }

  Future<UserRoleList?> userRoles({int? page, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query userRoles(\$page: Int, \$limit: Int) { userRoles(page: \$page, limit: \$limit) ${selectionSet ?? '{ roles { id name users { id username email name phone address sex active createTime updateTime } } limit page sort totalRows totalPages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
    }
    );
    return data['userRoles'] == null ? null : UserRoleList.fromJson((data['userRoles'] as Map).cast<String, dynamic>());
  }

  Future<AuthToken?> login({required LoginInput input, String? selectionSet}) async {
    final data = await _executor.public('mutation login(\$input: LoginInput!) { login(input: \$input) ${selectionSet ?? '{ accessToken refreshToken expiresIn }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['login'] == null ? null : AuthToken.fromJson((data['login'] as Map).cast<String, dynamic>());
  }

  Future<UserMessage?> signUp({required SignUpInput input, String? selectionSet}) async {
    final data = await _executor.public('mutation signUp(\$input: SignUpInput!) { signUp(input: \$input) ${selectionSet ?? '{ message user { id username email name phone address sex active createTime updateTime roles { id name } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['signUp'] == null ? null : UserMessage.fromJson((data['signUp'] as Map).cast<String, dynamic>());
  }

  Future<AuthToken?> refreshToken({required String refreshToken, String? selectionSet}) async {
    final data = await _executor.public('mutation refreshToken(\$refreshToken: String!) { refreshToken(refreshToken: \$refreshToken) ${selectionSet ?? '{ accessToken refreshToken expiresIn }'} }', variables: {
      'refreshToken': refreshToken,
    }
    );
    return data['refreshToken'] == null ? null : AuthToken.fromJson((data['refreshToken'] as Map).cast<String, dynamic>());
  }

  Future<bool?> logout() async {
    final data = await _executor.mutation('mutation logout { logout }'
    );
    return data['logout'] as bool?;
  }

  Future<UserMessage?> updateProfile({required UpdateProfileInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateProfile(\$input: UpdateProfileInput!) { updateProfile(input: \$input) ${selectionSet ?? '{ message user { id username email name phone address sex active createTime updateTime roles { id name } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['updateProfile'] == null ? null : UserMessage.fromJson((data['updateProfile'] as Map).cast<String, dynamic>());
  }

  Future<UserMessage?> adminUpdateProfile({required AdminUpdateProfileInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation adminUpdateProfile(\$input: AdminUpdateProfileInput!) { adminUpdateProfile(input: \$input) ${selectionSet ?? '{ message user { id username email name phone address sex active createTime updateTime roles { id name } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['adminUpdateProfile'] == null ? null : UserMessage.fromJson((data['adminUpdateProfile'] as Map).cast<String, dynamic>());
  }

  Future<UserRoleMessage?> createUserRole({required UserRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createUserRole(\$input: UserRoleInput!) { createUserRole(input: \$input) ${selectionSet ?? '{ message role { id name users { id username email name phone address sex active createTime updateTime } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['createUserRole'] == null ? null : UserRoleMessage.fromJson((data['createUserRole'] as Map).cast<String, dynamic>());
  }

  Future<UserRoleMessage?> updateUserRole({required int id, required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateUserRole(\$id: Int!, \$name: String!) { updateUserRole(id: \$id, name: \$name) ${selectionSet ?? '{ message role { id name users { id username email name phone address sex active createTime updateTime } } }'} }', variables: {
      'id': id,
      'name': name_,
    }
    );
    return data['updateUserRole'] == null ? null : UserRoleMessage.fromJson((data['updateUserRole'] as Map).cast<String, dynamic>());
  }

  Future<String?> deleteUserRole({required int id}) async {
    final data = await _executor.mutation('mutation deleteUserRole(\$id: Int!) { deleteUserRole(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['deleteUserRole']?.toString();
  }

  Future<UserRoleMessage?> assignUserRole({required AssignUserRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation assignUserRole(\$input: AssignUserRoleInput!) { assignUserRole(input: \$input) ${selectionSet ?? '{ message role { id name users { id username email name phone address sex active createTime updateTime } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['assignUserRole'] == null ? null : UserRoleMessage.fromJson((data['assignUserRole'] as Map).cast<String, dynamic>());
  }

  Future<UserRoleMessage?> addUserRole({required AssignUserRoleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addUserRole(\$input: AssignUserRoleInput!) { addUserRole(input: \$input) ${selectionSet ?? '{ message role { id name users { id username email name phone address sex active createTime updateTime } } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['addUserRole'] == null ? null : UserRoleMessage.fromJson((data['addUserRole'] as Map).cast<String, dynamic>());
  }

  Future<String?> removeUserRoles({required int userId}) async {
    final data = await _executor.mutation('mutation removeUserRoles(\$userId: Int!) { removeUserRoles(userId: \$userId) }', variables: {
      'userId': userId,
    }
    );
    return data['removeUserRoles']?.toString();
  }

  Future<MfaStatus?> mfaStatus({String? selectionSet}) async {
    final data = await _executor.query('query mfaStatus { mfaStatus ${selectionSet ?? '{ enabled method enabledAt recoveryCodesRemaining }'} }'
    );
    return data['mfaStatus'] == null ? null : MfaStatus.fromJson((data['mfaStatus'] as Map).cast<String, dynamic>());
  }

  Future<TotpSetup?> setupTotp({String? selectionSet}) async {
    final data = await _executor.mutation('mutation setupTotp { setupTotp ${selectionSet ?? '{ secret otpauthUri qrCode issuer account digits period algorithm }'} }'
    );
    return data['setupTotp'] == null ? null : TotpSetup.fromJson((data['setupTotp'] as Map).cast<String, dynamic>());
  }

  Future<MfaRecoveryCodes?> enableTotp({required String code, String? selectionSet}) async {
    final data = await _executor.mutation('mutation enableTotp(\$code: String!) { enableTotp(code: \$code) ${selectionSet ?? '{ recoveryCodes }'} }', variables: {
      'code': code,
    }
    );
    return data['enableTotp'] == null ? null : MfaRecoveryCodes.fromJson((data['enableTotp'] as Map).cast<String, dynamic>());
  }

  Future<bool?> disableTotp({required String code}) async {
    final data = await _executor.mutation('mutation disableTotp(\$code: String!) { disableTotp(code: \$code) }', variables: {
      'code': code,
    }
    );
    return data['disableTotp'] as bool?;
  }

  Future<MfaRecoveryCodes?> regenerateMfaRecoveryCodes({required String code, String? selectionSet}) async {
    final data = await _executor.mutation('mutation regenerateMfaRecoveryCodes(\$code: String!) { regenerateMfaRecoveryCodes(code: \$code) ${selectionSet ?? '{ recoveryCodes }'} }', variables: {
      'code': code,
    }
    );
    return data['regenerateMfaRecoveryCodes'] == null ? null : MfaRecoveryCodes.fromJson((data['regenerateMfaRecoveryCodes'] as Map).cast<String, dynamic>());
  }

  Future<AuthToken?> verifyTotpLogin({required String mfaToken, required String code, String? selectionSet}) async {
    final data = await _executor.public('mutation verifyTotpLogin(\$mfaToken: String!, \$code: String!) { verifyTotpLogin(mfaToken: \$mfaToken, code: \$code) ${selectionSet ?? '{ accessToken refreshToken expiresIn }'} }', variables: {
      'mfaToken': mfaToken,
      'code': code,
    }
    );
    return data['verifyTotpLogin'] == null ? null : AuthToken.fromJson((data['verifyTotpLogin'] as Map).cast<String, dynamic>());
  }

}
