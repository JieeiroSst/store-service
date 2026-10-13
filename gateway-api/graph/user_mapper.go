package graph

import (
	"context"
	"time"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/grpcerr"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/userpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (r *mutationResolver) updateProfile(ctx context.Context, req *userpb.UpdateProfileRequest) (*model.UserMessage, error) {
	ctx, cancel := r.Proxy.User.WithTimeout(ctx)
	defer cancel()
	res, err := r.Proxy.User.API().UpdateProfile(ctx, req)
	if err != nil {
		return nil, grpcerr.From(ctx, proxy.UserService, err)
	}
	return &model.UserMessage{Message: res.GetMessage(), User: toUser(res.GetUser())}, nil
}

func toUser(u *userpb.User) *model.User {
	if u == nil {
		return nil
	}
	out := &model.User{
		ID:         int(u.GetId()),
		Username:   u.GetUsername(),
		Email:      u.GetEmail(),
		Name:       u.GetName(),
		Phone:      u.GetPhone(),
		Address:    u.GetAddress(),
		Sex:        u.GetSex(),
		Active:     u.GetChecked(),
		CreateTime: formatTimestamp(u.GetCreateTime()),
		UpdateTime: formatTimestamp(u.GetUpdateTime()),
		Roles:      []*model.UserRole{},
	}
	for _, role := range u.GetRoles() {
		out.Roles = append(out.Roles, toUserRole(role))
	}
	return out
}

func toUserRole(role *userpb.Role) *model.UserRole {
	if role == nil {
		return nil
	}
	out := &model.UserRole{ID: int(role.GetId()), Name: role.GetName(), Users: []*model.User{}}
	for _, u := range role.GetUsers() {
		out.Users = append(out.Users, toUser(u))
	}
	return out
}
func formatTimestamp(ts *timestamppb.Timestamp) *string {
	if ts == nil {
		return nil
	}
	s := ts.AsTime().Format(time.RFC3339)
	return &s
}
