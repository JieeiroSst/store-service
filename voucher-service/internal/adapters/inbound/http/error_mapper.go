package http

import (
	"errors"
	"net/http"

	corporatedomain "github.com/JIeeiroSst/voucher-service/internal/domain/corporate"
	merchantdomain "github.com/JIeeiroSst/voucher-service/internal/domain/merchant"
	orderdomain "github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	userdomain "github.com/JIeeiroSst/voucher-service/internal/domain/user"
	voucherdomain "github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	walletdomain "github.com/JIeeiroSst/voucher-service/internal/domain/wallet"

	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func mapError(c *gin.Context, err error) {
	status, code, msg := classify(err)
	if status >= http.StatusInternalServerError {
		if log, ok := c.Get("logger"); ok {
			log.(*zap.Logger).Error("unhandled error", zap.Error(err), zap.String("path", c.FullPath()))
		}
	}
	c.JSON(status, gin.H{"error": gin.H{"code": code, "message": msg}})
}

func classify(err error) (status int, code, msg string) {
	switch {
	case errors.Is(err, shared.ErrValidation):
		return http.StatusBadRequest, "validation_error", err.Error()
	case errors.Is(err, shared.ErrNotFound),
		errors.Is(err, voucherdomain.ErrVoucherNotFound),
		errors.Is(err, orderdomain.ErrOrderNotFound),
		errors.Is(err, merchantdomain.ErrMerchantNotFound),
		errors.Is(err, walletdomain.ErrWalletNotFound),
		errors.Is(err, corporatedomain.ErrCorporateNotFound),
		errors.Is(err, userdomain.ErrUserNotFound):
		return http.StatusNotFound, "not_found", "resource not found"

	case errors.Is(err, voucherdomain.ErrVoucherExpired):
		return http.StatusGone, "voucher_expired", err.Error()
	case errors.Is(err, voucherdomain.ErrInvalidPIN):
		return http.StatusUnprocessableEntity, "invalid_pin", err.Error()
	case errors.Is(err, voucherdomain.ErrAlreadyRedeemed),
		errors.Is(err, voucherdomain.ErrInvalidTransition),
		errors.Is(err, orderdomain.ErrInvalidOrderTransition),
		errors.Is(err, voucherapp.ErrProviderRejected):
		return http.StatusConflict, "invalid_state", err.Error()

	case errors.Is(err, voucherdomain.ErrVersionConflict),
		errors.Is(err, orderdomain.ErrVersionConflict),
		errors.Is(err, walletdomain.ErrVersionConflict),
		errors.Is(err, merchantdomain.ErrVersionConflict),
		errors.Is(err, corporatedomain.ErrVersionConflict):
		return http.StatusConflict, "version_conflict", "resource was modified concurrently, please retry"

	case errors.Is(err, voucherapp.ErrRedeemInProgress),
		errors.Is(err, voucherapp.ErrDuplicateRequestInProgress):
		return http.StatusConflict, "duplicate_request", err.Error()
	case errors.Is(err, shared.ErrDuplicateRequest):
		return http.StatusConflict, "duplicate_request", "this idempotency key was previously used for a request that failed"

	case errors.Is(err, voucherapp.ErrLockUnavailable):
		return http.StatusServiceUnavailable, "lock_unavailable", "please retry shortly"
	case errors.Is(err, voucherapp.ErrProviderTimeout):
		return http.StatusBadGateway, "provider_timeout", "merchant provider did not respond in time, please retry"
	case errors.Is(err, voucherapp.ErrTransactionFailed):
		return http.StatusInternalServerError, "transaction_failed", "please retry"

	case errors.Is(err, walletdomain.ErrInsufficientFunds):
		return http.StatusUnprocessableEntity, "insufficient_funds", err.Error()
	case errors.Is(err, corporatedomain.ErrBudgetExceeded):
		return http.StatusUnprocessableEntity, "budget_exceeded", err.Error()
	case errors.Is(err, merchantdomain.ErrMerchantInactive):
		return http.StatusUnprocessableEntity, "merchant_inactive", err.Error()
	case errors.Is(err, userdomain.ErrInvalidCredentials):
		return http.StatusUnauthorized, "invalid_credentials", "invalid email or password"
	case errors.Is(err, userdomain.ErrUserInactive):
		return http.StatusForbidden, "user_inactive", err.Error()

	default:
		return http.StatusInternalServerError, "internal_error", "an unexpected error occurred"
	}
}
