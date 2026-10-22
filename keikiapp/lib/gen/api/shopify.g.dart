// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ShopifyApi {
  const ShopifyApi(this._executor);

  final GatewayExecutor _executor;

  Future<ShopifyProductList?> shopifyProducts({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.public('query shopifyProducts(\$limit: Int, \$offset: Int) { shopifyProducts(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id shopify_id title handle status vendor product_type tags variants { id product_id shopify_id title sku price inventory_quantity created_at updated_at } shopify_created_at shopify_updated_at created_at updated_at } total }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['shopifyProducts'] == null ? null : ShopifyProductList.fromJson((data['shopifyProducts'] as Map).cast<String, dynamic>());
  }

  Future<ShopifyProduct?> shopifyProduct({required int id, String? selectionSet}) async {
    final data = await _executor.public('query shopifyProduct(\$id: Int!) { shopifyProduct(id: \$id) ${selectionSet ?? '{ id shopify_id title handle status vendor product_type tags variants { id product_id shopify_id title sku price inventory_quantity created_at updated_at } shopify_created_at shopify_updated_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['shopifyProduct'] == null ? null : ShopifyProduct.fromJson((data['shopifyProduct'] as Map).cast<String, dynamic>());
  }

  Future<ShopifyOrderList?> shopifyOrders({int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query shopifyOrders(\$limit: Int, \$offset: Int) { shopifyOrders(limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id shopify_id name email financial_status fulfillment_status currency total_price cancelled_at processed_at line_items { id order_id shopify_id variant_shopify_id title sku quantity unit_price created_at updated_at } shopify_created_at shopify_updated_at created_at updated_at } total }'} }', variables: {
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['shopifyOrders'] == null ? null : ShopifyOrderList.fromJson((data['shopifyOrders'] as Map).cast<String, dynamic>());
  }

  Future<ShopifyOrder?> shopifyOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query shopifyOrder(\$id: Int!) { shopifyOrder(id: \$id) ${selectionSet ?? '{ id shopify_id name email financial_status fulfillment_status currency total_price cancelled_at processed_at line_items { id order_id shopify_id variant_shopify_id title sku quantity unit_price created_at updated_at } shopify_created_at shopify_updated_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['shopifyOrder'] == null ? null : ShopifyOrder.fromJson((data['shopifyOrder'] as Map).cast<String, dynamic>());
  }

  Future<ShopifyProduct?> shopifyCreateProduct({required String title, String? descriptionHtml, String? vendor, String? productType, List<String>? tags, String? status, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shopifyCreateProduct(\$title: String!, \$description_html: String, \$vendor: String, \$product_type: String, \$tags: [String!], \$status: String) { shopifyCreateProduct(title: \$title, description_html: \$description_html, vendor: \$vendor, product_type: \$product_type, tags: \$tags, status: \$status) ${selectionSet ?? '{ id shopify_id title handle status vendor product_type tags variants { id product_id shopify_id title sku price inventory_quantity created_at updated_at } shopify_created_at shopify_updated_at created_at updated_at }'} }', variables: {
      'title': title,
      'description_html': ?descriptionHtml,
      'vendor': ?vendor,
      'product_type': ?productType,
      'tags': ?tags,
      'status': ?status,
    }
    );
    return data['shopifyCreateProduct'] == null ? null : ShopifyProduct.fromJson((data['shopifyCreateProduct'] as Map).cast<String, dynamic>());
  }

  Future<Object?> shopifySyncProducts() async {
    final data = await _executor.mutation('mutation shopifySyncProducts { shopifySyncProducts }'
    );
    return data['shopifySyncProducts'];
  }

  Future<Object?> shopifySyncOrders() async {
    final data = await _executor.mutation('mutation shopifySyncOrders { shopifySyncOrders }'
    );
    return data['shopifySyncOrders'];
  }

}
