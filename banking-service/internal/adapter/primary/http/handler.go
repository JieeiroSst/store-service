package http

import (
	"errors"
	"net/http"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	person      port.PersonUsecase
	branch      port.BranchUsecase
	customer    port.CustomerUsecase
	employee    port.EmployeeUsecase
	account     port.AccountUsecase
	loan        port.LoanUsecase
	loanPayment port.LoanPaymentUsecase
	transaction port.TransactionUsecase
}

func NewHandler(
	person port.PersonUsecase,
	branch port.BranchUsecase,
	customer port.CustomerUsecase,
	employee port.EmployeeUsecase,
	account port.AccountUsecase,
	loan port.LoanUsecase,
	loanPayment port.LoanPaymentUsecase,
	transaction port.TransactionUsecase,
) *Handler {
	return &Handler{
		person:      person,
		branch:      branch,
		customer:    customer,
		employee:    employee,
		account:     account,
		loan:        loan,
		loanPayment: loanPayment,
		transaction: transaction,
	}
}

// writeError translates domain errors into the matching HTTP status code.
func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, common.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrInvalidRequest):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, common.ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
