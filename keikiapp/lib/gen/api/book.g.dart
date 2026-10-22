// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BookApi {
  const BookApi(this._executor);

  final GatewayExecutor _executor;

  Future<BookBookPage?> bookBooks({String? cursor, int? limit, String? q, String? author, int? year, String? category, String? selectionSet}) async {
    final data = await _executor.public('query bookBooks(\$cursor: String, \$limit: Int, \$q: String, \$author: String, \$year: Int, \$category: String) { bookBooks(cursor: \$cursor, limit: \$limit, q: \$q, author: \$author, year: \$year, category: \$category) ${selectionSet ?? '{ data { id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount } pagination { limit nextCursor hasMore } }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
      'q': ?q,
      'author': ?author,
      'year': ?year,
      'category': ?category,
    }
    );
    return data['bookBooks'] == null ? null : BookBookPage.fromJson((data['bookBooks'] as Map).cast<String, dynamic>());
  }

  Future<BookBook?> bookBook({required int id, String? selectionSet}) async {
    final data = await _executor.public('query bookBook(\$id: Int64!) { bookBook(id: \$id) ${selectionSet ?? '{ id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount }'} }', variables: {
      'id': id,
    }
    );
    return data['bookBook'] == null ? null : BookBook.fromJson((data['bookBook'] as Map).cast<String, dynamic>());
  }

  Future<BookChapterPage?> bookChapters({required int id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query bookChapters(\$id: Int64!, \$cursor: String, \$limit: Int) { bookChapters(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ data { bookId number title contentType sizeBytes content } pagination { limit nextCursor hasMore } }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['bookChapters'] == null ? null : BookChapterPage.fromJson((data['bookChapters'] as Map).cast<String, dynamic>());
  }

  Future<BookChapter?> bookChapter({required int id, required int number, String? selectionSet}) async {
    final data = await _executor.public('query bookChapter(\$id: Int64!, \$number: Int!) { bookChapter(id: \$id, number: \$number) ${selectionSet ?? '{ bookId number title contentType sizeBytes content }'} }', variables: {
      'id': id,
      'number': number,
    }
    );
    return data['bookChapter'] == null ? null : BookChapter.fromJson((data['bookChapter'] as Map).cast<String, dynamic>());
  }

  Future<List<BookCategory>?> bookCategories({String? selectionSet}) async {
    final data = await _executor.public('query bookCategories { bookCategories ${selectionSet ?? '{ slug name }'} }'
    );
    return (data['bookCategories'] as List?)?.map((e) => (e == null ? null : BookCategory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookCategory?> bookCategory({required String slug, String? selectionSet}) async {
    final data = await _executor.public('query bookCategory(\$slug: String!) { bookCategory(slug: \$slug) ${selectionSet ?? '{ slug name }'} }', variables: {
      'slug': slug,
    }
    );
    return data['bookCategory'] == null ? null : BookCategory.fromJson((data['bookCategory'] as Map).cast<String, dynamic>());
  }

  Future<BookBookPage?> bookCategoryBooks({required String slug, String? cursor, int? limit, String? q, String? selectionSet}) async {
    final data = await _executor.public('query bookCategoryBooks(\$slug: String!, \$cursor: String, \$limit: Int, \$q: String) { bookCategoryBooks(slug: \$slug, cursor: \$cursor, limit: \$limit, q: \$q) ${selectionSet ?? '{ data { id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount } pagination { limit nextCursor hasMore } }'} }', variables: {
      'slug': slug,
      'cursor': ?cursor,
      'limit': ?limit,
      'q': ?q,
    }
    );
    return data['bookCategoryBooks'] == null ? null : BookBookPage.fromJson((data['bookCategoryBooks'] as Map).cast<String, dynamic>());
  }

  Future<BookBook?> bookCreateBook({required BookBookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookCreateBook(\$input: BookBookInput!) { bookCreateBook(input: \$input) ${selectionSet ?? '{ id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookCreateBook'] == null ? null : BookBook.fromJson((data['bookCreateBook'] as Map).cast<String, dynamic>());
  }

  Future<BookBook?> bookUpdateBook({required int id, required BookBookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookUpdateBook(\$id: Int64!, \$input: BookBookInput!) { bookUpdateBook(id: \$id, input: \$input) ${selectionSet ?? '{ id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bookUpdateBook'] == null ? null : BookBook.fromJson((data['bookUpdateBook'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookDeleteBook({required int id}) async {
    final data = await _executor.mutation('mutation bookDeleteBook(\$id: Int64!) { bookDeleteBook(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookDeleteBook'] as bool?;
  }

  Future<BookBook?> bookPutCover({required int id, required GqlUpload file, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookPutCover(\$id: Int64!, \$file: Upload!) { bookPutCover(id: \$id, file: \$file) ${selectionSet ?? '{ id title author isbn description publishedYear categories { slug name } coverUrl createdAt updatedAt chapterCount }'} }', variables: {
      'id': id,
      'file': file,
    }
    );
    return data['bookPutCover'] == null ? null : BookBook.fromJson((data['bookPutCover'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookDeleteCover({required int id}) async {
    final data = await _executor.mutation('mutation bookDeleteCover(\$id: Int64!) { bookDeleteCover(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookDeleteCover'] as bool?;
  }

  Future<BookChapter?> bookPutChapter({required int id, required int number, required BookChapterInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookPutChapter(\$id: Int64!, \$number: Int!, \$input: BookChapterInput!) { bookPutChapter(id: \$id, number: \$number, input: \$input) ${selectionSet ?? '{ bookId number title contentType sizeBytes content }'} }', variables: {
      'id': id,
      'number': number,
      'input': input.toJson(),
    }
    );
    return data['bookPutChapter'] == null ? null : BookChapter.fromJson((data['bookPutChapter'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookDeleteChapter({required int id, required int number}) async {
    final data = await _executor.mutation('mutation bookDeleteChapter(\$id: Int64!, \$number: Int!) { bookDeleteChapter(id: \$id, number: \$number) }', variables: {
      'id': id,
      'number': number,
    }
    );
    return data['bookDeleteChapter'] as bool?;
  }

  Future<BookCategory?> bookCreateCategory({required BookCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookCreateCategory(\$input: BookCategoryInput!) { bookCreateCategory(input: \$input) ${selectionSet ?? '{ slug name }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookCreateCategory'] == null ? null : BookCategory.fromJson((data['bookCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<BookCategory?> bookRenameCategory({required String slug, required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookRenameCategory(\$slug: String!, \$name: String!) { bookRenameCategory(slug: \$slug, name: \$name) ${selectionSet ?? '{ slug name }'} }', variables: {
      'slug': slug,
      'name': name_,
    }
    );
    return data['bookRenameCategory'] == null ? null : BookCategory.fromJson((data['bookRenameCategory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookDeleteCategory({required String slug}) async {
    final data = await _executor.mutation('mutation bookDeleteCategory(\$slug: String!) { bookDeleteCategory(slug: \$slug) }', variables: {
      'slug': slug,
    }
    );
    return data['bookDeleteCategory'] as bool?;
  }

}
