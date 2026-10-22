// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PhotoApi {
  const PhotoApi(this._executor);

  final GatewayExecutor _executor;

  Future<PhotoComposition?> photoComposition({required String id, String? selectionSet}) async {
    final data = await _executor.query('query photoComposition(\$id: String!) { photoComposition(id: \$id) ${selectionSet ?? '{ id owner_id status layout { type rows cols width height spacing padding background cell_fit cells { x y width height image_index } format quality } sources { type url object_key order } object_key url width height format size_bytes error_message created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['photoComposition'] == null ? null : PhotoComposition.fromJson((data['photoComposition'] as Map).cast<String, dynamic>());
  }

  Future<List<PhotoComposition>?> myPhotoCompositions({int? limit, String? selectionSet}) async {
    final data = await _executor.query('query myPhotoCompositions(\$limit: Int) { myPhotoCompositions(limit: \$limit) ${selectionSet ?? '{ id owner_id status layout { type rows cols width height spacing padding background cell_fit cells { x y width height image_index } format quality } sources { type url object_key order } object_key url width height format size_bytes error_message created_at updated_at }'} }', variables: {
      'limit': ?limit,
    }
    );
    return (data['myPhotoCompositions'] as List?)?.map((e) => (e == null ? null : PhotoComposition.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PhotoComposition?> photoCreateComposition({List<GqlUpload>? images, List<String>? urls, String? layout, String? selectionSet}) async {
    final data = await _executor.mutation('mutation photoCreateComposition(\$images: [Upload!], \$urls: [String!], \$layout: String) { photoCreateComposition(images: \$images, urls: \$urls, layout: \$layout) ${selectionSet ?? '{ id owner_id status layout { type rows cols width height spacing padding background cell_fit cells { x y width height image_index } format quality } sources { type url object_key order } object_key url width height format size_bytes error_message created_at updated_at }'} }', variables: {
      'images': ?images,
      'urls': ?urls,
      'layout': ?layout,
    }
    );
    return data['photoCreateComposition'] == null ? null : PhotoComposition.fromJson((data['photoCreateComposition'] as Map).cast<String, dynamic>());
  }

}
