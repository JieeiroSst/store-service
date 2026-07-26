package port

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
)

type PersonUsecase interface {
	CreatePerson(ctx context.Context, person *model.Person) (*model.Person, error)
	GetPerson(ctx context.Context, id int) (*model.Person, error)
	ListPersons(ctx context.Context) ([]model.Person, error)
	UpdatePerson(ctx context.Context, person *model.Person) (*model.Person, error)
	DeletePerson(ctx context.Context, id int) error
}

type BranchUsecase interface {
	CreateBranch(ctx context.Context, branch *model.Branch) (*model.Branch, error)
	GetBranch(ctx context.Context, id int) (*model.Branch, error)
	ListBranches(ctx context.Context) ([]model.Branch, error)
	UpdateBranch(ctx context.Context, branch *model.Branch) (*model.Branch, error)
	DeleteBranch(ctx context.Context, id int) error
}

type CustomerUsecase interface {
	CreateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	GetCustomer(ctx context.Context, id int) (*model.Customer, error)
	ListCustomers(ctx context.Context) ([]model.Customer, error)
	UpdateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error)
	DeleteCustomer(ctx context.Context, id int) error
}

type EmployeeUsecase interface {
	CreateEmployee(ctx context.Context, employee *model.Employee) (*model.Employee, error)
	GetEmployee(ctx context.Context, id int) (*model.Employee, error)
	ListEmployees(ctx context.Context) ([]model.Employee, error)
	UpdateEmployee(ctx context.Context, employee *model.Employee) (*model.Employee, error)
	DeleteEmployee(ctx context.Context, id int) error
}

type AccountUsecase interface {
	CreateAccount(ctx context.Context, account *model.Account) (*model.Account, error)
	GetAccount(ctx context.Context, id int) (*model.Account, error)
	ListAccounts(ctx context.Context) ([]model.Account, error)
	UpdateAccount(ctx context.Context, account *model.Account) (*model.Account, error)
	DeleteAccount(ctx context.Context, id int) error
}

type LoanUsecase interface {
	CreateLoan(ctx context.Context, loan *model.Loan) (*model.Loan, error)
	GetLoan(ctx context.Context, id int) (*model.Loan, error)
	ListLoans(ctx context.Context) ([]model.Loan, error)
	UpdateLoan(ctx context.Context, loan *model.Loan) (*model.Loan, error)
	DeleteLoan(ctx context.Context, id int) error
}

type LoanPaymentUsecase interface {
	CreateLoanPayment(ctx context.Context, payment *model.LoanPayment) (*model.LoanPayment, error)
	GetLoanPayment(ctx context.Context, id int) (*model.LoanPayment, error)
	ListLoanPayments(ctx context.Context) ([]model.LoanPayment, error)
	UpdateLoanPayment(ctx context.Context, payment *model.LoanPayment) (*model.LoanPayment, error)
	DeleteLoanPayment(ctx context.Context, id int) error
}

type TransactionUsecase interface {
	CreateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
	GetTransaction(ctx context.Context, id int) (*model.Transaction, error)
	ListTransactions(ctx context.Context) ([]model.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *model.Transaction) (*model.Transaction, error)
	DeleteTransaction(ctx context.Context, id int) error
}
