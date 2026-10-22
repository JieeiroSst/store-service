// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SearchApi {
  const SearchApi(this._executor);

  final GatewayExecutor _executor;

  Future<SearchDocumentPage?> search({String? q, List<String>? index_, String? sort, List<String>? facets, String? cursor, int? size, Map<String, dynamic>? filters, String? selectionSet}) async {
    final data = await _executor.query('query search(\$q: String, \$index: [String!], \$sort: String, \$facets: [String!], \$cursor: String, \$size: Int, \$filters: Map) { search(q: \$q, index: \$index, sort: \$sort, facets: \$facets, cursor: \$cursor, size: \$size, filters: \$filters) ${selectionSet ?? '{ items { index id score source highlight } total next_cursor has_more facets suggestion }'} }', variables: {
      'q': ?q,
      'index': ?index_,
      'sort': ?sort,
      'facets': ?facets,
      'cursor': ?cursor,
      'size': ?size,
      'filters': ?filters,
    }
    );
    return data['search'] == null ? null : SearchDocumentPage.fromJson((data['search'] as Map).cast<String, dynamic>());
  }

  Future<Object?> searchAutocomplete({required String q, List<String>? index_, int? size}) async {
    final data = await _executor.query('query searchAutocomplete(\$q: String!, \$index: [String!], \$size: Int) { searchAutocomplete(q: \$q, index: \$index, size: \$size) }', variables: {
      'q': q,
      'index': ?index_,
      'size': ?size,
    }
    );
    return data['searchAutocomplete'];
  }

  Future<SearchDocumentPage?> searchDocuments({String? q, List<String>? index_, String? sort, List<String>? facets, String? cursor, int? size, Map<String, dynamic>? filters, String? selectionSet}) async {
    final data = await _executor.query('query searchDocuments(\$q: String, \$index: [String!], \$sort: String, \$facets: [String!], \$cursor: String, \$size: Int, \$filters: Map) { searchDocuments(q: \$q, index: \$index, sort: \$sort, facets: \$facets, cursor: \$cursor, size: \$size, filters: \$filters) ${selectionSet ?? '{ items { index id score source highlight } total next_cursor has_more facets suggestion }'} }', variables: {
      'q': ?q,
      'index': ?index_,
      'sort': ?sort,
      'facets': ?facets,
      'cursor': ?cursor,
      'size': ?size,
      'filters': ?filters,
    }
    );
    return data['searchDocuments'] == null ? null : SearchDocumentPage.fromJson((data['searchDocuments'] as Map).cast<String, dynamic>());
  }

  Future<SearchDocument?> searchDocument({required String index_, required String id, String? selectionSet}) async {
    final data = await _executor.query('query searchDocument(\$index: String!, \$id: String!) { searchDocument(index: \$index, id: \$id) ${selectionSet ?? '{ index id score source highlight }'} }', variables: {
      'index': index_,
      'id': id,
    }
    );
    return data['searchDocument'] == null ? null : SearchDocument.fromJson((data['searchDocument'] as Map).cast<String, dynamic>());
  }

  Future<Object?> searchSimilar({required String index_, required String id, int? size}) async {
    final data = await _executor.query('query searchSimilar(\$index: String!, \$id: String!, \$size: Int) { searchSimilar(index: \$index, id: \$id, size: \$size) }', variables: {
      'index': index_,
      'id': id,
      'size': ?size,
    }
    );
    return data['searchSimilar'];
  }

  Future<List<SearchIndexInfo>?> searchIndices({String? selectionSet}) async {
    final data = await _executor.query('query searchIndices { searchIndices ${selectionSet ?? '{ name index docs_count managed write_blocked }'} }'
    );
    return (data['searchIndices'] as List?)?.map((e) => (e == null ? null : SearchIndexInfo.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<SearchFieldInfo>?> searchFields({required String index_, String? selectionSet}) async {
    final data = await _executor.query('query searchFields(\$index: String!) { searchFields(index: \$index) ${selectionSet ?? '{ name types searchable aggregatable }'} }', variables: {
      'index': index_,
    }
    );
    return (data['searchFields'] as List?)?.map((e) => (e == null ? null : SearchFieldInfo.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SearchSchemaStatus?> searchSchemaStatus({String? selectionSet}) async {
    final data = await _executor.query('query searchSchemaStatus { searchSchemaStatus ${selectionSet ?? '{ template applied version indices { name index docs_count managed write_blocked } }'} }'
    );
    return data['searchSchemaStatus'] == null ? null : SearchSchemaStatus.fromJson((data['searchSchemaStatus'] as Map).cast<String, dynamic>());
  }

  Future<SearchSchemaStatus?> searchApplySchema({String? selectionSet}) async {
    final data = await _executor.mutation('mutation searchApplySchema { searchApplySchema ${selectionSet ?? '{ template applied version indices { name index docs_count managed write_blocked } }'} }'
    );
    return data['searchApplySchema'] == null ? null : SearchSchemaStatus.fromJson((data['searchApplySchema'] as Map).cast<String, dynamic>());
  }

  Future<List<SearchReindexResult>?> searchReindex({List<String>? index_, bool? unmanaged, String? selectionSet}) async {
    final data = await _executor.mutation('mutation searchReindex(\$index: [String!], \$unmanaged: Boolean) { searchReindex(index: \$index, unmanaged: \$unmanaged) ${selectionSet ?? '{ name from to copied error }'} }', variables: {
      'index': ?index_,
      'unmanaged': ?unmanaged,
    }
    );
    return (data['searchReindex'] as List?)?.map((e) => (e == null ? null : SearchReindexResult.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

}
