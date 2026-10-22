// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class UploadApi {
  const UploadApi(this._executor);

  final GatewayExecutor _executor;

  Future<UploadFileList?> uploadFiles({String? receiverId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query uploadFiles(\$receiver_id: String, \$limit: Int, \$offset: Int) { uploadFiles(receiver_id: \$receiver_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ files { id receiver_id file_name content_type size sha256 created_at updated_at url } total_count }'} }', variables: {
      'receiver_id': ?receiverId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['uploadFiles'] == null ? null : UploadFileList.fromJson((data['uploadFiles'] as Map).cast<String, dynamic>());
  }

  Future<UploadFile?> uploadFile({required String id, String? selectionSet}) async {
    final data = await _executor.query('query uploadFile(\$id: String!) { uploadFile(id: \$id) ${selectionSet ?? '{ id receiver_id file_name content_type size sha256 created_at updated_at url }'} }', variables: {
      'id': id,
    }
    );
    return data['uploadFile'] == null ? null : UploadFile.fromJson((data['uploadFile'] as Map).cast<String, dynamic>());
  }

  Future<UploadFile?> uploadFileMutation({required GqlUpload file, String? receiverId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation uploadFile(\$file: Upload!, \$receiver_id: String) { uploadFile(file: \$file, receiver_id: \$receiver_id) ${selectionSet ?? '{ id receiver_id file_name content_type size sha256 created_at updated_at url }'} }', variables: {
      'file': file,
      'receiver_id': ?receiverId,
    }
    );
    return data['uploadFile'] == null ? null : UploadFile.fromJson((data['uploadFile'] as Map).cast<String, dynamic>());
  }

  Future<UploadFile?> uploadReplaceFile({required String id, required GqlUpload file, String? selectionSet}) async {
    final data = await _executor.mutation('mutation uploadReplaceFile(\$id: String!, \$file: Upload!) { uploadReplaceFile(id: \$id, file: \$file) ${selectionSet ?? '{ id receiver_id file_name content_type size sha256 created_at updated_at url }'} }', variables: {
      'id': id,
      'file': file,
    }
    );
    return data['uploadReplaceFile'] == null ? null : UploadFile.fromJson((data['uploadReplaceFile'] as Map).cast<String, dynamic>());
  }

  Future<bool?> uploadDeleteFile({required String id}) async {
    final data = await _executor.mutation('mutation uploadDeleteFile(\$id: String!) { uploadDeleteFile(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['uploadDeleteFile'] as bool?;
  }

}
