// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BookStoreApi {
  const BookStoreApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BookStoreBook>?> bookStoreBooks({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreBooks { bookStoreBooks ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }'
    );
    return (data['bookStoreBooks'] as List?)?.map((e) => (e == null ? null : BookStoreBook.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookStoreBook?> bookStoreBook({required int id, String? selectionSet}) async {
    final data = await _executor.public('query bookStoreBook(\$id: Int!) { bookStoreBook(id: \$id) ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }', variables: {
      'id': id,
    }
    );
    return data['bookStoreBook'] == null ? null : BookStoreBook.fromJson((data['bookStoreBook'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreBook?> bookStoreHighestPriceBook({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreHighestPriceBook { bookStoreHighestPriceBook ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }'
    );
    return data['bookStoreHighestPriceBook'] == null ? null : BookStoreBook.fromJson((data['bookStoreHighestPriceBook'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreBook?> bookStoreMostPurchasedBook({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreMostPurchasedBook { bookStoreMostPurchasedBook ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }'
    );
    return data['bookStoreMostPurchasedBook'] == null ? null : BookStoreBook.fromJson((data['bookStoreMostPurchasedBook'] as Map).cast<String, dynamic>());
  }

  Future<List<BookStoreAuthor>?> bookStoreAuthors({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreAuthors { bookStoreAuthors ${selectionSet ?? '{ id name bio }'} }'
    );
    return (data['bookStoreAuthors'] as List?)?.map((e) => (e == null ? null : BookStoreAuthor.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookStoreAuthor?> bookStoreAuthor({required int id, String? selectionSet}) async {
    final data = await _executor.public('query bookStoreAuthor(\$id: Int!) { bookStoreAuthor(id: \$id) ${selectionSet ?? '{ id name bio }'} }', variables: {
      'id': id,
    }
    );
    return data['bookStoreAuthor'] == null ? null : BookStoreAuthor.fromJson((data['bookStoreAuthor'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreAuthorStat?> bookStoreMostReadAuthor({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreMostReadAuthor { bookStoreMostReadAuthor ${selectionSet ?? '{ author_id author_name total_reads total_purchases }'} }'
    );
    return data['bookStoreMostReadAuthor'] == null ? null : BookStoreAuthorStat.fromJson((data['bookStoreMostReadAuthor'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreAuthorStat?> bookStoreMostPurchasedAuthor({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreMostPurchasedAuthor { bookStoreMostPurchasedAuthor ${selectionSet ?? '{ author_id author_name total_reads total_purchases }'} }'
    );
    return data['bookStoreMostPurchasedAuthor'] == null ? null : BookStoreAuthorStat.fromJson((data['bookStoreMostPurchasedAuthor'] as Map).cast<String, dynamic>());
  }

  Future<List<BookStorePublisher>?> bookStorePublishers({String? selectionSet}) async {
    final data = await _executor.public('query bookStorePublishers { bookStorePublishers ${selectionSet ?? '{ id name address }'} }'
    );
    return (data['bookStorePublishers'] as List?)?.map((e) => (e == null ? null : BookStorePublisher.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookStorePublisher?> bookStorePublisher({required int id, String? selectionSet}) async {
    final data = await _executor.public('query bookStorePublisher(\$id: Int!) { bookStorePublisher(id: \$id) ${selectionSet ?? '{ id name address }'} }', variables: {
      'id': id,
    }
    );
    return data['bookStorePublisher'] == null ? null : BookStorePublisher.fromJson((data['bookStorePublisher'] as Map).cast<String, dynamic>());
  }

  Future<List<BookStoreCategory>?> bookStoreCategories({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreCategories { bookStoreCategories ${selectionSet ?? '{ id name }'} }'
    );
    return (data['bookStoreCategories'] as List?)?.map((e) => (e == null ? null : BookStoreCategory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookStoreCategory?> bookStoreCategory({required int id, String? selectionSet}) async {
    final data = await _executor.public('query bookStoreCategory(\$id: Int!) { bookStoreCategory(id: \$id) ${selectionSet ?? '{ id name }'} }', variables: {
      'id': id,
    }
    );
    return data['bookStoreCategory'] == null ? null : BookStoreCategory.fromJson((data['bookStoreCategory'] as Map).cast<String, dynamic>());
  }

  Future<List<BookStoreCategoryPriceStat>?> bookStoreCategoryPriceStats({String? selectionSet}) async {
    final data = await _executor.public('query bookStoreCategoryPriceStats { bookStoreCategoryPriceStats ${selectionSet ?? '{ category_id category_name min_price max_price avg_price book_count }'} }'
    );
    return (data['bookStoreCategoryPriceStats'] as List?)?.map((e) => (e == null ? null : BookStoreCategoryPriceStat.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BookStoreBook?> bookStoreCreateBook({required BookStoreBookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreCreateBook(\$input: BookStoreBookInput!) { bookStoreCreateBook(input: \$input) ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookStoreCreateBook'] == null ? null : BookStoreBook.fromJson((data['bookStoreCreateBook'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreBook?> bookStoreUpdateBook({required int id, required BookStoreBookInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreUpdateBook(\$id: Int!, \$input: BookStoreBookInput!) { bookStoreUpdateBook(id: \$id, input: \$input) ${selectionSet ?? '{ id name author_id author { id name bio } publisher_id publisher { id name address } category_id category { id name } price quantity unit_of_measurement purchase_count read_count }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bookStoreUpdateBook'] == null ? null : BookStoreBook.fromJson((data['bookStoreUpdateBook'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookStoreDeleteBook({required int id}) async {
    final data = await _executor.mutation('mutation bookStoreDeleteBook(\$id: Int!) { bookStoreDeleteBook(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookStoreDeleteBook'] as bool?;
  }

  Future<bool?> bookStorePurchaseBook({required int id, required int quantity}) async {
    final data = await _executor.mutation('mutation bookStorePurchaseBook(\$id: Int!, \$quantity: Int!) { bookStorePurchaseBook(id: \$id, quantity: \$quantity) }', variables: {
      'id': id,
      'quantity': quantity,
    }
    );
    return data['bookStorePurchaseBook'] as bool?;
  }

  Future<bool?> bookStoreReadBook({required int id}) async {
    final data = await _executor.mutation('mutation bookStoreReadBook(\$id: Int!) { bookStoreReadBook(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookStoreReadBook'] as bool?;
  }

  Future<BookStoreAuthor?> bookStoreCreateAuthor({required BookStoreAuthorInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreCreateAuthor(\$input: BookStoreAuthorInput!) { bookStoreCreateAuthor(input: \$input) ${selectionSet ?? '{ id name bio }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookStoreCreateAuthor'] == null ? null : BookStoreAuthor.fromJson((data['bookStoreCreateAuthor'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreAuthor?> bookStoreUpdateAuthor({required int id, required BookStoreAuthorInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreUpdateAuthor(\$id: Int!, \$input: BookStoreAuthorInput!) { bookStoreUpdateAuthor(id: \$id, input: \$input) ${selectionSet ?? '{ id name bio }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bookStoreUpdateAuthor'] == null ? null : BookStoreAuthor.fromJson((data['bookStoreUpdateAuthor'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookStoreDeleteAuthor({required int id}) async {
    final data = await _executor.mutation('mutation bookStoreDeleteAuthor(\$id: Int!) { bookStoreDeleteAuthor(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookStoreDeleteAuthor'] as bool?;
  }

  Future<BookStorePublisher?> bookStoreCreatePublisher({required BookStorePublisherInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreCreatePublisher(\$input: BookStorePublisherInput!) { bookStoreCreatePublisher(input: \$input) ${selectionSet ?? '{ id name address }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookStoreCreatePublisher'] == null ? null : BookStorePublisher.fromJson((data['bookStoreCreatePublisher'] as Map).cast<String, dynamic>());
  }

  Future<BookStorePublisher?> bookStoreUpdatePublisher({required int id, required BookStorePublisherInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreUpdatePublisher(\$id: Int!, \$input: BookStorePublisherInput!) { bookStoreUpdatePublisher(id: \$id, input: \$input) ${selectionSet ?? '{ id name address }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bookStoreUpdatePublisher'] == null ? null : BookStorePublisher.fromJson((data['bookStoreUpdatePublisher'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookStoreDeletePublisher({required int id}) async {
    final data = await _executor.mutation('mutation bookStoreDeletePublisher(\$id: Int!) { bookStoreDeletePublisher(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookStoreDeletePublisher'] as bool?;
  }

  Future<BookStoreCategory?> bookStoreCreateCategory({required BookStoreCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreCreateCategory(\$input: BookStoreCategoryInput!) { bookStoreCreateCategory(input: \$input) ${selectionSet ?? '{ id name }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bookStoreCreateCategory'] == null ? null : BookStoreCategory.fromJson((data['bookStoreCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<BookStoreCategory?> bookStoreUpdateCategory({required int id, required BookStoreCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bookStoreUpdateCategory(\$id: Int!, \$input: BookStoreCategoryInput!) { bookStoreUpdateCategory(id: \$id, input: \$input) ${selectionSet ?? '{ id name }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bookStoreUpdateCategory'] == null ? null : BookStoreCategory.fromJson((data['bookStoreUpdateCategory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bookStoreDeleteCategory({required int id}) async {
    final data = await _executor.mutation('mutation bookStoreDeleteCategory(\$id: Int!) { bookStoreDeleteCategory(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bookStoreDeleteCategory'] as bool?;
  }

}
