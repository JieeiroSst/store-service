// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final paymentWalletAdminOps = <AdminOp>[
  AdminOp(
    'wallet',
    resource: 'wallet',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.wallet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletByUser',
    resource: 'general',
    verb: '',
    fields: [
      KField('userId', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletByUser(userId: v.need(v.str('userId'), 'userId'))),
  ),
  AdminOp(
    'walletTransactions',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('limit', 'Limit', kind: FieldKind.number),
      KField('offset', 'Offset', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletTransactions(id: v.need(v.str('id'), 'id'), limit: v.integer('limit'), offset: v.integer('offset'))),
  ),
  AdminOp(
    'walletTransactionByReference',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
      KField('referenceId', 'Reference ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletTransactionByReference(id: v.need(v.str('id'), 'id'), referenceId: v.need(v.str('referenceId'), 'referenceId'))),
  ),
  AdminOp(
    'walletPockets',
    resource: 'pocket',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPockets(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletTransfer',
    resource: 'transfer',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletTransfer(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletTransaction',
    resource: 'transaction',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletTransaction(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletPaymentRequests',
    resource: 'payment request',
    verb: '',
    fields: [
      KField('wallet_id', 'Wallet ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPaymentRequests(walletId: v.need(v.str('wallet_id'), 'wallet_id'))),
  ),
  AdminOp(
    'walletPaymentRequest',
    resource: 'payment request',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPaymentRequest(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletPaymentRequestQR',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPaymentRequestQR(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletPaymentMethods',
    resource: 'payment method',
    verb: '',
    fields: [
      KField('user_id', 'User ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPaymentMethods(userId: v.need(v.str('user_id'), 'user_id'))),
  ),
  AdminOp(
    'walletCreateWallet',
    resource: 'wallet',
    verb: 'create',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('currency', 'Currency', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletCreateWallet(userId: v.need(v.str('user_id'), 'user_id'), currency: v.need(v.str('currency'), 'currency'))),
  ),
  AdminOp(
    'walletDeposit',
    resource: 'deposit',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('reference_id', 'Reference ID', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletDeposit(id: v.need(v.str('id'), 'id'), amount: v.need(v.integer('amount'), 'amount'), referenceId: v.need(v.str('reference_id'), 'reference_id'), description: v.str('description'))),
  ),
  AdminOp(
    'walletWithdraw',
    resource: 'withdraw',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('reference_id', 'Reference ID', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletWithdraw(id: v.need(v.str('id'), 'id'), amount: v.need(v.integer('amount'), 'amount'), referenceId: v.need(v.str('reference_id'), 'reference_id'), description: v.str('description'))),
  ),
  AdminOp(
    'walletFreeze',
    resource: 'general',
    verb: 'freeze',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletFreeze(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'walletUnfreeze',
    resource: 'general',
    verb: 'unfreeze',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletUnfreeze(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletClose',
    resource: 'general',
    verb: 'close',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletClose(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSetLimits',
    resource: 'general',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('daily_limit', 'Daily limit', kind: FieldKind.number, required: true),
      KField('per_transaction_limit', 'Per transaction limit', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletSetLimits(id: v.need(v.str('id'), 'id'), dailyLimit: v.need(v.integer('daily_limit'), 'daily_limit'), perTransactionLimit: v.need(v.integer('per_transaction_limit'), 'per_transaction_limit'))),
  ),
  AdminOp(
    'walletCreatePocket',
    resource: 'pocket',
    verb: 'create',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('name', 'Name', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletCreatePocket(id: v.need(v.str('id'), 'id'), name_: v.need(v.str('name'), 'name'))),
  ),
  AdminOp(
    'walletDepositToPocket',
    resource: 'deposit',
    verb: '',
    mutation: true,
    fields: [
      KField('pocketId', 'Pocket ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletDepositToPocket(pocketId: v.need(v.str('pocketId'), 'pocketId'), amount: v.need(v.integer('amount'), 'amount'))),
  ),
  AdminOp(
    'walletWithdrawFromPocket',
    resource: 'withdraw',
    verb: '',
    mutation: true,
    fields: [
      KField('pocketId', 'Pocket ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletWithdrawFromPocket(pocketId: v.need(v.str('pocketId'), 'pocketId'), amount: v.need(v.integer('amount'), 'amount'))),
  ),
  AdminOp(
    'walletClosePocket',
    resource: 'pocket',
    verb: 'close',
    mutation: true,
    fields: [
      KField('pocketId', 'Pocket ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletClosePocket(pocketId: v.need(v.str('pocketId'), 'pocketId'))),
  ),
  AdminOp(
    'walletTransfer',
    resource: 'transfer',
    verb: '',
    mutation: true,
    fields: [
      KField('sender_wallet_id', 'Sender wallet ID', required: true),
      KField('receiver_wallet_id', 'Receiver wallet ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('reference_id', 'Reference ID', required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletTransferMutation(senderWalletId: v.need(v.str('sender_wallet_id'), 'sender_wallet_id'), receiverWalletId: v.need(v.str('receiver_wallet_id'), 'receiver_wallet_id'), amount: v.need(v.integer('amount'), 'amount'), referenceId: v.need(v.str('reference_id'), 'reference_id'), description: v.str('description'))),
  ),
  AdminOp(
    'walletReverseTransfer',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletReverseTransfer(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'walletReverseTransaction',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('reason', 'Reason', kind: FieldKind.multiline),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletReverseTransaction(id: v.need(v.str('id'), 'id'), reason: v.str('reason'))),
  ),
  AdminOp(
    'walletCreatePaymentRequest',
    resource: 'payment request',
    verb: 'create',
    mutation: true,
    fields: [
      KField('requester_wallet_id', 'Requester wallet ID', required: true),
      KField('payer_wallet_id', 'Payer wallet ID'),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('description', 'Description', kind: FieldKind.multiline),
      KField('expires_in_seconds', 'Expires in seconds', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletCreatePaymentRequest(requesterWalletId: v.need(v.str('requester_wallet_id'), 'requester_wallet_id'), payerWalletId: v.str('payer_wallet_id'), amount: v.need(v.integer('amount'), 'amount'), description: v.str('description'), expiresInSeconds: v.integer('expires_in_seconds'))),
  ),
  AdminOp(
    'walletPayPaymentRequest',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
      KField('payer_wallet_id', 'Payer wallet ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletPayPaymentRequest(id: v.need(v.str('id'), 'id'), payerWalletId: v.need(v.str('payer_wallet_id'), 'payer_wallet_id'))),
  ),
  AdminOp(
    'walletCancelPaymentRequest',
    resource: 'payment request',
    verb: 'cancel',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletCancelPaymentRequest(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletAddPaymentMethod',
    resource: 'payment method',
    verb: 'add',
    mutation: true,
    fields: [
      KField('user_id', 'User ID', required: true),
      KField('type', 'Type', required: true),
      KField('provider', 'Provider', required: true),
      KField('account_number', 'Account number', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletAddPaymentMethod(userId: v.need(v.str('user_id'), 'user_id'), type_: v.need(v.str('type'), 'type'), provider: v.need(v.str('provider'), 'provider'), accountNumber: v.need(v.str('account_number'), 'account_number'))),
  ),
  AdminOp(
    'walletRemovePaymentMethod',
    resource: 'payment method',
    verb: 'remove',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletRemovePaymentMethod(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'walletSetDefaultPaymentMethod',
    resource: 'general',
    verb: 'set',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.paymentWallet.walletSetDefaultPaymentMethod(id: v.need(v.str('id'), 'id'))),
  ),
];
