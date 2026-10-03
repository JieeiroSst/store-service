package roleitem

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type fakeRoleRepo struct {
	output.RoleRepository
	roles map[int]string
}

func (f fakeRoleRepo) Role(_ context.Context, id int) (*domain.Role, error) {
	name, ok := f.roles[id]
	if !ok {
		return nil, domain.ErrRoleNotFound
	}
	return &domain.Role{Id: id, Name: name}, nil
}

type call struct {
	method string
	userID int
	roles  []string
}

type fakeAuthorizer struct {
	output.Authorizer
	calls []call
}

func (f *fakeAuthorizer) AssignRoles(_ context.Context, userID int, roles ...string) (domain.UserRoles, error) {
	f.calls = append(f.calls, call{"assign", userID, roles})
	return domain.UserRoles{}, nil
}

func (f *fakeAuthorizer) SetUserRoles(_ context.Context, userID int, roles ...string) (domain.UserRoles, error) {
	f.calls = append(f.calls, call{"set", userID, roles})
	return domain.UserRoles{}, nil
}

func (f *fakeAuthorizer) RemoveUser(_ context.Context, userID int) error {
	f.calls = append(f.calls, call{"remove", userID, nil})
	return nil
}

func TestRoleItemGoesThroughAuthorize(t *testing.T) {
	ctx := context.Background()
	authz := &fakeAuthorizer{}
	svc := New(fakeRoleRepo{roles: map[int]string{1: "operator", 2: "admin"}}, authz)

	if res, err := svc.AddRoleItem(ctx, dto.AddRoleItemRequest{UserId: 7, RoleId: 1}); err != nil || res.Role.Name != "operator" {
		t.Fatalf("AddRoleItem = %+v, %v", res, err)
	}
	if _, err := svc.UpdateItemRole(ctx, dto.UpdateRoleItemRequest{UserId: 7, RoleId: 2}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.RemoveRoleItem(ctx, dto.RemoveRoleItemRequest{UserId: 7}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.AddRoleItem(ctx, dto.AddRoleItemRequest{UserId: 7, RoleId: 99}); err == nil {
		t.Error("AddRoleItem with unknown role id should fail")
	}

	want := []call{{"assign", 7, []string{"operator"}}, {"set", 7, []string{"admin"}}, {"remove", 7, nil}}
	if len(authz.calls) != len(want) {
		t.Fatalf("calls = %+v, want %+v", authz.calls, want)
	}
	for i, c := range want {
		got := authz.calls[i]
		if got.method != c.method || got.userID != c.userID || len(got.roles) != len(c.roles) || (len(c.roles) > 0 && got.roles[0] != c.roles[0]) {
			t.Errorf("call %d = %+v, want %+v", i, got, c)
		}
	}
}
