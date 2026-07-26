package application

import "go.uber.org/fx"

var Module = fx.Options(
	fx.Provide(NewPersonService),
	fx.Provide(NewBranchService),
	fx.Provide(NewCustomerService),
	fx.Provide(NewEmployeeService),
	fx.Provide(NewAccountService),
	fx.Provide(NewLoanService),
	fx.Provide(NewLoanPaymentService),
	fx.Provide(NewTransactionService),
)
