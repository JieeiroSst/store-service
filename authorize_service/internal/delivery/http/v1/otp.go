package v1

import (
	"errors"

	"github.com/JIeeiroSst/utils/response"
	"github.com/JIeeiroSst/utils/trace_id"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

func (h *Handler) initOtpRouters(api *gin.RouterGroup) {
	otpGroup := api.Group("/otp")

	otpGroup.POST("/", h.middleware.AuthorizeControl(), h.CreateOtp)
	otpGroup.POST("/authorize", h.middleware.AuthorizeControl(), h.AuthorizeOTP)
}

func (h *Handler) CreateOtp(ctx *gin.Context) {
	newCtx := trace_id.TracerID("")
	username := ctx.Query("username")

	otp, err := h.usecase.Otps.CreateOtpByUser(username)
	if errors.Is(err, common.OTPFailed) {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		})
	}

	response.ResponseStatus(ctx, 200, response.MessageStatus{
		Message: common.CreateSuccess,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		Data:    otp.OTP,
	})
}

func (h *Handler) AuthorizeOTP(ctx *gin.Context) {
	newCtx := trace_id.TracerID("")
	username := ctx.Query("username")
	otp := ctx.Query("otp")

	err := h.usecase.Otps.Authorize(otp, username)
	if errors.Is(err, common.OTPFailed) {
		response.ResponseStatus(ctx, 401, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
			Data:    username,
		})
	}
	if err != nil {
		response.ResponseStatus(ctx, 500, response.MessageStatus{
			Message: err.Error(),
			Error:   true,
			TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
			Data:    username,
		})
	}

	response.ResponseStatus(ctx, 500, response.MessageStatus{
		Message: common.Authorized,
		Error:   true,
		TraceID: trace.SpanFromContext(newCtx).SpanContext().TraceID().String(),
		Data:    username,
	})
}
