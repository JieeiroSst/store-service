// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class DoordashApi {
  const DoordashApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<DoordashOrder>?> myDoordashOrders({String? selectionSet}) async {
    final data = await _executor.query('query myDoordashOrders { myDoordashOrders ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }'
    );
    return (data['myDoordashOrders'] as List?)?.map((e) => (e == null ? null : DoordashOrder.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<DoordashOrder?> doordashMyOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query doordashMyOrder(\$id: Int!) { doordashMyOrder(id: \$id) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['doordashMyOrder'] == null ? null : DoordashOrder.fromJson((data['doordashMyOrder'] as Map).cast<String, dynamic>());
  }

  Future<List<DoordashTracking>?> doordashMyOrderTracking({required int id, String? selectionSet}) async {
    final data = await _executor.query('query doordashMyOrderTracking(\$id: Int!) { doordashMyOrderTracking(id: \$id) ${selectionSet ?? '{ id order_id status latitude longitude timestamp notes }'} }', variables: {
      'id': id,
    }
    );
    return (data['doordashMyOrderTracking'] as List?)?.map((e) => (e == null ? null : DoordashTracking.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<DoordashOrder>?> doordashOrders({String? customerId, String? restaurantId, String? selectionSet}) async {
    final data = await _executor.query('query doordashOrders(\$customer_id: String, \$restaurant_id: String) { doordashOrders(customer_id: \$customer_id, restaurant_id: \$restaurant_id) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'customer_id': ?customerId,
      'restaurant_id': ?restaurantId,
    }
    );
    return (data['doordashOrders'] as List?)?.map((e) => (e == null ? null : DoordashOrder.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<DoordashOrder?> doordashOrder({required int id, String? selectionSet}) async {
    final data = await _executor.query('query doordashOrder(\$id: Int!) { doordashOrder(id: \$id) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['doordashOrder'] == null ? null : DoordashOrder.fromJson((data['doordashOrder'] as Map).cast<String, dynamic>());
  }

  Future<List<DoordashTracking>?> doordashOrderTracking({required int id, String? selectionSet}) async {
    final data = await _executor.query('query doordashOrderTracking(\$id: Int!) { doordashOrderTracking(id: \$id) ${selectionSet ?? '{ id order_id status latitude longitude timestamp notes }'} }', variables: {
      'id': id,
    }
    );
    return (data['doordashOrderTracking'] as List?)?.map((e) => (e == null ? null : DoordashTracking.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<DoordashAssignment>?> doordashDriverAssignments({required String driverId, String? selectionSet}) async {
    final data = await _executor.query('query doordashDriverAssignments(\$driver_id: String!) { doordashDriverAssignments(driver_id: \$driver_id) ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }', variables: {
      'driver_id': driverId,
    }
    );
    return (data['doordashDriverAssignments'] as List?)?.map((e) => (e == null ? null : DoordashAssignment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<DoordashAssignment>?> myDriverAssignments({String? selectionSet}) async {
    final data = await _executor.query('query myDriverAssignments { myDriverAssignments ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }'
    );
    return (data['myDriverAssignments'] as List?)?.map((e) => (e == null ? null : DoordashAssignment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<DoordashOrder?> doordashPlaceOrder({required DoordashOrderInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashPlaceOrder(\$input: DoordashOrderInput!) { doordashPlaceOrder(input: \$input) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['doordashPlaceOrder'] == null ? null : DoordashOrder.fromJson((data['doordashPlaceOrder'] as Map).cast<String, dynamic>());
  }

  Future<DoordashOrder?> doordashCancelMyOrder({required int id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashCancelMyOrder(\$id: Int!, \$reason: String) { doordashCancelMyOrder(id: \$id, reason: \$reason) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['doordashCancelMyOrder'] == null ? null : DoordashOrder.fromJson((data['doordashCancelMyOrder'] as Map).cast<String, dynamic>());
  }

  Future<DoordashOrder?> doordashUpdateOrderStatus({required int id, required String status, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashUpdateOrderStatus(\$id: Int!, \$status: String!) { doordashUpdateOrderStatus(id: \$id, status: \$status) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'id': id,
      'status': status,
    }
    );
    return data['doordashUpdateOrderStatus'] == null ? null : DoordashOrder.fromJson((data['doordashUpdateOrderStatus'] as Map).cast<String, dynamic>());
  }

  Future<DoordashOrder?> doordashCancelOrder({required int id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashCancelOrder(\$id: Int!, \$reason: String) { doordashCancelOrder(id: \$id, reason: \$reason) ${selectionSet ?? '{ id customer_id restaurant_id driver_id delivery_address_id status placed_at estimated_delivery_time actual_delivery_time subtotal delivery_fee service_fee tax tip total_amount payment_method_id payment_status special_instructions created_at updated_at items { id order_id menu_item_id name quantity unit_price customizations special_instructions created_at } }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['doordashCancelOrder'] == null ? null : DoordashOrder.fromJson((data['doordashCancelOrder'] as Map).cast<String, dynamic>());
  }

  Future<DoordashAssignment?> doordashAssignDriver({required int id, required String driverId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashAssignDriver(\$id: Int!, \$driver_id: String!) { doordashAssignDriver(id: \$id, driver_id: \$driver_id) ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }', variables: {
      'id': id,
      'driver_id': driverId,
    }
    );
    return data['doordashAssignDriver'] == null ? null : DoordashAssignment.fromJson((data['doordashAssignDriver'] as Map).cast<String, dynamic>());
  }

  Future<DoordashTracking?> doordashAddTrackingEvent({required int id, required DoordashTrackingInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashAddTrackingEvent(\$id: Int!, \$input: DoordashTrackingInput!) { doordashAddTrackingEvent(id: \$id, input: \$input) ${selectionSet ?? '{ id order_id status latitude longitude timestamp notes }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['doordashAddTrackingEvent'] == null ? null : DoordashTracking.fromJson((data['doordashAddTrackingEvent'] as Map).cast<String, dynamic>());
  }

  Future<DoordashAssignment?> doordashAcceptAssignment({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashAcceptAssignment(\$id: Int!) { doordashAcceptAssignment(id: \$id) ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }', variables: {
      'id': id,
    }
    );
    return data['doordashAcceptAssignment'] == null ? null : DoordashAssignment.fromJson((data['doordashAcceptAssignment'] as Map).cast<String, dynamic>());
  }

  Future<DoordashAssignment?> doordashRejectAssignment({required int id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashRejectAssignment(\$id: Int!, \$reason: String) { doordashRejectAssignment(id: \$id, reason: \$reason) ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['doordashRejectAssignment'] == null ? null : DoordashAssignment.fromJson((data['doordashRejectAssignment'] as Map).cast<String, dynamic>());
  }

  Future<DoordashAssignment?> doordashCompleteAssignment({required int id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation doordashCompleteAssignment(\$id: Int!) { doordashCompleteAssignment(id: \$id) ${selectionSet ?? '{ id driver_id order_id status assigned_at accepted_at completed_at rejection_reason }'} }', variables: {
      'id': id,
    }
    );
    return data['doordashCompleteAssignment'] == null ? null : DoordashAssignment.fromJson((data['doordashCompleteAssignment'] as Map).cast<String, dynamic>());
  }

}
