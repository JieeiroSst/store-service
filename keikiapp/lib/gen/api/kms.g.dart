// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class KmsApi {
  const KmsApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<KmsKey>?> kmsKeys({String? selectionSet}) async {
    final data = await _executor.query('query kmsKeys { kmsKeys ${selectionSet ?? '{ id alias algorithm key_length created_at updated_at expires_at last_rotated_at status version created_by tags use_count }'} }'
    );
    return (data['kmsKeys'] as List?)?.map((e) => (e == null ? null : KmsKey.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<KmsKey?> kmsKey({required String id, String? selectionSet}) async {
    final data = await _executor.query('query kmsKey(\$id: String!) { kmsKey(id: \$id) ${selectionSet ?? '{ id alias algorithm key_length created_at updated_at expires_at last_rotated_at status version created_by tags use_count }'} }', variables: {
      'id': id,
    }
    );
    return data['kmsKey'] == null ? null : KmsKey.fromJson((data['kmsKey'] as Map).cast<String, dynamic>());
  }

  Future<KmsKeyUse?> kmsKeyForUse({required String id, String? selectionSet}) async {
    final data = await _executor.query('query kmsKeyForUse(\$id: String!) { kmsKeyForUse(id: \$id) ${selectionSet ?? '{ key { id alias algorithm key_length created_at updated_at expires_at last_rotated_at status version created_by tags use_count } message }'} }', variables: {
      'id': id,
    }
    );
    return data['kmsKeyForUse'] == null ? null : KmsKeyUse.fromJson((data['kmsKeyForUse'] as Map).cast<String, dynamic>());
  }

  Future<KmsKeyStats?> kmsKeyUsageStats({required String id, String? selectionSet}) async {
    final data = await _executor.query('query kmsKeyUsageStats(\$id: String!) { kmsKeyUsageStats(id: \$id) ${selectionSet ?? '{ key_id alias use_count created_at last_rotated_at expires_at status version }'} }', variables: {
      'id': id,
    }
    );
    return data['kmsKeyUsageStats'] == null ? null : KmsKeyStats.fromJson((data['kmsKeyUsageStats'] as Map).cast<String, dynamic>());
  }

  Future<KmsKeyAuditLogs?> kmsKeyAuditLogs({required String id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query kmsKeyAuditLogs(\$id: String!, \$limit: Int, \$offset: Int) { kmsKeyAuditLogs(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ key_id logs { id actor_id actor_name action resource resource_id timestamp ip_address user_agent success error_msg metadata level } limit offset }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['kmsKeyAuditLogs'] == null ? null : KmsKeyAuditLogs.fromJson((data['kmsKeyAuditLogs'] as Map).cast<String, dynamic>());
  }

  Future<KmsAuditLogs?> kmsAuditLogs({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query kmsAuditLogs(\$limit: Int, \$offset: Int) { kmsAuditLogs(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ logs { id actor_id actor_name action resource resource_id timestamp ip_address user_agent success error_msg metadata level } limit offset }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['kmsAuditLogs'] == null ? null : KmsAuditLogs.fromJson((data['kmsAuditLogs'] as Map).cast<String, dynamic>());
  }

  Future<KmsKey?> kmsCreateKey({required KmsCreateKeyInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation kmsCreateKey(\$input: KmsCreateKeyInput!) { kmsCreateKey(input: \$input) ${selectionSet ?? '{ id alias algorithm key_length created_at updated_at expires_at last_rotated_at status version created_by tags use_count }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['kmsCreateKey'] == null ? null : KmsKey.fromJson((data['kmsCreateKey'] as Map).cast<String, dynamic>());
  }

  Future<String?> kmsRotateKey({required String id, bool? force}) async {
    final data = await _executor.mutation('mutation kmsRotateKey(\$id: String!, \$force: Boolean) { kmsRotateKey(id: \$id, force: \$force) }', variables: {
      'id': id,
      'force': ?force,
    }
    );
    return data['kmsRotateKey']?.toString();
  }

  Future<bool?> kmsDeleteKey({required String id}) async {
    final data = await _executor.mutation('mutation kmsDeleteKey(\$id: String!) { kmsDeleteKey(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['kmsDeleteKey'] as bool?;
  }

}
