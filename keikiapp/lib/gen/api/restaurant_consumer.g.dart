// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RestaurantConsumerApi {
  const RestaurantConsumerApi(this._executor);

  final GatewayExecutor _executor;

  Future<RConsumerPage?> rConsumerConsumers({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query rConsumerConsumers(\$Limit: Int, \$Page: Int, \$Sort: String) { rConsumerConsumers(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id name kitchen_id menu } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rConsumerConsumers'] == null ? null : RConsumerPage.fromJson((data['rConsumerConsumers'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rConsumerPlaceOrder({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rConsumerPlaceOrder(\$input: Map!) { rConsumerPlaceOrder(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rConsumerPlaceOrder'];
  }

}
