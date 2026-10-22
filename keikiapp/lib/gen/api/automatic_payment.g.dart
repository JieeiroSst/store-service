// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class AutomaticPaymentApi {
  const AutomaticPaymentApi(this._executor);

  final GatewayExecutor _executor;

  Future<AutoPaySubscription?> autoPaySubscription({required String id, String? selectionSet}) async {
    final data = await _executor.query('query autoPaySubscription(\$id: String!) { autoPaySubscription(id: \$id) ${selectionSet ?? '{ subscriptionId userId planId status amount currency startDate endDate autoRenewal trialEndDate nextBillingDate paymentFailureCount createdAt updatedAt }'} }', variables: {
      'id': id,
    }
    );
    return data['autoPaySubscription'] == null ? null : AutoPaySubscription.fromJson((data['autoPaySubscription'] as Map).cast<String, dynamic>());
  }

  Future<List<AutoPayTransaction>?> autoPaySubscriptionTransactions({required String id, String? selectionSet}) async {
    final data = await _executor.query('query autoPaySubscriptionTransactions(\$id: String!) { autoPaySubscriptionTransactions(id: \$id) ${selectionSet ?? '{ transactionId subscriptionId paymentMethodId amount currency status gatewayTransactionId errorMessage createdAt }'} }', variables: {
      'id': id,
    }
    );
    return (data['autoPaySubscriptionTransactions'] as List?)?.map((e) => (e == null ? null : AutoPayTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<AutoPayInvoice>?> autoPaySubscriptionInvoices({required String id, String? selectionSet}) async {
    final data = await _executor.query('query autoPaySubscriptionInvoices(\$id: String!) { autoPaySubscriptionInvoices(id: \$id) ${selectionSet ?? '{ invoiceId transactionId subscriptionId userId amount taxAmount totalAmount status dueDate paidDate invoiceNumber createdAt }'} }', variables: {
      'id': id,
    }
    );
    return (data['autoPaySubscriptionInvoices'] as List?)?.map((e) => (e == null ? null : AutoPayInvoice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<AutoPayPaymentMethod>?> autoPayPaymentMethods({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query autoPayPaymentMethods(\$userId: String!) { autoPayPaymentMethods(userId: \$userId) ${selectionSet ?? '{ paymentMethodId userId provider tokenId lastFourDigits expiryDate isDefault createdAt updatedAt }'} }', variables: {
      'userId': userId,
    }
    );
    return (data['autoPayPaymentMethods'] as List?)?.map((e) => (e == null ? null : AutoPayPaymentMethod.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<List<AutoPayInvoice>?> autoPayInvoices({required String userId, String? selectionSet}) async {
    final data = await _executor.query('query autoPayInvoices(\$userId: String!) { autoPayInvoices(userId: \$userId) ${selectionSet ?? '{ invoiceId transactionId subscriptionId userId amount taxAmount totalAmount status dueDate paidDate invoiceNumber createdAt }'} }', variables: {
      'userId': userId,
    }
    );
    return (data['autoPayInvoices'] as List?)?.map((e) => (e == null ? null : AutoPayInvoice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<AutoPayInvoice?> autoPayInvoice({required String id, String? selectionSet}) async {
    final data = await _executor.query('query autoPayInvoice(\$id: String!) { autoPayInvoice(id: \$id) ${selectionSet ?? '{ invoiceId transactionId subscriptionId userId amount taxAmount totalAmount status dueDate paidDate invoiceNumber createdAt }'} }', variables: {
      'id': id,
    }
    );
    return data['autoPayInvoice'] == null ? null : AutoPayInvoice.fromJson((data['autoPayInvoice'] as Map).cast<String, dynamic>());
  }

  Future<AutoPaySubscription?> autoPayCreateSubscription({required AutoPaySubscriptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation autoPayCreateSubscription(\$input: AutoPaySubscriptionInput!) { autoPayCreateSubscription(input: \$input) ${selectionSet ?? '{ subscriptionId userId planId status amount currency startDate endDate autoRenewal trialEndDate nextBillingDate paymentFailureCount createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['autoPayCreateSubscription'] == null ? null : AutoPaySubscription.fromJson((data['autoPayCreateSubscription'] as Map).cast<String, dynamic>());
  }

  Future<AutoPaySubscription?> autoPayCancelSubscription({required String id, String? selectionSet}) async {
    final data = await _executor.mutation('mutation autoPayCancelSubscription(\$id: String!) { autoPayCancelSubscription(id: \$id) ${selectionSet ?? '{ subscriptionId userId planId status amount currency startDate endDate autoRenewal trialEndDate nextBillingDate paymentFailureCount createdAt updatedAt }'} }', variables: {
      'id': id,
    }
    );
    return data['autoPayCancelSubscription'] == null ? null : AutoPaySubscription.fromJson((data['autoPayCancelSubscription'] as Map).cast<String, dynamic>());
  }

  Future<Object?> autoPayProcessRenewals() async {
    final data = await _executor.mutation('mutation autoPayProcessRenewals { autoPayProcessRenewals }'
    );
    return data['autoPayProcessRenewals'];
  }

  Future<AutoPayPaymentMethod?> autoPayAddPaymentMethod({required AutoPayPaymentMethodInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation autoPayAddPaymentMethod(\$input: AutoPayPaymentMethodInput!) { autoPayAddPaymentMethod(input: \$input) ${selectionSet ?? '{ paymentMethodId userId provider tokenId lastFourDigits expiryDate isDefault createdAt updatedAt }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['autoPayAddPaymentMethod'] == null ? null : AutoPayPaymentMethod.fromJson((data['autoPayAddPaymentMethod'] as Map).cast<String, dynamic>());
  }

  Future<bool?> autoPayDeletePaymentMethod({required String id}) async {
    final data = await _executor.mutation('mutation autoPayDeletePaymentMethod(\$id: String!) { autoPayDeletePaymentMethod(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['autoPayDeletePaymentMethod'] as bool?;
  }

  Future<Object?> autoPaySetDefaultPaymentMethod({required String id, required String userId}) async {
    final data = await _executor.mutation('mutation autoPaySetDefaultPaymentMethod(\$id: String!, \$userId: String!) { autoPaySetDefaultPaymentMethod(id: \$id, userId: \$userId) }', variables: {
      'id': id,
      'userId': userId,
    }
    );
    return data['autoPaySetDefaultPaymentMethod'];
  }

}
