// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

// ignore_for_file: unused_import

import '../../../apps/kit/kit.dart';
import '../../../../gen/models.g.dart';
import '../admin_op.dart';

final solanaAdminOps = <AdminOp>[
  AdminOp(
    'solAccount',
    resource: 'general',
    verb: '',
    fields: [
      KField('address', 'Address', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solAccount(address: v.need(v.str('address'), 'address'))),
  ),
  AdminOp(
    'solBalance',
    resource: 'general',
    verb: '',
    fields: [
      KField('address', 'Address', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solBalance(address: v.need(v.str('address'), 'address'))),
  ),
  AdminOp(
    'solTransaction',
    resource: 'general',
    verb: '',
    fields: [
      KField('signature', 'Signature', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solTransaction(signature: v.need(v.str('signature'), 'signature'))),
  ),
  AdminOp(
    'solProgram',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solProgram(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleWalletSet',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleWalletSet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleWalletSets',
    resource: 'circle',
    verb: '',
    run: (api, v) async => adminPlain(await api.solana.solCircleWalletSets()),
  ),
  AdminOp(
    'solCircleWallet',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleWallet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleWallets',
    resource: 'circle',
    verb: '',
    fields: [
      KField('walletSetId', 'Wallet set ID'),
      KField('blockchain', 'Blockchain'),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleWallets(walletSetId: v.str('walletSetId'), blockchain: v.str('blockchain'))),
  ),
  AdminOp(
    'solCircleWalletBalance',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleWalletBalance(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleWalletNFTs',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleWalletNFTs(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleTransaction',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleTransaction(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleEstimateFee',
    resource: 'circle',
    verb: '',
    fields: [
      KField('walletId', 'Wallet ID', required: true),
      KField('destinationAddress', 'Destination address', required: true),
      KField('tokenId', 'Token ID', required: true),
      KField('amount', 'Amount', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleEstimateFee(walletId: v.need(v.str('walletId'), 'walletId'), destinationAddress: v.need(v.str('destinationAddress'), 'destinationAddress'), tokenId: v.need(v.str('tokenId'), 'tokenId'), amount: v.need(v.str('amount'), 'amount'))),
  ),
  AdminOp(
    'solCircleToken',
    resource: 'circle',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleToken(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solCircleTokens',
    resource: 'circle',
    verb: '',
    fields: [
      KField('blockchain', 'Blockchain'),
    ],
    run: (api, v) async => adminPlain(await api.solana.solCircleTokens(blockchain: v.str('blockchain'))),
  ),
  AdminOp(
    'solBridgeTransfer',
    resource: 'general',
    verb: '',
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solBridgeTransfer(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'solEstimateTransferFee',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('from', 'From', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('to', 'To', required: true, hint: 'ISO 8601, ví dụ 2026-10-22T09:00:00Z'),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solEstimateTransferFee(from: v.need(v.str('from'), 'from'), to: v.need(v.str('to'), 'to'), amount: v.need(v.integer('amount'), 'amount'))),
  ),
  AdminOp(
    'solFindPda',
    resource: 'general',
    verb: 'find',
    mutation: true,
    fields: [
      KField('program_id', 'Program ID', required: true),
      KField('user_pubkey', 'User pubkey', required: true),
      KField('identifier', 'Identifier', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.solFindPda(programId: v.need(v.str('program_id'), 'program_id'), userPubkey: v.need(v.str('user_pubkey'), 'user_pubkey'), identifier: v.need(v.str('identifier'), 'identifier'))),
  ),
  AdminOp(
    'circleCreateWalletSet',
    resource: 'wallet',
    verb: 'create',
    mutation: true,
    fields: [
      KField('name', 'Name', required: true),
      KField('custodyType', 'Custody type'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleCreateWalletSet(name_: v.need(v.str('name'), 'name'), custodyType: v.str('custodyType'))),
  ),
  AdminOp(
    'circleCreateWallet',
    resource: 'wallet',
    verb: 'create',
    mutation: true,
    fields: [
      KField('walletSetId', 'Wallet set ID', required: true),
      KField('blockchain', 'Blockchain', required: true),
      KField('accountType', 'Account type'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleCreateWallet(walletSetId: v.need(v.str('walletSetId'), 'walletSetId'), blockchain: v.need(v.str('blockchain'), 'blockchain'), accountType: v.str('accountType'))),
  ),
  AdminOp(
    'circleFreezeWallet',
    resource: 'wallet',
    verb: 'freeze',
    mutation: true,
    destructive: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleFreezeWallet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'circleUnfreezeWallet',
    resource: 'wallet',
    verb: 'unfreeze',
    mutation: true,
    fields: [
      KField('id', 'ID', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleUnfreezeWallet(id: v.need(v.str('id'), 'id'))),
  ),
  AdminOp(
    'circleCreateTransfer',
    resource: 'general',
    verb: 'create',
    mutation: true,
    fields: [
      KField('wallet_id', 'Wallet ID', required: true),
      KField('destination_address', 'Destination address', required: true),
      KField('token_id', 'Token ID', required: true),
      KField('amount', 'Amount', required: true),
      KField('fee.type', 'Type'),
      KField('fee.config', 'Config', kind: FieldKind.multiline, hint: 'JSON, ví dụ {"key": "value"}'),
      KField('idempotency_key', 'Idempotency key'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleCreateTransfer(walletId: v.need(v.str('wallet_id'), 'wallet_id'), destinationAddress: v.need(v.str('destination_address'), 'destination_address'), tokenId: v.need(v.str('token_id'), 'token_id'), amount: v.need(v.str('amount'), 'amount'), fee: v.any('fee.') ? SolFeeInput(type_: v.str('fee.type'), config: v.jsonMap('fee.config')) : null, idempotencyKey: v.str('idempotency_key'))),
  ),
  AdminOp(
    'circleCreateNftTransfer',
    resource: 'general',
    verb: 'create',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleCreateNftTransfer(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'circleExecuteContract',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleExecuteContract(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'circleImportToken',
    resource: 'general',
    verb: 'import',
    mutation: true,
    fields: [
      KField('input', 'Input', kind: FieldKind.multiline, required: true, hint: 'JSON, ví dụ {"key": "value"}'),
    ],
    run: (api, v) async => adminPlain(await api.solana.circleImportToken(input: v.need(v.jsonMap('input'), 'input'))),
  ),
  AdminOp(
    'bridgeSolanaToCircle',
    resource: 'general',
    verb: '',
    mutation: true,
    fields: [
      KField('fromSolanaAddress', 'From solana address', required: true),
      KField('toCircleWalletId', 'To circle wallet ID', required: true),
      KField('amount', 'Amount', kind: FieldKind.number, required: true),
      KField('privateKey', 'Private key', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.bridgeSolanaToCircle(fromSolanaAddress: v.need(v.str('fromSolanaAddress'), 'fromSolanaAddress'), toCircleWalletId: v.need(v.str('toCircleWalletId'), 'toCircleWalletId'), amount: v.need(v.integer('amount'), 'amount'), privateKey: v.need(v.str('privateKey'), 'privateKey'))),
  ),
  AdminOp(
    'bridgeCircleToSolana',
    resource: 'circle',
    verb: '',
    mutation: true,
    fields: [
      KField('fromCircleWalletId', 'From circle wallet ID', required: true),
      KField('toSolanaAddress', 'To solana address', required: true),
      KField('amount', 'Amount', required: true),
    ],
    run: (api, v) async => adminPlain(await api.solana.bridgeCircleToSolana(fromCircleWalletId: v.need(v.str('fromCircleWalletId'), 'fromCircleWalletId'), toSolanaAddress: v.need(v.str('toSolanaAddress'), 'toSolanaAddress'), amount: v.need(v.str('amount'), 'amount'))),
  ),
];
