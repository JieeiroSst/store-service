package graph

import (
	"context"
	"strconv"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/grpcerr"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/userpb"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) Login(ctx context.Context, input model.LoginInput) (*model.AuthToken, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().Login(ctx, &userpb.LoginRequest{Username: input.Username, Password: input.Password})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.AuthToken{AccessToken: res.GetSessionToken(), RefreshToken: res.GetRefreshToken(), ExpiresIn: int(res.GetExpiryTime())}, nil
}

func (r *mutationResolver) SignUp(ctx context.Context, input model.SignUpInput) (*model.UserMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().SignUp(ctx, &userpb.SignUpRequest{
		Username: input.Username,
		Password: input.Password,
		Email:    input.Email,
		Name:     input.Name,
		Phone:    deref(input.Phone),
		Address:  deref(input.Address),
		Sex:      deref(input.Sex),
	})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserMessage{Message: res.GetMessage(), User: toUser(res.GetUser())}, nil
}

func (r *mutationResolver) RefreshToken(ctx context.Context, refreshToken string) (*model.AuthToken, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().RefreshToken(ctx, &userpb.RefreshRequest{RefreshToken: refreshToken})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.AuthToken{AccessToken: res.GetNewSessionToken(), RefreshToken: res.GetNewRefreshToken(), ExpiresIn: int(res.GetExpiryTime())}, nil
}

func (r *mutationResolver) Logout(ctx context.Context) (bool, error) {
	user, _ := auth.UserFromContext(ctx)
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().Logout(ctx, &userpb.LogoutRequest{SessionToken: user.Token})
	if err != nil {
		return false, grpcerr.From(ctx, proxy.UserService, err)
	}
	return res.GetMessage() == "true", nil
}

func (r *mutationResolver) UpdateProfile(ctx context.Context, input model.UpdateProfileInput) (*model.UserMessage, error) {
	user, _ := auth.UserFromContext(ctx)
	id, err := strconv.Atoi(user.ID)
	if err != nil {
		return nil, auth.NewError(ctx, auth.CodeUnauthenticated, "invalid session user")
	}
	return r.updateProfile(ctx, &userpb.UpdateProfileRequest{
		Id:      int32(id),
		Name:    deref(input.Name),
		Email:   deref(input.Email),
		Phone:   deref(input.Phone),
		Address: deref(input.Address),
		Sex:     deref(input.Sex),
	})
}

func (r *mutationResolver) AdminUpdateProfile(ctx context.Context, input model.AdminUpdateProfileInput) (*model.UserMessage, error) {
	return r.updateProfile(ctx, &userpb.UpdateProfileRequest{
		Id:      int32(input.ID),
		Name:    deref(input.Name),
		Email:   deref(input.Email),
		Phone:   deref(input.Phone),
		Address: deref(input.Address),
		Sex:     deref(input.Sex),
	})
}

func (r *mutationResolver) CreateUserRole(ctx context.Context, input model.UserRoleInput) (*model.UserRoleMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().CreateRole(ctx, &userpb.CreateRoleResquest{Id: derefInt32(input.ID), Name: input.Name})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserRoleMessage{Message: res.GetMessage(), Role: toUserRole(res.GetRole())}, nil
}

func (r *mutationResolver) UpdateUserRole(ctx context.Context, id int, name string) (*model.UserRoleMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().UpdateRole(ctx, &userpb.UpdateRoleRequest{Id: int32(id), Name: name})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserRoleMessage{Message: res.GetMessage(), Role: toUserRole(res.GetRole())}, nil
}

func (r *mutationResolver) DeleteUserRole(ctx context.Context, id int) (string, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().DeleteRole(ctx, &userpb.DeleteRoleRequest{Id: int32(id)})
	if err != nil {
		return "", grpcerr.From(ctx, proxy.UserService, err)
	}
	return res.GetMessage(), nil
}

func (r *mutationResolver) AssignUserRole(ctx context.Context, input model.AssignUserRoleInput) (*model.UserRoleMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().UpdateItemRole(ctx, &userpb.UpdateRoleItemRequest{UserId: int32(input.UserID), RoleId: int32(input.RoleID)})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserRoleMessage{Message: res.GetMessage(), Role: toUserRole(res.GetRole())}, nil
}

func (r *mutationResolver) AddUserRole(ctx context.Context, input model.AssignUserRoleInput) (*model.UserRoleMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().AddRole(ctx, &userpb.AddRoleItemRequest{UserId: int32(input.UserID), RoleId: int32(input.RoleID)})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserRoleMessage{Message: res.GetMessage(), Role: toUserRole(res.GetRole())}, nil
}

