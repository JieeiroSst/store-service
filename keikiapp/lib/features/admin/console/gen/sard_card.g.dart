// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final sardCardAdminOps = <AdminOp>[
  AdminOp(
    'sardPrograms',
    resource: 'general',
    verb: '',
    run: (api, v) async => adminPlain(await api.sardCard.sardPrograms()),
  ),
  AdminOp(
    'sardAccount',
    resource: 'account',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAccount(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardCustomerAccounts',
    resource: 'general',
    verb: '',
    fields: [
      KField('customerId', 'Customer ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCustomerAccounts(customerId: v.need(v.str('customerId'), 'customerId'))),
  ),
  AdminOp(
    'sardAccountTransactions',
    resource: 'account',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAccountTransactions(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'sardAccountCards',
    resource: 'account',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAccountCards(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardCard',
    resource: 'card',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCard(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardCardAuthorizations',
    resource: 'card',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('query', 'Query', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCardAuthorizations(id: v.need(v.str('id'), 'id'), query: v.jsonMap('query'))),
  ),
  AdminOp(
    'sardAuthorization',
    resource: 'authorization',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAuthorization(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'sardOpenAccount',
    resource: 'account',
    verb: 'open',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardOpenAccount(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardBlockAccount',
    resource: 'account',
    verb: 'block',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardBlockAccount(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'sardUnblockAccount',
    resource: 'account',
    verb: 'unblock',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardUnblockAccount(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'sardCancelAccount',
    resource: 'account',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCancelAccount(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'sardSetCreditLimit',
    resource: 'general',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardSetCreditLimit(id: v.need(v.str('id'), 'id'), input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardAccountPayment',
    resource: 'account',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAccountPayment(id: v.need(v.str('id'), 'id'), input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardIssueCard',
    resource: 'card',
    verb: 'issue',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardIssueCard(id: v.need(v.str('id'), 'id'), input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardActivateCard',
    resource: 'card',
    verb: 'activate',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardActivateCard(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardBlockCard',
    resource: 'card',
    verb: 'block',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardBlockCard(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardUnblockCard',
    resource: 'card',
    verb: 'unblock',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardUnblockCard(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardCancelCard',
    resource: 'card',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCancelCard(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardReportCard',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardReportCard(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardUpdateCardLimits',
    resource: 'card',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('per_transaction', 'Per transaction', kind: FieldKind.number),
      KField('daily', 'Daily', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardUpdateCardLimits(id: v.need(v.str('id'), 'id'), perTransaction: v.integer('per_transaction'), daily: v.integer('daily'))),
  ),
  AdminOp(
    'sardUpdateCardControls',
    resource: 'card',
    verb: 'update',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('pos', 'Pos', kind: FieldKind.toggle),
      KField('contactless', 'Contactless', kind: FieldKind.toggle),
      KField('ecommerce', 'Ecommerce', kind: FieldKind.toggle),
      KField('atm', 'Atm', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardUpdateCardControls(id: v.need(v.str('id'), 'id'), pos: v.optFlag('pos'), contactless: v.optFlag('contactless'), ecommerce: v.optFlag('ecommerce'), atm: v.optFlag('atm'))),
  ),
  AdminOp(
    'sardSetCardPin',
    resource: 'card',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardSetCardPin(id: v.need(v.str('id'), 'id'), input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardAuthorize',
    resource: 'general',
    verb: 'authorize',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardAuthorize(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'sardConfirmAuthorization',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardConfirmAuthorization(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
  AdminOp(
    'sardCancelAuthorization',
    resource: 'authorization',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('input', 'Input', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.sardCard.sardCancelAuthorization(id: v.need(v.str('id'), 'id'), input: v.jsonMap('input'))),
  ),
];
