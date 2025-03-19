package http

import (
	"context"
	"errors"

	authorizeServiceGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JIeeiroSst/utils/copy"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/usecase"
	"github.com/JieeiroSst/authorize-service/model"
	"github.com/JieeiroSst/authorize-service/pkg/log"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"github.com/JIeeiroSst/utils/trace_id"
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
	auth := model.CasbinAuth{
		Sub: in.Sub, // username
		Obj: in.Obj, // path api
		Act: in.Act, // method api
	}
	log.Info(auth)
	err := h.usecase.Casbins.EnforceCasbin(ctx, auth)
	if errors.Is(err, common.FailedDB) {
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.FailedDB.Error(),
		}, err
	}
	if errors.Is(err, common.Failedenforcer) {
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.Failedenforcer.Error(),
		}, err
	}
	if errors.Is(err, common.NotAllow) {
		return &authorizeServiceGrpc.AuthorizeResponse{
			Message: common.NotAllow.Error(),
		}, err
	}
	return nil, nil
}

func (h *Handler) GetCasbinRules(ctx context.Context, in *authorizeServiceGrpc.CasbinRequest) (*authorizeServiceGrpc.CasbinRuleList, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	var p pagination.Pagination
	if err := copy.CopyObject(&in, &p); err != nil {
		return nil, err
	}
	casbins, err := h.usecase.Casbins.CasbinRuleAll(ctx, p)
	if errors.Is(err, common.NotFound) {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var res *authorizeServiceGrpc.CasbinRuleList
	if err := copy.CopyObject(&casbins, &res); err != nil {
		return nil, err
	}

	return res, nil
}

func (h *Handler) GetCasbinRuleById(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.CasbinRule, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	casbin, err := h.usecase.Casbins.CasbinRuleById(ctx, int(in.GetId()))
	if errors.Is(err, common.NotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	var resp *authorizeServiceGrpc.CasbinRule
	if err := copy.CopyObject(&casbin, &resp); err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *Handler) CreateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRule) (*authorizeServiceGrpc.CasbinRule, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	var casbin model.CasbinRule
	if err := copy.CopyObject(&in, &casbin); err != nil {
		return nil, err
	}
	err := h.usecase.Casbins.CreateCasbinRule(ctx, casbin)
	if errors.Is(err, common.FailedDB) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return &authorizeServiceGrpc.CasbinRule{}, nil
}

func (h *Handler) DeleteCasbinRule(ctx context.Context, in *authorizeServiceGrpc.CasbinRuleId) (*authorizeServiceGrpc.DeleteCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	if err := h.usecase.Casbins.DeleteCasbinRule(ctx, int(in.GetId())); err != nil {
		return &authorizeServiceGrpc.DeleteCasbinRuleResponse{
			Success: false,
		}, err
	}
	return &authorizeServiceGrpc.DeleteCasbinRuleResponse{
		Success: true,
	}, nil // success
}

func (h *Handler) UpdateCasbinRule(ctx context.Context, in *authorizeServiceGrpc.UpdateCasbinRuleRequest) (*authorizeServiceGrpc.CasbinRule, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	if err := h.usecase.Casbins.UpdateCasbinRule(ctx, int(in.Id), in.GetField(), in.GetValue()); err != nil {	
		return nil, err
	}
	return nil, nil
}

func (h *Handler) CreateOTP(ctx context.Context, in *authorizeServiceGrpc.CreateOTPRequest) (*authorizeServiceGrpc.CreateOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
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
	ctx = trace_id.EnsureTracerID(ctx)
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
	ctx = trace_id.EnsureTracerID(ctx)
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
