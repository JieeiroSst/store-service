// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class MovieApi {
  const MovieApi(this._executor);

  final GatewayExecutor _executor;

  Future<MoviePage?> movieForUser({required String userId, int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query movieForUser(\$user_id: String!, \$page: Int, \$page_size: Int, \$snapshot: String) { movieForUser(user_id: \$user_id, page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'user_id': userId,
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieForUser'] == null ? null : MoviePage.fromJson((data['movieForUser'] as Map).cast<String, dynamic>());
  }

  Future<MovieHome?> movieHome({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query movieHome(\$user_id: String!) { movieHome(user_id: \$user_id) ${selectionSet ?? '{ sections { id title items { score reason progress at } } }'} }', variables: {
      'user_id': userId,
    }
    );
    return data['movieHome'] == null ? null : MovieHome.fromJson((data['movieHome'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> movieContinueWatching({required String userId, int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query movieContinueWatching(\$user_id: String!, \$page: Int, \$page_size: Int, \$snapshot: String) { movieContinueWatching(user_id: \$user_id, page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'user_id': userId,
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieContinueWatching'] == null ? null : MoviePage.fromJson((data['movieContinueWatching'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> movieHistory({required String userId, int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query movieHistory(\$user_id: String!, \$page: Int, \$page_size: Int, \$snapshot: String) { movieHistory(user_id: \$user_id, page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'user_id': userId,
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieHistory'] == null ? null : MoviePage.fromJson((data['movieHistory'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> movieNewReleases({int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.public('query movieNewReleases(\$page: Int, \$page_size: Int, \$snapshot: String) { movieNewReleases(page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieNewReleases'] == null ? null : MoviePage.fromJson((data['movieNewReleases'] as Map).cast<String, dynamic>());
  }

  Future<MovieVideo?> movieVideo({required String id, String? selectionSet}) async {
    final data = await _executor.public('query movieVideo(\$id: String!) { movieVideo(id: \$id) ${selectionSet ?? '{ id title description tags status duration views created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['movieVideo'] == null ? null : MovieVideo.fromJson((data['movieVideo'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> movieSimilar({required String id, int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.public('query movieSimilar(\$id: String!, \$page: Int, \$page_size: Int, \$snapshot: String) { movieSimilar(id: \$id, page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'id': id,
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieSimilar'] == null ? null : MoviePage.fromJson((data['movieSimilar'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> movieTrending({int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.public('query movieTrending(\$page: Int, \$page_size: Int, \$snapshot: String) { movieTrending(page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['movieTrending'] == null ? null : MoviePage.fromJson((data['movieTrending'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> myMovieForUser({int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query myMovieForUser(\$page: Int, \$page_size: Int, \$snapshot: String) { myMovieForUser(page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['myMovieForUser'] == null ? null : MoviePage.fromJson((data['myMovieForUser'] as Map).cast<String, dynamic>());
  }

  Future<MovieHome?> myMovieHome({String? selectionSet}) async {
    final data = await _executor.query('query myMovieHome { myMovieHome ${selectionSet ?? '{ sections { id title items { score reason progress at } } }'} }'
    );
    return data['myMovieHome'] == null ? null : MovieHome.fromJson((data['myMovieHome'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> myMovieContinueWatching({int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query myMovieContinueWatching(\$page: Int, \$page_size: Int, \$snapshot: String) { myMovieContinueWatching(page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['myMovieContinueWatching'] == null ? null : MoviePage.fromJson((data['myMovieContinueWatching'] as Map).cast<String, dynamic>());
  }

  Future<MoviePage?> myMovieHistory({int? page, int? pageSize, String? snapshot, String? selectionSet}) async {
    final data = await _executor.query('query myMovieHistory(\$page: Int, \$page_size: Int, \$snapshot: String) { myMovieHistory(page: \$page, page_size: \$page_size, snapshot: \$snapshot) ${selectionSet ?? '{ items { video { id title description tags status duration views created_at } score reason because { id title description tags status duration views created_at } progress at } total page page_size snapshot }'} }', variables: {
      'page': ?page,
      'page_size': ?pageSize,
      'snapshot': ?snapshot,
    }
    );
    return data['myMovieHistory'] == null ? null : MoviePage.fromJson((data['myMovieHistory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> movieRecordEvent({required String videoId, required String type_, double? value}) async {
    final data = await _executor.mutation('mutation movieRecordEvent(\$video_id: String!, \$type: String!, \$value: Float) { movieRecordEvent(video_id: \$video_id, type: \$type, value: \$value) }', variables: {
      'video_id': videoId,
      'type': type_,
      'value': ?value,
    }
    );
    return data['movieRecordEvent'] as bool?;
  }

  Future<bool?> movieRemoveFromMyHistory({required String videoId}) async {
    final data = await _executor.mutation('mutation movieRemoveFromMyHistory(\$video_id: String!) { movieRemoveFromMyHistory(video_id: \$video_id) }', variables: {
      'video_id': videoId,
    }
    );
    return data['movieRemoveFromMyHistory'] as bool?;
  }

  Future<bool?> movieUpsertVideo({required String id, required MovieVideoInput input}) async {
    final data = await _executor.mutation('mutation movieUpsertVideo(\$id: String!, \$input: MovieVideoInput!) { movieUpsertVideo(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['movieUpsertVideo'] as bool?;
  }

  Future<bool?> movieDeleteVideo({required String id}) async {
    final data = await _executor.mutation('mutation movieDeleteVideo(\$id: String!) { movieDeleteVideo(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['movieDeleteVideo'] as bool?;
  }

}
