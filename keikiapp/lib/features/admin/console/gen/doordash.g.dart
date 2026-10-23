// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final doordashAdminOps = <AdminOp>[
  AdminOp(
    'doordashOrders',
    resource: 'order',
    verb: '',
    fields: [
      KField('customer_id', 'Customer ID'),
      KField('restaurant_id', 'Restaurant ID'),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashOrders(customerId: v.str('customer_id'), restaurantId: v.str('restaurant_id'))),
  ),
  AdminOp(
    'doordashOrder',
    resource: 'order',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashOrder(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'doordashOrderTracking',
    resource: 'order',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashOrderTracking(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'doordashDriverAssignments',
    resource: 'driver',
    verb: '',
    fields: [
      KField('driver_id', 'Driver ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashDriverAssignments(driverId: v.need(v.str('driver_id'), 'driver_id'))),
  ),
  AdminOp(
    'doordashUpdateOrderStatus',
    resource: 'order',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('status', 'Status', required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashUpdateOrderStatus(id: v.need(v.integer('id'), 'id'), status: v.need(v.str('status'), 'status'))),
  ),
  AdminOp(
    'doordashCancelOrder',
    resource: 'order',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashCancelOrder(id: v.need(v.integer('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'doordashAssignDriver',
    resource: 'driver',
    verb: 'assign',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('driver_id', 'Driver ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashAssignDriver(id: v.need(v.integer('id'), 'id'), driverId: v.need(v.str('driver_id'), 'driver_id'))),
  ),
  AdminOp(
    'doordashAddTrackingEvent',
    resource: 'tracking event',
    verb: 'add',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('input.status', 'Status'),
      KField('input.latitude', 'Latitude', kind: FieldKind.number),
      KField('input.longitude', 'Longitude', kind: FieldKind.number),
      KField('input.notes', 'Notes', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashAddTrackingEvent(id: v.need(v.integer('id'), 'id'), input: DoordashTrackingInput(status: v.str('input.status'), latitude: v.number('input.latitude'), longitude: v.number('input.longitude'), notes: v.str('input.notes')))),
  ),
  AdminOp(
    'doordashAcceptAssignment',
    resource: 'accept assignment',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashAcceptAssignment(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'doordashRejectAssignment',
    resource: 'assignment',
    verb: 'reject',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashRejectAssignment(id: v.need(v.integer('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'doordashCompleteAssignment',
    resource: 'assignment',
    verb: 'complete',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.doordash.doordashCompleteAssignment(id: v.need(v.integer('id'), 'id'))),
  ),
];
