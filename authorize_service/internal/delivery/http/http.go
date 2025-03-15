package http

import (
	"context"
	"errors"

	authorizeServiceGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/model"
)

type Handler struct {
	usecase *usecase.Usecase
	authorizeServiceGrpc.UnimplementedAuthorizeServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Authorize(ctx context.Context, in *authorizeServiceGrpc.CasbinAuth) (*authorizeServiceGrpc.OTP, error) {
	return nil, nil
}

func (h *Handler) GetCasbinRules(ctx context.Context, in *authorizeServiceGrpc.CasbinRequest) (*authorizeServiceGrpc.CasbinRuleList, error) {
	return nil, nil
}

func (h *Handler) GetCasbinRuleById(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.CasbinRule, error) {
	return nil, nil
}

func (h *Handler) CreateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRule) (*authorizeServiceGrpc.CasbinRule, error) {
	return nil, nil
}

func (h *Handler) DeleteCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.DeleteCasbinRuleResponse, error) {
	return nil, nil
}

func (h *Handler) UpdateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.UpdateCasbinRuleRequest) (*authorizeServiceGrpc.CasbinRule, error) {
	return nil, nil
}

func (h *Handler) CreateOTP(ctx context.Context, in *authorizeServiceGrpc.CreateOTPRequest) (*authorizeServiceGrpc.CreateOTPResponse, error) {
	otp, err := h.usecase.Otps.CreateOtpByUser(ctx, in.Username)
	if errors.Is(err, common.OTPFailed) {
		return nil, err
	}
	if errors.Is(err, common.OTPLimmit) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &authorizeServiceGrpc.CreateOTPResponse{
		Otp:       otp.OTP,
		ExpiresAt: 70,
	}, nil
}

func (h *Handler) AuthorizeOTP(ctx context.Context, in *authorizeServiceGrpc.AuthorizeOTPRequest) (*authorizeServiceGrpc.AuthorizeOTPResponse, error) {
	err := h.usecase.Otps.Authorize(ctx, in.Otp, in.Username)
	if errors.Is(err, common.OTPFailed) {
		return &authorizeServiceGrpc.AuthorizeOTPResponse{
			Message: common.OTPFailed.Error(),
		}, err
	}
	return &authorizeServiceGrpc.AuthorizeOTPResponse{
		Message: "OTP AUTHORIZED",
	}, nil
}

func (h *Handler) EnforceCasbin(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleRequest) (*authorizeServiceGrpc.CasbinRuleReponse, error) {
	casbin := model.CasbinAuth{
		Sub: in.Sub,
		Obj: in.Obj,
		Act: in.Act,
	}

	err := h.usecase.EnforceCasbin(ctx, casbin)

	if errors.Is(err, common.FailedDB) {
		return &authorizeServiceGrpc.CasbinRuleReponse{
			Message: common.FailedDB.Error(),
			Error:   true,
		}, err
	}

	if errors.Is(err, common.Failedenforcer) {
		return &authorizeServiceGrpc.CasbinRuleReponse{
			Message: common.Failedenforcer.Error(),
			Error:   true,
		}, err
	}

	if errors.Is(err, common.NotAllow) {
		return &authorizeServiceGrpc.CasbinRuleReponse{
			Message: common.NotAllow.Error(),
			Error:   true,
		}, err
	}

	return &authorizeServiceGrpc.CasbinRuleReponse{
		Message: "THE CUSTOMER IS AUTHORIZED FOR THE CONTENT REQUESTED",
		Error:   false,
	}, err
}
