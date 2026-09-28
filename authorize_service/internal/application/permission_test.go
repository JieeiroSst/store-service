package application

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
)

// common.RBACModelPath is relative to the module root.
func TestMain(m *testing.M) {
	if err := os.Chdir(filepath.Join("..", "..")); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func newTestPermissionService(t *testing.T, superAdmins ...string) port.PermissionUsecase {
	t.Helper()
	path := filepath.Join(t.TempDir(), "policy.csv")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	svc := NewPermissionService(NewEnforcerProvider(fileadapter.NewAdapter(path)))
	if err := svc.SeedDefaults(context.Background(), superAdmins); err != nil {
		t.Fatalf("SeedDefaults: %v", err)
	}
	return svc
}

func mustAllow(t *testing.T, svc port.PermissionUsecase, userID, obj, act string, want bool) {
	t.Helper()
	got, err := svc.CheckPermission(context.Background(), userID, obj, act)
	if err != nil {
		t.Fatalf("CheckPermission(%s %s %s): %v", userID, obj, act, err)
	}
	if got != want {
		t.Errorf("CheckPermission(%s, %s, %s) = %v, want %v", userID, obj, act, got, want)
	}
}

func TestRoleHierarchy(t *testing.T) {
	ctx := context.Background()
	svc := newTestPermissionService(t, "1")

	for id, role := range map[string]string{"10": model.RoleUser, "20": model.RoleOperator, "30": model.RoleAdmin} {
		if _, err := svc.AssignRoles(ctx, id, []string{role}); err != nil {
			t.Fatalf("AssignRoles(%s, %s): %v", id, role, err)
		}
	}

	cases := []struct {
		obj, act              string
		user, op, admin, root bool
	}{
		{"/products/5", "GET", true, true, true, true},
		{"/products/5", "DELETE", false, false, false, true},
		{"/users/me", "PUT", true, true, true, true},
		{"/orders", "POST", true, true, true, true},
		{"/admin/orders/5", "PUT", false, true, true, true},
		{"/admin/products/5", "DELETE", false, true, true, true},
		{"/admin/users/5/lock", "POST", false, true, true, true},
		{"/admin/users/5", "DELETE", false, false, true, true},
		{"/admin/roles", "POST", false, false, true, true},
		{"/admin/roles", "GET", false, true, true, true},
		{"/admin/configs/x", "PUT", false, false, true, true},
		{"/anything/at/all", "DELETE", false, false, false, true},
	}
	for _, c := range cases {
		mustAllow(t, svc, "10", c.obj, c.act, c.user)
		mustAllow(t, svc, "20", c.obj, c.act, c.op)
		mustAllow(t, svc, "30", c.obj, c.act, c.admin)
		mustAllow(t, svc, "1", c.obj, c.act, c.root)
	}

	// A user without any role can do nothing.
	mustAllow(t, svc, "99", "/products/5", "GET", false)
}

func TestActionRegexIsAnchored(t *testing.T) {
	svc := newTestPermissionService(t)
	if _, err := svc.AssignRoles(context.Background(), "10", []string{model.RoleUser}); err != nil {
		t.Fatal(err)
	}
	mustAllow(t, svc, "10", "/products/5", "GET", true)
	mustAllow(t, svc, "10", "/products/5", "TARGET", false)
}

func TestAssignRolesValidation(t *testing.T) {
	ctx := context.Background()
	svc := newTestPermissionService(t)

	if _, err := svc.AssignRoles(ctx, "", []string{model.RoleUser}); !errors.Is(err, common.ErrInvalidUserID) {
		t.Errorf("empty user id: got %v", err)
	}
	if _, err := svc.AssignRoles(ctx, "10", []string{"ghost"}); !errors.Is(err, common.ErrUnknownRole) {
		t.Errorf("unknown role: got %v", err)
	}
	if _, err := svc.AssignRoles(ctx, "10", []string{model.RoleSuperAdmin}); !errors.Is(err, common.ErrRoleNotAssignable) {
		t.Errorf("super_admin: got %v", err)
	}

	// Assigning twice is idempotent.
	for i := 0; i < 2; i++ {
		res, err := svc.AssignRoles(ctx, "10", []string{model.RoleUser, model.RoleUser})
		if err != nil {
			t.Fatal(err)
		}
		if len(res.Roles) != 1 || res.PrimaryRole != model.RoleUser {
			t.Fatalf("unexpected roles: %+v", res)
		}
	}
}

func TestSetUserRolesAndPrimaryRole(t *testing.T) {
	ctx := context.Background()
	svc := newTestPermissionService(t, "1")

	res, err := svc.AssignRoles(ctx, "10", []string{model.RoleUser, model.RoleOperator})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrimaryRole != model.RoleOperator {
		t.Errorf("primary role = %q, want operator", res.PrimaryRole)
	}

	res, err = svc.SetUserRoles(ctx, "10", []string{model.RoleAdmin})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Roles) != 1 || res.Roles[0] != model.RoleAdmin {
		t.Errorf("roles after set = %v, want [admin]", res.Roles)
	}
	wantEffective := []string{model.RoleAdmin, model.RoleOperator, model.RoleUser}
	if len(res.EffectiveRoles) != len(wantEffective) {
		t.Fatalf("effective = %v, want %v", res.EffectiveRoles, wantEffective)
	}
	for i := range wantEffective {
		if res.EffectiveRoles[i] != wantEffective[i] {
			t.Errorf("effective = %v, want %v", res.EffectiveRoles, wantEffective)
		}
	}

	// Bootstrap super admin survives SetUserRoles and cannot be revoked.
	res, err = svc.SetUserRoles(ctx, "1", []string{model.RoleUser})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrimaryRole != model.RoleSuperAdmin {
		t.Errorf("super admin lost by SetUserRoles: %+v", res)
	}
	if _, err := svc.RevokeRoles(ctx, "1", []string{model.RoleSuperAdmin}); !errors.Is(err, common.ErrRoleNotAssignable) {
		t.Errorf("revoke super_admin: got %v", err)
	}

	if err := svc.RemoveUser(ctx, "10"); err != nil {
		t.Fatal(err)
	}
	res, err = svc.GetUserRoles(ctx, "10")
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Roles) != 0 || res.PrimaryRole != "" {
		t.Errorf("roles after RemoveUser = %+v", res)
	}
}

