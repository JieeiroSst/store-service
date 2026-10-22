// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class RestaurantOrderApi {
  const RestaurantOrderApi(this._executor);

  final GatewayExecutor _executor;

  Future<ROrderOrder?> rOrderOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query rOrderOrder(\$id: Int!) { rOrderOrder(id: \$id) ${selectionSet ?? '{ id table_name status kitchen_id total_amount items { id order_id food_id food_name quantity unit_price } created_date updated_time }'} }', variables: {
      'id': id,
    }
    );
    return data['rOrderOrder'] == null ? null : ROrderOrder.fromJson((data['rOrderOrder'] as Map).cast<String, dynamic>());
  }

  Future<ROrderOrderPage?> rOrderOrders({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query rOrderOrders(\$Limit: Int, \$Page: Int, \$Sort: String) { rOrderOrders(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id table_name status kitchen_id total_amount items { id order_id food_id food_name quantity unit_price } created_date updated_time } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rOrderOrders'] == null ? null : ROrderOrderPage.fromJson((data['rOrderOrders'] as Map).cast<String, dynamic>());
  }

  Future<ROrderReservationPage?> rOrderReservations({int? limit, int? page, String? sort, String? selectionSet}) async {
    final data = await _executor.query('query rOrderReservations(\$Limit: Int, \$Page: Int, \$Sort: String) { rOrderReservations(Limit: \$Limit, Page: \$Page, Sort: \$Sort) ${selectionSet ?? '{ limit page sort total_rows total_pages rows { id user_id table_name customer_name party_size reserved_at status created_date } }'} }', variables: {
      'Limit': ?limit,
      'Page': ?page,
      'Sort': ?sort,
    }
    );
    return data['rOrderReservations'] == null ? null : ROrderReservationPage.fromJson((data['rOrderReservations'] as Map).cast<String, dynamic>());
  }

  Future<List<ROrderReservation>?> myROrderReservations({int? limit, String? selectionSet}) async {
    final data = await _executor.query('query myROrderReservations(\$limit: Int) { myROrderReservations(limit: \$limit) ${selectionSet ?? '{ id user_id table_name customer_name party_size reserved_at status created_date }'} }', variables: {
      'limit': ?limit,
    }
    );
    return (data['myROrderReservations'] as List?)?.map((e) => (e == null ? null : ROrderReservation.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<ROrderReservation?> rOrderReservation({required int id, String? selectionSet}) async {
    final data = await _executor.query('query rOrderReservation(\$id: Int!) { rOrderReservation(id: \$id) ${selectionSet ?? '{ id user_id table_name customer_name party_size reserved_at status created_date }'} }', variables: {
      'id': id,
    }
    );
    return data['rOrderReservation'] == null ? null : ROrderReservation.fromJson((data['rOrderReservation'] as Map).cast<String, dynamic>());
  }

  Future<Object?> rOrderCreateReservation({required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation rOrderCreateReservation(\$input: Map!) { rOrderCreateReservation(input: \$input) }', variables: {
      'input': input,
    }
    );
    return data['rOrderCreateReservation'];
  }

  Future<Object?> rOrderCancelReservation({required String id}) async {
    final data = await _executor.mutation('mutation rOrderCancelReservation(\$id: String!) { rOrderCancelReservation(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['rOrderCancelReservation'];
  }

}
