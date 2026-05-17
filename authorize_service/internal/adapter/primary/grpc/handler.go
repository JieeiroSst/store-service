package grpc

import (
	"context"
	"errors"

	authorizeGrpc "github.com/JIeeiroSst/lib-gateway/authorize-service/gateway/authorize-service"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/trace_id"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/JieeiroSst/authorize-service/pkg/pagination"
	"go.uber.org/zap"
)

type Handler struct {
	casbin port.CasbinUsecase
	otp    port.OTPUsecase
	authorizeGrpc.UnimplementedAuthorizeServiceServer
}

func NewHandler(casbinUC port.CasbinUsecase, otpUC port.OTPUsecase) *Handler {
	return &Handler{casbin: casbinUC, otp: otpUC}
}

// ─── Enforce endpoints ────────────────────────────────────────────────────────

func (h *Handler) Authorize(ctx context.Context, in *authorizeGrpc.CasbinAuth) (*authorizeGrpc.AuthorizeResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	auth := model.CasbinAuth{Sub: in.Sub, Obj: in.Obj, Act: in.Act}
	logger.WithContext(ctx).Info("Authorize", zap.Any("auth", auth))

	if err := h.casbin.Enforce(ctx, auth); err != nil {
		return &authorizeGrpc.AuthorizeResponse{Message: err.Error()}, err
	}
	return &authorizeGrpc.AuthorizeResponse{Message: "success"}, nil
}

func (h *Handler) EnforceCasbin(ctx context.Context, in *authorizeGrpc.CasbinRuleRequest) (*authorizeGrpc.CasbinRuleReponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	auth := model.CasbinAuth{Sub: in.Sub, Obj: in.Obj, Act: in.Act}
	logger.WithContext(ctx).Info("EnforceCasbin", zap.Any("auth", auth))

	if err := h.casbin.Enforce(ctx, auth); err != nil {
		msg, isErr := mapEnforceError(err)
		return &authorizeGrpc.CasbinRuleReponse{Message: msg, Error: isErr}, err
	}
	return &authorizeGrpc.CasbinRuleReponse{
		Message: "THE CUSTOMER IS AUTHORIZED FOR THE CONTENT REQUESTED",
		Error:   false,
	}, nil
}

// ─── CRUD endpoints ───────────────────────────────────────────────────────────

func (h *Handler) GetCasbinRules(ctx context.Context, in *authorizeGrpc.CasbinRequest) (*authorizeGrpc.CasbinRuleList, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	p := pagination.Pagination{
		Limit: int(in.GetLimit()),
		Page:  int(in.GetPage()),
	}
	result, err := h.casbin.ListRules(ctx, p)
	if err != nil {
		lg.Error("GetCasbinRules", zap.Error(err))
		return nil, err
	}

	rules, _ := result.Rows.([]model.CasbinRule)
	var pbRules []*authorizeGrpc.CasbinRule
	for _, r := range rules {
		pbRules = append(pbRules, domainToPB(r))
	}
	// return &authorizeGrpc.CasbinRuleList{CasbinRules: pbRules}, nil
	return &authorizeGrpc.CasbinRuleList{Rows: pbRules}, nil
}

func (h *Handler) GetCasbinRuleById(ctx context.Context, in *authorizeGrpc.CasbinRuleId) (*authorizeGrpc.CasbinRule, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	rule, err := h.casbin.GetRule(ctx, int(in.GetId()))
	if err != nil {
		lg.Error("GetCasbinRuleById", zap.Error(err))
		return nil, err
	}
	return domainToPB(*rule), nil
}

func (h *Handler) CreateCasbinRule(ctx context.Context, in *authorizeGrpc.CasbinRule) (*authorizeGrpc.CreateCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	rule := model.CasbinRule{
		Ptype: in.GetPtype(),
		V0:    in.GetV0(), V1: in.GetV1(), V2: in.GetV2(),
		V3: in.GetV3(), V4: in.GetV4(), V5: in.GetV5(),
	}
	if err := h.casbin.CreateRule(ctx, rule); err != nil {
		lg.Error("CreateCasbinRule", zap.Error(err))
		return &authorizeGrpc.CreateCasbinRuleResponse{Message: "failed"}, err
	}
	return &authorizeGrpc.CreateCasbinRuleResponse{Message: "success"}, nil
}

func (h *Handler) DeleteCasbinRule(ctx context.Context, in *authorizeGrpc.CasbinRuleId) (*authorizeGrpc.DeleteCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	if err := h.casbin.DeleteRule(ctx, int(in.GetId())); err != nil {
		lg.Error("DeleteCasbinRule", zap.Error(err))
		return &authorizeGrpc.DeleteCasbinRuleResponse{Success: false}, err
	}
	return &authorizeGrpc.DeleteCasbinRuleResponse{Success: true}, nil
}

func (h *Handler) UpdateCasbinRule(ctx context.Context, in *authorizeGrpc.UpdateCasbinRuleRequest) (*authorizeGrpc.UpdateCasbinRuleResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	field := model.UpdateField(in.GetField())
	if err := h.casbin.UpdateRuleField(ctx, int(in.GetId()), field, in.GetValue()); err != nil {
		lg.Error("UpdateCasbinRule", zap.Error(err))
		return &authorizeGrpc.UpdateCasbinRuleResponse{Message: "failed"}, err
	}
	return &authorizeGrpc.UpdateCasbinRuleResponse{Message: "success"}, nil
}

// ─── OTP endpoints ────────────────────────────────────────────────────────────

func (h *Handler) CreateOTP(ctx context.Context, in *authorizeGrpc.CreateOTPRequest) (*authorizeGrpc.CreateOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	token, err := h.otp.CreateOtpByUser(ctx, in.Username)
	if err != nil {
		lg.Error("CreateOTP", zap.Error(err))
		return nil, err
	}
	return &authorizeGrpc.CreateOTPResponse{Otp: token, ExpiresAt: 70}, nil
}

func (h *Handler) AuthorizeOTP(ctx context.Context, in *authorizeGrpc.AuthorizeOTPRequest) (*authorizeGrpc.AuthorizeOTPResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	lg := logger.WithContext(ctx)

	if err := h.otp.Authorize(ctx, in.Otp, in.Username); err != nil {
		lg.Error("AuthorizeOTP", zap.Error(err))
		return &authorizeGrpc.AuthorizeOTPResponse{Message: err.Error()}, err
	}
	return &authorizeGrpc.AuthorizeOTPResponse{Message: "OTP authorized"}, nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

func domainToPB(r model.CasbinRule) *authorizeGrpc.CasbinRule {
	return &authorizeGrpc.CasbinRule{
		Id: int32(r.ID), Ptype: r.Ptype,
		V0: r.V0, V1: r.V1, V2: r.V2,
		V3: r.V3, V4: r.V4, V5: r.V5,
	}
}

func mapEnforceError(err error) (msg string, isErr bool) {
	switch {
	case errors.Is(err, common.ErrDBFailed):
		return common.ErrDBFailed.Error(), true
	case errors.Is(err, common.ErrEnforcerFailed):
		return common.ErrEnforcerFailed.Error(), true
	case errors.Is(err, common.ErrNotAllowed):
		return "THE CUSTOMER IS NOT AUTHORIZED FOR THE CONTENT REQUESTED", true
	default:
		return err.Error(), true
	}
}
