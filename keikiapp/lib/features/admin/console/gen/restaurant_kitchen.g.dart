// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final restaurantKitchenAdminOps = <AdminOp>[
  AdminOp(
    'rKitchenCreateCategory',
    resource: 'category',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantKitchen.rKitchenCreateCategory(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'rKitchenCreateFood',
    resource: 'food',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantKitchen.rKitchenCreateFood(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'rKitchenUpdateStatus',
    resource: 'status',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('status', 'Status', required: true),
    ],
    run: (api, v) async => adminPlain(await api.restaurantKitchen.rKitchenUpdateStatus(id: v.need(v.str('id'), 'id'), status: v.need(v.str('status'), 'status'))),
  ),
];
