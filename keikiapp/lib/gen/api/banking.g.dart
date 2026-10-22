// GENERATED CODE - DO NOT MODIFY BY HAND. Run: dart run tool/gen_gateway.dart
// ignore_for_file: type=lint

import '../gateway_executor.dart';
import '../models.g.dart';

class BankingApi {
  const BankingApi(this._executor);

  final GatewayExecutor _executor;

  Future<List<BankingPerson>?> bankingPersons({String? selectionSet}) async {
    final data = await _executor.query('query bankingPersons { bankingPersons ${selectionSet ?? '{ person_id last_name first_name date_of_birth email phone_number address tax_identifier }'} }'
    );
    return (data['bankingPersons'] as List?)?.map((e) => (e == null ? null : BankingPerson.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingPerson?> bankingPerson({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingPerson(\$id: Int!) { bankingPerson(id: \$id) ${selectionSet ?? '{ person_id last_name first_name date_of_birth email phone_number address tax_identifier }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingPerson'] == null ? null : BankingPerson.fromJson((data['bankingPerson'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingBranch>?> bankingBranches({String? selectionSet}) async {
    final data = await _executor.query('query bankingBranches { bankingBranches ${selectionSet ?? '{ branch_id branch_name branch_code address phone_number }'} }'
    );
    return (data['bankingBranches'] as List?)?.map((e) => (e == null ? null : BankingBranch.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingBranch?> bankingBranch({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingBranch(\$id: Int!) { bankingBranch(id: \$id) ${selectionSet ?? '{ branch_id branch_name branch_code address phone_number }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingBranch'] == null ? null : BankingBranch.fromJson((data['bankingBranch'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingCustomer>?> bankingCustomers({String? selectionSet}) async {
    final data = await _executor.query('query bankingCustomers { bankingCustomers ${selectionSet ?? '{ customer_id customer_type }'} }'
    );
    return (data['bankingCustomers'] as List?)?.map((e) => (e == null ? null : BankingCustomer.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingCustomer?> bankingCustomer({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingCustomer(\$id: Int!) { bankingCustomer(id: \$id) ${selectionSet ?? '{ customer_id customer_type }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingCustomer'] == null ? null : BankingCustomer.fromJson((data['bankingCustomer'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingEmployee>?> bankingEmployees({String? selectionSet}) async {
    final data = await _executor.query('query bankingEmployees { bankingEmployees ${selectionSet ?? '{ employee_id position }'} }'
    );
    return (data['bankingEmployees'] as List?)?.map((e) => (e == null ? null : BankingEmployee.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingEmployee?> bankingEmployee({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingEmployee(\$id: Int!) { bankingEmployee(id: \$id) ${selectionSet ?? '{ employee_id position }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingEmployee'] == null ? null : BankingEmployee.fromJson((data['bankingEmployee'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingAccount>?> bankingAccounts({String? selectionSet}) async {
    final data = await _executor.query('query bankingAccounts { bankingAccounts ${selectionSet ?? '{ account_id account_number account_type current_balance date_opened date_closed account_status }'} }'
    );
    return (data['bankingAccounts'] as List?)?.map((e) => (e == null ? null : BankingAccount.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingAccount?> bankingAccount({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingAccount(\$id: Int!) { bankingAccount(id: \$id) ${selectionSet ?? '{ account_id account_number account_type current_balance date_opened date_closed account_status }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingAccount'] == null ? null : BankingAccount.fromJson((data['bankingAccount'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingLoan>?> bankingLoans({String? selectionSet}) async {
    final data = await _executor.query('query bankingLoans { bankingLoans ${selectionSet ?? '{ loan_id loan_type loan_amount interes_rate term start_date end_date loan_status }'} }'
    );
    return (data['bankingLoans'] as List?)?.map((e) => (e == null ? null : BankingLoan.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingLoan?> bankingLoan({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingLoan(\$id: Int!) { bankingLoan(id: \$id) ${selectionSet ?? '{ loan_id loan_type loan_amount interes_rate term start_date end_date loan_status }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingLoan'] == null ? null : BankingLoan.fromJson((data['bankingLoan'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingLoanPayment>?> bankingLoanPayments({String? selectionSet}) async {
    final data = await _executor.query('query bankingLoanPayments { bankingLoanPayments ${selectionSet ?? '{ loan_payment_id scheduled_payment_date payment_account principal_amount interest_amount paid_amount paid_date }'} }'
    );
    return (data['bankingLoanPayments'] as List?)?.map((e) => (e == null ? null : BankingLoanPayment.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingLoanPayment?> bankingLoanPayment({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingLoanPayment(\$id: Int!) { bankingLoanPayment(id: \$id) ${selectionSet ?? '{ loan_payment_id scheduled_payment_date payment_account principal_amount interest_amount paid_amount paid_date }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingLoanPayment'] == null ? null : BankingLoanPayment.fromJson((data['bankingLoanPayment'] as Map).cast<String, dynamic>());
  }

  Future<List<BankingTransaction>?> bankingTransactions({String? selectionSet}) async {
    final data = await _executor.query('query bankingTransactions { bankingTransactions ${selectionSet ?? '{ transaction_id account_id external_ref transaction_type amount transaction_date }'} }'
    );
    return (data['bankingTransactions'] as List?)?.map((e) => (e == null ? null : BankingTransaction.fromJson((e as Map).cast<String, dynamic>()))!).toList();
  }

  Future<BankingTransaction?> bankingTransaction({required int id, String? selectionSet}) async {
    final data = await _executor.query('query bankingTransaction(\$id: Int!) { bankingTransaction(id: \$id) ${selectionSet ?? '{ transaction_id account_id external_ref transaction_type amount transaction_date }'} }', variables: {
      'id': id,
    }
    );
    return data['bankingTransaction'] == null ? null : BankingTransaction.fromJson((data['bankingTransaction'] as Map).cast<String, dynamic>());
  }

  Future<BankingPerson?> bankingCreatePerson({required BankingPersonInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreatePerson(\$input: BankingPersonInput!) { bankingCreatePerson(input: \$input) ${selectionSet ?? '{ person_id last_name first_name date_of_birth email phone_number address tax_identifier }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreatePerson'] == null ? null : BankingPerson.fromJson((data['bankingCreatePerson'] as Map).cast<String, dynamic>());
  }

  Future<BankingPerson?> bankingUpdatePerson({required int id, required BankingPersonInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdatePerson(\$id: Int!, \$input: BankingPersonInput!) { bankingUpdatePerson(id: \$id, input: \$input) ${selectionSet ?? '{ person_id last_name first_name date_of_birth email phone_number address tax_identifier }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdatePerson'] == null ? null : BankingPerson.fromJson((data['bankingUpdatePerson'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeletePerson({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeletePerson(\$id: Int!) { bankingDeletePerson(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeletePerson'] as bool?;
  }

  Future<BankingBranch?> bankingCreateBranch({required BankingBranchInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateBranch(\$input: BankingBranchInput!) { bankingCreateBranch(input: \$input) ${selectionSet ?? '{ branch_id branch_name branch_code address phone_number }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateBranch'] == null ? null : BankingBranch.fromJson((data['bankingCreateBranch'] as Map).cast<String, dynamic>());
  }

  Future<BankingBranch?> bankingUpdateBranch({required int id, required BankingBranchInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateBranch(\$id: Int!, \$input: BankingBranchInput!) { bankingUpdateBranch(id: \$id, input: \$input) ${selectionSet ?? '{ branch_id branch_name branch_code address phone_number }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateBranch'] == null ? null : BankingBranch.fromJson((data['bankingUpdateBranch'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteBranch({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteBranch(\$id: Int!) { bankingDeleteBranch(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteBranch'] as bool?;
  }

  Future<BankingCustomer?> bankingCreateCustomer({required BankingCustomerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateCustomer(\$input: BankingCustomerInput!) { bankingCreateCustomer(input: \$input) ${selectionSet ?? '{ customer_id customer_type }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateCustomer'] == null ? null : BankingCustomer.fromJson((data['bankingCreateCustomer'] as Map).cast<String, dynamic>());
  }

  Future<BankingCustomer?> bankingUpdateCustomer({required int id, required BankingCustomerInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateCustomer(\$id: Int!, \$input: BankingCustomerInput!) { bankingUpdateCustomer(id: \$id, input: \$input) ${selectionSet ?? '{ customer_id customer_type }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateCustomer'] == null ? null : BankingCustomer.fromJson((data['bankingUpdateCustomer'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteCustomer({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteCustomer(\$id: Int!) { bankingDeleteCustomer(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteCustomer'] as bool?;
  }

  Future<BankingEmployee?> bankingCreateEmployee({required BankingEmployeeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateEmployee(\$input: BankingEmployeeInput!) { bankingCreateEmployee(input: \$input) ${selectionSet ?? '{ employee_id position }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateEmployee'] == null ? null : BankingEmployee.fromJson((data['bankingCreateEmployee'] as Map).cast<String, dynamic>());
  }

  Future<BankingEmployee?> bankingUpdateEmployee({required int id, required BankingEmployeeInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateEmployee(\$id: Int!, \$input: BankingEmployeeInput!) { bankingUpdateEmployee(id: \$id, input: \$input) ${selectionSet ?? '{ employee_id position }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateEmployee'] == null ? null : BankingEmployee.fromJson((data['bankingUpdateEmployee'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteEmployee({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteEmployee(\$id: Int!) { bankingDeleteEmployee(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteEmployee'] as bool?;
  }

  Future<BankingAccount?> bankingCreateAccount({required BankingAccountInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateAccount(\$input: BankingAccountInput!) { bankingCreateAccount(input: \$input) ${selectionSet ?? '{ account_id account_number account_type current_balance date_opened date_closed account_status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateAccount'] == null ? null : BankingAccount.fromJson((data['bankingCreateAccount'] as Map).cast<String, dynamic>());
  }

  Future<BankingAccount?> bankingUpdateAccount({required int id, required BankingAccountInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateAccount(\$id: Int!, \$input: BankingAccountInput!) { bankingUpdateAccount(id: \$id, input: \$input) ${selectionSet ?? '{ account_id account_number account_type current_balance date_opened date_closed account_status }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateAccount'] == null ? null : BankingAccount.fromJson((data['bankingUpdateAccount'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteAccount({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteAccount(\$id: Int!) { bankingDeleteAccount(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteAccount'] as bool?;
  }

  Future<BankingLoan?> bankingCreateLoan({required BankingLoanInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateLoan(\$input: BankingLoanInput!) { bankingCreateLoan(input: \$input) ${selectionSet ?? '{ loan_id loan_type loan_amount interes_rate term start_date end_date loan_status }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateLoan'] == null ? null : BankingLoan.fromJson((data['bankingCreateLoan'] as Map).cast<String, dynamic>());
  }

  Future<BankingLoan?> bankingUpdateLoan({required int id, required BankingLoanInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateLoan(\$id: Int!, \$input: BankingLoanInput!) { bankingUpdateLoan(id: \$id, input: \$input) ${selectionSet ?? '{ loan_id loan_type loan_amount interes_rate term start_date end_date loan_status }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateLoan'] == null ? null : BankingLoan.fromJson((data['bankingUpdateLoan'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteLoan({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteLoan(\$id: Int!) { bankingDeleteLoan(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteLoan'] as bool?;
  }

  Future<BankingLoanPayment?> bankingCreateLoanPayment({required BankingLoanPaymentInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateLoanPayment(\$input: BankingLoanPaymentInput!) { bankingCreateLoanPayment(input: \$input) ${selectionSet ?? '{ loan_payment_id scheduled_payment_date payment_account principal_amount interest_amount paid_amount paid_date }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateLoanPayment'] == null ? null : BankingLoanPayment.fromJson((data['bankingCreateLoanPayment'] as Map).cast<String, dynamic>());
  }

  Future<BankingLoanPayment?> bankingUpdateLoanPayment({required int id, required BankingLoanPaymentInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateLoanPayment(\$id: Int!, \$input: BankingLoanPaymentInput!) { bankingUpdateLoanPayment(id: \$id, input: \$input) ${selectionSet ?? '{ loan_payment_id scheduled_payment_date payment_account principal_amount interest_amount paid_amount paid_date }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateLoanPayment'] == null ? null : BankingLoanPayment.fromJson((data['bankingUpdateLoanPayment'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteLoanPayment({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteLoanPayment(\$id: Int!) { bankingDeleteLoanPayment(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteLoanPayment'] as bool?;
  }

  Future<BankingTransaction?> bankingCreateTransaction({required BankingTransactionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingCreateTransaction(\$input: BankingTransactionInput!) { bankingCreateTransaction(input: \$input) ${selectionSet ?? '{ transaction_id account_id external_ref transaction_type amount transaction_date }'} }', variables: {
      'input': input.toJson(),
    }
    );
    return data['bankingCreateTransaction'] == null ? null : BankingTransaction.fromJson((data['bankingCreateTransaction'] as Map).cast<String, dynamic>());
  }

  Future<BankingTransaction?> bankingUpdateTransaction({required int id, required BankingTransactionInput input, String? selectionSet}) async {
    final data = await _executor.mutation('mutation bankingUpdateTransaction(\$id: Int!, \$input: BankingTransactionInput!) { bankingUpdateTransaction(id: \$id, input: \$input) ${selectionSet ?? '{ transaction_id account_id external_ref transaction_type amount transaction_date }'} }', variables: {
      'id': id,
      'input': input.toJson(),
    }
    );
    return data['bankingUpdateTransaction'] == null ? null : BankingTransaction.fromJson((data['bankingUpdateTransaction'] as Map).cast<String, dynamic>());
  }

  Future<bool?> bankingDeleteTransaction({required int id}) async {
    final data = await _executor.mutation('mutation bankingDeleteTransaction(\$id: Int!) { bankingDeleteTransaction(id: \$id) }', variables: {
      'id': id,
    }
    );
    return data['bankingDeleteTransaction'] as bool?;
  }

}
