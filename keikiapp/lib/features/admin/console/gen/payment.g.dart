// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final paymentAdminOps = <AdminOp>[
  AdminOp(
    'payments',
    resource: 'payment',
    verb: '',
    fields: [
      KField('provider', 'Provider'),
      KField('status', 'Status'),
      KField('from', 'From', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.payment.payments(provider: v.str('provider'), status: v.str('status'), from: v.str('from'), to: v.str('to'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'payment',
    resource: 'payment',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.payment.payment(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'paymentTransactions',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.payment.paymentTransactions(id: v.need(v.integer('id'), 'id'))),
  ),
  AdminOp(
    'paymentRefundPayment',
    resource: 'payment',
    verb: 'refund',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.payment.paymentRefundPayment(id: v.need(v.str('id'), 'id'), amount: v.integer('amount'))),
  ),
];
