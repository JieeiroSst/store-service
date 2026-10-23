// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final payerAdminOps = <AdminOp>[
  AdminOp(
    'payerTransaction',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('transaction_id', 'Transaction ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.payer.payerTransaction(transactionId: v.need(v.integer('transaction_id'), 'transaction_id'))),
  ),
];
