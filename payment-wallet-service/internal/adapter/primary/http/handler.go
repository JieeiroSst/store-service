package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	wallets         port.WalletUsecase
	transactions    port.TransactionUsecase
	paymentMethods  port.PaymentMethodUsecase
	pockets         port.PocketUsecase
	paymentRequests port.PaymentRequestUsecase
}

func NewHandler(
	wallets port.WalletUsecase,
	transactions port.TransactionUsecase,
	paymentMethods port.PaymentMethodUsecase,
	pockets port.PocketUsecase,
	paymentRequests port.PaymentRequestUsecase,
) *Handler {
	return &Handler{
		wallets:         wallets,
		transactions:    transactions,
		paymentMethods:  paymentMethods,
		pockets:         pockets,
		paymentRequests: paymentRequests,
	}
}

func writeError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, port.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrWalletAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrInsufficientBalance):
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
	case errors.Is(err, port.ErrWalletNotActive),
		errors.Is(err, port.ErrInvalidAmount),
		errors.Is(err, port.ErrCurrencyMismatch),
		errors.Is(err, port.ErrSameWallet),
		errors.Is(err, port.ErrInvalidPaymentMethodType),
		errors.Is(err, port.ErrTransactionNotReversible),
		errors.Is(err, port.ErrTransferNotReversible),
		errors.Is(err, port.ErrInvalidWalletStatusTransition),
		errors.Is(err, port.ErrWalletNotEmpty),
		errors.Is(err, port.ErrPerTransactionLimitExceeded),
		errors.Is(err, port.ErrDailyLimitExceeded),
		errors.Is(err, port.ErrPaymentRequestNotPending),
		errors.Is(err, port.ErrPaymentRequestExpired),
		errors.Is(err, port.ErrPaymentRequestPayerMismatch):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
