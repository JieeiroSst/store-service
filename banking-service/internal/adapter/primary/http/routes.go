package http

import "github.com/gin-gonic/gin"

func NewRouter(h *Handler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery(), CORSMiddleware())

	RegisterRoutes(engine, h)

	return engine
}

func RegisterRoutes(engine *gin.Engine, h *Handler) {
	engine.GET("/health", h.GetHealth)

	persons := engine.Group("/persons")
	{
		persons.POST("", h.CreatePerson)
		persons.GET("", h.ListPersons)
		persons.GET("/:id", h.GetPerson)
		persons.PUT("/:id", h.UpdatePerson)
		persons.DELETE("/:id", h.DeletePerson)
	}

	branches := engine.Group("/branches")
	{
		branches.POST("", h.CreateBranch)
		branches.GET("", h.ListBranches)
		branches.GET("/:id", h.GetBranch)
		branches.PUT("/:id", h.UpdateBranch)
		branches.DELETE("/:id", h.DeleteBranch)
	}

	customers := engine.Group("/customers")
	{
		customers.POST("", h.CreateCustomer)
		customers.GET("", h.ListCustomers)
		customers.GET("/:id", h.GetCustomer)
		customers.PUT("/:id", h.UpdateCustomer)
		customers.DELETE("/:id", h.DeleteCustomer)
	}

	employees := engine.Group("/employees")
	{
		employees.POST("", h.CreateEmployee)
		employees.GET("", h.ListEmployees)
		employees.GET("/:id", h.GetEmployee)
		employees.PUT("/:id", h.UpdateEmployee)
		employees.DELETE("/:id", h.DeleteEmployee)
	}

	accounts := engine.Group("/accounts")
	{
		accounts.POST("", h.CreateAccount)
		accounts.GET("", h.ListAccounts)
		accounts.GET("/:id", h.GetAccount)
		accounts.PUT("/:id", h.UpdateAccount)
		accounts.DELETE("/:id", h.DeleteAccount)
	}

	loans := engine.Group("/loans")
	{
		loans.POST("", h.CreateLoan)
		loans.GET("", h.ListLoans)
		loans.GET("/:id", h.GetLoan)
		loans.PUT("/:id", h.UpdateLoan)
		loans.DELETE("/:id", h.DeleteLoan)
	}

	loanPayments := engine.Group("/loan-payments")
	{
		loanPayments.POST("", h.CreateLoanPayment)
		loanPayments.GET("", h.ListLoanPayments)
		loanPayments.GET("/:id", h.GetLoanPayment)
		loanPayments.PUT("/:id", h.UpdateLoanPayment)
		loanPayments.DELETE("/:id", h.DeleteLoanPayment)
	}

	transactions := engine.Group("/transactions")
	{
		transactions.POST("", h.CreateTransaction)
		transactions.GET("", h.ListTransactions)
		transactions.GET("/:id", h.GetTransaction)
		transactions.PUT("/:id", h.UpdateTransaction)
		transactions.DELETE("/:id", h.DeleteTransaction)
	}
}
