package grpc

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/trace_id"
	permissionpb "github.com/JieeiroSst/authorize-service/api/permission/pb"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type PermissionHandler struct {
	perms port.PermissionUsecase
	permissionpb.UnimplementedPermissionServiceServer
}

func NewPermissionHandler(perms port.PermissionUsecase) *PermissionHandler {
	return &PermissionHandler{perms: perms}
}

func (h *PermissionHandler) AssignRoles(ctx context.Context, in *permissionpb.AssignRolesRequest) (*permissionpb.UserRolesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("AssignRoles", zap.String("user_id", in.GetUserId()), zap.Strings("roles", in.GetRoles()))

	res, err := h.perms.AssignRoles(ctx, in.GetUserId(), in.GetRoles())
	if err != nil {
		return nil, toStatus(err)
	}
	return userRolesToPB(res), nil
}

func (h *PermissionHandler) RevokeRoles(ctx context.Context, in *permissionpb.RevokeRolesRequest) (*permissionpb.UserRolesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("RevokeRoles", zap.String("user_id", in.GetUserId()), zap.Strings("roles", in.GetRoles()))

	res, err := h.perms.RevokeRoles(ctx, in.GetUserId(), in.GetRoles())
	if err != nil {
		return nil, toStatus(err)
	}
	return userRolesToPB(res), nil
}

func (h *PermissionHandler) SetUserRoles(ctx context.Context, in *permissionpb.SetUserRolesRequest) (*permissionpb.UserRolesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("SetUserRoles", zap.String("user_id", in.GetUserId()), zap.Strings("roles", in.GetRoles()))

	res, err := h.perms.SetUserRoles(ctx, in.GetUserId(), in.GetRoles())
	if err != nil {
		return nil, toStatus(err)
	}
	return userRolesToPB(res), nil
}

func (h *PermissionHandler) RemoveUser(ctx context.Context, in *permissionpb.RemoveUserRequest) (*permissionpb.RemoveUserResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)
	logger.WithContext(ctx).Info("RemoveUser", zap.String("user_id", in.GetUserId()))

	if err := h.perms.RemoveUser(ctx, in.GetUserId()); err != nil {
		return nil, toStatus(err)
	}
	return &permissionpb.RemoveUserResponse{Success: true}, nil
}

func (h *PermissionHandler) GetUserRoles(ctx context.Context, in *permissionpb.GetUserRolesRequest) (*permissionpb.UserRolesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	res, err := h.perms.GetUserRoles(ctx, in.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	return userRolesToPB(res), nil
}

func (h *PermissionHandler) GetUserPermissions(ctx context.Context, in *permissionpb.GetUserPermissionsRequest) (*permissionpb.GetUserPermissionsResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	perms, err := h.perms.GetUserPermissions(ctx, in.GetUserId())
	if err != nil {
		return nil, toStatus(err)
	}
	return &permissionpb.GetUserPermissionsResponse{
		UserId:      in.GetUserId(),
		Permissions: permissionsToPB(perms),
	}, nil
}

func (h *PermissionHandler) CheckPermission(ctx context.Context, in *permissionpb.CheckPermissionRequest) (*permissionpb.CheckPermissionResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	ok, err := h.perms.CheckPermission(ctx, in.GetUserId(), in.GetObject(), in.GetAction())
	if err != nil {
		return nil, toStatus(err)
	}
	return &permissionpb.CheckPermissionResponse{Allowed: ok}, nil
}

func (h *PermissionHandler) ListRoles(ctx context.Context, _ *permissionpb.ListRolesRequest) (*permissionpb.ListRolesResponse, error) {
	ctx = trace_id.EnsureTracerID(ctx)

	roles, err := h.perms.ListRoles(ctx)
	if err != nil {
		return nil, toStatus(err)
	}

	out := make([]*permissionpb.Role, 0, len(roles))
	for _, r := range roles {
		out = append(out, &permissionpb.Role{
			Name:        r.Name,
			Description: r.Description,
			Rank:        int32(r.Rank),
			Inherits:    r.Inherits,
			Permissions: permissionsToPB(r.Permissions),
			Assignable:  r.Assignable,
		})
	}
	return &permissionpb.ListRolesResponse{Roles: out}, nil
}

func userRolesToPB(r model.UserRoles) *permissionpb.UserRolesResponse {
	return &permissionpb.UserRolesResponse{
		UserId:         r.UserID,
		Roles:          r.Roles,
		EffectiveRoles: r.EffectiveRoles,
		PrimaryRole:    r.PrimaryRole,
	}
}

func permissionsToPB(perms []model.Permission) []*permissionpb.Permission {
	out := make([]*permissionpb.Permission, 0, len(perms))
	for _, p := range perms {
		out = append(out, &permissionpb.Permission{Object: p.Object, Action: p.Action, GrantedBy: p.GrantedBy})
	}
	return out
}

func toStatus(err error) error {
	switch {
	case errors.Is(err, common.ErrInvalidUserID), errors.Is(err, common.ErrUnknownRole):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, common.ErrRoleNotAssignable):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, common.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
