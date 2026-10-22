// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class CataloguesApi {
  const CataloguesApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<CatalogueCategory>?> catalogueCategories({String? selectionSet}) async {
    final data = await _executor.public('query catalogueCategories { catalogueCategories ${selectionSet ?? '{ id parent_id name slug description image is_public ancestors_are_public meta_title meta_description }'} }'
    );
    return (data['catalogueCategories'] as List?)?.map((e) => (e == null ? null : CatalogueCategory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CatalogueCategory?> catalogueCategory({required int id, String? selectionSet}) async {
    final data = await _executor.public('query catalogueCategory(\$id: Int!) { catalogueCategory(id: \$id) ${selectionSet ?? '{ id parent_id name slug description image is_public ancestors_are_public meta_title meta_description }'} }', variables: {
      'id': id,
    }
    );
    return data['catalogueCategory'] == null ? null : CatalogueCategory.fromJson((data['catalogueCategory'] as Map).cast<String, dynamic>());
  }

  Future<List<CatalogueOption>?> catalogueOptions({String? selectionSet}) async {
    final data = await _executor.public('query catalogueOptions { catalogueOptions ${selectionSet ?? '{ id name code type required }'} }'
    );
    return (data['catalogueOptions'] as List?)?.map((e) => (e == null ? null : CatalogueOption.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CatalogueOption?> catalogueOption({required int id, String? selectionSet}) async {
    final data = await _executor.public('query catalogueOption(\$id: Int!) { catalogueOption(id: \$id) ${selectionSet ?? '{ id name code type required }'} }', variables: {
      'id': id,
    }
    );
    return data['catalogueOption'] == null ? null : CatalogueOption.fromJson((data['catalogueOption'] as Map).cast<String, dynamic>());
  }

  Future<List<CatalogueProductClass>?> catalogueProductClasses({String? selectionSet}) async {
    final data = await _executor.public('query catalogueProductClasses { catalogueProductClasses ${selectionSet ?? '{ id name slug requires_shipping track_stock option_ids }'} }'
    );
    return (data['catalogueProductClasses'] as List?)?.map((e) => (e == null ? null : CatalogueProductClass.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CatalogueProductClass?> catalogueProductClass({required int id, String? selectionSet}) async {
    final data = await _executor.public('query catalogueProductClass(\$id: Int!) { catalogueProductClass(id: \$id) ${selectionSet ?? '{ id name slug requires_shipping track_stock option_ids }'} }', variables: {
      'id': id,
    }
    );
    return data['catalogueProductClass'] == null ? null : CatalogueProductClass.fromJson((data['catalogueProductClass'] as Map).cast<String, dynamic>());
  }

  Future<List<CatalogueProduct>?> catalogueProducts({String? q, int? categoryId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.public('query catalogueProducts(\$q: String, \$category_id: Int, \$limit: Int, \$offset: Int) { catalogueProducts(q: \$q, category_id: \$category_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'q': ?q,
      'category_id': ?categoryId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['catalogueProducts'] as List?)?.map((e) => (e == null ? null : CatalogueProduct.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<CatalogueProduct>?> catalogueAdminProducts({String? q, int? categoryId, bool? public, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query catalogueAdminProducts(\$q: String, \$category_id: Int, \$public: Boolean, \$limit: Int, \$offset: Int) { catalogueAdminProducts(q: \$q, category_id: \$category_id, public: \$public, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'q': ?q,
      'category_id': ?categoryId,
      'public': ?public,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['catalogueAdminProducts'] as List?)?.map((e) => (e == null ? null : CatalogueProduct.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CatalogueProduct?> catalogueProduct({required int id, String? selectionSet}) async {
    final data = await _executor.public('query catalogueProduct(\$id: Int!) { catalogueProduct(id: \$id) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'id': id,
    }
    );
    return data['catalogueProduct'] == null ? null : CatalogueProduct.fromJson((data['catalogueProduct'] as Map).cast<String, dynamic>());
  }

  Future<List<CatalogueStockRecord>?> catalogueStockRecords({required int productId, String? selectionSet}) async {
    final data = await _executor.public('query catalogueStockRecords(\$product_id: Int!) { catalogueStockRecords(product_id: \$product_id) ${selectionSet ?? '{ id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated }'} }', variables: {
      'product_id': productId,
    }
    );
    return (data['catalogueStockRecords'] as List?)?.map((e) => (e == null ? null : CatalogueStockRecord.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<CatalogueStockRecord?> catalogueStockRecord({required int id, String? selectionSet}) async {
    final data = await _executor.public('query catalogueStockRecord(\$id: Int!) { catalogueStockRecord(id: \$id) ${selectionSet ?? '{ id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated }'} }', variables: {
      'id': id,
    }
    );
    return data['catalogueStockRecord'] == null ? null : CatalogueStockRecord.fromJson((data['catalogueStockRecord'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueCategory?> catalogueCreateCategory({required CatalogueCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueCreateCategory(\$input: CatalogueCategoryInput!) { catalogueCreateCategory(input: \$input) ${selectionSet ?? '{ id parent_id name slug description image is_public ancestors_are_public meta_title meta_description }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['catalogueCreateCategory'] == null ? null : CatalogueCategory.fromJson((data['catalogueCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueCategory?> catalogueUpdateCategory({required int id, required CatalogueCategoryInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueUpdateCategory(\$id: Int!, \$input: CatalogueCategoryInput!) { catalogueUpdateCategory(id: \$id, input: \$input) ${selectionSet ?? '{ id parent_id name slug description image is_public ancestors_are_public meta_title meta_description }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueUpdateCategory'] == null ? null : CatalogueCategory.fromJson((data['catalogueUpdateCategory'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteCategory({required int id}) async {
    final data = await _executor.mutation('mutation catalogueDeleteCategory(\$id: Int!) { catalogueDeleteCategory(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['catalogueDeleteCategory'] as bool?;
  }

  Future<CatalogueOption?> catalogueCreateOption({required CatalogueOptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueCreateOption(\$input: CatalogueOptionInput!) { catalogueCreateOption(input: \$input) ${selectionSet ?? '{ id name code type required }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['catalogueCreateOption'] == null ? null : CatalogueOption.fromJson((data['catalogueCreateOption'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueOption?> catalogueUpdateOption({required int id, required CatalogueOptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueUpdateOption(\$id: Int!, \$input: CatalogueOptionInput!) { catalogueUpdateOption(id: \$id, input: \$input) ${selectionSet ?? '{ id name code type required }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueUpdateOption'] == null ? null : CatalogueOption.fromJson((data['catalogueUpdateOption'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteOption({required int id}) async {
    final data = await _executor.mutation('mutation catalogueDeleteOption(\$id: Int!) { catalogueDeleteOption(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['catalogueDeleteOption'] as bool?;
  }

  Future<CatalogueProductClass?> catalogueCreateProductClass({required CatalogueProductClassInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueCreateProductClass(\$input: CatalogueProductClassInput!) { catalogueCreateProductClass(input: \$input) ${selectionSet ?? '{ id name slug requires_shipping track_stock option_ids }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['catalogueCreateProductClass'] == null ? null : CatalogueProductClass.fromJson((data['catalogueCreateProductClass'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueProductClass?> catalogueUpdateProductClass({required int id, required CatalogueProductClassInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueUpdateProductClass(\$id: Int!, \$input: CatalogueProductClassInput!) { catalogueUpdateProductClass(id: \$id, input: \$input) ${selectionSet ?? '{ id name slug requires_shipping track_stock option_ids }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueUpdateProductClass'] == null ? null : CatalogueProductClass.fromJson((data['catalogueUpdateProductClass'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteProductClass({required int id}) async {
    final data = await _executor.mutation('mutation catalogueDeleteProductClass(\$id: Int!) { catalogueDeleteProductClass(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['catalogueDeleteProductClass'] as bool?;
  }

  Future<CatalogueProduct?> catalogueCreateProduct({required CatalogueProductInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueCreateProduct(\$input: CatalogueProductInput!) { catalogueCreateProduct(input: \$input) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['catalogueCreateProduct'] == null ? null : CatalogueProduct.fromJson((data['catalogueCreateProduct'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueProduct?> catalogueUpdateProduct({required int id, required CatalogueProductInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueUpdateProduct(\$id: Int!, \$input: CatalogueProductInput!) { catalogueUpdateProduct(id: \$id, input: \$input) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueUpdateProduct'] == null ? null : CatalogueProduct.fromJson((data['catalogueUpdateProduct'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteProduct({required int id}) async {
    final data = await _executor.mutation('mutation catalogueDeleteProduct(\$id: Int!) { catalogueDeleteProduct(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['catalogueDeleteProduct'] as bool?;
  }

  Future<CatalogueProductImage?> catalogueAddProductImage({required int id, required CatalogueImageInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueAddProductImage(\$id: Int!, \$input: CatalogueImageInput!) { catalogueAddProductImage(id: \$id, input: \$input) ${selectionSet ?? '{ id product_id original caption display_order date_created }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueAddProductImage'] == null ? null : CatalogueProductImage.fromJson((data['catalogueAddProductImage'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteProductImage({required int id, required int imageId}) async {
    final data = await _executor.mutation('mutation catalogueDeleteProductImage(\$id: Int!, \$image_id: Int!) { catalogueDeleteProductImage(id: \$id, image_id: \$image_id) }', variables: {
      'id': id,
      'image_id': imageId,
    }
    );
    return data['catalogueDeleteProductImage'] as bool?;
  }

  Future<CatalogueProduct?> catalogueSetRecommendations({required int id, required List<CatalogueRecommendationInput> recommendations, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueSetRecommendations(\$id: Int!, \$recommendations: [CatalogueRecommendationInput!]!) { catalogueSetRecommendations(id: \$id, recommendations: \$recommendations) ${selectionSet ?? '{ id structure upc title slug description rating is_discountable is_public parent_id product_class_id meta_title meta_description date_created date_updated category_ids option_ids images { id product_id original caption display_order date_created } recommendations { product_id ranking } stockrecords { id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated } price price_currency num_available in_stock }'} }', variables: {
      'id': id,
      'recommendations': recommendations.map((e) => e.toJson()).toList(),
    }
    );
    return data['catalogueSetRecommendations'] == null ? null : CatalogueProduct.fromJson((data['catalogueSetRecommendations'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueStockRecord?> catalogueCreateStockRecord({required int productId, required CatalogueStockRecordInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueCreateStockRecord(\$product_id: Int!, \$input: CatalogueStockRecordInput!) { catalogueCreateStockRecord(product_id: \$product_id, input: \$input) ${selectionSet ?? '{ id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated }'} }', variables: {
      'product_id': productId,
      'input': input.toJson(),
    }
    );
    return data['catalogueCreateStockRecord'] == null ? null : CatalogueStockRecord.fromJson((data['catalogueCreateStockRecord'] as Map).cast<String, dynamic>());
  }

  Future<CatalogueStockRecord?> catalogueUpdateStockRecord({required int id, required CatalogueStockRecordInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation catalogueUpdateStockRecord(\$id: Int!, \$input: CatalogueStockRecordInput!) { catalogueUpdateStockRecord(id: \$id, input: \$input) ${selectionSet ?? '{ id product_id partner_sku price_currency price num_in_stock num_allocated num_available date_created date_updated }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['catalogueUpdateStockRecord'] == null ? null : CatalogueStockRecord.fromJson((data['catalogueUpdateStockRecord'] as Map).cast<String, dynamic>());
  }

  Future<bool?> catalogueDeleteStockRecord({required int id}) async {
    final data = await _executor.mutation('mutation catalogueDeleteStockRecord(\$id: Int!) { catalogueDeleteStockRecord(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['catalogueDeleteStockRecord'] as bool?;
  }

}
