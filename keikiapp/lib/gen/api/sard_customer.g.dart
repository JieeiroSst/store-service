// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SardCustomerApi {
  const SardCustomerApi(this._executor);

  final GatewayExecutor _executor;

  Future<SardCustomer?> sardCustomer({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardCustomer(\$id: String!) { sardCustomer(id: \$id) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardCustomer'] == null ? null : SardCustomer.fromJson((data['sardCustomer'] as Map).cast<String, dynamic>());
  }

  Future<SardCustomer?> sardCustomerByUser({required int userId, String? selectionSet}) async {
    final data = await _executor.query('query sardCustomerByUser(\$userId: Int!) { sardCustomerByUser(userId: \$userId) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'userId': userId,
    }
    );
    return data['sardCustomerByUser'] == null ? null : SardCustomer.fromJson((data['sardCustomerByUser'] as Map).cast<String, dynamic>());
  }

  Future<SardCustomer?> sardOnboardCustomer({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardOnboardCustomer(\$input: Map!) { sardOnboardCustomer(input: \$input) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'input': input,
    }
    );
    return data['sardOnboardCustomer'] == null ? null : SardCustomer.fromJson((data['sardOnboardCustomer'] as Map).cast<String, dynamic>());
  }

  Future<SardCustomer?> sardSyncCustomer({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardSyncCustomer(\$id: String!) { sardSyncCustomer(id: \$id) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardSyncCustomer'] == null ? null : SardCustomer.fromJson((data['sardSyncCustomer'] as Map).cast<String, dynamic>());
  }

  Future<SardCustomer?> sardSubmitKyc({required String id, Map<String, dynamic>? input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardSubmitKyc(\$id: String!, \$input: Map) { sardSubmitKyc(id: \$id, input: \$input) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardSubmitKyc'] == null ? null : SardCustomer.fromJson((data['sardSubmitKyc'] as Map).cast<String, dynamic>());
  }

  Future<SardCustomer?> sardRefreshKyc({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardRefreshKyc(\$id: String!) { sardRefreshKyc(id: \$id) ${selectionSet ?? '{ id user_id username full_name email phone address gender status kyc { status face_verified document { type source number_last4 full_name date_of_birth gender nationality expiry_date checksum_valid nfc_verified confidence } reasons submitted_at verified_at } card_eligibility { eligible reasons } created_at updated_at synced_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardRefreshKyc'] == null ? null : SardCustomer.fromJson((data['sardRefreshKyc'] as Map).cast<String, dynamic>());
  }

}
