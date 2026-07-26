package grpcadapter

import (
	"context"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/JIeeiroSst/utils/copy"
)

type Handler struct {
	auth     input.AuthService
	user     input.UserService
	role     input.RoleService
	roleItem input.RoleItemService
	userServiceGrpc.UnimplementedUserServiceServer
}

func NewHandler(auth input.AuthService, user input.UserService, role input.RoleService, roleItem input.RoleItemService) *Handler {
	return &Handler{
		auth:     auth,
		user:     user,
		role:     role,
		roleItem: roleItem,
	}
}

func (h *Handler) Login(ctx context.Context, in *userServiceGrpc.LoginRequest) (*userServiceGrpc.LoginResponse, error) {
	var req dto.LoginRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	session, err := h.auth.Login(ctx, req)
	if err != nil {
		return nil, err
	}

	var res userServiceGrpc.LoginResponse
	if err := copy.CopyObject(&session, &res); err != nil {
		return nil, err
	}
	return &res, nil
}

func (h *Handler) Logout(ctx context.Context, in *userServiceGrpc.LogoutRequest) (*userServiceGrpc.LogoutResponse, error) {
	var req dto.LogoutRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.auth.Logout(ctx, req)
	if err != nil {
		return &userServiceGrpc.LogoutResponse{Message: resp.Message}, err
	}

	return &userServiceGrpc.LogoutResponse{Message: resp.Message}, nil
}

func (h *Handler) ValidateSession(ctx context.Context, in *userServiceGrpc.ValidateRequest) (*userServiceGrpc.ValidateResponse, error) {
	var req dto.ValidateRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.auth.ValidateSession(ctx, req)
	if err != nil {
		return &userServiceGrpc.ValidateResponse{}, err
	}

	var res userServiceGrpc.ValidateResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) RefreshToken(ctx context.Context, in *userServiceGrpc.RefreshRequest) (*userServiceGrpc.RefreshResponse, error) {
	var req dto.RefreshRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.auth.RefreshToken(ctx, req)
	if err != nil {
		return &userServiceGrpc.RefreshResponse{}, err
	}

	var res userServiceGrpc.RefreshResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) SignUp(ctx context.Context, in *userServiceGrpc.SignUpRequest) (*userServiceGrpc.SignUpResponse, error) {
	var req dto.SignUpRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.user.SignUp(ctx, req)
	if err != nil {
		return &userServiceGrpc.SignUpResponse{}, err
	}

	var res userServiceGrpc.SignUpResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) UpdateProfile(ctx context.Context, in *userServiceGrpc.UpdateProfileRequest) (*userServiceGrpc.UpdateProfileResponse, error) {
	var req dto.UpdateProfileRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.user.UpdateProfile(ctx, req)
	if err != nil {
		return &userServiceGrpc.UpdateProfileResponse{}, err
	}

	var res userServiceGrpc.UpdateProfileResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) FindUser(ctx context.Context, in *userServiceGrpc.FindUserRequest) (*userServiceGrpc.FindUserResponse, error) {
	var req dto.FindUserRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.user.FindUser(ctx, req)
	if err != nil {
		return &userServiceGrpc.FindUserResponse{}, err
	}

	var res userServiceGrpc.FindUserResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) AddRoleItem(ctx context.Context, in *userServiceGrpc.AddRoleItemRequest) (*userServiceGrpc.AddRoleItemResponse, error) {
	var req dto.AddRoleItemRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.roleItem.AddRoleItem(ctx, req)
	if err != nil {
		return &userServiceGrpc.AddRoleItemResponse{}, err
	}

	var res userServiceGrpc.AddRoleItemResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) UpdateItemRole(ctx context.Context, in *userServiceGrpc.UpdateRoleItemRequest) (*userServiceGrpc.UpdateRoleItemResponse, error) {
	var req dto.UpdateRoleItemRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.roleItem.UpdateItemRole(ctx, req)
	if err != nil {
		return &userServiceGrpc.UpdateRoleItemResponse{}, err
	}

	var res userServiceGrpc.UpdateRoleItemResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) RemoveRoleItem(ctx context.Context, in *userServiceGrpc.RemoveRoleItemRequest) (*userServiceGrpc.RemoveRoleItemResponse, error) {
	var req dto.RemoveRoleItemRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.roleItem.RemoveRoleItem(ctx, req)
	if err != nil {
		return &userServiceGrpc.RemoveRoleItemResponse{}, err
	}

	var res userServiceGrpc.RemoveRoleItemResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) GetRole(ctx context.Context, in *userServiceGrpc.GetRoleRequest) (*userServiceGrpc.GetRoleResponse, error) {
	var req dto.GetRoleRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.role.GetRole(ctx, req)
	if err != nil {
		return &userServiceGrpc.GetRoleResponse{}, err
	}

	var res userServiceGrpc.GetRoleResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) ListRoles(ctx context.Context, in *userServiceGrpc.ListRolesRequest) (*userServiceGrpc.ListRolesResponse, error) {
	var req dto.ListRolesRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.role.ListRoles(ctx, req)
	if err != nil {
		return &userServiceGrpc.ListRolesResponse{}, err
	}

	var res userServiceGrpc.ListRolesResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) CreateRole(ctx context.Context, in *userServiceGrpc.CreateRoleResquest) (*userServiceGrpc.CreateRoleResponse, error) {
	var req dto.CreateRoleResquest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.role.CreateRole(ctx, req)
	if err != nil {
		return &userServiceGrpc.CreateRoleResponse{}, err
	}

	var res userServiceGrpc.CreateRoleResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) UpdateRole(ctx context.Context, in *userServiceGrpc.UpdateRoleRequest) (*userServiceGrpc.UpdateRoleResponse, error) {
	var req dto.UpdateRoleRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.role.UpdateRole(ctx, req)
	if err != nil {
		return &userServiceGrpc.UpdateRoleResponse{}, err
	}

	var res userServiceGrpc.UpdateRoleResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

func (h *Handler) DeleteRole(ctx context.Context, in *userServiceGrpc.DeleteRoleRequest) (*userServiceGrpc.DeleteRoleResponse, error) {
	var req dto.DeleteRoleRequest
	if err := copy.CopyObject(&in, &req); err != nil {
		return nil, err
	}

	resp, err := h.role.DeleteRole(ctx, req)
	if err != nil {
		return &userServiceGrpc.DeleteRoleResponse{}, err
	}

	var res userServiceGrpc.DeleteRoleResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
