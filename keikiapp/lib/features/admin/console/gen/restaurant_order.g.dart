// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final restaurantOrderAdminOps = <AdminOp>[
  AdminOp(
    'rOrderOrder',
    resource: 'order',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.restaurantOrder.rOrderOrder(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'rOrderOrders',
    resource: 'order',
    verb: '',
    fields: [
      KField('Limit', 'Limit', kind: FieldKind.number),
      KField('Page', 'Page', kind: FieldKind.number),
      KField('Sort', 'Sort'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantOrder.rOrderOrders(limit: v.integer('Limit'), page: v.integer('Page'), sort: v.str('Sort'))),
  ),
  AdminOp(
    'rOrderReservations',
    resource: 'reservation',
    verb: '',
    fields: [
      KField('Limit', 'Limit', kind: FieldKind.number),
      KField('Page', 'Page', kind: FieldKind.number),
      KField('Sort', 'Sort'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantOrder.rOrderReservations(limit: v.integer('Limit'), page: v.integer('Page'), sort: v.str('Sort'))),
  ),
  AdminOp(
    'rOrderReservation',
    resource: 'reservation',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.restaurantOrder.rOrderReservation(id: v.need(v.integer('id'), 'id'))),
  ),
];
