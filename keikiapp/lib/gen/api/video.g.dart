// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class VideoApi {
  const VideoApi(this._executor);

  final GatewayExecutor _executor;

  Future<VideoVideoList?> videos({int? page, int? pageSize, String? sort, String? q, String? selectionSet}) async {
    final data = await _executor.public('query videos(\$page: Int, \$page_size: Int, \$sort: String, \$q: String) { videos(page: \$page, page_size: \$page_size, sort: \$sort, q: \$q) ${selectionSet ?? '{ items { id title description content_type size status duration width height created_at views } total }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'sort': ?sort,
      'q': ?q,
    }
    );
    return data['videos'] == null ? null : VideoVideoList.fromJson((data['videos'] as Map).cast<String, dynamic>());
  }

  Future<VideoVideo?> video({required String id, String? selectionSet}) async {
    final data = await _executor.public('query video(\$id: String!) { video(id: \$id) ${selectionSet ?? '{ id title description content_type size status duration width height created_at views }'} }', variables: {
      'id': id,
    }
    );
    return data['video'] == null ? null : VideoVideo.fromJson((data['video'] as Map).cast<String, dynamic>());
  }

  Future<List<VideoVideo>?> videoRelated({required String id, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query videoRelated(\$id: String!, \$limit: Int) { videoRelated(id: \$id, limit: \$limit) ${selectionSet ?? '{ id title description content_type size status duration width height created_at views }'} }', variables: {
      'id': id,
      'limit': ?limit,
    }
    );
    return (data['videoRelated'] as List?)?.map((e) => (e == null ? null : VideoVideo.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VideoVideo?> videoUpload({required GqlUpload file, String? title, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation videoUpload(\$file: Upload!, \$title: String, \$description: String) { videoUpload(file: \$file, title: \$title, description: \$description) ${selectionSet ?? '{ id title description content_type size status duration width height created_at views }'} }', variables: {
      'file': file,
      'title': ?title,
      'description': ?description,
    }
    );
    return data['videoUpload'] == null ? null : VideoVideo.fromJson((data['videoUpload'] as Map).cast<String, dynamic>());
  }

  Future<Object?> videoView({required String id}) async {
    final data = await _executor.public('mutation videoView(\$id: String!) { videoView(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['videoView'];
  }

  Future<bool?> videoDelete({required String id}) async {
    final data = await _executor.mutation('mutation videoDelete(\$id: String!) { videoDelete(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['videoDelete'] as bool?;
  }

}
