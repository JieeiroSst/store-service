// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class SardCardApi {
  const SardCardApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<SardProgram>?> sardPrograms({String? selectionSet}) async {
    final data = await _executor.query('query sardPrograms { sardPrograms ${selectionSet ?? '{ code name mode bin currency validity_years card_types max_active_virtual default_credit_limit max_credit_limit default_limits { per_transaction daily } max_limits { per_transaction daily } supported_channels { pos contactless ecommerce atm } step_up_threshold }'} }'
    );
    return (data['sardPrograms'] as List?)?.map((e) => (e == null ? null : SardProgram.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SardAccount?> sardAccount({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardAccount(\$id: String!) { sardAccount(id: \$id) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardAccount'] == null ? null : SardAccount.fromJson((data['sardAccount'] as Map).cast<String, dynamic>());
  }

  Future<List<SardAccount>?> sardCustomerAccounts({required String customerId, String? selectionSet}) async {
    final data = await _executor.query('query sardCustomerAccounts(\$customerId: String!) { sardCustomerAccounts(customerId: \$customerId) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'customerId': customerId,
    }
    );
    return (data['sardCustomerAccounts'] as List?)?.map((e) => (e == null ? null : SardAccount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<SardTransaction>?> sardAccountTransactions({required String id, Map<String, dynamic>? query, String? selectionSet}) async {
    final data = await _executor.query('query sardAccountTransactions(\$id: String!, \$query: Map) { sardAccountTransactions(id: \$id, query: \$query) ${selectionSet ?? '{ id card_id authorization_id type processing_code amount balance_after description created_at }'} }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return (data['sardAccountTransactions'] as List?)?.map((e) => (e == null ? null : SardTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<SardCard>?> sardAccountCards({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardAccountCards(\$id: String!) { sardAccountCards(id: \$id) ${selectionSet ?? '{ id account_id customer_id program_code type cardholder_name masked_pan last4 expiry valid_until status status_reason limits { per_transaction daily } controls { pos contactless ecommerce atm } spent_today pin_set pin_locked replaces_card_id replaced_by_card_id created_at updated_at activated_at }'} }', variables: {
      'id': id,
    }
    );
    return (data['sardAccountCards'] as List?)?.map((e) => (e == null ? null : SardCard.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SardCard?> sardCard({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardCard(\$id: String!) { sardCard(id: \$id) ${selectionSet ?? '{ id account_id customer_id program_code type cardholder_name masked_pan last4 expiry valid_until status status_reason limits { per_transaction daily } controls { pos contactless ecommerce atm } spent_today pin_set pin_locked replaces_card_id replaced_by_card_id created_at updated_at activated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardCard'] == null ? null : SardCard.fromJson((data['sardCard'] as Map).cast<String, dynamic>());
  }

  Future<List<SardAuthorization>?> sardCardAuthorizations({required String id, Map<String, dynamic>? query, String? selectionSet}) async {
    final data = await _executor.query('query sardCardAuthorizations(\$id: String!, \$query: Map) { sardCardAuthorizations(id: \$id, query: \$query) ${selectionSet ?? '{ id account_id card_id processing_code status approved response_code reason auth_code amount confirmed_amount currency channel merchant mcc authentication_id created_at updated_at }'} }', variables: {
      'id': id,
      'query': ?query,
    }
    );
    return (data['sardCardAuthorizations'] as List?)?.map((e) => (e == null ? null : SardAuthorization.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<SardAuthorization?> sardAuthorization({required String id, String? selectionSet}) async {
    final data = await _executor.query('query sardAuthorization(\$id: String!) { sardAuthorization(id: \$id) ${selectionSet ?? '{ id account_id card_id processing_code status approved response_code reason auth_code amount confirmed_amount currency channel merchant mcc authentication_id created_at updated_at }'} }', variables: {
      'id': id,
    }
    );
    return data['sardAuthorization'] == null ? null : SardAuthorization.fromJson((data['sardAuthorization'] as Map).cast<String, dynamic>());
  }

  Future<SardAccount?> sardOpenAccount({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardOpenAccount(\$input: Map!) { sardOpenAccount(input: \$input) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'input': input,
    }
    );
    return data['sardOpenAccount'] == null ? null : SardAccount.fromJson((data['sardOpenAccount'] as Map).cast<String, dynamic>());
  }

  Future<SardAccount?> sardBlockAccount({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardBlockAccount(\$id: String!, \$reason: String) { sardBlockAccount(id: \$id, reason: \$reason) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['sardBlockAccount'] == null ? null : SardAccount.fromJson((data['sardBlockAccount'] as Map).cast<String, dynamic>());
  }

  Future<SardAccount?> sardUnblockAccount({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardUnblockAccount(\$id: String!, \$reason: String) { sardUnblockAccount(id: \$id, reason: \$reason) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['sardUnblockAccount'] == null ? null : SardAccount.fromJson((data['sardUnblockAccount'] as Map).cast<String, dynamic>());
  }

  Future<SardAccount?> sardCancelAccount({required String id, String? reason, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardCancelAccount(\$id: String!, \$reason: String) { sardCancelAccount(id: \$id, reason: \$reason) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'id': id,
      'reason': ?reason,
    }
    );
    return data['sardCancelAccount'] == null ? null : SardAccount.fromJson((data['sardCancelAccount'] as Map).cast<String, dynamic>());
  }

  Future<SardAccount?> sardSetCreditLimit({required String id, required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardSetCreditLimit(\$id: String!, \$input: Map!) { sardSetCreditLimit(id: \$id, input: \$input) ${selectionSet ?? '{ id customer_id program_code mode currency holder_name status status_reason credit_limit balance held available created_at updated_at }'} }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['sardSetCreditLimit'] == null ? null : SardAccount.fromJson((data['sardSetCreditLimit'] as Map).cast<String, dynamic>());
  }

  Future<Object?> sardAccountPayment({required String id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation sardAccountPayment(\$id: String!, \$input: Map!) { sardAccountPayment(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['sardAccountPayment'];
  }

  Future<SardIssuedCard?> sardIssueCard({required String id, required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardIssueCard(\$id: String!, \$input: Map!) { sardIssueCard(id: \$id, input: \$input) ${selectionSet ?? '{ card { id account_id customer_id program_code type cardholder_name masked_pan last4 expiry valid_until status status_reason limits { per_transaction daily } controls { pos contactless ecommerce atm } spent_today pin_set pin_locked replaces_card_id replaced_by_card_id created_at updated_at activated_at } pan cvv notice }'} }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['sardIssueCard'] == null ? null : SardIssuedCard.fromJson((data['sardIssueCard'] as Map).cast<String, dynamic>());
  }

  Future<Object?> sardActivateCard({required String id, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation sardActivateCard(\$id: String!, \$input: Map) { sardActivateCard(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardActivateCard'];
  }

  Future<Object?> sardBlockCard({required String id, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation sardBlockCard(\$id: String!, \$input: Map) { sardBlockCard(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardBlockCard'];
  }

  Future<Object?> sardUnblockCard({required String id, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation sardUnblockCard(\$id: String!, \$input: Map) { sardUnblockCard(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardUnblockCard'];
  }

  Future<Object?> sardCancelCard({required String id, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation sardCancelCard(\$id: String!, \$input: Map) { sardCancelCard(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardCancelCard'];
  }

  Future<Object?> sardReportCard({required String id, Map<String, dynamic>? input}) async {
    final data = await _executor.mutation('mutation sardReportCard(\$id: String!, \$input: Map) { sardReportCard(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardReportCard'];
  }

  Future<SardCard?> sardUpdateCardLimits({required String id, int? perTransaction, int? daily, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardUpdateCardLimits(\$id: String!, \$per_transaction: Int64, \$daily: Int64) { sardUpdateCardLimits(id: \$id, per_transaction: \$per_transaction, daily: \$daily) ${selectionSet ?? '{ id account_id customer_id program_code type cardholder_name masked_pan last4 expiry valid_until status status_reason limits { per_transaction daily } controls { pos contactless ecommerce atm } spent_today pin_set pin_locked replaces_card_id replaced_by_card_id created_at updated_at activated_at }'} }', variables: {
      'id': id,
      'per_transaction': ?perTransaction,
      'daily': ?daily,
    }
    );
    return data['sardUpdateCardLimits'] == null ? null : SardCard.fromJson((data['sardUpdateCardLimits'] as Map).cast<String, dynamic>());
  }

  Future<SardCard?> sardUpdateCardControls({required String id, bool? pos, bool? contactless, bool? ecommerce, bool? atm, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardUpdateCardControls(\$id: String!, \$pos: Boolean, \$contactless: Boolean, \$ecommerce: Boolean, \$atm: Boolean) { sardUpdateCardControls(id: \$id, pos: \$pos, contactless: \$contactless, ecommerce: \$ecommerce, atm: \$atm) ${selectionSet ?? '{ id account_id customer_id program_code type cardholder_name masked_pan last4 expiry valid_until status status_reason limits { per_transaction daily } controls { pos contactless ecommerce atm } spent_today pin_set pin_locked replaces_card_id replaced_by_card_id created_at updated_at activated_at }'} }', variables: {
      'id': id,
      'pos': ?pos,
      'contactless': ?contactless,
      'ecommerce': ?ecommerce,
      'atm': ?atm,
    }
    );
    return data['sardUpdateCardControls'] == null ? null : SardCard.fromJson((data['sardUpdateCardControls'] as Map).cast<String, dynamic>());
  }

  Future<Object?> sardSetCardPin({required String id, required Map<String, dynamic> input}) async {
    final data = await _executor.mutation('mutation sardSetCardPin(\$id: String!, \$input: Map!) { sardSetCardPin(id: \$id, input: \$input) }', variables: {
      'id': id,
      'input': input,
    }
    );
    return data['sardSetCardPin'];
  }

  Future<SardAuthorization?> sardAuthorize({required Map<String, dynamic> input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardAuthorize(\$input: Map!) { sardAuthorize(input: \$input) ${selectionSet ?? '{ id account_id card_id processing_code status approved response_code reason auth_code amount confirmed_amount currency channel merchant mcc authentication_id created_at updated_at }'} }', variables: {
      'input': input,
    }
    );
    return data['sardAuthorize'] == null ? null : SardAuthorization.fromJson((data['sardAuthorize'] as Map).cast<String, dynamic>());
  }

  Future<SardAuthorization?> sardConfirmAuthorization({required String id, Map<String, dynamic>? input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardConfirmAuthorization(\$id: String!, \$input: Map) { sardConfirmAuthorization(id: \$id, input: \$input) ${selectionSet ?? '{ id account_id card_id processing_code status approved response_code reason auth_code amount confirmed_amount currency channel merchant mcc authentication_id created_at updated_at }'} }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardConfirmAuthorization'] == null ? null : SardAuthorization.fromJson((data['sardConfirmAuthorization'] as Map).cast<String, dynamic>());
  }

  Future<SardAuthorization?> sardCancelAuthorization({required String id, Map<String, dynamic>? input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation sardCancelAuthorization(\$id: String!, \$input: Map) { sardCancelAuthorization(id: \$id, input: \$input) ${selectionSet ?? '{ id account_id card_id processing_code status approved response_code reason auth_code amount confirmed_amount currency channel merchant mcc authentication_id created_at updated_at }'} }', variables: {
      'id': id,
      'input': ?input,
    }
    );
    return data['sardCancelAuthorization'] == null ? null : SardAuthorization.fromJson((data['sardCancelAuthorization'] as Map).cast<String, dynamic>());
  }

}
