// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final sardAuthAdminOps = <AdminOp>[
  AdminOp(
    'sardPaymentAuthentication',
    resource: 'payment authentication',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardAuth.sardPaymentAuthentication(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardInitiatePaymentAuthentication',
    resource: 'initiate payment authentication',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardAuth.sardInitiatePaymentAuthentication(input: v.need(v.jsonMap('input'), 'input'))),
  ),
];
