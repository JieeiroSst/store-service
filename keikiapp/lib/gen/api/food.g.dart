// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class FoodApi {
  const FoodApi(this._executor);

  final GatewayExecutor _executor;

  Future<FoodFood?> food({required int id, String? selectionSet}) async {
    final data = await _executor.public('query food(\$id: Int!) { food(id: \$id) ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }', variables: {
      'id': id,
    }
    );
    return data['food'] == null ? null : FoodFood.fromJson((data['food'] as Map).cast<String, dynamic>());
  }

  Future<List<FoodFood>?> foods({String? selectionSet}) async {
    final data = await _executor.public('query foods { foods ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }'
    );
    return (data['foods'] as List?)?.map((e) => (e == null ? null : FoodFood.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<FoodFood>?> foodSearchFoods({required String name_, String? selectionSet}) async {
    final data = await _executor.public('query foodSearchFoods(\$name: String!) { foodSearchFoods(name: \$name) ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }', variables: {
      'name': name_,
    }
    );
    return (data['foodSearchFoods'] as List?)?.map((e) => (e == null ? null : FoodFood.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<FoodRestaurant>?> foodRestaurants({String? q, String? selectionSet}) async {
    final data = await _executor.public('query foodRestaurants(\$q: String) { foodRestaurants(q: \$q) ${selectionSet ?? '{ id name description address image_url is_active commission_rate }'} }', variables: {
      'q': ?q,
    }
    );
    return (data['foodRestaurants'] as List?)?.map((e) => (e == null ? null : FoodRestaurant.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<FoodRestaurant?> foodRestaurant({required int id, String? selectionSet}) async {
    final data = await _executor.public('query foodRestaurant(\$id: Int!) { foodRestaurant(id: \$id) ${selectionSet ?? '{ id name description address image_url is_active commission_rate }'} }', variables: {
      'id': id,
    }
    );
    return data['foodRestaurant'] == null ? null : FoodRestaurant.fromJson((data['foodRestaurant'] as Map).cast<String, dynamic>());
  }

  Future<List<FoodFood>?> foodRestaurantMenu({required int id, String? selectionSet}) async {
    final data = await _executor.public('query foodRestaurantMenu(\$id: Int!) { foodRestaurantMenu(id: \$id) ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }', variables: {
      'id': id,
    }
    );
    return (data['foodRestaurantMenu'] as List?)?.map((e) => (e == null ? null : FoodFood.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<FoodFood?> foodCreateFood({required FoodFoodInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodCreateFood(\$input: FoodFoodInput!) { foodCreateFood(input: \$input) ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['foodCreateFood'] == null ? null : FoodFood.fromJson((data['foodCreateFood'] as Map).cast<String, dynamic>());
  }

  Future<FoodFood?> foodUpdateFood({required int id, required FoodFoodInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodUpdateFood(\$id: Int!, \$input: FoodFoodInput!) { foodUpdateFood(id: \$id, input: \$input) ${selectionSet ?? '{ id name description price category restaurant_id is_active }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['foodUpdateFood'] == null ? null : FoodFood.fromJson((data['foodUpdateFood'] as Map).cast<String, dynamic>());
  }

  Future<bool?> foodDeleteFood({required int id}) async {
    final data = await _executor.mutation('mutation foodDeleteFood(\$id: Int!) { foodDeleteFood(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['foodDeleteFood'] as bool?;
  }

  Future<bool?> foodDeleteAllFoods() async {
    final data = await _executor.mutation('mutation foodDeleteAllFoods { foodDeleteAllFoods }'
    );
    return data['foodDeleteAllFoods'] as bool?;
  }

  Future<FoodAllergen?> foodCreateAllergen({required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodCreateAllergen(\$name: String!) { foodCreateAllergen(name: \$name) ${selectionSet ?? '{ id name }'} }', variables: {
      'name': name_,
    }
    );
    return data['foodCreateAllergen'] == null ? null : FoodAllergen.fromJson((data['foodCreateAllergen'] as Map).cast<String, dynamic>());
  }

  Future<FoodIngredient?> foodCreateIngredient({required String name_, String? unit, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodCreateIngredient(\$name: String!, \$unit: String) { foodCreateIngredient(name: \$name, unit: \$unit) ${selectionSet ?? '{ id name unit }'} }', variables: {
      'name': name_,
      'unit': ?unit,
    }
    );
    return data['foodCreateIngredient'] == null ? null : FoodIngredient.fromJson((data['foodCreateIngredient'] as Map).cast<String, dynamic>());
  }

  Future<FoodRestaurant?> foodCreateRestaurant({required FoodRestaurantInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodCreateRestaurant(\$input: FoodRestaurantInput!) { foodCreateRestaurant(input: \$input) ${selectionSet ?? '{ id name description address image_url is_active commission_rate }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['foodCreateRestaurant'] == null ? null : FoodRestaurant.fromJson((data['foodCreateRestaurant'] as Map).cast<String, dynamic>());
  }

  Future<FoodRestaurant?> foodUpdateRestaurant({required int id, required FoodRestaurantInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation foodUpdateRestaurant(\$id: Int!, \$input: FoodRestaurantInput!) { foodUpdateRestaurant(id: \$id, input: \$input) ${selectionSet ?? '{ id name description address image_url is_active commission_rate }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['foodUpdateRestaurant'] == null ? null : FoodRestaurant.fromJson((data['foodUpdateRestaurant'] as Map).cast<String, dynamic>());
  }

}
