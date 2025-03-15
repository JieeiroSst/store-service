package http

import (
	"context"

	userServiceGrpc "github.com/JIeeiroSst/lib-gateway/user-service/gateway/user-service"
	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/usecase"
	"github.com/JIeeiroSst/utils/copy"
)

type Handler struct {
	usecase *usecase.Usecase
	userServiceGrpc.UnimplementedUserServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) Login(ctx context.Context, in *userServiceGrpc.LoginRequest) (*userServiceGrpc.LoginResponse, error) {
	var user dto.LoginRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	session, err := h.usecase.Login(ctx, user)
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
	var user dto.LogoutRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.Logout(ctx, user)
	if err != nil {
		return &userServiceGrpc.LogoutResponse{Message: resp.Message}, err
	}

	return &userServiceGrpc.LogoutResponse{Message: resp.Message}, nil
}

func (h *Handler) ValidateSession(ctx context.Context, in *userServiceGrpc.ValidateRequest) (*userServiceGrpc.ValidateResponse, error) {
	var user dto.ValidateRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.ValidateSession(ctx, user)
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
	var user dto.RefreshRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.RefreshToken(ctx, user)
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
	var user dto.SignUpRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.SignUp(ctx, user)
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
	var user dto.UpdateProfileRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.UpdateProfile(ctx, user)
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
	var user dto.FindUserRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.FindUser(ctx, user)
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
	var user dto.AddRoleItemRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.AddRoleItem(ctx, user)
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
	var user dto.UpdateRoleItemRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.UpdateItemRole(ctx, user)
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
	var user dto.RemoveRoleItemRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.RemoveRoleItem(ctx, user)
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
	var user dto.GetRoleRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.GetRole(ctx, user)
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
	var user dto.ListRolesRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.ListRoles(ctx, user)
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
	var user dto.CreateRoleResquest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.CreateRole(ctx, user)
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
	var user dto.UpdateRoleRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.UpdateRole(ctx, user)
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
	var user dto.DeleteRoleRequest
	if err := copy.CopyObject(&in, &user); err != nil {
		return nil, err
	}

	resp, err := h.usecase.DeleteRole(ctx, user)
	if err != nil {
		return &userServiceGrpc.DeleteRoleResponse{}, err
	}

	var res userServiceGrpc.DeleteRoleResponse
	if err := copy.CopyObject(&resp, &res); err != nil {
		return nil, err
	}

	return &res, nil
}
