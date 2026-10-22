// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class VendingApi {
  const VendingApi(this._executor);

  final GatewayExecutor _executor;

  Future<VendMachinePage?> vendMachines({String? status, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query vendMachines(\$status: String, \$cursor: String, \$limit: Int) { vendMachines(status: \$status, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id location model status last_maintenance_at created_at updated_at } next_cursor is_last_page limit }'} }', variables: {
      'status': ?status,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['vendMachines'] == null ? null : VendMachinePage.fromJson((data['vendMachines'] as Map).cast<String, dynamic>());
  }

  Future<VendMachine?> vendMachine({required String id, String? selectionSet}) async {
    final data = await _executor.public('query vendMachine(\$id: String!) { vendMachine(id: \$id) ${selectionSet ?? '{ id location model status last_maintenance_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['vendMachine'] == null ? null : VendMachine.fromJson((data['vendMachine'] as Map).cast<String, dynamic>());
  }

  Future<List<VendMaintenance>?> vendMaintenance({required String id, String? selectionSet}) async {
    final data = await _executor.query('query vendMaintenance(\$id: String!) { vendMaintenance(id: \$id) ${selectionSet ?? '{ id machine_id technician_id maintenance_type notes performed_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['vendMaintenance'] as List?)?.map((e) => (e == null ? null : VendMaintenance.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<VendEvent>?> vendEvents({required String id, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query vendEvents(\$id: String!, \$limit: Int) { vendEvents(id: \$id, limit: \$limit) ${selectionSet ?? '{ id event_type related_entity entity_id machine_id data occurred_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
    }
    );
    return (data['vendEvents'] as List?)?.map((e) => (e == null ? null : VendEvent.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VendSalesReport?> vendSalesReport({required String id, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query vendSalesReport(\$id: String!, \$from: String, \$to: String) { vendSalesReport(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ machine_id from to orders revenue_cents discount_cents products { product_id product_name units revenue_cents discount_cents } }'} }', variables: {
      'id': id,
      'from': ?from,
      'to': ?to,
    }
    );
    return data['vendSalesReport'] == null ? null : VendSalesReport.fromJson((data['vendSalesReport'] as Map).cast<String, dynamic>());
  }

  Future<VendInventoryPage?> vendInventory({required String id, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query vendInventory(\$id: String!, \$cursor: String, \$limit: Int) { vendInventory(id: \$id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id machine_id slot product_id product { id name description price_cents category_id image_url barcode is_active attributes created_at } quantity max_capacity low_threshold is_low last_restocked_at updated_at } next_cursor is_last_page limit }'} }', variables: {
      'id': id,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['vendInventory'] == null ? null : VendInventoryPage.fromJson((data['vendInventory'] as Map).cast<String, dynamic>());
  }

  Future<VendInventoryPage?> vendLowInventory({String? machineId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.query('query vendLowInventory(\$machine_id: String, \$cursor: String, \$limit: Int) { vendLowInventory(machine_id: \$machine_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id machine_id slot product_id product { id name description price_cents category_id image_url barcode is_active attributes created_at } quantity max_capacity low_threshold is_low last_restocked_at updated_at } next_cursor is_last_page limit }'} }', variables: {
      'machine_id': ?machineId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['vendLowInventory'] == null ? null : VendInventoryPage.fromJson((data['vendLowInventory'] as Map).cast<String, dynamic>());
  }

  Future<VendCategoryPage?> vendCategories({String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query vendCategories(\$cursor: String, \$limit: Int) { vendCategories(cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id name description display_order created_at } next_cursor is_last_page limit }'} }', variables: {
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['vendCategories'] == null ? null : VendCategoryPage.fromJson((data['vendCategories'] as Map).cast<String, dynamic>());
  }

  Future<VendProductPage?> vendProducts({String? categoryId, String? cursor, int? limit, String? selectionSet}) async {
    final data = await _executor.public('query vendProducts(\$category_id: String, \$cursor: String, \$limit: Int) { vendProducts(category_id: \$category_id, cursor: \$cursor, limit: \$limit) ${selectionSet ?? '{ items { id name description price_cents category_id image_url barcode is_active attributes created_at } next_cursor is_last_page limit }'} }', variables: {
      'category_id': ?categoryId,
      'cursor': ?cursor,
      'limit': ?limit,
    }
    );
    return data['vendProducts'] == null ? null : VendProductPage.fromJson((data['vendProducts'] as Map).cast<String, dynamic>());
  }

  Future<VendProduct?> vendProduct({required String id, String? selectionSet}) async {
    final data = await _executor.public('query vendProduct(\$id: String!) { vendProduct(id: \$id) ${selectionSet ?? '{ id name description price_cents category_id image_url barcode is_active attributes created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['vendProduct'] == null ? null : VendProduct.fromJson((data['vendProduct'] as Map).cast<String, dynamic>());
  }

  Future<VendSession?> vendSession({required String id, String? selectionSet}) async {
    final data = await _executor.query('query vendSession(\$id: String!) { vendSession(id: \$id) ${selectionSet ?? '{ id machine_id status started_at expires_at }'} }', variables: {
      'id': id,
    }
    );
    return data['vendSession'] == null ? null : VendSession.fromJson((data['vendSession'] as Map).cast<String, dynamic>());
  }

  Future<List<VendOrder>?> vendSessionOrders({required String id, String? selectionSet}) async {
    final data = await _executor.query('query vendSessionOrders(\$id: String!) { vendSessionOrders(id: \$id) ${selectionSet ?? '{ id order_no session_id reservation_id payment_id status fulfilled_at created_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['vendSessionOrders'] as List?)?.map((e) => (e == null ? null : VendOrder.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VendPaymentStatus?> vendPayment({required String id, String? selectionSet}) async {
    final data = await _executor.query('query vendPayment(\$id: String!) { vendPayment(id: \$id) ${selectionSet ?? '{ payment { id amount_cents discount_cents coupon_code currency payment_method status transaction_id completed_at } order { id order_no session_id reservation_id payment_id status fulfilled_at created_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['vendPayment'] == null ? null : VendPaymentStatus.fromJson((data['vendPayment'] as Map).cast<String, dynamic>());
  }

  Future<VendOrder?> vendOrder({required String id, String? selectionSet}) async {
    final data = await _executor.query('query vendOrder(\$id: String!) { vendOrder(id: \$id) ${selectionSet ?? '{ id order_no session_id reservation_id payment_id status fulfilled_at created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['vendOrder'] == null ? null : VendOrder.fromJson((data['vendOrder'] as Map).cast<String, dynamic>());
  }

  Future<VendMachine?> vendRegisterMachine({required String location, required String machineModel, String? status, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendRegisterMachine(\$location: String!, \$machineModel: String!, \$status: String) { vendRegisterMachine(location: \$location, machineModel: \$machineModel, status: \$status) ${selectionSet ?? '{ id location model status last_maintenance_at created_at updated_at }'} }', variables: {
      'location': location,
      'machineModel': machineModel,
      'status': ?status,
    }
    );
    return data['vendRegisterMachine'] == null ? null : VendMachine.fromJson((data['vendRegisterMachine'] as Map).cast<String, dynamic>());
  }

  Future<VendMachine?> vendUpdateMachine({required String id, String? location, String? machineModel, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendUpdateMachine(\$id: String!, \$location: String, \$machineModel: String) { vendUpdateMachine(id: \$id, location: \$location, machineModel: \$machineModel) ${selectionSet ?? '{ id location model status last_maintenance_at created_at updated_at }'} }', variables: {
      'id': id,
      'location': ?location,
      'machineModel': ?machineModel,
    }
    );
    return data['vendUpdateMachine'] == null ? null : VendMachine.fromJson((data['vendUpdateMachine'] as Map).cast<String, dynamic>());
  }

  Future<VendMachine?> vendChangeMachineStatus({required String id, required String status, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendChangeMachineStatus(\$id: String!, \$status: String!) { vendChangeMachineStatus(id: \$id, status: \$status) ${selectionSet ?? '{ id location model status last_maintenance_at created_at updated_at }'} }', variables: {
      'id': id,
      'status': status,
    }
    );
    return data['vendChangeMachineStatus'] == null ? null : VendMachine.fromJson((data['vendChangeMachineStatus'] as Map).cast<String, dynamic>());
  }

  Future<VendMaintenance?> vendRecordMaintenance({required String id, required String technicianId, required String maintenanceType, String? notes, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendRecordMaintenance(\$id: String!, \$technician_id: String!, \$maintenance_type: String!, \$notes: String) { vendRecordMaintenance(id: \$id, technician_id: \$technician_id, maintenance_type: \$maintenance_type, notes: \$notes) ${selectionSet ?? '{ id machine_id technician_id maintenance_type notes performed_at }'} }', variables: {
      'id': id,
      'technician_id': technicianId,
      'maintenance_type': maintenanceType,
      'notes': ?notes,
    }
    );
    return data['vendRecordMaintenance'] == null ? null : VendMaintenance.fromJson((data['vendRecordMaintenance'] as Map).cast<String, dynamic>());
  }

  Future<VendInventory?> vendAssignSlot({required String id, required String slot, required VendSlotInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendAssignSlot(\$id: String!, \$slot: String!, \$input: VendSlotInput!) { vendAssignSlot(id: \$id, slot: \$slot, input: \$input) ${selectionSet ?? '{ id machine_id slot product_id product { id name description price_cents category_id image_url barcode is_active attributes created_at } quantity max_capacity low_threshold is_low last_restocked_at updated_at }'} }', variables: {
      'id': id,
      'slot': slot,
      'input': input.toJson(),
    }
    );
    return data['vendAssignSlot'] == null ? null : VendInventory.fromJson((data['vendAssignSlot'] as Map).cast<String, dynamic>());
  }

  Future<VendInventory?> vendRestock({required String id, required int quantity, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendRestock(\$id: String!, \$quantity: Int!) { vendRestock(id: \$id, quantity: \$quantity) ${selectionSet ?? '{ id machine_id slot product_id product { id name description price_cents category_id image_url barcode is_active attributes created_at } quantity max_capacity low_threshold is_low last_restocked_at updated_at }'} }', variables: {
      'id': id,
      'quantity': quantity,
    }
    );
    return data['vendRestock'] == null ? null : VendInventory.fromJson((data['vendRestock'] as Map).cast<String, dynamic>());
  }

  Future<VendCategory?> vendCreateCategory({required String name_, String? description, int? displayOrder, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendCreateCategory(\$name: String!, \$description: String, \$display_order: Int) { vendCreateCategory(name: \$name, description: \$description, display_order: \$display_order) ${selectionSet ?? '{ id name description display_order created_at }'} }', variables: {
      'name': name_,
      'description': ?description,
      'display_order': ?displayOrder,
    }
    );
    return data['vendCreateCategory'] == null ? null : VendCategory.fromJson((data['vendCreateCategory'] as Map).cast<String, dynamic>());
  }

  Future<VendProduct?> vendCreateProduct({required VendProductInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendCreateProduct(\$input: VendProductInput!) { vendCreateProduct(input: \$input) ${selectionSet ?? '{ id name description price_cents category_id image_url barcode is_active attributes created_at }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['vendCreateProduct'] == null ? null : VendProduct.fromJson((data['vendCreateProduct'] as Map).cast<String, dynamic>());
  }

  Future<VendProduct?> vendUpdateProduct({required String id, required VendProductPatchInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendUpdateProduct(\$id: String!, \$input: VendProductPatchInput!) { vendUpdateProduct(id: \$id, input: \$input) ${selectionSet ?? '{ id name description price_cents category_id image_url barcode is_active attributes created_at }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['vendUpdateProduct'] == null ? null : VendProduct.fromJson((data['vendUpdateProduct'] as Map).cast<String, dynamic>());
  }

  Future<VendSession?> vendStartSession({required String machineId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendStartSession(\$machine_id: String!) { vendStartSession(machine_id: \$machine_id) ${selectionSet ?? '{ id machine_id status started_at expires_at }'} }', variables: {
      'machine_id': machineId,
    }
    );
    return data['vendStartSession'] == null ? null : VendSession.fromJson((data['vendStartSession'] as Map).cast<String, dynamic>());
  }

  Future<VendSession?> vendEndSession({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendEndSession(\$id: String!) { vendEndSession(id: \$id) ${selectionSet ?? '{ id machine_id status started_at expires_at }'} }', variables: {
      'id': id,
    }
    );
    return data['vendEndSession'] == null ? null : VendSession.fromJson((data['vendEndSession'] as Map).cast<String, dynamic>());
  }

  Future<Object?> vendReserve({required String id, required String slot}) async {
    final data = await _executor.mutation('mutation vendReserve(\$id: String!, \$slot: String!) { vendReserve(id: \$id, slot: \$slot) }', variables: {
      'id': id,
      'slot': slot,
    }
    );
    return data['vendReserve'];
  }

  Future<Object?> vendCancelReservation({required String id}) async {
    final data = await _executor.mutation('mutation vendCancelReservation(\$id: String!) { vendCancelReservation(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['vendCancelReservation'];
  }

  Future<VendOrder?> vendCheckout({required String id, required String paymentMethod, String? walletId, String? couponCode, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendCheckout(\$id: String!, \$payment_method: String!, \$wallet_id: String, \$coupon_code: String) { vendCheckout(id: \$id, payment_method: \$payment_method, wallet_id: \$wallet_id, coupon_code: \$coupon_code) ${selectionSet ?? '{ id order_no session_id reservation_id payment_id status fulfilled_at created_at }'} }', variables: {
      'id': id,
      'payment_method': paymentMethod,
      'wallet_id': ?walletId,
      'coupon_code': ?couponCode,
    }
    );
    return data['vendCheckout'] == null ? null : VendOrder.fromJson((data['vendCheckout'] as Map).cast<String, dynamic>());
  }

  Future<VendOrder?> vendReportDispense({required String id, required bool dispensed, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendReportDispense(\$id: String!, \$dispensed: Boolean!) { vendReportDispense(id: \$id, dispensed: \$dispensed) ${selectionSet ?? '{ id order_no session_id reservation_id payment_id status fulfilled_at created_at }'} }', variables: {
      'id': id,
      'dispensed': dispensed,
    }
    );
    return data['vendReportDispense'] == null ? null : VendOrder.fromJson((data['vendReportDispense'] as Map).cast<String, dynamic>());
  }

  Future<VendOrder?> vendRefundOrder({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation vendRefundOrder(\$id: String!, \$reason: String) { vendRefundOrder(id: \$id, reason: \$reason) ${selectionSet ?? '{ id order_no session_id reservation_id payment_id status fulfilled_at created_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['vendRefundOrder'] == null ? null : VendOrder.fromJson((data['vendRefundOrder'] as Map).cast<String, dynamic>());
  }

}
