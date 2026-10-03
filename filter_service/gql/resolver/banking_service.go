package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BankQuery() generated.BankQueryResolver { return &bankQueryResolver{r} }

type bankQueryResolver struct{ *Resolver }

func (r *bankQueryResolver) Persons(ctx context.Context, obj *model.BankQuery) ([]*model.BankPerson, error) {
	return r.Clients.BankingService.Persons(ctx)
}

func (r *bankQueryResolver) Person(ctx context.Context, obj *model.BankQuery, id int) (*model.BankPerson, error) {
	return r.Clients.BankingService.Person(ctx, id)
}

func (r *bankQueryResolver) Branches(ctx context.Context, obj *model.BankQuery) ([]*model.BankBranch, error) {
	return r.Clients.BankingService.Branches(ctx)
}

func (r *bankQueryResolver) Branch(ctx context.Context, obj *model.BankQuery, id int) (*model.BankBranch, error) {
	return r.Clients.BankingService.Branch(ctx, id)
}

func (r *bankQueryResolver) Customers(ctx context.Context, obj *model.BankQuery) ([]*model.BankCustomer, error) {
	return r.Clients.BankingService.Customers(ctx)
}

func (r *bankQueryResolver) Customer(ctx context.Context, obj *model.BankQuery, id int) (*model.BankCustomer, error) {
	return r.Clients.BankingService.Customer(ctx, id)
}

func (r *bankQueryResolver) Employees(ctx context.Context, obj *model.BankQuery) ([]*model.BankEmployee, error) {
	return r.Clients.BankingService.Employees(ctx)
}

func (r *bankQueryResolver) Employee(ctx context.Context, obj *model.BankQuery, id int) (*model.BankEmployee, error) {
	return r.Clients.BankingService.Employee(ctx, id)
}

func (r *bankQueryResolver) Accounts(ctx context.Context, obj *model.BankQuery) ([]*model.BankAccount, error) {
	return r.Clients.BankingService.Accounts(ctx)
}

func (r *bankQueryResolver) Account(ctx context.Context, obj *model.BankQuery, id int) (*model.BankAccount, error) {
	return r.Clients.BankingService.Account(ctx, id)
}

func (r *bankQueryResolver) Loans(ctx context.Context, obj *model.BankQuery) ([]*model.BankLoan, error) {
	return r.Clients.BankingService.Loans(ctx)
}

func (r *bankQueryResolver) Loan(ctx context.Context, obj *model.BankQuery, id int) (*model.BankLoan, error) {
	return r.Clients.BankingService.Loan(ctx, id)
}

func (r *bankQueryResolver) LoanPayments(ctx context.Context, obj *model.BankQuery) ([]*model.BankLoanPayment, error) {
	return r.Clients.BankingService.LoanPayments(ctx)
}

func (r *bankQueryResolver) LoanPayment(ctx context.Context, obj *model.BankQuery, id int) (*model.BankLoanPayment, error) {
	return r.Clients.BankingService.LoanPayment(ctx, id)
}

func (r *bankQueryResolver) Transactions(ctx context.Context, obj *model.BankQuery) ([]*model.BankTransaction, error) {
	return r.Clients.BankingService.Transactions(ctx)
}

func (r *bankQueryResolver) Transaction(ctx context.Context, obj *model.BankQuery, id int) (*model.BankTransaction, error) {
	return r.Clients.BankingService.Transaction(ctx, id)
}
