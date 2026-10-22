// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class IntegratedPaymentApi {
  const IntegratedPaymentApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<IntPayPayment>?> myIntPayPayments({String? selectionSet}) async {
    final data = await _executor.query('query myIntPayPayments { myIntPayPayments ${selectionSet ?? '{ id user_id amount currency payment_method status description metadata external_id processor_response created_at updated_at }'} }'
    );
    return (data['myIntPayPayments'] as List?)?.map((e) => (e == null ? null : IntPayPayment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<IntPayPayment?> intPayPayment({required String id, String? selectionSet}) async {
    final data = await _executor.query('query intPayPayment(\$id: String!) { intPayPayment(id: \$id) ${selectionSet ?? '{ id user_id amount currency payment_method status description metadata external_id processor_response created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['intPayPayment'] == null ? null : IntPayPayment.fromJson((data['intPayPayment'] as Map).cast<String, dynamic>());
  }

  Future<String?> intPayPaymentStatus({required String id}) async {
    final data = await _executor.query('query intPayPaymentStatus(\$id: String!) { intPayPaymentStatus(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['intPayPaymentStatus']?.toString();
  }

  Future<IntPayPaymentResult?> intPayCreatePayment({required IntPayCreatePaymentInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation intPayCreatePayment(\$input: IntPayCreatePaymentInput!) { intPayCreatePayment(input: \$input) ${selectionSet ?? '{ id status amount currency payment_url transaction_id processor_data message }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['intPayCreatePayment'] == null ? null : IntPayPaymentResult.fromJson((data['intPayCreatePayment'] as Map).cast<String, dynamic>());
  }

  Future<Object?> intPayProcessPayment({required String id}) async {
    final data = await _executor.mutation('mutation intPayProcessPayment(\$id: String!) { intPayProcessPayment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['intPayProcessPayment'];
  }

  Future<Object?> intPayRefundPayment({required String id, required double amount}) async {
    final data = await _executor.mutation('mutation intPayRefundPayment(\$id: String!, \$amount: Float!) { intPayRefundPayment(id: \$id, amount: \$amount) }', variables: {
      'id': id,
      'amount': amount,
    }
    );
    return data['intPayRefundPayment'];
  }

}