func TestListRolesAndPermissions(t *testing.T) {
	ctx := context.Background()
	svc := newTestPermissionService(t)

	roles, err := svc.ListRoles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(roles) != len(model.DefaultRoles) {
		t.Fatalf("ListRoles returned %d roles, want %d", len(roles), len(model.DefaultRoles))
	}
	byName := map[string]model.RoleDefinition{}
	for _, r := range roles {
		byName[r.Name] = r
	}
	for _, def := range model.DefaultRoles {
		got, ok := byName[def.Name]
		if !ok || len(got.Permissions) != len(def.Permissions) {
			t.Errorf("role %s: got %d perms (listed %v), want %d", def.Name, len(got.Permissions), ok, len(def.Permissions))
		}
	}
	for i := 1; i < len(roles); i++ {
		if roles[i-1].Rank > roles[i].Rank {
			t.Errorf("roles not sorted by rank: %s(%d) before %s(%d)", roles[i-1].Name, roles[i-1].Rank, roles[i].Name, roles[i].Rank)
		}
	}

	if _, err := svc.AssignRoles(ctx, "20", []string{model.RoleOperator}); err != nil {
		t.Fatal(err)
	}
	perms, err := svc.GetUserPermissions(ctx, "20")
	if err != nil {
		t.Fatal(err)
	}
	var fromUser, fromOperator bool
	for _, p := range perms {
		switch p.GrantedBy {
		case model.RoleUser:
			fromUser = true
		case model.RoleOperator:
			fromOperator = true
		}
	}
	if !fromUser || !fromOperator {
		t.Errorf("operator permissions should include user + operator rules, got %d rules", len(perms))
	}
}

// A CRM role on top of a platform role must never become the primary role:
// other services (car-rental's staff check, livestream's admin check) read
// only the primary role.
func TestCRMRolesStayFunctional(t *testing.T) {
	ctx := context.Background()
	svc := newTestPermissionService(t)

	res, err := svc.AssignRoles(ctx, "20", []string{model.RoleOperator, model.RoleCRMManager})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrimaryRole != model.RoleOperator {
		t.Errorf("primary role = %q, want operator", res.PrimaryRole)
	}
	want := map[string]bool{model.RoleCRMManager: true, model.RoleCRMStaff: true, model.RoleCRMViewer: true, model.RoleOperator: true, model.RoleUser: true}
	for _, r := range res.EffectiveRoles {
		delete(want, r)
	}
	if len(want) != 0 {
		t.Errorf("effective roles %v missing %v", res.EffectiveRoles, want)
	}

	mustAllow(t, svc, "20", "/crm/contracts/1/files", "DELETE", true)
	mustAllow(t, svc, "20", "/admin/users/5", "DELETE", false)

	// A plain user with only a CRM role has that role as primary.
	res, err = svc.AssignRoles(ctx, "21", []string{model.RoleCRMViewer})
	if err != nil {
		t.Fatal(err)
	}
	if res.PrimaryRole != model.RoleCRMViewer {
		t.Errorf("primary role = %q, want crm-viewer", res.PrimaryRole)
	}
	mustAllow(t, svc, "21", "/crm/contracts", "GET", true)
	mustAllow(t, svc, "21", "/crm/contracts", "POST", false)
}
