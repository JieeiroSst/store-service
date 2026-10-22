// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PostApi {
  const PostApi(this._executor);

  final GatewayExecutor _executor;

  Future<PostPostPage?> posts({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query posts(\$cursor: String, \$limit: Int) { posts(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id author_id name content description media_id created_at updated_at categories { id name description created_at updated_at } media { id url description created_at updated_at } } next_cursor has_more }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['posts'] == null ? null : PostPostPage.fromJson((data['posts'] as Map).cast<String, dynamic>());
  }

  Future<PostPost?> post({required String id, String? selectionSet}) async {
    final data = await _executor.public('query post(\$id: String!) { post(id: \$id) ${selectionSet ?? '{ id author_id name content description media_id created_at updated_at categories { id name description created_at updated_at } media { id url description created_at updated_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['post'] == null ? null : PostPost.fromJson((data['post'] as Map).cast<String, dynamic>());
  }

  Future<PostCategoryPage?> postCategories({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query postCategories(\$cursor: String, \$limit: Int) { postCategories(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ rows { id name description created_at updated_at posts { id author_id name content description media_id created_at updated_at } } next_cursor has_more }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['postCategories'] == null ? null : PostCategoryPage.fromJson((data['postCategories'] as Map).cast<String, dynamic>());
  }

  Future<PostCategory?> postCategory({required String id, String? selectionSet}) async {
    final data = await _executor.public('query postCategory(\$id: String!) { postCategory(id: \$id) ${selectionSet ?? '{ id name description created_at updated_at posts { id author_id name content description media_id created_at updated_at media { id url description created_at updated_at } } }'} }', variables: {
      'id': id,
    }
    );
    return data['postCategory'] == null ? null : PostCategory.fromJson((data['postCategory'] as Map).cast<String, dynamic>());
  }

  Future<PostPost?> postCreatePost({required String name_, required String content, String? description, String? categoryId, GqlUpload? file, String? selectionSet}) async {
    final data = await _executor.mutation('mutation postCreatePost(\$name: String!, \$content: String!, \$description: String, \$category_id: String, \$file: Upload) { postCreatePost(name: \$name, content: \$content, description: \$description, category_id: \$category_id, file: \$file) ${selectionSet ?? '{ id author_id name content description media_id created_at updated_at categories { id name description created_at updated_at } media { id url description created_at updated_at } }'} }', variables: {
      'name': name_,
      'content': content,
      'description': ?description,
      'category_id': ?categoryId,
      'file': ?file,
    }
    );
    return data['postCreatePost'] == null ? null : PostPost.fromJson((data['postCreatePost'] as Map).cast<String, dynamic>());
  }

  Future<PostPost?> postUpdatePost({required String id, String? name_, String? content, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation postUpdatePost(\$id: String!, \$name: String, \$content: String, \$description: String) { postUpdatePost(id: \$id, name: \$name, content: \$content, description: \$description) ${selectionSet ?? '{ id author_id name content description media_id created_at updated_at categories { id name description created_at updated_at } media { id url description created_at updated_at } }'} }', variables: {
      'id': id,
      'name': ?name_,
      'content': ?content,
      'description': ?description,
    }
    );
    return data['postUpdatePost'] == null ? null : PostPost.fromJson((data['postUpdatePost'] as Map).cast<String, dynamic>());
  }

  Future<PostCategory?> postCreateCategory({required String name_, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation postCreateCategory(\$name: String!, \$description: String) { postCreateCategory(name: \$name, description: \$description) ${selectionSet ?? '{ id name description created_at updated_at posts { id author_id name content description media_id created_at updated_at media { id url description created_at updated_at } } }'} }', variables: {
      'name': name_,
      'description': ?description,
    }
    );
    return data['postCreateCategory'] == null ? null : PostCategory.fromJson((data['postCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<PostCategory?> postUpdateCategory({required String id, String? name_, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation postUpdateCategory(\$id: String!, \$name: String, \$description: String) { postUpdateCategory(id: \$id, name: \$name, description: \$description) ${selectionSet ?? '{ id name description created_at updated_at posts { id author_id name content description media_id created_at updated_at media { id url description created_at updated_at } } }'} }', variables: {
      'id': id,
      'name': ?name_,
      'description': ?description,
    }
    );
    return data['postUpdateCategory'] == null ? null : PostCategory.fromJson((data['postUpdateCategory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> postDeleteCategory({required String id}) async {
    final data = await _executor.mutation('mutation postDeleteCategory(\$id: String!) { postDeleteCategory(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['postDeleteCategory'] as bool?;
  }

}
