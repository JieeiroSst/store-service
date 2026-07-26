package port

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
)

type PersonRepository interface {
	Create(ctx context.Context, person *model.Person) error
	GetByID(ctx context.Context, id int) (*model.Person, error)
	List(ctx context.Context) ([]model.Person, error)
	Update(ctx context.Context, person *model.Person) error
	Delete(ctx context.Context, id int) error
}

type BranchRepository interface {
	Create(ctx context.Context, branch *model.Branch) error
	GetByID(ctx context.Context, id int) (*model.Branch, error)
	List(ctx context.Context) ([]model.Branch, error)
	Update(ctx context.Context, branch *model.Branch) error
	Delete(ctx context.Context, id int) error
}

type CustomerRepository interface {
	Create(ctx context.Context, customer *model.Customer) error
	GetByID(ctx context.Context, id int) (*model.Customer, error)
	List(ctx context.Context) ([]model.Customer, error)
	Update(ctx context.Context, customer *model.Customer) error
	Delete(ctx context.Context, id int) error
}

type EmployeeRepository interface {
	Create(ctx context.Context, employee *model.Employee) error
	GetByID(ctx context.Context, id int) (*model.Employee, error)
	List(ctx context.Context) ([]model.Employee, error)
	Update(ctx context.Context, employee *model.Employee) error
	Delete(ctx context.Context, id int) error
}

type AccountRepository interface {
	Create(ctx context.Context, account *model.Account) error
	GetByID(ctx context.Context, id int) (*model.Account, error)
	List(ctx context.Context) ([]model.Account, error)
	Update(ctx context.Context, account *model.Account) error
	Delete(ctx context.Context, id int) error
}

type LoanRepository interface {
	Create(ctx context.Context, loan *model.Loan) error
	GetByID(ctx context.Context, id int) (*model.Loan, error)
	List(ctx context.Context) ([]model.Loan, error)
	Update(ctx context.Context, loan *model.Loan) error
	Delete(ctx context.Context, id int) error
}

type LoanPaymentRepository interface {
	Create(ctx context.Context, payment *model.LoanPayment) error
	GetByID(ctx context.Context, id int) (*model.LoanPayment, error)
	List(ctx context.Context) ([]model.LoanPayment, error)
	Update(ctx context.Context, payment *model.LoanPayment) error
	Delete(ctx context.Context, id int) error
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *model.Transaction) error
	GetByID(ctx context.Context, id int) (*model.Transaction, error)
	List(ctx context.Context) ([]model.Transaction, error)
	Update(ctx context.Context, transaction *model.Transaction) error
	Delete(ctx context.Context, id int) error
}
