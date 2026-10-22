// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SardAuthApi {
  const SardAuthApi(this._executor);

  final GatewayExecutor _executor;

  Future<Object?> mySardPaymentAuthentications() async {
    final data = await _executor.query('query mySardPaymentAuthentications { mySardPaymentAuthentications }'
    );
    return data['mySardPaymentAuthentications'];
  }

  Future<SardChallenge?> sardPaymentAuthentication({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardPaymentAuthentication(\$id: String!) { sardPaymentAuthentication(id: \$id) ${selectionSet ?? '{ id status card_id masked_pan amount currency merchant mcc attempts_left resends otp_expires_at failure_reason expires_at authenticated_at valid_until used_at created_at otp }'} }', variables: {
      'id': id,
    }
    );
    return data['sardPaymentAuthentication'] == null ? null : SardChallenge.fromJson((data['sardPaymentAuthentication'] as Map).cast<String, dynamic>());
  }

  Future<SardChallenge?> sardVerifyPaymentAuthentication({required String id, required String otp, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardVerifyPaymentAuthentication(\$id: String!, \$otp: String!) { sardVerifyPaymentAuthentication(id: \$id, otp: \$otp) ${selectionSet ?? '{ id status card_id masked_pan amount currency merchant mcc attempts_left resends otp_expires_at failure_reason expires_at authenticated_at valid_until used_at created_at otp }'} }', variables: {
      'id': id,
      'otp': otp,
    }
    );
    return data['sardVerifyPaymentAuthentication'] == null ? null : SardChallenge.fromJson((data['sardVerifyPaymentAuthentication'] as Map).cast<String, dynamic>());
  }

  Future<SardChallenge?> sardResendPaymentAuthentication({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardResendPaymentAuthentication(\$id: String!) { sardResendPaymentAuthentication(id: \$id) ${selectionSet ?? '{ id status card_id masked_pan amount currency merchant mcc attempts_left resends otp_expires_at failure_reason expires_at authenticated_at valid_until used_at created_at otp }'} }', variables: {
      'id': id,
    }
    );
    return data['sardResendPaymentAuthentication'] == null ? null : SardChallenge.fromJson((data['sardResendPaymentAuthentication'] as Map).cast<String, dynamic>());
  }

  Future<SardChallenge?> sardDeclinePaymentAuthentication({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardDeclinePaymentAuthentication(\$id: String!) { sardDeclinePaymentAuthentication(id: \$id) ${selectionSet ?? '{ id status card_id masked_pan amount currency merchant mcc attempts_left resends otp_expires_at failure_reason expires_at authenticated_at valid_until used_at created_at otp }'} }', variables: {
      'id': id,
    }
    );
    return data['sardDeclinePaymentAuthentication'] == null ? null : SardChallenge.fromJson((data['sardDeclinePaymentAuthentication'] as Map).cast<String, dynamic>());
  }

  Future<SardChallenge?> sardInitiatePaymentAuthentication({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardInitiatePaymentAuthentication(\$input: Map!) { sardInitiatePaymentAuthentication(input: \$input) ${selectionSet ?? '{ id status card_id masked_pan amount currency merchant mcc attempts_left resends otp_expires_at failure_reason expires_at authenticated_at valid_until used_at created_at otp }'} }', variables: {
      'input': input,
    }
    );
    return data['sardInitiatePaymentAuthentication'] == null ? null : SardChallenge.fromJson((data['sardInitiatePaymentAuthentication'] as Map).cast<String, dynamic>());
  }

}
