package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/authorizepb"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/permissionpb"
	"google.golang.org/grpc"
)

func (r *mutationResolver) AuthzAssignRoles(ctx context.Context, userID string, roles []string) (*model.AuthzUserRoles, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.UserRolesResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).AssignRoles(ctx, &permissionpb.AssignRolesRequest{UserId: userID, Roles: roles})
	})
	if err != nil {
		return nil, err
	}
	return toAuthzUserRoles(res), nil
}

func (r *mutationResolver) AuthzRevokeRoles(ctx context.Context, userID string, roles []string) (*model.AuthzUserRoles, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.UserRolesResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).RevokeRoles(ctx, &permissionpb.RevokeRolesRequest{UserId: userID, Roles: roles})
	})
	if err != nil {
		return nil, err
	}
	return toAuthzUserRoles(res), nil
}

func (r *mutationResolver) AuthzSetUserRoles(ctx context.Context, userID string, roles []string) (*model.AuthzUserRoles, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.UserRolesResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).SetUserRoles(ctx, &permissionpb.SetUserRolesRequest{UserId: userID, Roles: roles})
	})
	if err != nil {
		return nil, err
	}
	return toAuthzUserRoles(res), nil
}

func (r *mutationResolver) AuthzRemoveUser(ctx context.Context, userID string) (bool, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.RemoveUserResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).RemoveUser(ctx, &permissionpb.RemoveUserRequest{UserId: userID})
	})
	if err != nil {
		return false, err
	}
	return res.GetSuccess(), nil
}

func (r *mutationResolver) CreateCasbinRule(ctx context.Context, input model.CasbinRuleInput) (*model.AuthzResult, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.CreateCasbinRuleResponse, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).CreateCasbinRule(ctx, &authorizepb.CasbinRule{
			Ptype: input.Ptype,
			V0:    deref(input.V0),
			V1:    deref(input.V1),
			V2:    deref(input.V2),
			V3:    deref(input.V3),
			V4:    deref(input.V4),
			V5:    deref(input.V5),
		})
	})
	if err != nil {
		return nil, err
	}
	return &model.AuthzResult{Message: res.GetMessage(), Error: res.GetError()}, nil
}

func (r *mutationResolver) UpdateCasbinRule(ctx context.Context, id int, field string, value string) (*model.AuthzResult, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.UpdateCasbinRuleResponse, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).UpdateCasbinRule(ctx, &authorizepb.UpdateCasbinRuleRequest{Id: int32(id), Field: field, Value: value})
	})
	if err != nil {
		return nil, err
	}
	return &model.AuthzResult{Message: res.GetMessage(), Error: res.GetError()}, nil
}

func (r *mutationResolver) DeleteCasbinRule(ctx context.Context, id int) (bool, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.DeleteCasbinRuleResponse, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).DeleteCasbinRule(ctx, &authorizepb.CasbinRuleId{Id: int32(id)})
	})
	if err != nil {
		return false, err
	}
	return res.GetSuccess(), nil
}

func (r *mutationResolver) CreateOtp(ctx context.Context, userID string, username string) (*model.OtpToken, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.CreateOTPResponse, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).CreateOTP(ctx, &authorizepb.CreateOTPRequest{UserId: userID, Username: username})
	})
	if err != nil {
		return nil, err
	}
	return &model.OtpToken{Otp: res.GetOtp(), ExpiresAt: int(res.GetExpiresAt())}, nil
}

func (r *mutationResolver) VerifyOtp(ctx context.Context, otp string) (*model.OtpVerification, error) {
	user, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.AuthorizeOTPResponse, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).AuthorizeOTP(ctx, &authorizepb.AuthorizeOTPRequest{Otp: otp, UserId: user.ID, Username: user.Username})
	})
	if err != nil {
		return nil, err
	}
	return &model.OtpVerification{Valid: res.GetValid(), Message: res.GetMessage()}, nil
}

func (r *queryResolver) MyRoles(ctx context.Context) (*model.AuthzUserRoles, error) {
	user, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	return r.AuthzUserRoles(ctx, user.ID)
}

func (r *queryResolver) MyPermissions(ctx context.Context) ([]*model.AuthzPermission, error) {
	user, err := currentUser(ctx)
	if err != nil {
		return nil, err
	}
	return r.AuthzUserPermissions(ctx, user.ID)
}

func (r *queryResolver) CheckMyPermission(ctx context.Context, object string, action string) (bool, error) {
	user, err := currentUser(ctx)
	if err != nil {
		return false, err
	}
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.CheckPermissionResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).CheckPermission(ctx, &permissionpb.CheckPermissionRequest{UserId: user.ID, Object: object, Action: action})
	})
	if err != nil {
		return false, err
	}
	return res.GetAllowed(), nil
}

func (r *queryResolver) AuthzRoles(ctx context.Context) ([]*model.AuthzRole, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.ListRolesResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).ListRoles(ctx, &permissionpb.ListRolesRequest{})
	})
	if err != nil {
		return nil, err
	}
	out := make([]*model.AuthzRole, 0, len(res.GetRoles()))
	for _, role := range res.GetRoles() {
		out = append(out, &model.AuthzRole{
			Name:        role.GetName(),
			Description: role.GetDescription(),
			Rank:        int(role.GetRank()),
			Inherits:    nonNil(role.GetInherits()),
			Permissions: toAuthzPermissions(role.GetPermissions()),
			Assignable:  role.GetAssignable(),
		})
	}
	return out, nil
}

func (r *queryResolver) AuthzUserRoles(ctx context.Context, userID string) (*model.AuthzUserRoles, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.UserRolesResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).GetUserRoles(ctx, &permissionpb.GetUserRolesRequest{UserId: userID})
	})
	if err != nil {
		return nil, err
	}
	return toAuthzUserRoles(res), nil
}

func (r *queryResolver) AuthzUserPermissions(ctx context.Context, userID string) ([]*model.AuthzPermission, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*permissionpb.GetUserPermissionsResponse, error) {
		return permissionpb.NewPermissionServiceClient(conn).GetUserPermissions(ctx, &permissionpb.GetUserPermissionsRequest{UserId: userID})
	})
	if err != nil {
		return nil, err
	}
	return toAuthzPermissions(res.GetPermissions()), nil
}

func (r *queryResolver) CasbinRules(ctx context.Context, page *int, limit *int, sort *string) (*model.CasbinRuleList, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.CasbinRuleList, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).GetCasbinRules(ctx, &authorizepb.CasbinRequest{Page: derefInt32(page), Limit: derefInt32(limit), Sort: deref(sort)})
	})
	if err != nil {
		return nil, err
	}
	out := &model.CasbinRuleList{
		Rows:       []*model.CasbinRule{},
		Limit:      int(res.GetLimit()),
		Page:       int(res.GetPage()),
		Sort:       res.GetSort(),
		TotalRows:  int(res.GetTotalRows()),
		TotalPages: int(res.GetTotalPages()),
	}
	for _, row := range res.GetRows() {
		out.Rows = append(out.Rows, toCasbinRule(row))
	}
	return out, nil
}

func (r *queryResolver) CasbinRule(ctx context.Context, id int) (*model.CasbinRule, error) {
	res, err := grpcCall(ctx, r.Proxy, proxy.AuthorizeService, func(ctx context.Context, conn *grpc.ClientConn) (*authorizepb.CasbinRule, error) {
		return authorizepb.NewAuthorizeServiceClient(conn).GetCasbinRuleById(ctx, &authorizepb.CasbinRuleId{Id: int32(id)})
	})
	if err != nil {
		return nil, err
	}
	return toCasbinRule(res), nil
}
