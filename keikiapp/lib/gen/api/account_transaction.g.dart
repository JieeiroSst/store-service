// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AccountTransactionApi {
  const AccountTransactionApi(this._executor);

  final GatewayExecutor _executor;

  Future<AccountTxAccount?> accountTxGetAccount({required AccountTxGetAccountRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query accountTxGetAccount(\$input: AccountTxGetAccountRequestInput!) { accountTxGetAccount(input: \$input) ${selectionSet ?? '{ id firstName lastName dateCreated }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxGetAccount'] == null ? null : AccountTxAccount.fromJson((data['accountTxGetAccount'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxListAccountsResponse?> accountTxListAccounts({required AccountTxListAccountsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query accountTxListAccounts(\$input: AccountTxListAccountsRequestInput!) { accountTxListAccounts(input: \$input) ${selectionSet ?? '{ accounts { id firstName lastName dateCreated } nextPageToken }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxListAccounts'] == null ? null : AccountTxListAccountsResponse.fromJson((data['accountTxListAccounts'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxTransaction?> accountTxGetTransaction({required AccountTxGetTransactionRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query accountTxGetTransaction(\$input: AccountTxGetTransactionRequestInput!) { accountTxGetTransaction(input: \$input) ${selectionSet ?? '{ id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxGetTransaction'] == null ? null : AccountTxTransaction.fromJson((data['accountTxGetTransaction'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxListTransactionsResponse?> accountTxListTransactions({required AccountTxListTransactionsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query accountTxListTransactions(\$input: AccountTxListTransactionsRequestInput!) { accountTxListTransactions(input: \$input) ${selectionSet ?? '{ transactions { id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } } nextPageToken }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxListTransactions'] == null ? null : AccountTxListTransactionsResponse.fromJson((data['accountTxListTransactions'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxListTransactionsResponse?> accountTxGetAccountTransactions({required AccountTxGetAccountTransactionsRequestInput input, String? selectionSet}) async {
    final data = await _executor.query('query accountTxGetAccountTransactions(\$input: AccountTxGetAccountTransactionsRequestInput!) { accountTxGetAccountTransactions(input: \$input) ${selectionSet ?? '{ transactions { id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } } nextPageToken }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxGetAccountTransactions'] == null ? null : AccountTxListTransactionsResponse.fromJson((data['accountTxGetAccountTransactions'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxAccount?> accountTxCreateAccount({required AccountTxCreateAccountRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation accountTxCreateAccount(\$input: AccountTxCreateAccountRequestInput!) { accountTxCreateAccount(input: \$input) ${selectionSet ?? '{ id firstName lastName dateCreated }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxCreateAccount'] == null ? null : AccountTxAccount.fromJson((data['accountTxCreateAccount'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxTransaction?> accountTxCreateDepositTransaction({required AccountTxCreateDepositTransactionRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation accountTxCreateDepositTransaction(\$input: AccountTxCreateDepositTransactionRequestInput!) { accountTxCreateDepositTransaction(input: \$input) ${selectionSet ?? '{ id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxCreateDepositTransaction'] == null ? null : AccountTxTransaction.fromJson((data['accountTxCreateDepositTransaction'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxTransaction?> accountTxCreateWithdrawalTransaction({required AccountTxCreateWithdrawalTransactionRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation accountTxCreateWithdrawalTransaction(\$input: AccountTxCreateWithdrawalTransactionRequestInput!) { accountTxCreateWithdrawalTransaction(input: \$input) ${selectionSet ?? '{ id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxCreateWithdrawalTransaction'] == null ? null : AccountTxTransaction.fromJson((data['accountTxCreateWithdrawalTransaction'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxTransaction?> accountTxCreateAccountToAccountTransaction({required AccountTxCreateAccountToAccountTransactionRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation accountTxCreateAccountToAccountTransaction(\$input: AccountTxCreateAccountToAccountTransactionRequestInput!) { accountTxCreateAccountToAccountTransaction(input: \$input) ${selectionSet ?? '{ id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxCreateAccountToAccountTransaction'] == null ? null : AccountTxTransaction.fromJson((data['accountTxCreateAccountToAccountTransaction'] as Map).cast<String, dynamic>());
  }

  Future<AccountTxTransaction?> accountTxCreatePaymentForServiceTransaction({required AccountTxCreatePaymentForServiceTransactionRequestInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation accountTxCreatePaymentForServiceTransaction(\$input: AccountTxCreatePaymentForServiceTransactionRequestInput!) { accountTxCreatePaymentForServiceTransaction(input: \$input) ${selectionSet ?? '{ id type amount dateCreated deposit { accountId } withdrawal { accountId } transfer { senderId receiverId } payment { accountId serviceName } }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['accountTxCreatePaymentForServiceTransaction'] == null ? null : AccountTxTransaction.fromJson((data['accountTxCreatePaymentForServiceTransaction'] as Map).cast<String, dynamic>());
  }

}
