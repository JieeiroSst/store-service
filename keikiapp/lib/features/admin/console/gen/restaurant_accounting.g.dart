// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final restaurantAccountingAdminOps = <AdminOp>[
  AdminOp(
    'rAccountingPayment',
    resource: 'payment',
    verb: '',
    fields: [
      KField('orderID', 'Order ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.restaurantAccounting.rAccountingPayment(orderID: v.need(v.integer('orderID'), 'orderID'))),
  ),
  AdminOp(
    'rAccountingAuthCart',
    resource: 'auth cart',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.restaurantAccounting.rAccountingAuthCart(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'rAccountingMarkPaid',
    resource: 'paid',
    verb: 'mark',
    mutation: true,
    fields: [
      KField('orderID', 'Order ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.restaurantAccounting.rAccountingMarkPaid(orderID: v.need(v.str('orderID'), 'orderID'))),
  ),
];
