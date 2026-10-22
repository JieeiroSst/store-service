// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class DrawImageApi {
  const DrawImageApi(this._executor);

  final GatewayExecutor _executor;

  Future<DrawImageCollage?> drawImageCollage({required String id, String? selectionSet}) async {
    final data = await _executor.query('query drawImageCollage(\$id: String!) { drawImageCollage(id: \$id) ${selectionSet ?? '{ id url bucket object_key content_type size width height image_count layout columns created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['drawImageCollage'] == null ? null : DrawImageCollage.fromJson((data['drawImageCollage'] as Map).cast<String, dynamic>());
  }

  Future<DrawImageCollage?> drawImageCreateCollage({required List<GqlUpload> files, GqlUpload? center, String? layout, int? columns, String? selectionSet}) async {
    final data = await _executor.mutation('mutation drawImageCreateCollage(\$files: [Upload!]!, \$center: Upload, \$layout: String, \$columns: Int) { drawImageCreateCollage(files: \$files, center: \$center, layout: \$layout, columns: \$columns) ${selectionSet ?? '{ id url bucket object_key content_type size width height image_count layout columns created_at }'} }', variables: {
      'files': files,
      'center': ?center,
      'layout': ?layout,
      'columns': ?columns,
    }
    );
    return data['drawImageCreateCollage'] == null ? null : DrawImageCollage.fromJson((data['drawImageCreateCollage'] as Map).cast<String, dynamic>());
  }

}
