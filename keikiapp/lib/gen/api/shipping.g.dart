// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class ShippingApi {
  const ShippingApi(this._executor);

  final GatewayExecutor _executor;

  Future<ShipShipmentList?> shipments({String? status, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query shipments(\$status: String, \$limit: Int, \$offset: Int) { shipments(status: \$status, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ items { id code client_service client_order_code status carrier_status carrier_order_code warehouse_id carrier_shop_id service_id service_type_id service_name strategy recipient { name phone street ward_code district_id province_id } parcel { weight_gram length_cm width_cm height_cm content } customer { user_id email } cod_amount insurance_value payment_type required_note note quoted_fee shipping_fee expected_delivery_at delivered_at callback_url place_attempts last_error created_at updated_at } total }'} }', variables: {
      'status': ?status,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['shipments'] == null ? null : ShipShipmentList.fromJson((data['shipments'] as Map).cast<String, dynamic>());
  }

  Future<ShipShipmentDetail?> shipment({required int id, String? selectionSet}) async {
    final data = await _executor.query('query shipment(\$id: Int!) { shipment(id: \$id) ${selectionSet ?? '{ shipment { id code client_service client_order_code status carrier_status carrier_order_code warehouse_id carrier_shop_id service_id service_type_id service_name strategy recipient { name phone street ward_code district_id province_id } parcel { weight_gram length_cm width_cm height_cm content } customer { user_id email } cod_amount insurance_value payment_type required_note note quoted_fee shipping_fee expected_delivery_at delivered_at callback_url place_attempts last_error created_at updated_at } events { id shipment_id source from_status to_status carrier_status reason applied occurred_at created_at } }'} }', variables: {
      'id': id,
    }
    );
    return data['shipment'] == null ? null : ShipShipmentDetail.fromJson((data['shipment'] as Map).cast<String, dynamic>());
  }

  Future<ShipShipmentDetail?> shipmentByClientCode({required String code, String? selectionSet}) async {
    final data = await _executor.query('query shipmentByClientCode(\$code: String!) { shipmentByClientCode(code: \$code) ${selectionSet ?? '{ shipment { id code client_service client_order_code status carrier_status carrier_order_code warehouse_id carrier_shop_id service_id service_type_id service_name strategy recipient { name phone street ward_code district_id province_id } parcel { weight_gram length_cm width_cm height_cm content } customer { user_id email } cod_amount insurance_value payment_type required_note note quoted_fee shipping_fee expected_delivery_at delivered_at callback_url place_attempts last_error created_at updated_at } events { id shipment_id source from_status to_status carrier_status reason applied occurred_at created_at } }'} }', variables: {
      'code': code,
    }
    );
    return data['shipmentByClientCode'] == null ? null : ShipShipmentDetail.fromJson((data['shipmentByClientCode'] as Map).cast<String, dynamic>());
  }

  Future<List<ShipWarehouse>?> shipWarehouses({String? selectionSet}) async {
    final data = await _executor.query('query shipWarehouses { shipWarehouses ${selectionSet ?? '{ id code name phone street ward_code district_id province_id carrier_shop_id active created_at updated_at }'} }'
    );
    return (data['shipWarehouses'] as List?)?.map((e) => (e == null ? null : ShipWarehouse.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ShipProvince>?> shipProvinces({String? selectionSet}) async {
    final data = await _executor.public('query shipProvinces { shipProvinces ${selectionSet ?? '{ id name }'} }'
    );
    return (data['shipProvinces'] as List?)?.map((e) => (e == null ? null : ShipProvince.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ShipDistrict>?> shipDistricts({required int provinceId, String? selectionSet}) async {
    final data = await _executor.public('query shipDistricts(\$province_id: Int!) { shipDistricts(province_id: \$province_id) ${selectionSet ?? '{ id province_id name }'} }', variables: {
      'province_id': provinceId,
    }
    );
    return (data['shipDistricts'] as List?)?.map((e) => (e == null ? null : ShipDistrict.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<ShipWard>?> shipWards({required int districtId, String? selectionSet}) async {
    final data = await _executor.public('query shipWards(\$district_id: Int!) { shipWards(district_id: \$district_id) ${selectionSet ?? '{ code district_id name }'} }', variables: {
      'district_id': districtId,
    }
    );
    return (data['shipWards'] as List?)?.map((e) => (e == null ? null : ShipWard.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<Object?> shipQuote({required ShipQuoteInput input}) async {
    final data = await _executor.query('query shipQuote(\$input: ShipQuoteInput!) { shipQuote(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shipQuote'];
  }

  Future<Object?> shipCreateShipment({required ShipCreateShipmentInput input}) async {
    final data = await _executor.mutation('mutation shipCreateShipment(\$input: ShipCreateShipmentInput!) { shipCreateShipment(input: \$input) }', variables: {
      'input': input.toJson(),
    }
    );
    return data['shipCreateShipment'];
  }

  Future<Object?> shipPlaceShipment({required int id}) async {
    final data = await _executor.mutation('mutation shipPlaceShipment(\$id: Int64!) { shipPlaceShipment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shipPlaceShipment'];
  }

  Future<Object?> shipCancelShipment({required int id, String? reason}) async {
    final data = await _executor.mutation('mutation shipCancelShipment(\$id: Int64!, \$reason: String) { shipCancelShipment(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['shipCancelShipment'];
  }

  Future<Object?> shipSyncShipment({required int id}) async {
    final data = await _executor.mutation('mutation shipSyncShipment(\$id: Int64!) { shipSyncShipment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['shipSyncShipment'];
  }

  Future<ShipWarehouse?> shipCreateWarehouse({required String code, required String name_, String? phone, String? street, String? wardCode, int? districtId, int? provinceId, int? carrierShopId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shipCreateWarehouse(\$code: String!, \$name: String!, \$phone: String, \$street: String, \$ward_code: String, \$district_id: Int, \$province_id: Int, \$carrier_shop_id: Int64) { shipCreateWarehouse(code: \$code, name: \$name, phone: \$phone, street: \$street, ward_code: \$ward_code, district_id: \$district_id, province_id: \$province_id, carrier_shop_id: \$carrier_shop_id) ${selectionSet ?? '{ id code name phone street ward_code district_id province_id carrier_shop_id active created_at updated_at }'} }', variables: {
      'code': code,
      'name': name_,
      'phone': ?phone,
      'street': ?street,
      'ward_code': ?wardCode,
      'district_id': ?districtId,
      'province_id': ?provinceId,
      'carrier_shop_id': ?carrierShopId,
    }
    );
    return data['shipCreateWarehouse'] == null ? null : ShipWarehouse.fromJson((data['shipCreateWarehouse'] as Map).cast<String, dynamic>());
  }

  Future<ShipWarehouse?> shipSetWarehouseActive({required int id, required bool active, String? selectionSet}) async {
    final data = await _executor.mutation('mutation shipSetWarehouseActive(\$id: Int64!, \$active: Boolean!) { shipSetWarehouseActive(id: \$id, active: \$active) ${selectionSet ?? '{ id code name phone street ward_code district_id province_id carrier_shop_id active created_at updated_at }'} }', variables: {
      'id': id,
      'active': active,
    }
    );
    return data['shipSetWarehouseActive'] == null ? null : ShipWarehouse.fromJson((data['shipSetWarehouseActive'] as Map).cast<String, dynamic>());
  }

}
