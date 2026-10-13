package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BankingCreatePerson(ctx context.Context, input model.BankingPersonInput) (*model.BankingPerson, error) {
	return restCall[*model.BankingPerson](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/persons", Body: input}, "")
}

func (r *mutationResolver) BankingUpdatePerson(ctx context.Context, id int, input model.BankingPersonInput) (*model.BankingPerson, error) {
	return restCall[*model.BankingPerson](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/persons/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeletePerson(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/persons/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateBranch(ctx context.Context, input model.BankingBranchInput) (*model.BankingBranch, error) {
	return restCall[*model.BankingBranch](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/branches", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateBranch(ctx context.Context, id int, input model.BankingBranchInput) (*model.BankingBranch, error) {
	return restCall[*model.BankingBranch](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/branches/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteBranch(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/branches/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateCustomer(ctx context.Context, input model.BankingCustomerInput) (*model.BankingCustomer, error) {
	return restCall[*model.BankingCustomer](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/customers", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateCustomer(ctx context.Context, id int, input model.BankingCustomerInput) (*model.BankingCustomer, error) {
	return restCall[*model.BankingCustomer](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/customers/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteCustomer(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/customers/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateEmployee(ctx context.Context, input model.BankingEmployeeInput) (*model.BankingEmployee, error) {
	return restCall[*model.BankingEmployee](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/employees", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateEmployee(ctx context.Context, id int, input model.BankingEmployeeInput) (*model.BankingEmployee, error) {
	return restCall[*model.BankingEmployee](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/employees/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteEmployee(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/employees/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateAccount(ctx context.Context, input model.BankingAccountInput) (*model.BankingAccount, error) {
	return restCall[*model.BankingAccount](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/accounts", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateAccount(ctx context.Context, id int, input model.BankingAccountInput) (*model.BankingAccount, error) {
	return restCall[*model.BankingAccount](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/accounts/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteAccount(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/accounts/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateLoan(ctx context.Context, input model.BankingLoanInput) (*model.BankingLoan, error) {
	return restCall[*model.BankingLoan](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/loans", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateLoan(ctx context.Context, id int, input model.BankingLoanInput) (*model.BankingLoan, error) {
	return restCall[*model.BankingLoan](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/loans/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteLoan(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/loans/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateLoanPayment(ctx context.Context, input model.BankingLoanPaymentInput) (*model.BankingLoanPayment, error) {
	return restCall[*model.BankingLoanPayment](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/loan-payments", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateLoanPayment(ctx context.Context, id int, input model.BankingLoanPaymentInput) (*model.BankingLoanPayment, error) {
	return restCall[*model.BankingLoanPayment](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/loan-payments/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteLoanPayment(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/loan-payments/%v", id)}, "")
}

func (r *mutationResolver) BankingCreateTransaction(ctx context.Context, input model.BankingTransactionInput) (*model.BankingTransaction, error) {
	return restCall[*model.BankingTransaction](ctx, r.Proxy, "banking-service", rest.Request{Method: "POST", Path: "/transactions", Body: input}, "")
}

func (r *mutationResolver) BankingUpdateTransaction(ctx context.Context, id int, input model.BankingTransactionInput) (*model.BankingTransaction, error) {
	return restCall[*model.BankingTransaction](ctx, r.Proxy, "banking-service", rest.Request{Method: "PUT", Path: rest.P("/transactions/%v", id), Body: input}, "")
}

func (r *mutationResolver) BankingDeleteTransaction(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "banking-service", rest.Request{Method: "DELETE", Path: rest.P("/transactions/%v", id)}, "")
}

func (r *queryResolver) BankingPersons(ctx context.Context) ([]*model.BankingPerson, error) {
	return restCall[[]*model.BankingPerson](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/persons"}, "")
}

func (r *queryResolver) BankingPerson(ctx context.Context, id int) (*model.BankingPerson, error) {
	return restCall[*model.BankingPerson](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/persons/%v", id)}, "")
}

func (r *queryResolver) BankingBranches(ctx context.Context) ([]*model.BankingBranch, error) {
	return restCall[[]*model.BankingBranch](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/branches"}, "")
}

func (r *queryResolver) BankingBranch(ctx context.Context, id int) (*model.BankingBranch, error) {
	return restCall[*model.BankingBranch](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/branches/%v", id)}, "")
}

func (r *queryResolver) BankingCustomers(ctx context.Context) ([]*model.BankingCustomer, error) {
	return restCall[[]*model.BankingCustomer](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/customers"}, "")
}

func (r *queryResolver) BankingCustomer(ctx context.Context, id int) (*model.BankingCustomer, error) {
	return restCall[*model.BankingCustomer](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/customers/%v", id)}, "")
}

func (r *queryResolver) BankingEmployees(ctx context.Context) ([]*model.BankingEmployee, error) {
	return restCall[[]*model.BankingEmployee](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/employees"}, "")
}

func (r *queryResolver) BankingEmployee(ctx context.Context, id int) (*model.BankingEmployee, error) {
	return restCall[*model.BankingEmployee](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/employees/%v", id)}, "")
}

func (r *queryResolver) BankingAccounts(ctx context.Context) ([]*model.BankingAccount, error) {
	return restCall[[]*model.BankingAccount](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/accounts"}, "")
}

func (r *queryResolver) BankingAccount(ctx context.Context, id int) (*model.BankingAccount, error) {
	return restCall[*model.BankingAccount](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/accounts/%v", id)}, "")
}

func (r *queryResolver) BankingLoans(ctx context.Context) ([]*model.BankingLoan, error) {
	return restCall[[]*model.BankingLoan](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/loans"}, "")
}

func (r *queryResolver) BankingLoan(ctx context.Context, id int) (*model.BankingLoan, error) {
	return restCall[*model.BankingLoan](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/loans/%v", id)}, "")
}

func (r *queryResolver) BankingLoanPayments(ctx context.Context) ([]*model.BankingLoanPayment, error) {
	return restCall[[]*model.BankingLoanPayment](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/loan-payments"}, "")
}

func (r *queryResolver) BankingLoanPayment(ctx context.Context, id int) (*model.BankingLoanPayment, error) {
	return restCall[*model.BankingLoanPayment](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/loan-payments/%v", id)}, "")
}

func (r *queryResolver) BankingTransactions(ctx context.Context) ([]*model.BankingTransaction, error) {
	return restCall[[]*model.BankingTransaction](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: "/transactions"}, "")
}

func (r *queryResolver) BankingTransaction(ctx context.Context, id int) (*model.BankingTransaction, error) {
	return restCall[*model.BankingTransaction](ctx, r.Proxy, "banking-service", rest.Request{Method: "GET", Path: rest.P("/transactions/%v", id)}, "")
}
