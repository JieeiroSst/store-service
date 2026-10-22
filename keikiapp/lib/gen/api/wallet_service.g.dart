// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class WalletServiceApi {
  const WalletServiceApi(this._executor);

  final GatewayExecutor _executor;

  Future<WalletSvcWallet?> walletSvcWallet({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcWallet(\$id: String!) { walletSvcWallet(id: \$id) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcWallet'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcWallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcTransaction({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcTransaction(\$id: String!) { walletSvcTransaction(id: \$id) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcTransaction'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcTransaction'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletSvcWallet>?> walletSvcMyWallets({String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyWallets { walletSvcMyWallets ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }'
    );
    return (data['walletSvcMyWallets'] as List?)?.map((e) => (e == null ? null : WalletSvcWallet.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletSvcWallet?> walletSvcMyWallet({required String currency, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyWallet(\$currency: String!) { walletSvcMyWallet(currency: \$currency) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'currency': currency,
    }
    );
    return data['walletSvcMyWallet'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcMyWallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcBalance?> walletSvcMyBalance({required String walletId, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyBalance(\$wallet_id: String!) { walletSvcMyBalance(wallet_id: \$wallet_id) ${selectionSet ?? '{ wallet_id currency balance held available status }'} }', variables: {
      'wallet_id': walletId,
    }
    );
    return data['walletSvcMyBalance'] == null ? null : WalletSvcBalance.fromJson((data['walletSvcMyBalance'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcLimits?> walletSvcMyLimits({required String walletId, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyLimits(\$wallet_id: String!) { walletSvcMyLimits(wallet_id: \$wallet_id) ${selectionSet ?? '{ wallet_id tier currency limits { per_transaction daily } used_today remaining_today resets_at }'} }', variables: {
      'wallet_id': walletId,
    }
    );
    return data['walletSvcMyLimits'] == null ? null : WalletSvcLimits.fromJson((data['walletSvcMyLimits'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransactionPage?> walletSvcMyTransactions({required String walletId, String? type_, String? status, String? from, String? to, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyTransactions(\$wallet_id: String!, \$type: String, \$status: String, \$from: String, \$to: String, \$page: Int, \$page_size: Int) { walletSvcMyTransactions(wallet_id: \$wallet_id, type: \$type, status: \$status, from: \$from, to: \$to, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ items { id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at } total page page_size }'} }', variables: {
      'wallet_id': walletId,
      'type': ?type_,
      'status': ?status,
      'from': ?from,
      'to': ?to,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['walletSvcMyTransactions'] == null ? null : WalletSvcTransactionPage.fromJson((data['walletSvcMyTransactions'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcLedgerPage?> walletSvcMyLedger({required String walletId, String? from, String? to, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyLedger(\$wallet_id: String!, \$from: String, \$to: String, \$page: Int, \$page_size: Int) { walletSvcMyLedger(wallet_id: \$wallet_id, from: \$from, to: \$to, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ items { id wallet_id transaction_id direction amount balance_after created_at } total page page_size }'} }', variables: {
      'wallet_id': walletId,
      'from': ?from,
      'to': ?to,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['walletSvcMyLedger'] == null ? null : WalletSvcLedgerPage.fromJson((data['walletSvcMyLedger'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcStatement?> walletSvcMyStatement({required String walletId, required String from, required String to, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyStatement(\$wallet_id: String!, \$from: String!, \$to: String!) { walletSvcMyStatement(wallet_id: \$wallet_id, from: \$from, to: \$to) ${selectionSet ?? '{ wallet_id currency from to opening_balance closing_balance total_credit total_debit net_change credit_count debit_count }'} }', variables: {
      'wallet_id': walletId,
      'from': from,
      'to': to,
    }
    );
    return data['walletSvcMyStatement'] == null ? null : WalletSvcStatement.fromJson((data['walletSvcMyStatement'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletSvcCard>?> walletSvcMyCards({required String walletId, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyCards(\$wallet_id: String!) { walletSvcMyCards(wallet_id: \$wallet_id) ${selectionSet ?? '{ id wallet_id issuer_bank_id card_number holder_name network expiry_month expiry_year status created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
    }
    );
    return (data['walletSvcMyCards'] as List?)?.map((e) => (e == null ? null : WalletSvcCard.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<WalletSvcBankAccount>?> walletSvcMyBankAccounts({String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyBankAccounts { walletSvcMyBankAccounts ${selectionSet ?? '{ id user_id bank_id account_number account_name is_default status created_at updated_at }'} }'
    );
    return (data['walletSvcMyBankAccounts'] as List?)?.map((e) => (e == null ? null : WalletSvcBankAccount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletSvcPaymentRequestPage?> walletSvcMyPaymentRequests({String? role, String? status, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcMyPaymentRequests(\$role: String, \$status: String, \$page: Int, \$page_size: Int) { walletSvcMyPaymentRequests(role: \$role, status: \$status, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ items { id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at } total page page_size }'} }', variables: {
      'role': ?role,
      'status': ?status,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['walletSvcMyPaymentRequests'] == null ? null : WalletSvcPaymentRequestPage.fromJson((data['walletSvcMyPaymentRequests'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletSvcWallet>?> walletSvcUserWallets({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcUserWallets(\$user_id: String!) { walletSvcUserWallets(user_id: \$user_id) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['walletSvcUserWallets'] as List?)?.map((e) => (e == null ? null : WalletSvcWallet.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletSvcBalance?> walletSvcBalance({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcBalance(\$id: String!) { walletSvcBalance(id: \$id) ${selectionSet ?? '{ wallet_id currency balance held available status }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcBalance'] == null ? null : WalletSvcBalance.fromJson((data['walletSvcBalance'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransactionPage?> walletSvcTransactions({required String id, String? type_, String? status, String? from, String? to, int? page, int? pageSize, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcTransactions(\$id: String!, \$type: String, \$status: String, \$from: String, \$to: String, \$page: Int, \$page_size: Int) { walletSvcTransactions(id: \$id, type: \$type, status: \$status, from: \$from, to: \$to, page: \$page, page_size: \$page_size) ${selectionSet ?? '{ items { id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at } total page page_size }'} }', variables: {
      'id': id,
      'type': ?type_,
      'status': ?status,
      'from': ?from,
      'to': ?to,
      'page': ?page,
      'page_size': ?pageSize,
    }
    );
    return data['walletSvcTransactions'] == null ? null : WalletSvcTransactionPage.fromJson((data['walletSvcTransactions'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcStatement?> walletSvcStatement({required String id, required String from, required String to, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcStatement(\$id: String!, \$from: String!, \$to: String!) { walletSvcStatement(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ wallet_id currency from to opening_balance closing_balance total_credit total_debit net_change credit_count debit_count }'} }', variables: {
      'id': id,
      'from': from,
      'to': to,
    }
    );
    return data['walletSvcStatement'] == null ? null : WalletSvcStatement.fromJson((data['walletSvcStatement'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcCard?> walletSvcCard({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcCard(\$id: String!) { walletSvcCard(id: \$id) ${selectionSet ?? '{ id wallet_id issuer_bank_id card_number holder_name network expiry_month expiry_year status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcCard'] == null ? null : WalletSvcCard.fromJson((data['walletSvcCard'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcPaymentRequest?> walletSvcPaymentRequest({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletSvcPaymentRequest(\$id: String!) { walletSvcPaymentRequest(id: \$id) ${selectionSet ?? '{ id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcPaymentRequest'] == null ? null : WalletSvcPaymentRequest.fromJson((data['walletSvcPaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcWallet?> walletSvcCreateWallet({required String userId, required String currency, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcCreateWallet(\$user_id: String!, \$currency: String!) { walletSvcCreateWallet(user_id: \$user_id, currency: \$currency) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'user_id': userId,
      'currency': currency,
    }
    );
    return data['walletSvcCreateWallet'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcCreateWallet'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletSvcCredit({required String id, required String amount, String? description}) async {
    final data = await _executor.mutation('mutation walletSvcCredit(\$id: String!, \$amount: String!, \$description: String) { walletSvcCredit(id: \$id, amount: \$amount, description: \$description) }', variables: {
      'id': id,
      'amount': amount,
      'description': ?description,
    }
    );
    return data['walletSvcCredit'];
  }

  Future<Object?> walletSvcFreeze({required String id}) async {
    final data = await _executor.mutation('mutation walletSvcFreeze(\$id: String!) { walletSvcFreeze(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletSvcFreeze'];
  }

  Future<WalletSvcTransaction?> walletSvcAuthorize({required String idempotencyKey, required String cardId, required String merchantId, required String amount, required String currency, String? description, Map<String, dynamic>? metadata, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcAuthorize(\$idempotency_key: String!, \$card_id: String!, \$merchant_id: String!, \$amount: String!, \$currency: String!, \$description: String, \$metadata: Map) { walletSvcAuthorize(idempotency_key: \$idempotency_key, card_id: \$card_id, merchant_id: \$merchant_id, amount: \$amount, currency: \$currency, description: \$description, metadata: \$metadata) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'idempotency_key': idempotencyKey,
      'card_id': cardId,
      'merchant_id': merchantId,
      'amount': amount,
      'currency': currency,
      'description': ?description,
      'metadata': ?metadata,
    }
    );
    return data['walletSvcAuthorize'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcAuthorize'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcCapture({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcCapture(\$id: String!) { walletSvcCapture(id: \$id) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcCapture'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcCapture'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcVoid({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcVoid(\$id: String!) { walletSvcVoid(id: \$id) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcVoid'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcVoid'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletSvcCreateSettlementBatch({required String merchantId}) async {
    final data = await _executor.mutation('mutation walletSvcCreateSettlementBatch(\$merchant_id: String!) { walletSvcCreateSettlementBatch(merchant_id: \$merchant_id) }', variables: {
      'merchant_id': merchantId,
    }
    );
    return data['walletSvcCreateSettlementBatch'];
  }

  Future<Object?> walletSvcProcessClearing({required String batchId}) async {
    final data = await _executor.mutation('mutation walletSvcProcessClearing(\$batch_id: String!) { walletSvcProcessClearing(batch_id: \$batch_id) }', variables: {
      'batch_id': batchId,
    }
    );
    return data['walletSvcProcessClearing'];
  }

  Future<Object?> walletSvcProcessSettlement({required String batchId}) async {
    final data = await _executor.mutation('mutation walletSvcProcessSettlement(\$batch_id: String!) { walletSvcProcessSettlement(batch_id: \$batch_id) }', variables: {
      'batch_id': batchId,
    }
    );
    return data['walletSvcProcessSettlement'];
  }

  Future<WalletSvcWallet?> walletSvcMyCreateWallet({required String currency, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyCreateWallet(\$currency: String!) { walletSvcMyCreateWallet(currency: \$currency) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'currency': currency,
    }
    );
    return data['walletSvcMyCreateWallet'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcMyCreateWallet'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletSvcMySetPin({required String walletId, String? currentPin, required String newPin}) async {
    final data = await _executor.mutation('mutation walletSvcMySetPin(\$wallet_id: String!, \$current_pin: String, \$new_pin: String!) { walletSvcMySetPin(wallet_id: \$wallet_id, current_pin: \$current_pin, new_pin: \$new_pin) }', variables: {
      'wallet_id': walletId,
      'current_pin': ?currentPin,
      'new_pin': newPin,
    }
    );
    return data['walletSvcMySetPin'];
  }

  Future<WalletSvcWallet?> walletSvcMyCloseWallet({required String walletId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyCloseWallet(\$wallet_id: String!) { walletSvcMyCloseWallet(wallet_id: \$wallet_id) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
    }
    );
    return data['walletSvcMyCloseWallet'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcMyCloseWallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcMyTopUp({required String walletId, required String bankAccountId, required String amount, String? description, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyTopUp(\$wallet_id: String!, \$bank_account_id: String!, \$amount: String!, \$description: String, \$idempotency_key: String!) { walletSvcMyTopUp(wallet_id: \$wallet_id, bank_account_id: \$bank_account_id, amount: \$amount, description: \$description, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
      'bank_account_id': bankAccountId,
      'amount': amount,
      'description': ?description,
      'idempotency_key': idempotencyKey,
    }
    );
    return data['walletSvcMyTopUp'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcMyTopUp'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcMyWithdraw({required String walletId, required String bankAccountId, required String amount, required String pin, String? description, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyWithdraw(\$wallet_id: String!, \$bank_account_id: String!, \$amount: String!, \$pin: String!, \$description: String, \$idempotency_key: String!) { walletSvcMyWithdraw(wallet_id: \$wallet_id, bank_account_id: \$bank_account_id, amount: \$amount, pin: \$pin, description: \$description, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
      'bank_account_id': bankAccountId,
      'amount': amount,
      'pin': pin,
      'description': ?description,
      'idempotency_key': idempotencyKey,
    }
    );
    return data['walletSvcMyWithdraw'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcMyWithdraw'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcMyTransfer({required String fromWalletId, String? toWalletId, String? toUserId, required String amount, required String pin, String? description, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyTransfer(\$from_wallet_id: String!, \$to_wallet_id: String, \$to_user_id: String, \$amount: String!, \$pin: String!, \$description: String, \$idempotency_key: String!) { walletSvcMyTransfer(from_wallet_id: \$from_wallet_id, to_wallet_id: \$to_wallet_id, to_user_id: \$to_user_id, amount: \$amount, pin: \$pin, description: \$description, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'from_wallet_id': fromWalletId,
      'to_wallet_id': ?toWalletId,
      'to_user_id': ?toUserId,
      'amount': amount,
      'pin': pin,
      'description': ?description,
      'idempotency_key': idempotencyKey,
    }
    );
    return data['walletSvcMyTransfer'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcMyTransfer'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcMyPay({required String walletId, required String merchantId, required String amount, required String pin, String? description, Map<String, dynamic>? metadata, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyPay(\$wallet_id: String!, \$merchant_id: String!, \$amount: String!, \$pin: String!, \$description: String, \$metadata: Map, \$idempotency_key: String!) { walletSvcMyPay(wallet_id: \$wallet_id, merchant_id: \$merchant_id, amount: \$amount, pin: \$pin, description: \$description, metadata: \$metadata, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
      'merchant_id': merchantId,
      'amount': amount,
      'pin': pin,
      'description': ?description,
      'metadata': ?metadata,
      'idempotency_key': idempotencyKey,
    }
    );
    return data['walletSvcMyPay'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcMyPay'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcBankAccount?> walletSvcMyLinkBankAccount({required String bankId, required String accountNumber, required String accountName, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyLinkBankAccount(\$bank_id: String!, \$account_number: String!, \$account_name: String!) { walletSvcMyLinkBankAccount(bank_id: \$bank_id, account_number: \$account_number, account_name: \$account_name) ${selectionSet ?? '{ id user_id bank_id account_number account_name is_default status created_at updated_at }'} }', variables: {
      'bank_id': bankId,
      'account_number': accountNumber,
      'account_name': accountName,
    }
    );
    return data['walletSvcMyLinkBankAccount'] == null ? null : WalletSvcBankAccount.fromJson((data['walletSvcMyLinkBankAccount'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletSvcMyRemoveBankAccount({required String accountId}) async {
    final data = await _executor.mutation('mutation walletSvcMyRemoveBankAccount(\$account_id: String!) { walletSvcMyRemoveBankAccount(account_id: \$account_id) }', variables: {
      'account_id': accountId,
    }
    );
    return data['walletSvcMyRemoveBankAccount'];
  }

  Future<WalletSvcBankAccount?> walletSvcMySetDefaultBankAccount({required String accountId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMySetDefaultBankAccount(\$account_id: String!) { walletSvcMySetDefaultBankAccount(account_id: \$account_id) ${selectionSet ?? '{ id user_id bank_id account_number account_name is_default status created_at updated_at }'} }', variables: {
      'account_id': accountId,
    }
    );
    return data['walletSvcMySetDefaultBankAccount'] == null ? null : WalletSvcBankAccount.fromJson((data['walletSvcMySetDefaultBankAccount'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcCard?> walletSvcMyIssueCard({required String walletId, required String issuerBankId, required String network, required String holderName, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyIssueCard(\$wallet_id: String!, \$issuer_bank_id: String!, \$network: String!, \$holder_name: String!) { walletSvcMyIssueCard(wallet_id: \$wallet_id, issuer_bank_id: \$issuer_bank_id, network: \$network, holder_name: \$holder_name) ${selectionSet ?? '{ id wallet_id issuer_bank_id card_number holder_name network expiry_month expiry_year status created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
      'issuer_bank_id': issuerBankId,
      'network': network,
      'holder_name': holderName,
    }
    );
    return data['walletSvcMyIssueCard'] == null ? null : WalletSvcCard.fromJson((data['walletSvcMyIssueCard'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcPaymentRequest?> walletSvcMyRequestPayment({required String walletId, required String payerUserId, required String amount, String? note, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyRequestPayment(\$wallet_id: String!, \$payer_user_id: String!, \$amount: String!, \$note: String) { walletSvcMyRequestPayment(wallet_id: \$wallet_id, payer_user_id: \$payer_user_id, amount: \$amount, note: \$note) ${selectionSet ?? '{ id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at }'} }', variables: {
      'wallet_id': walletId,
      'payer_user_id': payerUserId,
      'amount': amount,
      'note': ?note,
    }
    );
    return data['walletSvcMyRequestPayment'] == null ? null : WalletSvcPaymentRequest.fromJson((data['walletSvcMyRequestPayment'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcPaymentRequest?> walletSvcMyPayPaymentRequest({required String id, required String walletId, required String pin, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyPayPaymentRequest(\$id: String!, \$wallet_id: String!, \$pin: String!) { walletSvcMyPayPaymentRequest(id: \$id, wallet_id: \$wallet_id, pin: \$pin) ${selectionSet ?? '{ id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at }'} }', variables: {
      'id': id,
      'wallet_id': walletId,
      'pin': pin,
    }
    );
    return data['walletSvcMyPayPaymentRequest'] == null ? null : WalletSvcPaymentRequest.fromJson((data['walletSvcMyPayPaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcPaymentRequest?> walletSvcMyDeclinePaymentRequest({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyDeclinePaymentRequest(\$id: String!) { walletSvcMyDeclinePaymentRequest(id: \$id) ${selectionSet ?? '{ id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcMyDeclinePaymentRequest'] == null ? null : WalletSvcPaymentRequest.fromJson((data['walletSvcMyDeclinePaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcPaymentRequest?> walletSvcMyCancelPaymentRequest({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcMyCancelPaymentRequest(\$id: String!) { walletSvcMyCancelPaymentRequest(id: \$id) ${selectionSet ?? '{ id requester_user_id requester_wallet_id payer_user_id payer_wallet_id transaction_id amount currency note status expires_at created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcMyCancelPaymentRequest'] == null ? null : WalletSvcPaymentRequest.fromJson((data['walletSvcMyCancelPaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletSvcUnfreeze({required String id}) async {
    final data = await _executor.mutation('mutation walletSvcUnfreeze(\$id: String!) { walletSvcUnfreeze(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletSvcUnfreeze'];
  }

  Future<WalletSvcWallet?> walletSvcChangeTier({required String id, required String tier, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcChangeTier(\$id: String!, \$tier: String!) { walletSvcChangeTier(id: \$id, tier: \$tier) ${selectionSet ?? '{ id user_id balance frozen_amount currency status tier has_pin version created_at updated_at }'} }', variables: {
      'id': id,
      'tier': tier,
    }
    );
    return data['walletSvcChangeTier'] == null ? null : WalletSvcWallet.fromJson((data['walletSvcChangeTier'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcTransaction?> walletSvcRefund({required String id, String? amount, String? reason, required String idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcRefund(\$id: String!, \$amount: String, \$reason: String, \$idempotency_key: String!) { walletSvcRefund(id: \$id, amount: \$amount, reason: \$reason, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id idempotency_key wallet_id counterparty_wallet_id bank_account_id reference_id merchant_id acquirer_bank_id issuer_bank_id card_id card_network amount currency fee refunded_amount type status description authorization_code external_ref failure_reason batch_id metadata created_at updated_at }'} }', variables: {
      'id': id,
      'amount': ?amount,
      'reason': ?reason,
      'idempotency_key': idempotencyKey,
    }
    );
    return data['walletSvcRefund'] == null ? null : WalletSvcTransaction.fromJson((data['walletSvcRefund'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcCard?> walletSvcBlockCard({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcBlockCard(\$id: String!) { walletSvcBlockCard(id: \$id) ${selectionSet ?? '{ id wallet_id issuer_bank_id card_number holder_name network expiry_month expiry_year status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcBlockCard'] == null ? null : WalletSvcCard.fromJson((data['walletSvcBlockCard'] as Map).cast<String, dynamic>());
  }

  Future<WalletSvcCard?> walletSvcUnblockCard({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletSvcUnblockCard(\$id: String!) { walletSvcUnblockCard(id: \$id) ${selectionSet ?? '{ id wallet_id issuer_bank_id card_number holder_name network expiry_month expiry_year status created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletSvcUnblockCard'] == null ? null : WalletSvcCard.fromJson((data['walletSvcUnblockCard'] as Map).cast<String, dynamic>());
  }

}
