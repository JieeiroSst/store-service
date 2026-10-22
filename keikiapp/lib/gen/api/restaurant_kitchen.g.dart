// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RestaurantKitchenApi {
  const RestaurantKitchenApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<RKitchenCategory>?> rKitchenCategories({String? selectionSet}) async {
    final data = await _executor.public('query rKitchenCategories { rKitchenCategories ${selectionSet ?? '{ id name }'} }'
    );
    return (data['rKitchenCategories'] as List?)?.map((e) => (e == null ? null : RKitchenCategory.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RKitchenFoodPage?> rKitchenFoods({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.public('query rKitchenFoods(\$Limit: Int, \$Page: Int, \$Sort: String) { rKitchenFoods(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id name category_id price status category { id name } } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rKitchenFoods'] == null ? null : RKitchenFoodPage.fromJson((data['rKitchenFoods'] as Map).cast<String, dynamic>());
  }

  Future<List<RKitchenFood>?> rKitchenFoodsByIDs({required String ids, String? selectionSet}) async {
    final data = await _executor.public('query rKitchenFoodsByIDs(\$ids: String!) { rKitchenFoodsByIDs(ids: \$ids) ${selectionSet ?? '{ id name category_id price status category { id name } }'} }', variables: {
      'ids': ids,
    }
    );
    return (data['rKitchenFoodsByIDs'] as List?)?.map((e) => (e == null ? null : RKitchenFood.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<RKitchenKitchenPage?> rKitchenKitchens({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.public('query rKitchenKitchens(\$Limit: Int, \$Page: Int, \$Sort: String) { rKitchenKitchens(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id name status foods { id name category_id price status } } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rKitchenKitchens'] == null ? null : RKitchenKitchenPage.fromJson((data['rKitchenKitchens'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rKitchenCreateCategory({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rKitchenCreateCategory(\$input: Map!) { rKitchenCreateCategory(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rKitchenCreateCategory'];
  }

  Future<Object?> rKitchenCreateFood({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rKitchenCreateFood(\$input: Map!) { rKitchenCreateFood(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rKitchenCreateFood'];
  }

  Future<Object?> rKitchenUpdateStatus({required String id, required String status}) async {
    final data = await _executor.mutation('mutation rKitchenUpdateStatus(\$id: String!, \$status: String!) { rKitchenUpdateStatus(id: \$id, status: \$status) }', variables: {
      'id': id,
      'status': status,
    }
    );
    return data['rKitchenUpdateStatus'];
  }

}
