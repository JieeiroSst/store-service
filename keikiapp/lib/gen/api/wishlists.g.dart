// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class WishlistsApi {
  const WishlistsApi(this._executor);

  final GatewayExecutor _executor;

  Future<WishlistPage?> wishlists({String? ownerId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query wishlists(\$owner_id: String, \$limit: Int, \$offset: Int) { wishlists(owner_id: \$owner_id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } } total limit offset }'} }', variables: {
      'owner_id': ?ownerId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['wishlists'] == null ? null : WishlistPage.fromJson((data['wishlists'] as Map).cast<String, dynamic>());
  }

  Future<Wishlist?> wishlist({required int id, String? selectionSet}) async {
    final data = await _executor.query('query wishlist(\$id: Int64!) { wishlist(id: \$id) ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }', variables: {
      'id': id,
    }
    );
    return data['wishlist'] == null ? null : Wishlist.fromJson((data['wishlist'] as Map).cast<String, dynamic>());
  }

  Future<Wishlist?> wishlistDefault({String? selectionSet}) async {
    final data = await _executor.query('query wishlistDefault { wishlistDefault ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }'
    );
    return data['wishlistDefault'] == null ? null : Wishlist.fromJson((data['wishlistDefault'] as Map).cast<String, dynamic>());
  }

  Future<WishlistPage?> wishlistsOfUser({required String userId, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query wishlistsOfUser(\$userId: String!, \$limit: Int, \$offset: Int) { wishlistsOfUser(userId: \$userId, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } } total limit offset }'} }', variables: {
      'userId': userId,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['wishlistsOfUser'] == null ? null : WishlistPage.fromJson((data['wishlistsOfUser'] as Map).cast<String, dynamic>());
  }

  Future<Wishlist?> wishlistShared({required String key, String? selectionSet}) async {
    final data = await _executor.public('query wishlistShared(\$key: String!) { wishlistShared(key: \$key) ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }', variables: {
      'key': key,
    }
    );
    return data['wishlistShared'] == null ? null : Wishlist.fromJson((data['wishlistShared'] as Map).cast<String, dynamic>());
  }

  Future<WishlistProductLookup?> wishlistProductLookup({required int productId, String? selectionSet}) async {
    final data = await _executor.query('query wishlistProductLookup(\$productId: Int64!) { wishlistProductLookup(productId: \$productId) ${selectionSet ?? '{ product_id in_wishlist wishlists { wishlist_id wishlist_name line_id quantity } }'} }', variables: {
      'productId': productId,
    }
    );
    return data['wishlistProductLookup'] == null ? null : WishlistProductLookup.fromJson((data['wishlistProductLookup'] as Map).cast<String, dynamic>());
  }

  Future<List<WishlistProductStat>?> wishlistPopularProducts({int? limit, String? selectionSet}) async {
    final data = await _executor.public('query wishlistPopularProducts(\$limit: Int) { wishlistPopularProducts(limit: \$limit) ${selectionSet ?? '{ product_id title wishlist_count total_quantity }'} }', variables: {
      'limit': ?limit,
    }
    );
    return (data['wishlistPopularProducts'] as List?)?.map((e) => (e == null ? null : WishlistProductStat.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<Wishlist?> wishlistCreate({required WishlistCreateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistCreate(\$input: WishlistCreateInput!) { wishlistCreate(input: \$input) ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['wishlistCreate'] == null ? null : Wishlist.fromJson((data['wishlistCreate'] as Map).cast<String, dynamic>());
  }

  Future<Wishlist?> wishlistUpdate({required int id, required WishlistUpdateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistUpdate(\$id: Int64!, \$input: WishlistUpdateInput!) { wishlistUpdate(id: \$id, input: \$input) ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['wishlistUpdate'] == null ? null : Wishlist.fromJson((data['wishlistUpdate'] as Map).cast<String, dynamic>());
  }

  Future<bool?> wishlistDelete({required int id}) async {
    final data = await _executor.mutation('mutation wishlistDelete(\$id: Int64!) { wishlistDelete(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['wishlistDelete'] as bool?;
  }

  Future<WishlistLine?> wishlistAddLine({required int id, required WishlistLineInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistAddLine(\$id: Int64!, \$input: WishlistLineInput!) { wishlistAddLine(id: \$id, input: \$input) ${selectionSet ?? '{ id wishlist_id product_id title quantity }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['wishlistAddLine'] == null ? null : WishlistLine.fromJson((data['wishlistAddLine'] as Map).cast<String, dynamic>());
  }

  Future<WishlistLine?> wishlistUpdateLine({required int id, required int lineId, required int quantity, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistUpdateLine(\$id: Int64!, \$lineId: Int64!, \$quantity: Int!) { wishlistUpdateLine(id: \$id, lineId: \$lineId, quantity: \$quantity) ${selectionSet ?? '{ id wishlist_id product_id title quantity }'} }', variables: {
      'id': id,
      'lineId': lineId,
      'quantity': quantity,
    }
    );
    return data['wishlistUpdateLine'] == null ? null : WishlistLine.fromJson((data['wishlistUpdateLine'] as Map).cast<String, dynamic>());
  }

  Future<bool?> wishlistRemoveLine({required int id, required int lineId}) async {
    final data = await _executor.mutation('mutation wishlistRemoveLine(\$id: Int64!, \$lineId: Int64!) { wishlistRemoveLine(id: \$id, lineId: \$lineId) }', variables: {
      'id': id,
      'lineId': lineId,
    }
    );
    return data['wishlistRemoveLine'] as bool?;
  }

  Future<WishlistClearResult?> wishlistClearLines({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistClearLines(\$id: Int64!) { wishlistClearLines(id: \$id) ${selectionSet ?? '{ removed }'} }', variables: {
      'id': id,
    }
    );
    return data['wishlistClearLines'] == null ? null : WishlistClearResult.fromJson((data['wishlistClearLines'] as Map).cast<String, dynamic>());
  }

  Future<WishlistLine?> wishlistMoveLine({required int id, required int lineId, required int targetWishlistId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistMoveLine(\$id: Int64!, \$lineId: Int64!, \$target_wishlist_id: Int64!) { wishlistMoveLine(id: \$id, lineId: \$lineId, target_wishlist_id: \$target_wishlist_id) ${selectionSet ?? '{ id wishlist_id product_id title quantity }'} }', variables: {
      'id': id,
      'lineId': lineId,
      'target_wishlist_id': targetWishlistId,
    }
    );
    return data['wishlistMoveLine'] == null ? null : WishlistLine.fromJson((data['wishlistMoveLine'] as Map).cast<String, dynamic>());
  }

  Future<Wishlist?> wishlistCopyShared({required String key, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistCopyShared(\$key: String!) { wishlistCopyShared(key: \$key) ${selectionSet ?? '{ id name key visibility date_created owner_id owner { id username name active } lines { id wishlist_id product_id title quantity } }'} }', variables: {
      'key': key,
    }
    );
    return data['wishlistCopyShared'] == null ? null : Wishlist.fromJson((data['wishlistCopyShared'] as Map).cast<String, dynamic>());
  }

  Future<WishlistLine?> wishlistQuickAdd({required WishlistLineInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation wishlistQuickAdd(\$input: WishlistLineInput!) { wishlistQuickAdd(input: \$input) ${selectionSet ?? '{ id wishlist_id product_id title quantity }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['wishlistQuickAdd'] == null ? null : WishlistLine.fromJson((data['wishlistQuickAdd'] as Map).cast<String, dynamic>());
  }

}
