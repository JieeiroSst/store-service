// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final sardCustomerAdminOps = <AdminOp>[
  AdminOp(
    'sardCustomer',
    resource: 'customer',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardCustomer(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardCustomerByUser',
    resource: 'customer',
    verb: '',
    fields: [
      KField('userId', 'User ID', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardCustomerByUser(userId: v.need(v.integer('userId'), 'userId'))),
  ),
  AdminOp(
    'sardOnboardCustomer',
    resource: 'onboard customer',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardOnboardCustomer(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardSyncCustomer',
    resource: 'customer',
    verb: 'sync',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardSyncCustomer(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardSubmitKyc',
    resource: 'kyc',
    verb: 'submit',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardSubmitKyc(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardRefreshKyc',
    resource: 'kyc',
    verb: 'refresh',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCustomer.sardRefreshKyc(id: v.need(v.str('id'), 'id'))),
  ),
];
