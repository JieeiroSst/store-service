// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final shippingAdminOps = <AdminOp>[
  AdminOp(
    'shipments',
    resource: 'shipment',
    verb: '',
    fields: [
      KField('status', 'Status'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipments(status: v.str('status'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'shipment',
    resource: 'shipment',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipment(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'shipmentByClientCode',
    resource: 'general',
    verb: '',
    fields: [
      KField('code', 'Code', required: true),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipmentByClientCode(code: v.need(v.str('code'), 'code'))),
  ),
  AdminOp(
    'shipWarehouses',
    resource: 'warehouse',
    verb: '',
    run: (api, v) async => adminPlain(await api.shipping.shipWarehouses()),
  ),
  AdminOp(
    'shipCreateShipment',
    resource: 'shipment',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.client_order_code', 'Client order code'),
      KField('input.recipient', 'Recipient', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"name":"","phone":"","street":"","ward_code":"","district_id":0,"province_id":0}'),
      KField('input.parcel', 'Parcel', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"weight_gram":0,"length_cm":0,"width_cm":0,"height_cm":0,"content":"","items":[{"name":"","code":"","quantity":0,"weight_gram":0,"price":0}]}'),
      KField('input.cod_amount', 'Cod amount', kind: FieldKind.number),
      KField('input.insurance_value', 'Insurance value', kind: FieldKind.number),
      KField('input.payment_type', 'Payment type', kind: FieldKind.number),
      KField('input.required_note', 'Required note', kind: FieldKind.multiline),
      KField('input.note', 'Note', kind: FieldKind.multiline),
      KField('input.strategy', 'Strategy'),
      KField('input.warehouse_ids', 'Warehouse ids', hint: 'Nhiều giá trị, cách nhau bởi dấu phẩy'),
      KField('input.warehouse_id', 'Warehouse ID', kind: FieldKind.number),
      KField('input.service_id', 'Service ID', kind: FieldKind.number),
      KField('input.customer', 'Customer', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('input.callback_url', 'Callback URL'),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipCreateShipment(input: ShipCreateShipmentInput(clientOrderCode: v.str('input.client_order_code'), recipient: v.json('input.recipient') == null ? null : ShipAddressInput.fromJson((v.json('input.recipient') as Map).cast<String, dynamic>()), parcel: v.json('input.parcel') == null ? null : ShipParcelInput.fromJson((v.json('input.parcel') as Map).cast<String, dynamic>()), codAmount: v.integer('input.cod_amount'), insuranceValue: v.integer('input.insurance_value'), paymentType: v.integer('input.payment_type'), requiredNote: v.str('input.required_note'), note: v.str('input.note'), strategy: v.str('input.strategy'), warehouseIds: v.intList('input.warehouse_ids'), warehouseId: v.integer('input.warehouse_id'), serviceId: v.integer('input.service_id'), customer: v.jsonMap('input.customer'), callbackUrl: v.str('input.callback_url')))),
  ),
  AdminOp(
    'shipPlaceShipment',
    resource: 'place shipment',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipPlaceShipment(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'shipCancelShipment',
    resource: 'shipment',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipCancelShipment(id: v.need(v.integer('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'shipSyncShipment',
    resource: 'shipment',
    verb: 'sync',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipSyncShipment(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'shipCreateWarehouse',
    resource: 'warehouse',
    verb: 'create',
    mutation: true,
    fields: [
      KField('code', 'Code', required: true),
      KField('name', 'Name', required: true),
      KField('phone', 'Phone', kind: FieldKind.phone),
      KField('street', 'Street'),
      KField('ward_code', 'Ward code'),
      KField('district_id', 'District ID', kind: FieldKind.number),
      KField('province_id', 'Province ID', kind: FieldKind.number),
      KField('carrier_shop_id', 'Carrier shop ID', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipCreateWarehouse(code: v.need(v.str('code'), 'code'), name_: v.need(v.str('name'), 'name'), phone: v.str('phone'), street: v.str('street'), wardCode: v.str('ward_code'), districtId: v.integer('district_id'), provinceId: v.integer('province_id'), carrierShopId: v.integer('carrier_shop_id'))),
  ),
  AdminOp(
    'shipSetWarehouseActive',
    resource: 'warehouse',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('active', 'Active', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.shipping.shipSetWarehouseActive(id: v.need(v.integer('id'), 'id'), active: v.flag('active'))),
  ),
];
