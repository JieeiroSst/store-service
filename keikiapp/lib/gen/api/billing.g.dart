// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BillingApi {
  const BillingApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BillingAddress>?> billingAddresses({String? selectionSet}) async {
    final data = await _executor.query('query billingAddresses { billingAddresses ${selectionSet ?? '{ address_id line1 line2 city state postal_code country }'} }'
    );
    return (data['billingAddresses'] as List?)?.map((e) => (e == null ? null : BillingAddress.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingAddress?> billingAddress({required int id, String? selectionSet}) async {
    final data = await _executor.query('query billingAddress(\$id: Int!) { billingAddress(id: \$id) ${selectionSet ?? '{ address_id line1 line2 city state postal_code country }'} }', variables: {
      'id': id,
    }
    );
    return data['billingAddress'] == null ? null : BillingAddress.fromJson((data['billingAddress'] as Map).cast<String, dynamic>());
  }

  Future<List<BillingCustomer>?> billingCustomers({String? selectionSet}) async {
    final data = await _executor.query('query billingCustomers { billingCustomers ${selectionSet ?? '{ customer_id name company email phone_number address_id }'} }'
    );
    return (data['billingCustomers'] as List?)?.map((e) => (e == null ? null : BillingCustomer.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingCustomer?> billingCustomer({required int id, String? selectionSet}) async {
    final data = await _executor.query('query billingCustomer(\$id: Int!) { billingCustomer(id: \$id) ${selectionSet ?? '{ customer_id name company email phone_number address_id }'} }', variables: {
      'id': id,
    }
    );
    return data['billingCustomer'] == null ? null : BillingCustomer.fromJson((data['billingCustomer'] as Map).cast<String, dynamic>());
  }

  Future<List<BillingPlan>?> billingPlans({String? selectionSet}) async {
    final data = await _executor.public('query billingPlans { billingPlans ${selectionSet ?? '{ plan_id name description price billing_cycle }'} }'
    );
    return (data['billingPlans'] as List?)?.map((e) => (e == null ? null : BillingPlan.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingPlan?> billingPlan({required int id, String? selectionSet}) async {
    final data = await _executor.public('query billingPlan(\$id: Int!) { billingPlan(id: \$id) ${selectionSet ?? '{ plan_id name description price billing_cycle }'} }', variables: {
      'id': id,
    }
    );
    return data['billingPlan'] == null ? null : BillingPlan.fromJson((data['billingPlan'] as Map).cast<String, dynamic>());
  }

  Future<List<BillingSubscription>?> billingSubscriptions({String? selectionSet}) async {
    final data = await _executor.query('query billingSubscriptions { billingSubscriptions ${selectionSet ?? '{ subscription_id customer_id plan_id start_date end_date status }'} }'
    );
    return (data['billingSubscriptions'] as List?)?.map((e) => (e == null ? null : BillingSubscription.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingSubscription?> billingSubscription({required int id, String? selectionSet}) async {
    final data = await _executor.query('query billingSubscription(\$id: Int!) { billingSubscription(id: \$id) ${selectionSet ?? '{ subscription_id customer_id plan_id start_date end_date status }'} }', variables: {
      'id': id,
    }
    );
    return data['billingSubscription'] == null ? null : BillingSubscription.fromJson((data['billingSubscription'] as Map).cast<String, dynamic>());
  }

  Future<List<BillingInvoice>?> billingInvoices({String? selectionSet}) async {
    final data = await _executor.query('query billingInvoices { billingInvoices ${selectionSet ?? '{ invoice_id subscription_id invoice_date due_date amount tax status }'} }'
    );
    return (data['billingInvoices'] as List?)?.map((e) => (e == null ? null : BillingInvoice.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingInvoice?> billingInvoice({required int id, String? selectionSet}) async {
    final data = await _executor.query('query billingInvoice(\$id: Int!) { billingInvoice(id: \$id) ${selectionSet ?? '{ invoice_id subscription_id invoice_date due_date amount tax status }'} }', variables: {
      'id': id,
    }
    );
    return data['billingInvoice'] == null ? null : BillingInvoice.fromJson((data['billingInvoice'] as Map).cast<String, dynamic>());
  }

  Future<List<BillingTransaction>?> billingTransactions({String? selectionSet}) async {
    final data = await _executor.query('query billingTransactions { billingTransactions ${selectionSet ?? '{ transaction_id invoice_id payment_method transaction_date amount status }'} }'
    );
    return (data['billingTransactions'] as List?)?.map((e) => (e == null ? null : BillingTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BillingTransaction?> billingTransaction({required int id, String? selectionSet}) async {
    final data = await _executor.query('query billingTransaction(\$id: Int!) { billingTransaction(id: \$id) ${selectionSet ?? '{ transaction_id invoice_id payment_method transaction_date amount status }'} }', variables: {
      'id': id,
    }
    );
    return data['billingTransaction'] == null ? null : BillingTransaction.fromJson((data['billingTransaction'] as Map).cast<String, dynamic>());
  }

  Future<BillingAddress?> billingCreateAddress({required BillingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreateAddress(\$input: BillingAddressInput!) { billingCreateAddress(input: \$input) ${selectionSet ?? '{ address_id line1 line2 city state postal_code country }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreateAddress'] == null ? null : BillingAddress.fromJson((data['billingCreateAddress'] as Map).cast<String, dynamic>());
  }

  Future<BillingAddress?> billingUpdateAddress({required int id, required BillingAddressInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdateAddress(\$id: Int!, \$input: BillingAddressInput!) { billingUpdateAddress(id: \$id, input: \$input) ${selectionSet ?? '{ address_id line1 line2 city state postal_code country }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdateAddress'] == null ? null : BillingAddress.fromJson((data['billingUpdateAddress'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeleteAddress({required int id}) async {
    final data = await _executor.mutation('mutation billingDeleteAddress(\$id: Int!) { billingDeleteAddress(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeleteAddress'] as bool?;
  }

  Future<BillingCustomer?> billingCreateCustomer({required BillingCustomerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreateCustomer(\$input: BillingCustomerInput!) { billingCreateCustomer(input: \$input) ${selectionSet ?? '{ customer_id name company email phone_number address_id }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreateCustomer'] == null ? null : BillingCustomer.fromJson((data['billingCreateCustomer'] as Map).cast<String, dynamic>());
  }

  Future<BillingCustomer?> billingUpdateCustomer({required int id, required BillingCustomerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdateCustomer(\$id: Int!, \$input: BillingCustomerInput!) { billingUpdateCustomer(id: \$id, input: \$input) ${selectionSet ?? '{ customer_id name company email phone_number address_id }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdateCustomer'] == null ? null : BillingCustomer.fromJson((data['billingUpdateCustomer'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeleteCustomer({required int id}) async {
    final data = await _executor.mutation('mutation billingDeleteCustomer(\$id: Int!) { billingDeleteCustomer(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeleteCustomer'] as bool?;
  }

  Future<BillingPlan?> billingCreatePlan({required BillingPlanInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreatePlan(\$input: BillingPlanInput!) { billingCreatePlan(input: \$input) ${selectionSet ?? '{ plan_id name description price billing_cycle }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreatePlan'] == null ? null : BillingPlan.fromJson((data['billingCreatePlan'] as Map).cast<String, dynamic>());
  }

  Future<BillingPlan?> billingUpdatePlan({required int id, required BillingPlanInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdatePlan(\$id: Int!, \$input: BillingPlanInput!) { billingUpdatePlan(id: \$id, input: \$input) ${selectionSet ?? '{ plan_id name description price billing_cycle }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdatePlan'] == null ? null : BillingPlan.fromJson((data['billingUpdatePlan'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeletePlan({required int id}) async {
    final data = await _executor.mutation('mutation billingDeletePlan(\$id: Int!) { billingDeletePlan(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeletePlan'] as bool?;
  }

  Future<BillingSubscription?> billingCreateSubscription({required BillingSubscriptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreateSubscription(\$input: BillingSubscriptionInput!) { billingCreateSubscription(input: \$input) ${selectionSet ?? '{ subscription_id customer_id plan_id start_date end_date status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreateSubscription'] == null ? null : BillingSubscription.fromJson((data['billingCreateSubscription'] as Map).cast<String, dynamic>());
  }

  Future<BillingSubscription?> billingUpdateSubscription({required int id, required BillingSubscriptionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdateSubscription(\$id: Int!, \$input: BillingSubscriptionInput!) { billingUpdateSubscription(id: \$id, input: \$input) ${selectionSet ?? '{ subscription_id customer_id plan_id start_date end_date status }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdateSubscription'] == null ? null : BillingSubscription.fromJson((data['billingUpdateSubscription'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeleteSubscription({required int id}) async {
    final data = await _executor.mutation('mutation billingDeleteSubscription(\$id: Int!) { billingDeleteSubscription(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeleteSubscription'] as bool?;
  }

  Future<BillingInvoice?> billingCreateInvoice({required BillingInvoiceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreateInvoice(\$input: BillingInvoiceInput!) { billingCreateInvoice(input: \$input) ${selectionSet ?? '{ invoice_id subscription_id invoice_date due_date amount tax status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreateInvoice'] == null ? null : BillingInvoice.fromJson((data['billingCreateInvoice'] as Map).cast<String, dynamic>());
  }

  Future<BillingInvoice?> billingUpdateInvoice({required int id, required BillingInvoiceInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdateInvoice(\$id: Int!, \$input: BillingInvoiceInput!) { billingUpdateInvoice(id: \$id, input: \$input) ${selectionSet ?? '{ invoice_id subscription_id invoice_date due_date amount tax status }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdateInvoice'] == null ? null : BillingInvoice.fromJson((data['billingUpdateInvoice'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeleteInvoice({required int id}) async {
    final data = await _executor.mutation('mutation billingDeleteInvoice(\$id: Int!) { billingDeleteInvoice(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeleteInvoice'] as bool?;
  }

  Future<BillingTransaction?> billingCreateTransaction({required BillingTransactionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingCreateTransaction(\$input: BillingTransactionInput!) { billingCreateTransaction(input: \$input) ${selectionSet ?? '{ transaction_id invoice_id payment_method transaction_date amount status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['billingCreateTransaction'] == null ? null : BillingTransaction.fromJson((data['billingCreateTransaction'] as Map).cast<String, dynamic>());
  }

  Future<BillingTransaction?> billingUpdateTransaction({required int id, required BillingTransactionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation billingUpdateTransaction(\$id: Int!, \$input: BillingTransactionInput!) { billingUpdateTransaction(id: \$id, input: \$input) ${selectionSet ?? '{ transaction_id invoice_id payment_method transaction_date amount status }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['billingUpdateTransaction'] == null ? null : BillingTransaction.fromJson((data['billingUpdateTransaction'] as Map).cast<String, dynamic>());
  }

  Future<bool?> billingDeleteTransaction({required int id}) async {
    final data = await _executor.mutation('mutation billingDeleteTransaction(\$id: Int!) { billingDeleteTransaction(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['billingDeleteTransaction'] as bool?;
  }

}
