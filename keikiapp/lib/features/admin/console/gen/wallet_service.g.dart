// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final walletServiceAdminOps = <AdminOp>[
  AdminOp(
    'walletSvcWallet',
    resource: 'wallet',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcWallet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcTransaction',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcTransaction(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcUserWallets',
    resource: 'general',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcUserWallets(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'walletSvcBalance',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcBalance(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcTransactions',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('type', 'Type'),
      KField('status', 'Status'),
      KField('from', 'From', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('page', 'Page', kind: FieldKind.number),
      KField('page_size', 'Page size', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcTransactions(id: v.need(v.str('id'), 'id'), type_: v.str('type'), status: v.str('status'), from: v.str('from'), to: v.str('to'), page: v.integer('page'), pageSize: v.integer('page_size'))),
  ),
  AdminOp(
    'walletSvcStatement',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('from', 'From', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcStatement(id: v.need(v.str('id'), 'id'), from: v.need(v.str('from'), 'from'), to: v.need(v.str('to'), 'to'))),
  ),
  AdminOp(
    'walletSvcCard',
    resource: 'card',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcCard(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcPaymentRequest',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcPaymentRequest(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcCreateWallet',
    resource: 'wallet',
    verb: 'create',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('currency', 'Currency', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcCreateWallet(userId: v.need(v.str('user_id'), 'user_id'), currency: v.need(v.str('currency'), 'currency'))),
  ),
  AdminOp(
    'walletSvcCredit',
    resource: 'general',
    verb: 'credit',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('amount', 'Amount', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcCredit(id: v.need(v.str('id'), 'id'), amount: v.need(v.str('amount'), 'amount'), description: v.str('description'))),
  ),
  AdminOp(
    'walletSvcFreeze',
    resource: 'general',
    verb: 'freeze',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcFreeze(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcAuthorize',
    resource: 'general',
    verb: 'authorize',
    mutation: true,
    fields: [
      KField('idempotency_key', 'Idempotency key', required: true),
      KField('card_id', 'Card ID', required: true),
      KField('merchant_id', 'Merchant ID', required: true),
      KField('amount', 'Amount', required: true),
      KField('currency', 'Currency', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
      KField('metadata', 'Metadata', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcAuthorize(idempotencyKey: v.need(v.str('idempotency_key'), 'idempotency_key'), cardId: v.need(v.str('card_id'), 'card_id'), merchantId: v.need(v.str('merchant_id'), 'merchant_id'), amount: v.need(v.str('amount'), 'amount'), currency: v.need(v.str('currency'), 'currency'), description: v.str('description'), metadata: v.jsonMap('metadata'))),
  ),
  AdminOp(
    'walletSvcCapture',
    resource: 'general',
    verb: 'capture',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcCapture(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcVoid',
    resource: 'general',
    verb: 'void',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcVoid(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcCreateSettlementBatch',
    resource: 'settlement',
    verb: 'create',
    mutation: true,
    fields: [
      KField('merchant_id', 'Merchant ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcCreateSettlementBatch(merchantId: v.need(v.str('merchant_id'), 'merchant_id'))),
  ),
  AdminOp(
    'walletSvcProcessClearing',
    resource: 'general',
    verb: 'process',
    mutation: true,
    fields: [
      KField('batch_id', 'Batch ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcProcessClearing(batchId: v.need(v.str('batch_id'), 'batch_id'))),
  ),
  AdminOp(
    'walletSvcProcessSettlement',
    resource: 'settlement',
    verb: 'process',
    mutation: true,
    fields: [
      KField('batch_id', 'Batch ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcProcessSettlement(batchId: v.need(v.str('batch_id'), 'batch_id'))),
  ),
  AdminOp(
    'walletSvcUnfreeze',
    resource: 'general',
    verb: 'unfreeze',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcUnfreeze(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcChangeTier',
    resource: 'general',
    verb: 'change',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('tier', 'Tier', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcChangeTier(id: v.need(v.str('id'), 'id'), tier: v.need(v.str('tier'), 'tier'))),
  ),
  AdminOp(
    'walletSvcRefund',
    resource: 'general',
    verb: 'refund',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('amount', 'Amount'),
      KField('reason', 'Reason', kind: FieldKind.multiline),
      KField('idempotency_key', 'Idempotency key', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcRefund(id: v.need(v.str('id'), 'id'), amount: v.str('amount'), reason: v.str('reason'), idempotencyKey: v.need(v.str('idempotency_key'), 'idempotency_key'))),
  ),
  AdminOp(
    'walletSvcBlockCard',
    resource: 'card',
    verb: 'block',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcBlockCard(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSvcUnblockCard',
    resource: 'card',
    verb: 'unblock',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.walletService.walletSvcUnblockCard(id: v.need(v.str('id'), 'id'))),
  ),
];
