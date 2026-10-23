// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final vendingAdminOps = <AdminOp>[
  AdminOp(
    'vendMaintenance',
    resource: 'maintenance',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendMaintenance(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'vendEvents',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendEvents(id: v.need(v.str('id'), 'id'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'vendSalesReport',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('from', 'From', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendSalesReport(id: v.need(v.str('id'), 'id'), from: v.str('from'), to: v.str('to'))),
  ),
  AdminOp(
    'vendLowInventory',
    resource: 'general',
    verb: '',
    fields: [
      KField('machine_id', 'Machine ID'),
      KField('cursor', 'Cursor'),
      KField('limit', 'Limit', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendLowInventory(machineId: v.str('machine_id'), cursor: v.str('cursor'), limit: v.integer('limit'))),
  ),
  AdminOp(
    'vendSession',
    resource: 'session',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendSession(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'vendSessionOrders',
    resource: 'session',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendSessionOrders(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'vendPayment',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendPayment(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'vendOrder',
    resource: 'order',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendOrder(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'vendRegisterMachine',
    resource: 'machine',
    verb: 'register',
    mutation: true,
    fields: [
      KField('location', 'Location', required: true),
      KField('machineModel', 'Machine model', required: true),
      KField('status', 'Status'),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendRegisterMachine(location: v.need(v.str('location'), 'location'), machineModel: v.need(v.str('machineModel'), 'machineModel'), status: v.str('status'))),
  ),
  AdminOp(
    'vendUpdateMachine',
    resource: 'machine',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('location', 'Location'),
      KField('machineModel', 'Machine model'),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendUpdateMachine(id: v.need(v.str('id'), 'id'), location: v.str('location'), machineModel: v.str('machineModel'))),
  ),
  AdminOp(
    'vendChangeMachineStatus',
    resource: 'machine',
    verb: 'change',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('status', 'Status', required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendChangeMachineStatus(id: v.need(v.str('id'), 'id'), status: v.need(v.str('status'), 'status'))),
  ),
  AdminOp(
    'vendRecordMaintenance',
    resource: 'maintenance',
    verb: 'record',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('technician_id', 'Technician ID', required: true),
      KField('maintenance_type', 'Maintenance type', required: true),
      KField('notes', 'Notes', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendRecordMaintenance(id: v.need(v.str('id'), 'id'), technicianId: v.need(v.str('technician_id'), 'technician_id'), maintenanceType: v.need(v.str('maintenance_type'), 'maintenance_type'), notes: v.str('notes'))),
  ),
  AdminOp(
    'vendAssignSlot',
    resource: 'general',
    verb: 'assign',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('slot', 'Slot', required: true),
      KField('input.product_id', 'Product ID'),
      KField('input.quantity', 'Quantity', kind: FieldKind.number),
      KField('input.max_capacity', 'Max capacity', kind: FieldKind.number),
      KField('input.low_threshold', 'Low threshold', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendAssignSlot(id: v.need(v.str('id'), 'id'), slot: v.need(v.str('slot'), 'slot'), input: VendSlotInput(productId: v.str('input.product_id'), quantity: v.integer('input.quantity'), maxCapacity: v.integer('input.max_capacity'), lowThreshold: v.integer('input.low_threshold')))),
  ),
  AdminOp(
    'vendRestock',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('quantity', 'Quantity', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendRestock(id: v.need(v.str('id'), 'id'), quantity: v.need(v.integer('quantity'), 'quantity'))),
  ),
  AdminOp(
    'vendCreateCategory',
    resource: 'general',
    verb: 'create',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
      KField('display_order', 'Display order', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendCreateCategory(name_: v.need(v.str('name'), 'name'), description: v.str('description'), displayOrder: v.integer('display_order'))),
  ),
  AdminOp(
    'vendCreateProduct',
    resource: 'product',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.price_cents', 'Price cents', kind: FieldKind.number),
      KField('input.category_id', 'Category ID'),
      KField('input.image_url', 'Image URL'),
      KField('input.barcode', 'Barcode'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
      KField('input.attributes', 'Attributes', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendCreateProduct(input: VendProductInput(name_: v.str('input.name'), description: v.str('input.description'), priceCents: v.integer('input.price_cents'), categoryId: v.str('input.category_id'), imageUrl: v.str('input.image_url'), barcode: v.str('input.barcode'), isActive: v.optFlag('input.is_active'), attributes: v.jsonMap('input.attributes')))),
  ),
  AdminOp(
    'vendUpdateProduct',
    resource: 'product',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input.name', 'Name'),
      KField('input.description', 'Description', kind: FieldKind.multiline),
      KField('input.price_cents', 'Price cents', kind: FieldKind.number),
      KField('input.image_url', 'Image URL'),
      KField('input.barcode', 'Barcode'),
      KField('input.is_active', 'Is active', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendUpdateProduct(id: v.need(v.str('id'), 'id'), input: VendProductPatchInput(name_: v.str('input.name'), description: v.str('input.description'), priceCents: v.integer('input.price_cents'), imageUrl: v.str('input.image_url'), barcode: v.str('input.barcode'), isActive: v.optFlag('input.is_active')))),
  ),
  AdminOp(
    'vendReportDispense',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('dispensed', 'Dispensed', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendReportDispense(id: v.need(v.str('id'), 'id'), dispensed: v.flag('dispensed'))),
  ),
  AdminOp(
    'vendRefundOrder',
    resource: 'order',
    verb: 'refund',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.vending.vendRefundOrder(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
];