func (r *mutationResolver) RemoveUserRoles(ctx context.Context, userID int) (string, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().RemoveRole(ctx, &userpb.RemoveRoleItemRequest{UserId: int32(userID)})
	if err != nil {
		return "", grpcerr.From(ctx, proxy.UserService, err)
	}
	return res.GetMessage(), nil
}

func (r *mutationResolver) SetupTotp(ctx context.Context) (*model.TotpSetup, error) {
	return restCall[*model.TotpSetup](ctx, r.Proxy, "user-service", rest.Request{Method: "POST", Path: "/api/v1/mfa/totp/setup", Body: map[string]any{}}, "")
}

func (r *mutationResolver) EnableTotp(ctx context.Context, code string) (*model.MfaRecoveryCodes, error) {
	return restCall[*model.MfaRecoveryCodes](ctx, r.Proxy, "user-service", rest.Request{Method: "POST", Path: "/api/v1/mfa/totp/enable", Body: map[string]any{"code": code}}, "")
}

func (r *mutationResolver) DisableTotp(ctx context.Context, code string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "user-service", rest.Request{Method: "POST", Path: "/api/v1/mfa/totp/disable", Body: map[string]any{"code": code}}, "")
}

func (r *mutationResolver) RegenerateMfaRecoveryCodes(ctx context.Context, code string) (*model.MfaRecoveryCodes, error) {
	return restCall[*model.MfaRecoveryCodes](ctx, r.Proxy, "user-service", rest.Request{Method: "POST", Path: "/api/v1/mfa/totp/recovery-codes", Body: map[string]any{"code": code}}, "")
}

func (r *mutationResolver) VerifyTotpLogin(ctx context.Context, mfaToken string, code string) (*model.AuthToken, error) {
	var res struct {
		SessionToken string `json:"session_token"`
		RefreshToken string `json:"refresh_token"`
		ExpiryTime   int64  `json:"expiry_time"`
	}
	body := map[string]string{"mfa_token": mfaToken, "code": code}
	if err := r.Proxy.REST(proxy.UserService).Post(ctx, "/api/v1/mfa/totp/verify", body, &res); err != nil {
		return nil, err
	}
	return &model.AuthToken{AccessToken: res.SessionToken, RefreshToken: res.RefreshToken, ExpiresIn: int(res.ExpiryTime)}, nil
}

func (r *queryResolver) Me(ctx context.Context) (*model.User, error) {
	user, _ := auth.UserFromContext(ctx)
	if user.Username == "" {
		return nil, auth.NewError(ctx, auth.CodeUnauthenticated, "session has no username")
	}
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().FindUser(ctx, &userpb.FindUserRequest{Username: &user.Username})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	for _, u := range res.GetUsers() {
		if strconv.Itoa(int(u.GetId())) == user.ID {
			return toUser(u), nil
		}
	}
	return nil, auth.NewError(ctx, auth.CodeNotFound, "user not found")
}

func (r *queryResolver) Users(ctx context.Context, input *model.FindUsersInput) (*model.UserList, error) {
	req := &userpb.FindUserRequest{}
	if input != nil {
		req.Username, req.Email, req.Page, req.Limit = input.Username, input.Email, int32Ptr(input.Page), int32Ptr(input.Limit)
	}
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().FindUser(ctx, req)
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	out := &model.UserList{Total: int(res.GetTotal()), Users: []*model.User{}}
	for _, u := range res.GetUsers() {
		out.Users = append(out.Users, toUser(u))
	}
	return out, nil
}

func (r *queryResolver) UserRole(ctx context.Context, id int) (*model.UserRole, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().GetRole(ctx, &userpb.GetRoleRequest{Id: int32(id)})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return toUserRole(res.GetRole()), nil
}

func (r *queryResolver) UserRoles(ctx context.Context, page *int, limit *int) (*model.UserRoleList, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().ListRoles(ctx, &userpb.ListRolesRequest{Page: derefInt32(page), Limit: derefInt32(limit)})
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	out := &model.UserRoleList{
		Roles:      []*model.UserRole{},
		Limit:      int(res.GetLimit()),
		Page:       int(res.GetPage()),
		Sort:       res.GetSort(),
		TotalRows:  int(res.GetTotalRows()),
		TotalPages: int(res.GetTotalPages()),
	}
	for _, role := range res.GetRoles() {
		out.Roles = append(out.Roles, toUserRole(role))
	}
	return out, nil
}

func (r *queryResolver) MfaStatus(ctx context.Context) (*model.MfaStatus, error) {
	return restCall[*model.MfaStatus](ctx, r.Proxy, "user-service", rest.Request{Method: "GET", Path: "/api/v1/mfa/totp"}, "")
}
