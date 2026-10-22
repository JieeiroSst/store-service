// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SerpapiApi {
  const SerpapiApi(this._executor);

  final GatewayExecutor _executor;

  Future<Object?> serpEngines() async {
    final data = await _executor.query('query serpEngines { serpEngines }'
    );
    return data['serpEngines'];
  }

  Future<Object?> serpEngine({required String engine}) async {
    final data = await _executor.query('query serpEngine(\$engine: String!) { serpEngine(engine: \$engine) }', variables: {
      'engine': engine,
    }
    );
    return data['serpEngine'];
  }

  Future<Object?> serpSearch({String? engine, Map<String, dynamic>? params}) async {
    final data = await _executor.query('query serpSearch(\$engine: String, \$params: Map) { serpSearch(engine: \$engine, params: \$params) }', variables: {
      'engine': ?engine,
      'params': ?params,
    }
    );
    return data['serpSearch'];
  }

  Future<Object?> serpSearchBody({required SerpSearchInput input}) async {
    final data = await _executor.query('query serpSearchBody(\$input: SerpSearchInput!) { serpSearchBody(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['serpSearchBody'];
  }

  Future<SerpBatch?> serpSearchBatch({required List<SerpSearchInput> searches, String? selectionSet}) async {
    final data = await _executor.query('query serpSearchBatch(\$searches: [SerpSearchInput!]!) { serpSearchBatch(searches: \$searches) ${selectionSet ?? '{ results { status search_id cached error data } }'} }', variables: {
      'searches': searches.map((e) => e.toJson()).toList(),
    }
    );
    return data['serpSearchBatch'] == null ? null : SerpBatch.fromJson((data['serpSearchBatch'] as Map).cast<String, dynamic>());
  }

  Future<Object?> serpArchive({required String id}) async {
    final data = await _executor.query('query serpArchive(\$id: String!) { serpArchive(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['serpArchive'];
  }

  Future<Object?> serpAccount() async {
    final data = await _executor.query('query serpAccount { serpAccount }'
    );
    return data['serpAccount'];
  }

  Future<Object?> serpLocations({String? q, int? limit}) async {
    final data = await _executor.query('query serpLocations(\$q: String, \$limit: Int) { serpLocations(q: \$q, limit: \$limit) }', variables: {
      'q': ?q,
      'limit': ?limit,
    }
    );
    return data['serpLocations'];
  }

  Future<Object?> serpUploadImage({required GqlUpload image}) async {
    final data = await _executor.mutation('mutation serpUploadImage(\$image: Upload!) { serpUploadImage(image: \$image) }', variables: {
      'image': image,
    }
    );
    return data['serpUploadImage'];
  }

}
