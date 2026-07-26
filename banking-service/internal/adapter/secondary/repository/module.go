package repository

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPersonRepository),
	fx.Provide(NewBranchRepository),
	fx.Provide(NewCustomerRepository),
	fx.Provide(NewEmployeeRepository),
	fx.Provide(NewAccountRepository),
	fx.Provide(NewLoanRepository),
	fx.Provide(NewLoanPaymentRepository),
	fx.Provide(NewTransactionRepository),
)
