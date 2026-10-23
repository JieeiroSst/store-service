// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final automaticPaymentAdminOps = <AdminOp>[
  AdminOp(
    'autoPaySubscription',
    resource: 'subscription',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPaySubscription(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPaySubscriptionTransactions',
    resource: 'subscription',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPaySubscriptionTransactions(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPaySubscriptionInvoices',
    resource: 'subscription',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPaySubscriptionInvoices(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPayPaymentMethods',
    resource: 'payment method',
    verb: '',
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayPaymentMethods(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'autoPayInvoices',
    resource: 'invoice',
    verb: '',
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayInvoices(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'autoPayInvoice',
    resource: 'invoice',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayInvoice(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPayCreateSubscription',
    resource: 'subscription',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.userId', 'User ID'),
      KField('input.planId', 'Plan ID'),
      KField('input.amount', 'Amount', kind: FieldKind.number),
      KField('input.currency', 'Currency'),
      KField('input.autoRenewal', 'Auto renewal', kind: FieldKind.toggle),
      KField('input.trialDays', 'Trial days', kind: FieldKind.number),
      KField('input.paymentMethodId', 'Payment method ID'),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayCreateSubscription(input: AutoPaySubscriptionInput(userId: v.str('input.userId'), planId: v.str('input.planId'), amount: v.number('input.amount'), currency: v.str('input.currency'), autoRenewal: v.optFlag('input.autoRenewal'), trialDays: v.integer('input.trialDays'), paymentMethodId: v.str('input.paymentMethodId')))),
  ),
  AdminOp(
    'autoPayCancelSubscription',
    resource: 'subscription',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayCancelSubscription(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPayProcessRenewals',
    resource: 'renewal',
    verb: 'process',
    mutation: true,
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayProcessRenewals()),
  ),
  AdminOp(
    'autoPayAddPaymentMethod',
    resource: 'payment method',
    verb: 'add',
    mutation: true,
    fields: [
      KField('input.userId', 'User ID'),
      KField('input.provider', 'Provider'),
      KField('input.tokenId', 'Token ID'),
      KField('input.lastFourDigits', 'Last four digits'),
      KField('input.expiryDate', 'Expiry date', hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('input.isDefault', 'Is default', kind: FieldKind.toggle),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayAddPaymentMethod(input: AutoPayPaymentMethodInput(userId: v.str('input.userId'), provider: v.str('input.provider'), tokenId: v.str('input.tokenId'), lastFourDigits: v.str('input.lastFourDigits'), expiryDate: v.str('input.expiryDate'), isDefault: v.optFlag('input.isDefault')))),
  ),
  AdminOp(
    'autoPayDeletePaymentMethod',
    resource: 'payment method',
    verb: 'delete',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPayDeletePaymentMethod(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'autoPaySetDefaultPaymentMethod',
    resource: 'default payment method',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.automaticPayment.autoPaySetDefaultPaymentMethod(id: v.need(v.str('id'), 'id'), userId: v.need(v.str('userId'), 'userId'))),
  ),
];
