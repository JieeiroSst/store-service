package http

import (
	"context"
	"errors"

	authorizeServiceGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/trace_id"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"go.uber.org/zap"
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

func (h *Handler) Authorize(ctx context.Context, in *authorizeServiceGrpc.CasbinAuth) (*authorizeServiceGrpc.AuthorizeResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	auth := model.CasbinAuth{
		Sub: in.Sub, // username
		Obj: in.Obj, // path api
		Act: in.Act, // method api
	}
	lg.Info("Authorize request: %v", zap.Any("auth", auth))
	err := h.usecase.Casbins.EnforceCasbin(ctx, auth)
	if errors.Is(err, common.FailedDB) {
		lg.Error("FailedDB", zap.Error(err))
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.FailedDB.Error(),
		}, err
	}
	if errors.Is(err, common.Failedenforcer) {
		lg.Error("Failedenforcer", zap.Error(err))
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.Failedenforcer.Error(),
		}, err
	}
	if errors.Is(err, common.NotAllow) {
		lg.Error("NotAllow", zap.Error(err))
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.NotAllow.Error(),
		}, err
	}
	return &authorizeServiceGrpc.AuthorizeResponse{Message: "susscess"}, nil
}

func (h *Handler) GetCasbinRules(ctx context.Context, in *authorizeServiceGrpc.CasbinRequest) (*authorizeServiceGrpc.CasbinRuleList, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	var p pagination.Pagination
	if err := copy.CopyObject(&in, &p); err != nil {
		lg.Error("copy object failed", zap.Error(err))
		return nil, err
	}
	casbins, err := h.usecase.Casbins.CasbinRuleAll(ctx, p)
	if errors.Is(err, common.NotFound) {
		lg.Error("casbin not found", zap.Error(err))
		return nil, err
	}

	if err != nil {
		lg.Error("casbin not found", zap.Error(err))
		return nil, err
	}

	var res *authorizeServiceGrpc.CasbinRuleList
	if err := copy.CopyObject(&casbins, &res); err != nil {
		lg.Error("copy object failed", zap.Error(err))
		return nil, err
	}
	lg.Info("GetCasbinRules", zap.Any("casbins", casbins))
	return res, nil
}

func (h *Handler) GetCasbinRuleById(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.CasbinRule, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	casbin, err := h.usecase.Casbins.CasbinRuleById(ctx, int(in.GetId()))
	if errors.Is(err, common.NotFound) {
		lg.Error("casbin not found", zap.Error(err))
		return nil, err
	}
	if err != nil {
		lg.Error("casbin not found", zap.Error(err))
		return nil, err
	}

	var resp *authorizeServiceGrpc.CasbinRule
	if err := copy.CopyObject(&casbin, &resp); err != nil {
		lg.Error("copy object failed", zap.Error(err))
		return nil, err
	}
	lg.Error("GetCasbinRuleById", zap.Any("casbin", casbin))
	return resp, nil
}

func (h *Handler) CreateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRule) (*authorizeServiceGrpc.CreateCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	var casbin model.CasbinRule
	if err := copy.CopyObject(&in, &casbin); err != nil {
		lg.Error("copy object failed", zap.Error(err))
		return &authorizeServiceGrpc.CreateCasbinRuleResponse{
			Message: "failed",
		}, err
	}
	err := h.usecase.Casbins.CreateCasbinRule(ctx, casbin)
	if errors.Is(err, common.FailedDB) {
		lg.Error("FailedDB", zap.Error(err))
		return &authorizeServiceGrpc.CreateCasbinRuleResponse{
			Message: "failed",
		}, err
	}
	if err != nil {
		lg.Error("FailedDB", zap.Error(err))
		return nil, err
	}
	return &authorizeServiceGrpc.CreateCasbinRuleResponse{
		Message: "success",
	}, nil
}

func (h *Handler) DeleteCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.DeleteCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	if err := h.usecase.Casbins.DeleteCasbinRule(ctx, int(in.GetId())); err != nil {
		lg.Error("FailedDB", zap.Error(err))
		return &authorizeServiceGrpc.DeleteCasbinRuleResponse{
			Success: false,
		}, err
	}
	return &authorizeServiceGrpc.DeleteCasbinRuleResponse{
		Success: true,
	}, nil // success
}

func (h *Handler) UpdateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.UpdateCasbinRuleRequest) (*authorizeServiceGrpc.UpdateCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	if err := h.usecase.Casbins.UpdateCasbinRule(ctx, int(in.Id), in.GetField(), in.GetValue()); err != nil {
		lg.Error("FailedDB", zap.Error(err))
		return &authorizeServiceGrpc.UpdateCasbinRuleResponse{
			Message: "failed",
		}, err
	}
	return &authorizeServiceGrpc.UpdateCasbinRuleResponse{
		Message: "success",
	}, nil
}

func (h *Handler) CreateOTP(ctx context.Context, in *authorizeServiceGrpc.CreateOTPRequest) (*authorizeServiceGrpc.CreateOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	otp, err := h.usecase.Otps.CreateOtpByUser(ctx, in.Username)
	if errors.Is(err, common.OTPFailed) {
		lg.Error("OTPFailed", zap.Error(err))
		return nil, err
	}
	if errors.Is(err, common.OTPLimmit) {
		lg.Error("OTPLimmit", zap.Error(err))
		return nil, err
	}
	if err != nil {
		lg.Error("FailedDB", zap.Error(err))
		return nil, err
	}
	return &authorizeServiceGrpc.CreateOTPResponse{
		Otp:       otp.OTP,
		ExpiresAt: 70,
	}, nil
}

func (h *Handler) AuthorizeOTP(ctx context.Context, in *authorizeServiceGrpc.AuthorizeOTPRequest) (*authorizeServiceGrpc.AuthorizeOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	err := h.usecase.Otps.Authorize(ctx, in.Otp, in.Username)
	if errors.Is(err, common.OTPFailed) {
		lg.Error("OTPFailed", zap.Error(err))
		return &authorizeServiceGrpc.AuthorizeOTPResponse{
			Message: common.OTPFailed.Error(),
		}, err
	}
	return &authorizeServiceGrpc.AuthorizeOTPResponse{
		Message: "OTP AUTHORIZED",
	}, nil
}

func (h *Handler) EnforceCasbin(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleRequest) (*authorizeServiceGrpc.CasbinRuleReponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)
	casbin := model.CasbinAuth{
		Sub: in.Sub,
		Obj: in.Obj,
		Act: in.Act,
	}
	lg.Info("EnforceCasbin", zap.Any("casbin", casbin))

	err := h.usecase.EnforceCasbin(ctx, casbin)
	if errors.Is(err, common.FailedDB) {
		lg.Error("FailedDB", zap.Error(err))
		return &authorizeServiceGrpc.CasbinRuleReponse{
			Message: common.FailedDB.Error(),
			Error:   true,
		}, err
	}

	if errors.Is(err, common.Failedenforcer) {
		lg.Error("Failedenforcer", zap.Error(err))
		return &authorizeServiceGrpc.CasbinRuleReponse{
			Message: common.Failedenforcer.Error(),
			Error:   true,
		}, err
	}

	if errors.Is(err, common.NotAllow) {
		lg.Error("NotAllow", zap.Error(err))
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
