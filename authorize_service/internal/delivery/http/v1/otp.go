package v1

import (
	"errors"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initOtpRouters(api *gin.RouterGroup) {
	otpGroup := api.Group("/otp")

	otpGroup.POST("/", h.middleware.AuthorizeControl(), h.CreateOtp)
	otpGroup.POST("/authorize", h.middleware.AuthorizeControl(), h.AuthorizeOTP)
}

func (h *Handler) CreateOtp(ctx *gin.Context) {
	username := ctx.Query("username")

	otp, err := h.usecase.Otps.CreateOtpByUser(username)
	if errors.Is(err, common.OTPFailed) {
		ctx.JSON(401, gin.H{
			"message":        err,
			"quotaRemaining": 0,
			"otp":            "",
		})
		return
	}
	if err != nil {
		ctx.JSON(400, gin.H{
			"message":        err,
			"quotaRemaining": 0,
			"otp":            "",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"otp":            otp.OTP,
		"success":        true,
		"textId":         "1234",
		"quotaRemaining": 70,
	})
}

func (h *Handler) AuthorizeOTP(ctx *gin.Context) {
	usernname := ctx.Query("username")
	otp := ctx.Query("otp")

	err := h.usecase.Otps.Authorize(otp, usernname)
	if errors.Is(err, common.OTPFailed) {
		ctx.JSON(401, gin.H{
			"message":        err.Error(),
			"success":        false,
			"quotaRemaining": 0,
			"otp":            "",
		})
		return
	}
	if err != nil {
		ctx.JSON(400, gin.H{
			"message":        err.Error(),
			"success":        false,
			"quotaRemaining": 0,
			"otp":            "",
		})
		return
	}

	ctx.JSON(200, gin.H{
		"success":    true,
		"isValidOtp": true,
	})
}
