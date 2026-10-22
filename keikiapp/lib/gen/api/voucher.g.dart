// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class VoucherApi {
  const VoucherApi(this._executor);

  final GatewayExecutor _executor;

  Future<VoucherVoucher?> voucher({required String id, String? selectionSet}) async {
    final data = await _executor.query('query voucher(\$id: String!) { voucher(id: \$id) ${selectionSet ?? '{ id merchant_id code status version }'} }', variables: {
      'id': id,
    }
    );
    return data['voucher'] == null ? null : VoucherVoucher.fromJson((data['voucher'] as Map).cast<String, dynamic>());
  }

  Future<VoucherValidationResult?> voucherValidate({required String id, String? pin, String? selectionSet}) async {
    final data = await _executor.query('query voucherValidate(\$id: String!, \$pin: String) { voucherValidate(id: \$id, pin: \$pin) ${selectionSet ?? '{ valid balance { amount currency } reason }'} }', variables: {
      'id': id,
      'pin': ?pin,
    }
    );
    return data['voucherValidate'] == null ? null : VoucherValidationResult.fromJson((data['voucherValidate'] as Map).cast<String, dynamic>());
  }

  Future<List<VoucherVoucher>?> vouchers({String? ownerId, String? selectionSet}) async {
    final data = await _executor.query('query vouchers(\$owner_id: String) { vouchers(owner_id: \$owner_id) ${selectionSet ?? '{ id merchant_id code status version }'} }', variables: {
      'owner_id': ?ownerId,
    }
    );
    return (data['vouchers'] as List?)?.map((e) => (e == null ? null : VoucherVoucher.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VoucherOrder?> voucherOrder({required String id, String? selectionSet}) async {
    final data = await _executor.query('query voucherOrder(\$id: String!) { voucherOrder(id: \$id) ${selectionSet ?? '{ id status total_amount currency }'} }', variables: {
      'id': id,
    }
    );
    return data['voucherOrder'] == null ? null : VoucherOrder.fromJson((data['voucherOrder'] as Map).cast<String, dynamic>());
  }

  Future<VoucherMerchant?> voucherMerchant({required String id, String? selectionSet}) async {
    final data = await _executor.query('query voucherMerchant(\$id: String!) { voucherMerchant(id: \$id) ${selectionSet ?? '{ id name provider_type status }'} }', variables: {
      'id': id,
    }
    );
    return data['voucherMerchant'] == null ? null : VoucherMerchant.fromJson((data['voucherMerchant'] as Map).cast<String, dynamic>());
  }

  Future<List<VoucherMerchant>?> voucherMerchants({String? selectionSet}) async {
    final data = await _executor.query('query voucherMerchants { voucherMerchants ${selectionSet ?? '{ id name provider_type status }'} }'
    );
    return (data['voucherMerchants'] as List?)?.map((e) => (e == null ? null : VoucherMerchant.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VoucherBalance?> voucherWalletBalance({required String ownerType, required String ownerId, String? selectionSet}) async {
    final data = await _executor.query('query voucherWalletBalance(\$ownerType: String!, \$ownerId: String!) { voucherWalletBalance(ownerType: \$ownerType, ownerId: \$ownerId) ${selectionSet ?? '{ amount currency }'} }', variables: {
      'ownerType': ownerType,
      'ownerId': ownerId,
    }
    );
    return data['voucherWalletBalance'] == null ? null : VoucherBalance.fromJson((data['voucherWalletBalance'] as Map).cast<String, dynamic>());
  }

  Future<List<VoucherRedemptionReport>?> voucherRedemptionReports({String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query voucherRedemptionReports(\$from: String, \$to: String) { voucherRedemptionReports(from: \$from, to: \$to) ${selectionSet ?? '{ merchant_id total_issued total_redeemed redemption_rate period_start period_end }'} }', variables: {
      'from': ?from,
      'to': ?to,
    }
    );
    return (data['voucherRedemptionReports'] as List?)?.map((e) => (e == null ? null : VoucherRedemptionReport.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<VoucherCorporateSpendReport?> voucherCorporateSpend({required String id, String? from, String? to, String? selectionSet}) async {
    final data = await _executor.query('query voucherCorporateSpend(\$id: String!, \$from: String, \$to: String) { voucherCorporateSpend(id: \$id, from: \$from, to: \$to) ${selectionSet ?? '{ corporate_id total_spent currency vouchers_sent }'} }', variables: {
      'id': id,
      'from': ?from,
      'to': ?to,
    }
    );
    return data['voucherCorporateSpend'] == null ? null : VoucherCorporateSpendReport.fromJson((data['voucherCorporateSpend'] as Map).cast<String, dynamic>());
  }

  Future<Object?> voucherIssue({required String merchantId, required String productSku, required int denominationAmount, required String currency, required int quantity, int? expiresInDays, String? idempotencyKey}) async {
    final data = await _executor.mutation('mutation voucherIssue(\$merchant_id: String!, \$product_sku: String!, \$denomination_amount: Int64!, \$currency: String!, \$quantity: Int!, \$expires_in_days: Int, \$idempotency_key: String) { voucherIssue(merchant_id: \$merchant_id, product_sku: \$product_sku, denomination_amount: \$denomination_amount, currency: \$currency, quantity: \$quantity, expires_in_days: \$expires_in_days, idempotency_key: \$idempotency_key) }', variables: {
      'merchant_id': merchantId,
      'product_sku': productSku,
      'denomination_amount': denominationAmount,
      'currency': currency,
      'quantity': quantity,
      'expires_in_days': ?expiresInDays,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherIssue'];
  }

  Future<VoucherVoucher?> voucherActivate({required String id, required String ownerType, required String ownerId, String? selectionSet}) async {
    final data = await _executor.mutation('mutation voucherActivate(\$id: String!, \$owner_type: String!, \$owner_id: String!) { voucherActivate(id: \$id, owner_type: \$owner_type, owner_id: \$owner_id) ${selectionSet ?? '{ id merchant_id code status version }'} }', variables: {
      'id': id,
      'owner_type': ownerType,
      'owner_id': ownerId,
    }
    );
    return data['voucherActivate'] == null ? null : VoucherVoucher.fromJson((data['voucherActivate'] as Map).cast<String, dynamic>());
  }

  Future<Object?> voucherRedeem({required String id, required String pin, required int amount, required String currency, String? idempotencyKey}) async {
    final data = await _executor.mutation('mutation voucherRedeem(\$id: String!, \$pin: String!, \$amount: Int64!, \$currency: String!, \$idempotency_key: String) { voucherRedeem(id: \$id, pin: \$pin, amount: \$amount, currency: \$currency, idempotency_key: \$idempotency_key) }', variables: {
      'id': id,
      'pin': pin,
      'amount': amount,
      'currency': currency,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherRedeem'];
  }

  Future<VoucherVoucher?> voucherRevoke({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation voucherRevoke(\$id: String!) { voucherRevoke(id: \$id) ${selectionSet ?? '{ id merchant_id code status version }'} }', variables: {
      'id': id,
    }
    );
    return data['voucherRevoke'] == null ? null : VoucherVoucher.fromJson((data['voucherRevoke'] as Map).cast<String, dynamic>());
  }

  Future<VoucherOrder?> voucherCreateOrder({required String buyerType, required String buyerId, required String currency, required List<VoucherOrderItemInput> items, String? idempotencyKey, String? selectionSet}) async {
    final data = await _executor.mutation('mutation voucherCreateOrder(\$buyer_type: String!, \$buyer_id: String!, \$currency: String!, \$items: [VoucherOrderItemInput!]!, \$idempotency_key: String) { voucherCreateOrder(buyer_type: \$buyer_type, buyer_id: \$buyer_id, currency: \$currency, items: \$items, idempotency_key: \$idempotency_key) ${selectionSet ?? '{ id status total_amount currency }'} }', variables: {
      'buyer_type': buyerType,
      'buyer_id': buyerId,
      'currency': currency,
      'items': items.map((e) => e.toJson()).toList(),
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherCreateOrder'] == null ? null : VoucherOrder.fromJson((data['voucherCreateOrder'] as Map).cast<String, dynamic>());
  }

  Future<Object?> voucherCheckoutOrder({required String id, required String paymentMethod, String? idempotencyKey}) async {
    final data = await _executor.mutation('mutation voucherCheckoutOrder(\$id: String!, \$payment_method: String!, \$idempotency_key: String) { voucherCheckoutOrder(id: \$id, payment_method: \$payment_method, idempotency_key: \$idempotency_key) }', variables: {
      'id': id,
      'payment_method': paymentMethod,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherCheckoutOrder'];
  }

  Future<VoucherOrder?> voucherCancelOrder({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation voucherCancelOrder(\$id: String!) { voucherCancelOrder(id: \$id) ${selectionSet ?? '{ id status total_amount currency }'} }', variables: {
      'id': id,
    }
    );
    return data['voucherCancelOrder'] == null ? null : VoucherOrder.fromJson((data['voucherCancelOrder'] as Map).cast<String, dynamic>());
  }

  Future<VoucherMerchant?> voucherRegisterMerchant({required String name_, required String providerType, Map<String, dynamic>? config, String? selectionSet}) async {
    final data = await _executor.mutation('mutation voucherRegisterMerchant(\$name: String!, \$provider_type: String!, \$config: Map) { voucherRegisterMerchant(name: \$name, provider_type: \$provider_type, config: \$config) ${selectionSet ?? '{ id name provider_type status }'} }', variables: {
      'name': name_,
      'provider_type': providerType,
      'config': ?config,
    }
    );
    return data['voucherRegisterMerchant'] == null ? null : VoucherMerchant.fromJson((data['voucherRegisterMerchant'] as Map).cast<String, dynamic>());
  }

  Future<Object?> voucherRegisterCorporate({required String name_, String? taxCode, String? contactEmail}) async {
    final data = await _executor.mutation('mutation voucherRegisterCorporate(\$name: String!, \$tax_code: String, \$contact_email: String) { voucherRegisterCorporate(name: \$name, tax_code: \$tax_code, contact_email: \$contact_email) }', variables: {
      'name': name_,
      'tax_code': ?taxCode,
      'contact_email': ?contactEmail,
    }
    );
    return data['voucherRegisterCorporate'];
  }

  Future<Object?> voucherSetCorporateBudget({required String id, required int amount, required String currency}) async {
    final data = await _executor.mutation('mutation voucherSetCorporateBudget(\$id: String!, \$amount: Int64!, \$currency: String!) { voucherSetCorporateBudget(id: \$id, amount: \$amount, currency: \$currency) }', variables: {
      'id': id,
      'amount': amount,
      'currency': currency,
    }
    );
    return data['voucherSetCorporateBudget'];
  }

  Future<Object?> voucherCreateDistributionJob({required String corporateId, required String merchantId, required String productSku, required int denominationAmount, required String currency, required List<String> recipients}) async {
    final data = await _executor.mutation('mutation voucherCreateDistributionJob(\$corporate_id: String!, \$merchant_id: String!, \$product_sku: String!, \$denomination_amount: Int64!, \$currency: String!, \$recipients: [String!]!) { voucherCreateDistributionJob(corporate_id: \$corporate_id, merchant_id: \$merchant_id, product_sku: \$product_sku, denomination_amount: \$denomination_amount, currency: \$currency, recipients: \$recipients) }', variables: {
      'corporate_id': corporateId,
      'merchant_id': merchantId,
      'product_sku': productSku,
      'denomination_amount': denominationAmount,
      'currency': currency,
      'recipients': recipients,
    }
    );
    return data['voucherCreateDistributionJob'];
  }

  Future<Object?> voucherInitPayment({required String orderId, required int amount, required String currency, required String provider, String? returnUrl, String? idempotencyKey}) async {
    final data = await _executor.mutation('mutation voucherInitPayment(\$order_id: String!, \$amount: Int64!, \$currency: String!, \$provider: String!, \$return_url: String, \$idempotency_key: String) { voucherInitPayment(order_id: \$order_id, amount: \$amount, currency: \$currency, provider: \$provider, return_url: \$return_url, idempotency_key: \$idempotency_key) }', variables: {
      'order_id': orderId,
      'amount': amount,
      'currency': currency,
      'provider': provider,
      'return_url': ?returnUrl,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherInitPayment'];
  }

  Future<Object?> voucherImportStock({required String id, required GqlUpload file}) async {
    final data = await _executor.mutation('mutation voucherImportStock(\$id: String!, \$file: Upload!) { voucherImportStock(id: \$id, file: \$file) }', variables: {
      'id': id,
      'file': file,
    }
    );
    return data['voucherImportStock'];
  }

  Future<Object?> voucherCreditWallet({required String ownerType, required String ownerId, required int amount, required String currency, String? refType, String? refId, String? idempotencyKey}) async {
    final data = await _executor.mutation('mutation voucherCreditWallet(\$ownerType: String!, \$ownerId: String!, \$amount: Int64!, \$currency: String!, \$ref_type: String, \$ref_id: String, \$idempotency_key: String) { voucherCreditWallet(ownerType: \$ownerType, ownerId: \$ownerId, amount: \$amount, currency: \$currency, ref_type: \$ref_type, ref_id: \$ref_id, idempotency_key: \$idempotency_key) }', variables: {
      'ownerType': ownerType,
      'ownerId': ownerId,
      'amount': amount,
      'currency': currency,
      'ref_type': ?refType,
      'ref_id': ?refId,
      'idempotency_key': ?idempotencyKey,
    }
    );
    return data['voucherCreditWallet'];
  }

  Future<Object?> voucherClaim({required String token}) async {
    final data = await _executor.public('mutation voucherClaim(\$token: String!) { voucherClaim(token: \$token) }', variables: {
      'token': token,
    }
    );
    return data['voucherClaim'];
  }

}
