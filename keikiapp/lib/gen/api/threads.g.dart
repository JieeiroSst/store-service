// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ThreadsApi {
  const ThreadsApi(this._executor);

  final GatewayExecutor _executor;

  Future<ThreadsPostPage?> threadsPosts({String? authorId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query threadsPosts(\$author_id: String, \$cursor: String, \$limit: Int) { threadsPosts(author_id: \$author_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags original_post { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at tags } } next_cursor has_more }'} }', variables: {
      'author_id': ?authorId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsPosts'] == null ? null : ThreadsPostPage.fromJson((data['threadsPosts'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsPost?> threadsPost({required String id, String? selectionSet}) async {
    final data = await _executor.public('query threadsPost(\$id: String!) { threadsPost(id: \$id) ${selectionSet ?? '{ id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags original_post { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags } }'} }', variables: {
      'id': id,
    }
    );
    return data['threadsPost'] == null ? null : ThreadsPost.fromJson((data['threadsPost'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsCommentPage?> threadsComments({required String id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query threadsComments(\$id: String!, \$cursor: String, \$limit: Int) { threadsComments(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id post_id user_id parent_comment_id content media_urls created_at updated_at author { id username profile_image_url } } next_cursor has_more }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsComments'] == null ? null : ThreadsCommentPage.fromJson((data['threadsComments'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsAuthorPage?> threadsFollowers({required String id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query threadsFollowers(\$id: String!, \$cursor: String, \$limit: Int) { threadsFollowers(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id username profile_image_url } next_cursor has_more }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsFollowers'] == null ? null : ThreadsAuthorPage.fromJson((data['threadsFollowers'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsAuthorPage?> threadsFollowing({required String id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query threadsFollowing(\$id: String!, \$cursor: String, \$limit: Int) { threadsFollowing(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id username profile_image_url } next_cursor has_more }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsFollowing'] == null ? null : ThreadsAuthorPage.fromJson((data['threadsFollowing'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsPostPage?> threadsHomeFeed({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query threadsHomeFeed(\$cursor: String, \$limit: Int) { threadsHomeFeed(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags original_post { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at tags } } next_cursor has_more }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsHomeFeed'] == null ? null : ThreadsPostPage.fromJson((data['threadsHomeFeed'] as Map).cast<String, dynamic>());
  }

  Future<ThreadsPostPage?> threadsBookmarks({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query threadsBookmarks(\$cursor: String, \$limit: Int) { threadsBookmarks(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags original_post { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at tags } } next_cursor has_more }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['threadsBookmarks'] == null ? null : ThreadsPostPage.fromJson((data['threadsBookmarks'] as Map).cast<String, dynamic>());
  }

  Future<List<ThreadsTagCount>?> threadsTrendingTags({int? limit, String? selectionSet}) async {
    final data = await _executor.public('query threadsTrendingTags(\$limit: Int) { threadsTrendingTags(limit: \$limit) ${selectionSet ?? '{ name count }'} }', variables: {
      'limit': ?limit,
    }
    );
    return (data['threadsTrendingTags'] as List?)?.map((e) => (e == null ? null : ThreadsTagCount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ThreadsPost?> threadsCreatePost({required String content, List<String>? mediaUrls, List<String>? tags, String? selectionSet}) async {
    final data = await _executor.mutation('mutation threadsCreatePost(\$content: String!, \$media_urls: [String!], \$tags: [String!]) { threadsCreatePost(content: \$content, media_urls: \$media_urls, tags: \$tags) ${selectionSet ?? '{ id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags original_post { id user_id content media_urls like_count comment_count repost_of_id repost_count created_at updated_at author { id username profile_image_url } tags } }'} }', variables: {
      'content': content,
      'media_urls': ?mediaUrls,
      'tags': ?tags,
    }
    );
    return data['threadsCreatePost'] == null ? null : ThreadsPost.fromJson((data['threadsCreatePost'] as Map).cast<String, dynamic>());
  }

  Future<bool?> threadsDeletePost({required String id}) async {
    final data = await _executor.mutation('mutation threadsDeletePost(\$id: String!) { threadsDeletePost(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsDeletePost'] as bool?;
  }

  Future<Object?> threadsLikePost({required String id}) async {
    final data = await _executor.mutation('mutation threadsLikePost(\$id: String!) { threadsLikePost(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsLikePost'];
  }

  Future<Object?> threadsLikeComment({required String id}) async {
    final data = await _executor.mutation('mutation threadsLikeComment(\$id: String!) { threadsLikeComment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsLikeComment'];
  }

  Future<Object?> threadsUnlikePost({required String id}) async {
    final data = await _executor.mutation('mutation threadsUnlikePost(\$id: String!) { threadsUnlikePost(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsUnlikePost'];
  }

  Future<Object?> threadsUnlikeComment({required String id}) async {
    final data = await _executor.mutation('mutation threadsUnlikeComment(\$id: String!) { threadsUnlikeComment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsUnlikeComment'];
  }

  Future<ThreadsComment?> threadsCreateComment({required String id, required String content, List<String>? mediaUrls, String? parentCommentId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation threadsCreateComment(\$id: String!, \$content: String!, \$media_urls: [String!], \$parent_comment_id: String) { threadsCreateComment(id: \$id, content: \$content, media_urls: \$media_urls, parent_comment_id: \$parent_comment_id) ${selectionSet ?? '{ id post_id user_id parent_comment_id content media_urls created_at updated_at author { id username profile_image_url } }'} }', variables: {
      'id': id,
      'content': content,
      'media_urls': ?mediaUrls,
      'parent_comment_id': ?parentCommentId,
    }
    );
    return data['threadsCreateComment'] == null ? null : ThreadsComment.fromJson((data['threadsCreateComment'] as Map).cast<String, dynamic>());
  }

  Future<Object?> threadsRepost({required String id}) async {
    final data = await _executor.mutation('mutation threadsRepost(\$id: String!) { threadsRepost(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsRepost'];
  }

  Future<Object?> threadsUnrepost({required String id}) async {
    final data = await _executor.mutation('mutation threadsUnrepost(\$id: String!) { threadsUnrepost(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsUnrepost'];
  }

  Future<Object?> threadsBookmark({required String id}) async {
    final data = await _executor.mutation('mutation threadsBookmark(\$id: String!) { threadsBookmark(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsBookmark'];
  }

  Future<Object?> threadsUnbookmark({required String id}) async {
    final data = await _executor.mutation('mutation threadsUnbookmark(\$id: String!) { threadsUnbookmark(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsUnbookmark'];
  }

  Future<bool?> threadsDeleteComment({required String id}) async {
    final data = await _executor.mutation('mutation threadsDeleteComment(\$id: String!) { threadsDeleteComment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsDeleteComment'] as bool?;
  }

  Future<Object?> threadsFollowUser({required String id}) async {
    final data = await _executor.mutation('mutation threadsFollowUser(\$id: String!) { threadsFollowUser(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsFollowUser'];
  }

  Future<Object?> threadsUnfollowUser({required String id}) async {
    final data = await _executor.mutation('mutation threadsUnfollowUser(\$id: String!) { threadsUnfollowUser(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['threadsUnfollowUser'];
  }

}
