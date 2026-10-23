// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final arrangeAdminOps = <AdminOp>[
  AdminOp(
    'arrangeDispatchAssignments',
    resource: 'dispatch assignment',
    verb: '',
    fields: [
      KField('driverId', 'Driver ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeDispatchAssignments(driverId: v.need(v.integer('driverId'), 'driverId'))),
  ),
  AdminOp(
    'arrangeShifts',
    resource: 'shift',
    verb: '',
    fields: [
      KField('storeId', 'Store ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeShifts(storeId: v.need(v.integer('storeId'), 'storeId'))),
  ),
  AdminOp(
    'arrangeSpaces',
    resource: 'space',
    verb: '',
    fields: [
      KField('storeId', 'Store ID', kind: FieldKind.number, required: true),
      KField('minCapacity', 'Min capacity', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeSpaces(storeId: v.need(v.integer('storeId'), 'storeId'), minCapacity: v.integer('minCapacity'))),
  ),
  AdminOp(
    'arrangeSpaceAssignments',
    resource: 'space',
    verb: '',
    fields: [
      KField('storeId', 'Store ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeSpaceAssignments(storeId: v.need(v.integer('storeId'), 'storeId'))),
  ),
  AdminOp(
    'arrangeDispatch',
    resource: 'general',
    verb: 'dispatch',
    mutation: true,
    fields: [
      KField('input.orderId', 'Order ID', kind: FieldKind.number),
      KField('input.driverId', 'Driver ID', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeDispatch(input: ArrangeDispatchInput(orderId: v.integer('input.orderId'), driverId: v.integer('input.driverId')))),
  ),
  AdminOp(
    'arrangeMarkDelivered',
    resource: 'delivered',
    verb: 'mark',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeMarkDelivered(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'arrangeShift',
    resource: 'shift',
    verb: '',
    mutation: true,
    fields: [
      KField('input.employeeId', 'Employee ID', kind: FieldKind.number),
      KField('input.storeId', 'Store ID', kind: FieldKind.number),
      KField('input.role', 'Role'),
      KField('input.startTime', 'Start time', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.endTime', 'End time', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeShift(input: ArrangeShiftInput(employeeId: v.integer('input.employeeId'), storeId: v.integer('input.storeId'), role: v.str('input.role'), startTime: v.str('input.startTime'), endTime: v.str('input.endTime')))),
  ),
  AdminOp(
    'arrangeCancelShift',
    resource: 'shift',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeCancelShift(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'arrangeSpace',
    resource: 'space',
    verb: '',
    mutation: true,
    fields: [
      KField('input.storeId', 'Store ID', kind: FieldKind.number),
      KField('input.referenceId', 'Reference ID', kind: FieldKind.number),
      KField('input.partySize', 'Party size', kind: FieldKind.number),
      KField('input.from', 'From', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.to', 'To', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeSpace(input: ArrangeSpaceInput(storeId: v.integer('input.storeId'), referenceId: v.integer('input.referenceId'), partySize: v.integer('input.partySize'), from: v.str('input.from'), to: v.str('input.to')))),
  ),
  AdminOp(
    'arrangeReleaseSpace',
    resource: 'space',
    verb: 'release',
    mutation: true,
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.arrange.arrangeReleaseSpace(id: v.need(v.integer('id'), 'id'))),
  ),
];
