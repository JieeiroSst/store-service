// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RestaurantDeliveryApi {
  const RestaurantDeliveryApi(this._executor);

  final GatewayExecutor _executor;

  Future<RDeliveryPage?> rDeliveryDeliveries({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query rDeliveryDeliveries(\$Limit: Int, \$Page: Int, \$Sort: String) { rDeliveryDeliveries(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { ship_id name address kitchen_id status distance_km fee eta_minutes } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rDeliveryDeliveries'] == null ? null : RDeliveryPage.fromJson((data['rDeliveryDeliveries'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rDeliveryRegister({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rDeliveryRegister(\$input: Map!) { rDeliveryRegister(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rDeliveryRegister'];
  }

}
