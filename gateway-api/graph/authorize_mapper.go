package graph

import (
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/authorizepb"
	"github.com/JIeeiroSst/gateway-api/proxy/pb/permissionpb"
)

func toAuthzUserRoles(res *permissionpb.UserRolesResponse) *model.AuthzUserRoles {
	return &model.AuthzUserRoles{
		UserID:         res.GetUserId(),
		Roles:          nonNil(res.GetRoles()),
		EffectiveRoles: nonNil(res.GetEffectiveRoles()),
		PrimaryRole:    res.GetPrimaryRole(),
	}
}

func toAuthzPermissions(in []*permissionpb.Permission) []*model.AuthzPermission {
	out := make([]*model.AuthzPermission, 0, len(in))
	for _, p := range in {
		out = append(out, &model.AuthzPermission{Object: p.GetObject(), Action: p.GetAction(), GrantedBy: p.GetGrantedBy()})
	}
	return out
}

func toCasbinRule(r *authorizepb.CasbinRule) *model.CasbinRule {
	if r == nil {
		return nil
	}
	return &model.CasbinRule{
		ID:    int(r.GetId()),
		Ptype: r.GetPtype(),
		V0:    r.GetV0(),
		V1:    r.GetV1(),
		V2:    r.GetV2(),
		V3:    r.GetV3(),
		V4:    r.GetV4(),
		V5:    r.GetV5(),
	}
}

func nonNil[T any](s []T) []T {
	if s == nil {
		return []T{}
	}
	return s
}
