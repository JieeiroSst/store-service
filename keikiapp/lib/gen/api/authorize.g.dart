// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AuthorizeApi {
  const AuthorizeApi(this._executor);

  final GatewayExecutor _executor;

  Future<AuthzUserRoles?> myRoles({String? selectionSet}) async {
    final data = await _executor.query('query myRoles { myRoles ${selectionSet ?? '{ userId roles effectiveRoles primaryRole }'} }'
    );
    return data['myRoles'] == null ? null : AuthzUserRoles.fromJson((data['myRoles'] as Map).cast<String, dynamic>());
  }

  Future<List<AuthzPermission>?> myPermissions({String? selectionSet}) async {
    final data = await _executor.query('query myPermissions { myPermissions ${selectionSet ?? '{ object action grantedBy }'} }'
    );
    return (data['myPermissions'] as List?)?.map((e) => (e == null ? null : AuthzPermission.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<bool?> checkMyPermission({required String object, required String action}) async {
    final data = await _executor.query('query checkMyPermission(\$object: String!, \$action: String!) { checkMyPermission(object: \$object, action: \$action) }', variables: {
      'object': object,
      'action': action,
    }
    );
    return data['checkMyPermission'] as bool?;
  }

  Future<List<AuthzRole>?> authzRoles({String? selectionSet}) async {
    final data = await _executor.query('query authzRoles { authzRoles ${selectionSet ?? '{ name description rank inherits permissions { object action grantedBy } assignable }'} }'
    );
    return (data['authzRoles'] as List?)?.map((e) => (e == null ? null : AuthzRole.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AuthzUserRoles?> authzUserRoles({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query authzUserRoles(\$userId: String!) { authzUserRoles(userId: \$userId) ${selectionSet ?? '{ userId roles effectiveRoles primaryRole }'} }', variables: {
      'userId': userId,
    }
    );
    return data['authzUserRoles'] == null ? null : AuthzUserRoles.fromJson((data['authzUserRoles'] as Map).cast<String, dynamic>());
  }

  Future<List<AuthzPermission>?> authzUserPermissions({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query authzUserPermissions(\$userId: String!) { authzUserPermissions(userId: \$userId) ${selectionSet ?? '{ object action grantedBy }'} }', variables: {
      'userId': userId,
    }
    );
    return (data['authzUserPermissions'] as List?)?.map((e) => (e == null ? null : AuthzPermission.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CasbinRuleList?> casbinRules({int? page, int? limit, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query casbinRules(\$page: Int, \$limit: Int, \$sort: String) { casbinRules(page: \$page, limit: \$limit, sort: \$sort) ${selectionSet ?? '{ rows { id ptype v0 v1 v2 v3 v4 v5 } limit page sort totalRows totalPages }'} }', variables: {
      'page': ?page,
      'limit': ?limit,
      'sort': ?sort,
    }
    );
    return data['casbinRules'] == null ? null : CasbinRuleList.fromJson((data['casbinRules'] as Map).cast<String, dynamic>());
  }

  Future<CasbinRule?> casbinRule({required int id, String? selectionSet}) async {
    final data = await _executor.query('query casbinRule(\$id: Int!) { casbinRule(id: \$id) ${selectionSet ?? '{ id ptype v0 v1 v2 v3 v4 v5 }'} }', variables: {
      'id': id,
    }
    );
    return data['casbinRule'] == null ? null : CasbinRule.fromJson((data['casbinRule'] as Map).cast<String, dynamic>());
  }

  Future<AuthzUserRoles?> authzAssignRoles({required String userId, required List<String> roles, String? selectionSet}) async {
    final data = await _executor.mutation('mutation authzAssignRoles(\$userId: String!, \$roles: [String!]!) { authzAssignRoles(userId: \$userId, roles: \$roles) ${selectionSet ?? '{ userId roles effectiveRoles primaryRole }'} }', variables: {
      'userId': userId,
      'roles': roles,
    }
    );
    return data['authzAssignRoles'] == null ? null : AuthzUserRoles.fromJson((data['authzAssignRoles'] as Map).cast<String, dynamic>());
  }

  Future<AuthzUserRoles?> authzRevokeRoles({required String userId, required List<String> roles, String? selectionSet}) async {
    final data = await _executor.mutation('mutation authzRevokeRoles(\$userId: String!, \$roles: [String!]!) { authzRevokeRoles(userId: \$userId, roles: \$roles) ${selectionSet ?? '{ userId roles effectiveRoles primaryRole }'} }', variables: {
      'userId': userId,
      'roles': roles,
    }
    );
    return data['authzRevokeRoles'] == null ? null : AuthzUserRoles.fromJson((data['authzRevokeRoles'] as Map).cast<String, dynamic>());
  }

  Future<AuthzUserRoles?> authzSetUserRoles({required String userId, required List<String> roles, String? selectionSet}) async {
    final data = await _executor.mutation('mutation authzSetUserRoles(\$userId: String!, \$roles: [String!]!) { authzSetUserRoles(userId: \$userId, roles: \$roles) ${selectionSet ?? '{ userId roles effectiveRoles primaryRole }'} }', variables: {
      'userId': userId,
      'roles': roles,
    }
    );
    return data['authzSetUserRoles'] == null ? null : AuthzUserRoles.fromJson((data['authzSetUserRoles'] as Map).cast<String, dynamic>());
  }

  Future<bool?> authzRemoveUser({required String userId}) async {
    final data = await _executor.mutation('mutation authzRemoveUser(\$userId: String!) { authzRemoveUser(userId: \$userId) }', variables: {
      'userId': userId,
    }
    );
    return data['authzRemoveUser'] as bool?;
  }

  Future<AuthzResult?> createCasbinRule({required CasbinRuleInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createCasbinRule(\$input: CasbinRuleInput!) { createCasbinRule(input: \$input) ${selectionSet ?? '{ message error }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['createCasbinRule'] == null ? null : AuthzResult.fromJson((data['createCasbinRule'] as Map).cast<String, dynamic>());
  }

  Future<AuthzResult?> updateCasbinRule({required int id, required String field, required String value, String? selectionSet}) async {
    final data = await _executor.mutation('mutation updateCasbinRule(\$id: Int!, \$field: String!, \$value: String!) { updateCasbinRule(id: \$id, field: \$field, value: \$value) ${selectionSet ?? '{ message error }'} }', variables: {
      'id': id,
      'field': field,
      'value': value,
    }
    );
    return data['updateCasbinRule'] == null ? null : AuthzResult.fromJson((data['updateCasbinRule'] as Map).cast<String, dynamic>());
  }

  Future<bool?> deleteCasbinRule({required int id}) async {
    final data = await _executor.mutation('mutation deleteCasbinRule(\$id: Int!) { deleteCasbinRule(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['deleteCasbinRule'] as bool?;
  }

  Future<OtpToken?> createOtp({required String userId, required String username, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createOtp(\$userId: String!, \$username: String!) { createOtp(userId: \$userId, username: \$username) ${selectionSet ?? '{ otp expiresAt }'} }', variables: {
      'userId': userId,
      'username': username,
    }
    );
    return data['createOtp'] == null ? null : OtpToken.fromJson((data['createOtp'] as Map).cast<String, dynamic>());
  }

  Future<OtpVerification?> verifyOtp({required String otp, String? selectionSet}) async {
    final data = await _executor.mutation('mutation verifyOtp(\$otp: String!) { verifyOtp(otp: \$otp) ${selectionSet ?? '{ valid message }'} }', variables: {
      'otp': otp,
    }
    );
    return data['verifyOtp'] == null ? null : OtpVerification.fromJson((data['verifyOtp'] as Map).cast<String, dynamic>());
  }

}
