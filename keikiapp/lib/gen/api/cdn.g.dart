// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CdnApi {
  const CdnApi(this._executor);

  final GatewayExecutor _executor;

  Future<CdnFileList?> cdnFiles({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query cdnFiles(\$limit: Int, \$offset: Int) { cdnFiles(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ files { id file_name content_type size_bytes status created_at updated_at } }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['cdnFiles'] == null ? null : CdnFileList.fromJson((data['cdnFiles'] as Map).cast<String, dynamic>());
  }

  Future<CdnFile?> cdnFile({required String id, String? selectionSet}) async {
    final data = await _executor.query('query cdnFile(\$id: String!) { cdnFile(id: \$id) ${selectionSet ?? '{ id file_name content_type size_bytes status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['cdnFile'] == null ? null : CdnFile.fromJson((data['cdnFile'] as Map).cast<String, dynamic>());
  }

  Future<CdnDownload?> cdnDownloadUrl({required String id, bool? direct, String? selectionSet}) async {
    final data = await _executor.query('query cdnDownloadUrl(\$id: String!, \$direct: Boolean) { cdnDownloadUrl(id: \$id, direct: \$direct) ${selectionSet ?? '{ location }'} }', variables: {
      'id': id,
      'direct': ?direct,
    }
    );
    return data['cdnDownloadUrl'] == null ? null : CdnDownload.fromJson((data['cdnDownloadUrl'] as Map).cast<String, dynamic>());
  }

  Future<CdnPresign?> cdnPresign({required CdnPresignInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation cdnPresign(\$input: CdnPresignInput!) { cdnPresign(input: \$input) ${selectionSet ?? '{ file_id object_key upload_url expires_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['cdnPresign'] == null ? null : CdnPresign.fromJson((data['cdnPresign'] as Map).cast<String, dynamic>());
  }

  Future<CdnFile?> cdnConfirm({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation cdnConfirm(\$id: String!) { cdnConfirm(id: \$id) ${selectionSet ?? '{ id file_name content_type size_bytes status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['cdnConfirm'] == null ? null : CdnFile.fromJson((data['cdnConfirm'] as Map).cast<String, dynamic>());
  }

  Future<bool?> cdnDeleteFile({required String id}) async {
    final data = await _executor.mutation('mutation cdnDeleteFile(\$id: String!) { cdnDeleteFile(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['cdnDeleteFile'] as bool?;
  }

}
