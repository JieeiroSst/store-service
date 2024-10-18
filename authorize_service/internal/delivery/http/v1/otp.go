package v1

import (
	"errors"
	"net/http"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initOtpRouters(api *gin.RouterGroup) {
	otpGroup := api.Group("/otp")

	otpGroup.POST("", h.middleware.AuthorizeControl(), h.CreateOtp)
	otpGroup.POST("/authorize", h.middleware.AuthorizeControl(), h.AuthorizeOTP)
}

func (h *Handler) CreateOtp(ctx *gin.Context) {
	username := ctx.Query("username")

	otp, err := h.usecase.Otps.CreateOtpByUser(ctx, username)
	if errors.Is(err, common.OTPFailed) {
		Response(ctx, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	}
	if err != nil {
		Response(ctx, http.StatusInternalServerError, Message{Message: err.Error()})
		return
	}

	Response(ctx, http.StatusOK, Message{
		OTP:            otp.OTP,
		Message:        "Success",
		QuotaRemaining: 70,
		TextID:         username,
	})
}

func (h *Handler) AuthorizeOTP(ctx *gin.Context) {
	username := ctx.Query("username")
	otp := ctx.Query("otp")

	err := h.usecase.Otps.Authorize(ctx, otp, username)
	if errors.Is(err, common.OTPFailed) {
		Response(ctx, http.StatusUnauthorized, Message{
			OTP:            err.Error(),
			Message:        "Faield",
			QuotaRemaining: 0,
			TextID:         username,
		})
		return
	}
	if err != nil {
		Response(ctx, http.StatusInternalServerError, Message{
			OTP:            err.Error(),
			Message:        "Faield",
			QuotaRemaining: 0,
			TextID:         username,
		})
		return
	}

	Response(ctx, http.StatusOK, Message{
		Message:        "Success",
		QuotaRemaining: 0,
		TextID:         username,
	})
}
