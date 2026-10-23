// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final accountTransactionAdminOps = <AdminOp>[
  AdminOp(
    'accountTxGetAccount',
    resource: 'account',
    verb: 'get',
    fields: [
      KField('input.id', 'ID'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxGetAccount(input: AccountTxGetAccountRequestInput(id: v.str('input.id')))),
  ),
  AdminOp(
    'accountTxListAccounts',
    resource: 'account',
    verb: 'list',
    fields: [
      KField('input.pageSize', 'Page size', kind: FieldKind.number),
      KField('input.pageToken', 'Page token'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxListAccounts(input: AccountTxListAccountsRequestInput(pageSize: v.integer('input.pageSize'), pageToken: v.str('input.pageToken')))),
  ),
  AdminOp(
    'accountTxGetTransaction',
    resource: 'transaction',
    verb: 'get',
    fields: [
      KField('input.id', 'ID'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxGetTransaction(input: AccountTxGetTransactionRequestInput(id: v.str('input.id')))),
  ),
  AdminOp(
    'accountTxListTransactions',
    resource: 'transaction',
    verb: 'list',
    fields: [
      KField('input.pageSize', 'Page size', kind: FieldKind.number),
      KField('input.pageToken', 'Page token'),
      KField('input.type', 'Type'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxListTransactions(input: AccountTxListTransactionsRequestInput(pageSize: v.integer('input.pageSize'), pageToken: v.str('input.pageToken'), type_: v.str('input.type')))),
  ),
  AdminOp(
    'accountTxGetAccountTransactions',
    resource: 'account',
    verb: 'get',
    fields: [
      KField('input.accountId', 'Account ID'),
      KField('input.pageSize', 'Page size', kind: FieldKind.number),
      KField('input.pageToken', 'Page token'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxGetAccountTransactions(input: AccountTxGetAccountTransactionsRequestInput(accountId: v.str('input.accountId'), pageSize: v.integer('input.pageSize'), pageToken: v.str('input.pageToken')))),
  ),
  AdminOp(
    'accountTxCreateAccount',
    resource: 'account',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.firstName', 'First name'),
      KField('input.lastName', 'Last name'),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxCreateAccount(input: AccountTxCreateAccountRequestInput(firstName: v.str('input.firstName'), lastName: v.str('input.lastName')))),
  ),
  AdminOp(
    'accountTxCreateDepositTransaction',
    resource: 'deposit transaction',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.accountId', 'Account ID'),
      KField('input.amount', 'Amount', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxCreateDepositTransaction(input: AccountTxCreateDepositTransactionRequestInput(accountId: v.str('input.accountId'), amount: v.number('input.amount')))),
  ),
  AdminOp(
    'accountTxCreateWithdrawalTransaction',
    resource: 'withdrawal transaction',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.accountId', 'Account ID'),
      KField('input.amount', 'Amount', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxCreateWithdrawalTransaction(input: AccountTxCreateWithdrawalTransactionRequestInput(accountId: v.str('input.accountId'), amount: v.number('input.amount')))),
  ),
  AdminOp(
    'accountTxCreateAccountToAccountTransaction',
    resource: 'account',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.senderId', 'Sender ID'),
      KField('input.receiverId', 'Receiver ID'),
      KField('input.amount', 'Amount', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxCreateAccountToAccountTransaction(input: AccountTxCreateAccountToAccountTransactionRequestInput(senderId: v.str('input.senderId'), receiverId: v.str('input.receiverId'), amount: v.number('input.amount')))),
  ),
  AdminOp(
    'accountTxCreatePaymentForServiceTransaction',
    resource: 'payment',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input.accountId', 'Account ID'),
      KField('input.serviceName', 'Service name'),
      KField('input.amount', 'Amount', kind: FieldKind.number),
    ],
    run: (api, v) async => adminPlain(await api.accountTransaction.accountTxCreatePaymentForServiceTransaction(input: AccountTxCreatePaymentForServiceTransactionRequestInput(accountId: v.str('input.accountId'), serviceName: v.str('input.serviceName'), amount: v.number('input.amount')))),
  ),
];
