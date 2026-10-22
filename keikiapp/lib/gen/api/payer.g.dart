// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PayerApi {
  const PayerApi(this._executor);

  final GatewayExecutor _executor;

  Future<PayerTransactionResponse?> payerTransaction({required int transactionId, String? selectionSet}) async {
    final data = await _executor.query('query payerTransaction(\$transaction_id: Int!) { payerTransaction(transaction_id: \$transaction_id) ${selectionSet ?? '{ transaction { transaction_id payer_id buyer_id amount transaction_date transaction_type description status } payers { payer_id name email phone_number } buyers { buyer_id name email phone_number } }'} }', variables: {
      'transaction_id': transactionId,
    }
    );
    return data['payerTransaction'] == null ? null : PayerTransactionResponse.fromJson((data['payerTransaction'] as Map).cast<String, dynamic>());
  }

}
