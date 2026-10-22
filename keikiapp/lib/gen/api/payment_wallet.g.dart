// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PaymentWalletApi {
  const PaymentWalletApi(this._executor);

  final GatewayExecutor _executor;

  Future<WalletWallet?> wallet({required String id, String? selectionSet}) async {
    final data = await _executor.query('query wallet(\$id: String!) { wallet(id: \$id) ${selectionSet ?? '{ wallet_id user_id balance currency status daily_limit per_transaction_limit frozen_reason created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['wallet'] == null ? null : WalletWallet.fromJson((data['wallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletWallet?> walletByUser({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query walletByUser(\$userId: String!) { walletByUser(userId: \$userId) ${selectionSet ?? '{ wallet_id user_id balance currency status daily_limit per_transaction_limit frozen_reason created_at updated_at }'} }', variables: {
      'userId': userId,
    }
    );
    return data['walletByUser'] == null ? null : WalletWallet.fromJson((data['walletByUser'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletTransaction>?> walletTransactions({required String id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query walletTransactions(\$id: String!, \$limit: Int, \$offset: Int) { walletTransactions(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ transaction_id wallet_id type amount currency status reference_id counterparty_wallet_id pocket_id description created_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['walletTransactions'] as List?)?.map((e) => (e == null ? null : WalletTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletTransaction?> walletTransactionByReference({required String id, required String referenceId, String? selectionSet}) async {
    final data = await _executor.query('query walletTransactionByReference(\$id: String!, \$referenceId: String!) { walletTransactionByReference(id: \$id, referenceId: \$referenceId) ${selectionSet ?? '{ transaction_id wallet_id type amount currency status reference_id counterparty_wallet_id pocket_id description created_at }'} }', variables: {
      'id': id,
      'referenceId': referenceId,
    }
    );
    return data['walletTransactionByReference'] == null ? null : WalletTransaction.fromJson((data['walletTransactionByReference'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletPocket>?> walletPockets({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletPockets(\$id: String!) { walletPockets(id: \$id) ${selectionSet ?? '{ pocket_id wallet_id name balance currency created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['walletPockets'] as List?)?.map((e) => (e == null ? null : WalletPocket.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletTransfer?> walletTransfer({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletTransfer(\$id: String!) { walletTransfer(id: \$id) ${selectionSet ?? '{ transfer_id sender_wallet_id receiver_wallet_id amount currency status reference_id out_transaction_id in_transaction_id description created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletTransfer'] == null ? null : WalletTransfer.fromJson((data['walletTransfer'] as Map).cast<String, dynamic>());
  }

  Future<WalletTransaction?> walletTransaction({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletTransaction(\$id: String!) { walletTransaction(id: \$id) ${selectionSet ?? '{ transaction_id wallet_id type amount currency status reference_id counterparty_wallet_id pocket_id description created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletTransaction'] == null ? null : WalletTransaction.fromJson((data['walletTransaction'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletPaymentRequest>?> walletPaymentRequests({required String walletId, String? selectionSet}) async {
    final data = await _executor.query('query walletPaymentRequests(\$wallet_id: String!) { walletPaymentRequests(wallet_id: \$wallet_id) ${selectionSet ?? '{ payment_request_id requester_wallet_id payer_wallet_id amount currency status description transfer_id expires_at created_at }'} }', variables: {
      'wallet_id': walletId,
    }
    );
    return (data['walletPaymentRequests'] as List?)?.map((e) => (e == null ? null : WalletPaymentRequest.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletPaymentRequest?> walletPaymentRequest({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletPaymentRequest(\$id: String!) { walletPaymentRequest(id: \$id) ${selectionSet ?? '{ payment_request_id requester_wallet_id payer_wallet_id amount currency status description transfer_id expires_at created_at }'} }', variables: {
      'id': id,
    }
    );
    return data['walletPaymentRequest'] == null ? null : WalletPaymentRequest.fromJson((data['walletPaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<WalletPaymentRequestQR?> walletPaymentRequestQR({required String id, String? selectionSet}) async {
    final data = await _executor.query('query walletPaymentRequestQR(\$id: String!) { walletPaymentRequestQR(id: \$id) ${selectionSet ?? '{ payment_request_id payload }'} }', variables: {
      'id': id,
    }
    );
    return data['walletPaymentRequestQR'] == null ? null : WalletPaymentRequestQR.fromJson((data['walletPaymentRequestQR'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletPaymentMethod>?> walletPaymentMethods({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query walletPaymentMethods(\$user_id: String!) { walletPaymentMethods(user_id: \$user_id) ${selectionSet ?? '{ payment_method_id user_id type provider account_number is_default is_active created_at }'} }', variables: {
      'user_id': userId,
    }
    );
    return (data['walletPaymentMethods'] as List?)?.map((e) => (e == null ? null : WalletPaymentMethod.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletWallet?> myWallet({String? selectionSet}) async {
    final data = await _executor.query('query myWallet { myWallet ${selectionSet ?? '{ wallet_id user_id balance currency status daily_limit per_transaction_limit frozen_reason created_at updated_at }'} }'
    );
    return data['myWallet'] == null ? null : WalletWallet.fromJson((data['myWallet'] as Map).cast<String, dynamic>());
  }

  Future<List<WalletTransaction>?> myWalletTransactions({required String id, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query myWalletTransactions(\$id: String!, \$limit: Int, \$offset: Int) { myWalletTransactions(id: \$id, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ transaction_id wallet_id type amount currency status reference_id counterparty_wallet_id pocket_id description created_at }'} }', variables: {
      'id': id,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return (data['myWalletTransactions'] as List?)?.map((e) => (e == null ? null : WalletTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<WalletPocket>?> myWalletPockets({required String id, String? selectionSet}) async {
    final data = await _executor.query('query myWalletPockets(\$id: String!) { myWalletPockets(id: \$id) ${selectionSet ?? '{ pocket_id wallet_id name balance currency created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['myWalletPockets'] as List?)?.map((e) => (e == null ? null : WalletPocket.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<WalletPaymentMethod>?> myPaymentMethods({String? selectionSet}) async {
    final data = await _executor.query('query myPaymentMethods { myPaymentMethods ${selectionSet ?? '{ payment_method_id user_id type provider account_number is_default is_active created_at }'} }'
    );
    return (data['myPaymentMethods'] as List?)?.map((e) => (e == null ? null : WalletPaymentMethod.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<WalletWallet?> createMyWallet({required String currency, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createMyWallet(\$currency: String!) { createMyWallet(currency: \$currency) ${selectionSet ?? '{ wallet_id user_id balance currency status daily_limit per_transaction_limit frozen_reason created_at updated_at }'} }', variables: {
      'currency': currency,
    }
    );
    return data['createMyWallet'] == null ? null : WalletWallet.fromJson((data['createMyWallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletPocket?> createMyPocket({required String id, required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createMyPocket(\$id: String!, \$name: String!) { createMyPocket(id: \$id, name: \$name) ${selectionSet ?? '{ pocket_id wallet_id name balance currency created_at updated_at }'} }', variables: {
      'id': id,
      'name': name_,
    }
    );
    return data['createMyPocket'] == null ? null : WalletPocket.fromJson((data['createMyPocket'] as Map).cast<String, dynamic>());
  }

  Future<WalletTransfer?> transferFromMyWallet({required String senderWalletId, required String receiverWalletId, required int amount, required String referenceId, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation transferFromMyWallet(\$sender_wallet_id: String!, \$receiver_wallet_id: String!, \$amount: Int64!, \$reference_id: String!, \$description: String) { transferFromMyWallet(sender_wallet_id: \$sender_wallet_id, receiver_wallet_id: \$receiver_wallet_id, amount: \$amount, reference_id: \$reference_id, description: \$description) ${selectionSet ?? '{ transfer_id sender_wallet_id receiver_wallet_id amount currency status reference_id out_transaction_id in_transaction_id description created_at }'} }', variables: {
      'sender_wallet_id': senderWalletId,
      'receiver_wallet_id': receiverWalletId,
      'amount': amount,
      'reference_id': referenceId,
      'description': ?description,
    }
    );
    return data['transferFromMyWallet'] == null ? null : WalletTransfer.fromJson((data['transferFromMyWallet'] as Map).cast<String, dynamic>());
  }

  Future<WalletPaymentRequest?> createMyPaymentRequest({required String requesterWalletId, String? payerWalletId, required int amount, String? description, int? expiresInSeconds, String? selectionSet}) async {
    final data = await _executor.mutation('mutation createMyPaymentRequest(\$requester_wallet_id: String!, \$payer_wallet_id: String, \$amount: Int64!, \$description: String, \$expires_in_seconds: Int64) { createMyPaymentRequest(requester_wallet_id: \$requester_wallet_id, payer_wallet_id: \$payer_wallet_id, amount: \$amount, description: \$description, expires_in_seconds: \$expires_in_seconds) ${selectionSet ?? '{ payment_request_id requester_wallet_id payer_wallet_id amount currency status description transfer_id expires_at created_at }'} }', variables: {
      'requester_wallet_id': requesterWalletId,
      'payer_wallet_id': ?payerWalletId,
      'amount': amount,
      'description': ?description,
      'expires_in_seconds': ?expiresInSeconds,
    }
    );
    return data['createMyPaymentRequest'] == null ? null : WalletPaymentRequest.fromJson((data['createMyPaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<Object?> payPaymentRequestFromMyWallet({required String id, required String payerWalletId}) async {
    final data = await _executor.mutation('mutation payPaymentRequestFromMyWallet(\$id: String!, \$payer_wallet_id: String!) { payPaymentRequestFromMyWallet(id: \$id, payer_wallet_id: \$payer_wallet_id) }', variables: {
      'id': id,
      'payer_wallet_id': payerWalletId,
    }
    );
    return data['payPaymentRequestFromMyWallet'];
  }

  Future<WalletPaymentMethod?> addMyPaymentMethod({required String type_, required String provider, required String accountNumber, String? selectionSet}) async {
    final data = await _executor.mutation('mutation addMyPaymentMethod(\$type: String!, \$provider: String!, \$account_number: String!) { addMyPaymentMethod(type: \$type, provider: \$provider, account_number: \$account_number) ${selectionSet ?? '{ payment_method_id user_id type provider account_number is_default is_active created_at }'} }', variables: {
      'type': type_,
      'provider': provider,
      'account_number': accountNumber,
    }
    );
    return data['addMyPaymentMethod'] == null ? null : WalletPaymentMethod.fromJson((data['addMyPaymentMethod'] as Map).cast<String, dynamic>());
  }

  Future<WalletWallet?> walletCreateWallet({required String userId, required String currency, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletCreateWallet(\$user_id: String!, \$currency: String!) { walletCreateWallet(user_id: \$user_id, currency: \$currency) ${selectionSet ?? '{ wallet_id user_id balance currency status daily_limit per_transaction_limit frozen_reason created_at updated_at }'} }', variables: {
      'user_id': userId,
      'currency': currency,
    }
    );
    return data['walletCreateWallet'] == null ? null : WalletWallet.fromJson((data['walletCreateWallet'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletDeposit({required String id, required int amount, required String referenceId, String? description}) async {
    final data = await _executor.mutation('mutation walletDeposit(\$id: String!, \$amount: Int64!, \$reference_id: String!, \$description: String) { walletDeposit(id: \$id, amount: \$amount, reference_id: \$reference_id, description: \$description) }', variables: {
      'id': id,
      'amount': amount,
      'reference_id': referenceId,
      'description': ?description,
    }
    );
    return data['walletDeposit'];
  }

  Future<Object?> walletWithdraw({required String id, required int amount, required String referenceId, String? description}) async {
    final data = await _executor.mutation('mutation walletWithdraw(\$id: String!, \$amount: Int64!, \$reference_id: String!, \$description: String) { walletWithdraw(id: \$id, amount: \$amount, reference_id: \$reference_id, description: \$description) }', variables: {
      'id': id,
      'amount': amount,
      'reference_id': referenceId,
      'description': ?description,
    }
    );
    return data['walletWithdraw'];
  }

  Future<Object?> walletFreeze({required String id, String? reason}) async {
    final data = await _executor.mutation('mutation walletFreeze(\$id: String!, \$reason: String) { walletFreeze(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['walletFreeze'];
  }

  Future<Object?> walletUnfreeze({required String id}) async {
    final data = await _executor.mutation('mutation walletUnfreeze(\$id: String!) { walletUnfreeze(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletUnfreeze'];
  }

  Future<Object?> walletClose({required String id}) async {
    final data = await _executor.mutation('mutation walletClose(\$id: String!) { walletClose(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletClose'];
  }

  Future<Object?> walletSetLimits({required String id, required int dailyLimit, required int perTransactionLimit}) async {
    final data = await _executor.mutation('mutation walletSetLimits(\$id: String!, \$daily_limit: Int64!, \$per_transaction_limit: Int64!) { walletSetLimits(id: \$id, daily_limit: \$daily_limit, per_transaction_limit: \$per_transaction_limit) }', variables: {
      'id': id,
      'daily_limit': dailyLimit,
      'per_transaction_limit': perTransactionLimit,
    }
    );
    return data['walletSetLimits'];
  }

  Future<WalletPocket?> walletCreatePocket({required String id, required String name_, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletCreatePocket(\$id: String!, \$name: String!) { walletCreatePocket(id: \$id, name: \$name) ${selectionSet ?? '{ pocket_id wallet_id name balance currency created_at updated_at }'} }', variables: {
      'id': id,
      'name': name_,
    }
    );
    return data['walletCreatePocket'] == null ? null : WalletPocket.fromJson((data['walletCreatePocket'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletDepositToPocket({required String pocketId, required int amount}) async {
    final data = await _executor.mutation('mutation walletDepositToPocket(\$pocketId: String!, \$amount: Int64!) { walletDepositToPocket(pocketId: \$pocketId, amount: \$amount) }', variables: {
      'pocketId': pocketId,
      'amount': amount,
    }
    );
    return data['walletDepositToPocket'];
  }

  Future<Object?> walletWithdrawFromPocket({required String pocketId, required int amount}) async {
    final data = await _executor.mutation('mutation walletWithdrawFromPocket(\$pocketId: String!, \$amount: Int64!) { walletWithdrawFromPocket(pocketId: \$pocketId, amount: \$amount) }', variables: {
      'pocketId': pocketId,
      'amount': amount,
    }
    );
    return data['walletWithdrawFromPocket'];
  }

  Future<Object?> walletClosePocket({required String pocketId}) async {
    final data = await _executor.mutation('mutation walletClosePocket(\$pocketId: String!) { walletClosePocket(pocketId: \$pocketId) }', variables: {
      'pocketId': pocketId,
    }
    );
    return data['walletClosePocket'];
  }

  Future<WalletTransfer?> walletTransferMutation({required String senderWalletId, required String receiverWalletId, required int amount, required String referenceId, String? description, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletTransfer(\$sender_wallet_id: String!, \$receiver_wallet_id: String!, \$amount: Int64!, \$reference_id: String!, \$description: String) { walletTransfer(sender_wallet_id: \$sender_wallet_id, receiver_wallet_id: \$receiver_wallet_id, amount: \$amount, reference_id: \$reference_id, description: \$description) ${selectionSet ?? '{ transfer_id sender_wallet_id receiver_wallet_id amount currency status reference_id out_transaction_id in_transaction_id description created_at }'} }', variables: {
      'sender_wallet_id': senderWalletId,
      'receiver_wallet_id': receiverWalletId,
      'amount': amount,
      'reference_id': referenceId,
      'description': ?description,
    }
    );
    return data['walletTransfer'] == null ? null : WalletTransfer.fromJson((data['walletTransfer'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletReverseTransfer({required String id, String? reason}) async {
    final data = await _executor.mutation('mutation walletReverseTransfer(\$id: String!, \$reason: String) { walletReverseTransfer(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['walletReverseTransfer'];
  }

  Future<Object?> walletReverseTransaction({required String id, String? reason}) async {
    final data = await _executor.mutation('mutation walletReverseTransaction(\$id: String!, \$reason: String) { walletReverseTransaction(id: \$id, reason: \$reason) }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['walletReverseTransaction'];
  }

  Future<WalletPaymentRequest?> walletCreatePaymentRequest({required String requesterWalletId, String? payerWalletId, required int amount, String? description, int? expiresInSeconds, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletCreatePaymentRequest(\$requester_wallet_id: String!, \$payer_wallet_id: String, \$amount: Int64!, \$description: String, \$expires_in_seconds: Int64) { walletCreatePaymentRequest(requester_wallet_id: \$requester_wallet_id, payer_wallet_id: \$payer_wallet_id, amount: \$amount, description: \$description, expires_in_seconds: \$expires_in_seconds) ${selectionSet ?? '{ payment_request_id requester_wallet_id payer_wallet_id amount currency status description transfer_id expires_at created_at }'} }', variables: {
      'requester_wallet_id': requesterWalletId,
      'payer_wallet_id': ?payerWalletId,
      'amount': amount,
      'description': ?description,
      'expires_in_seconds': ?expiresInSeconds,
    }
    );
    return data['walletCreatePaymentRequest'] == null ? null : WalletPaymentRequest.fromJson((data['walletCreatePaymentRequest'] as Map).cast<String, dynamic>());
  }

  Future<Object?> walletPayPaymentRequest({required String id, required String payerWalletId}) async {
    final data = await _executor.mutation('mutation walletPayPaymentRequest(\$id: String!, \$payer_wallet_id: String!) { walletPayPaymentRequest(id: \$id, payer_wallet_id: \$payer_wallet_id) }', variables: {
      'id': id,
      'payer_wallet_id': payerWalletId,
    }
    );
    return data['walletPayPaymentRequest'];
  }

  Future<Object?> walletCancelPaymentRequest({required String id}) async {
    final data = await _executor.mutation('mutation walletCancelPaymentRequest(\$id: String!) { walletCancelPaymentRequest(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletCancelPaymentRequest'];
  }

  Future<WalletPaymentMethod?> walletAddPaymentMethod({required String userId, required String type_, required String provider, required String accountNumber, String? selectionSet}) async {
    final data = await _executor.mutation('mutation walletAddPaymentMethod(\$user_id: String!, \$type: String!, \$provider: String!, \$account_number: String!) { walletAddPaymentMethod(user_id: \$user_id, type: \$type, provider: \$provider, account_number: \$account_number) ${selectionSet ?? '{ payment_method_id user_id type provider account_number is_default is_active created_at }'} }', variables: {
      'user_id': userId,
      'type': type_,
      'provider': provider,
      'account_number': accountNumber,
    }
    );
    return data['walletAddPaymentMethod'] == null ? null : WalletPaymentMethod.fromJson((data['walletAddPaymentMethod'] as Map).cast<String, dynamic>());
  }

  Future<bool?> walletRemovePaymentMethod({required String id}) async {
    final data = await _executor.mutation('mutation walletRemovePaymentMethod(\$id: String!) { walletRemovePaymentMethod(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletRemovePaymentMethod'] as bool?;
  }

  Future<Object?> walletSetDefaultPaymentMethod({required String id}) async {
    final data = await _executor.mutation('mutation walletSetDefaultPaymentMethod(\$id: String!) { walletSetDefaultPaymentMethod(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['walletSetDefaultPaymentMethod'];
  }

}
