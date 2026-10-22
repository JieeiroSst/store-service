// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class PaymentApi {
  const PaymentApi(this._executor);

  final GatewayExecutor _executor;

  Future<PaymentPaymentList?> payments({String? provider, String? status, String? from, String? to, int? limit, int? offset, String? selectionSet}) async {
    final data = await _executor.query('query payments(\$provider: String, \$status: String, \$from: String, \$to: String, \$limit: Int, \$offset: Int) { payments(provider: \$provider, status: \$status, from: \$from, to: \$to, limit: \$limit, offset: \$offset) ${selectionSet ?? '{ Items { ID Provider ExternalID Amount Currency Status PayerEmail Description FailureReason IdempotencyKey RefundedAmount CreatedAt UpdatedAt } Total }'} }', variables: {
      'provider': ?provider,
      'status': ?status,
      'from': ?from,
      'to': ?to,
      'limit': ?limit,
      'offset': ?offset,
    }
    );
    return data['payments'] == null ? null : PaymentPaymentList.fromJson((data['payments'] as Map).cast<String, dynamic>());
  }

  Future<PaymentPayment?> payment({required int id, String? selectionSet}) async {
    final data = await _executor.query('query payment(\$id: Int!) { payment(id: \$id) ${selectionSet ?? '{ ID Provider ExternalID Amount Currency Status PayerEmail Description FailureReason IdempotencyKey RefundedAmount CreatedAt UpdatedAt }'} }', variables: {
      'id': id,
    }
    );
    return data['payment'] == null ? null : PaymentPayment.fromJson((data['payment'] as Map).cast<String, dynamic>());
  }

  Future<List<PaymentTransaction>?> paymentTransactions({required int id, String? selectionSet}) async {
    final data = await _executor.query('query paymentTransactions(\$id: Int!) { paymentTransactions(id: \$id) ${selectionSet ?? '{ ID PaymentID Type Status Amount ExternalID RawStatus ErrorMessage CreatedAt }'} }', variables: {
      'id': id,
    }
    );
    return (data['paymentTransactions'] as List?)?.map((e) => (e == null ? null : PaymentTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<PaymentPayment?> paymentCreatePayment({required PaymentCreateInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation paymentCreatePayment(\$input: PaymentCreateInput!) { paymentCreatePayment(input: \$input) ${selectionSet ?? '{ ID Provider ExternalID Amount Currency Status PayerEmail Description FailureReason IdempotencyKey RefundedAmount CreatedAt UpdatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['paymentCreatePayment'] == null ? null : PaymentPayment.fromJson((data['paymentCreatePayment'] as Map).cast<String, dynamic>());
  }

  Future<PaymentPayment?> paymentRefundPayment({required String id, int? amount, String? selectionSet}) async {
    final data = await _executor.mutation('mutation paymentRefundPayment(\$id: String!, \$amount: Int64) { paymentRefundPayment(id: \$id, amount: \$amount) ${selectionSet ?? '{ ID Provider ExternalID Amount Currency Status PayerEmail Description FailureReason IdempotencyKey RefundedAmount CreatedAt UpdatedAt }'} }', variables: {
      'id': id,
      'amount': ?amount,
    }
    );
    return data['paymentRefundPayment'] == null ? null : PaymentPayment.fromJson((data['paymentRefundPayment'] as Map).cast<String, dynamic>());
  }

}
