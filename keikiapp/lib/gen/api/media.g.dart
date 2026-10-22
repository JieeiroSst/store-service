// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class MediaApi {
  const MediaApi(this._executor);

  final GatewayExecutor _executor;

  Future<MediaVideo?> mediaVideo({required int videoId, required int userId, String? selectionSet}) async {
    final data = await _executor.query('query mediaVideo(\$video_id: Int!, \$user_id: Int!) { mediaVideo(video_id: \$video_id, user_id: \$user_id) ${selectionSet ?? '{ video_id description thumbnail_url stream_url tag_id uploaded_on tag { tag_id video_id value } }'} }', variables: {
      'video_id': videoId,
      'user_id': userId,
    }
    );
    return data['mediaVideo'] == null ? null : MediaVideo.fromJson((data['mediaVideo'] as Map).cast<String, dynamic>());
  }

  Future<MediaSearchResult?> mediaSearchVideos({String? query, int? page, int? size, String? selectionSet}) async {
    final data = await _executor.query('query mediaSearchVideos(\$query: String, \$page: Int, \$size: Int) { mediaSearchVideos(query: \$query, page: \$page, size: \$size) ${selectionSet ?? '{ videos { video_id description thumbnail_url stream_url tag_id uploaded_on tag { tag_id video_id value } } total page size pages }'} }', variables: {
      'query': ?query,
      'page': ?page,
      'size': ?size,
    }
    );
    return data['mediaSearchVideos'] == null ? null : MediaSearchResult.fromJson((data['mediaSearchVideos'] as Map).cast<String, dynamic>());
  }

  Future<String?> mediaUploadVideo({required GqlUpload video}) async {
    final data = await _executor.mutation('mutation mediaUploadVideo(\$video: Upload!) { mediaUploadVideo(video: \$video) }', variables: {
      'video': video,
    }
    );
    return data['mediaUploadVideo']?.toString();
  }

  Future<bool?> mediaSaveVideo({required Map<String, dynamic> video, Map<String, dynamic>? tag}) async {
    final data = await _executor.mutation('mutation mediaSaveVideo(\$video: Map!, \$tag: Map) { mediaSaveVideo(video: \$video, tag: \$tag) }', variables: {
      'video': video,
      'tag': ?tag,
    }
    );
    return data['mediaSaveVideo'] as bool?;
  }

}
