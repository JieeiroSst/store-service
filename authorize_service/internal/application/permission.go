package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
	"github.com/casbin/casbin/v2"
	"go.uber.org/zap"
)

type permissionService struct {
	enforcers *EnforcerProvider
}

func NewPermissionService(enforcers *EnforcerProvider) port.PermissionUsecase {
	return &permissionService{enforcers: enforcers}
}

// ─── assignments ──────────────────────────────────────────────────────────────

func (s *permissionService) AssignRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error) {
	e, sub, roles, err := s.prepare(ctx, userID, roles)
	if err != nil {
		return model.UserRoles{}, err
	}

	if len(roles) > 0 {
		if _, err := e.AddGroupingPoliciesEx(groupingRules(sub, roles)); err != nil {
			logger.WithContext(ctx).Error("AssignRoles", zap.String("user_id", userID), zap.Error(err))
			return model.UserRoles{}, common.ErrDBFailed
		}
	}
	return s.userRoles(e, userID)
}

func (s *permissionService) RevokeRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error) {
	if userID == "" {
		return model.UserRoles{}, common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return model.UserRoles{}, err
	}

	roles = dedupe(roles)
	for _, r := range roles {
		if def, ok := model.FindRoleDefinition(r); ok && !def.Assignable {
			return model.UserRoles{}, fmt.Errorf("%w: %q", common.ErrRoleNotAssignable, r)
		}
	}

	sub := model.UserSubject(userID)
	for _, role := range roles {
		if _, err := e.DeleteRoleForUser(sub, role); err != nil {
			logger.WithContext(ctx).Error("RevokeRoles", zap.String("user_id", userID), zap.Error(err))
			return model.UserRoles{}, common.ErrDBFailed
		}
	}
	return s.userRoles(e, userID)
}

func (s *permissionService) SetUserRoles(ctx context.Context, userID string, roles []string) (model.UserRoles, error) {
	e, sub, roles, err := s.prepare(ctx, userID, roles)
	if err != nil {
		return model.UserRoles{}, err
	}

	current, err := e.GetRolesForUser(sub)
	if err != nil {
		return model.UserRoles{}, common.ErrEnforcerFailed
	}

	// A super_admin granted through bootstrap is never dropped by a
	// SetUserRoles call coming from another service.
	want := toSet(roles)
	for _, r := range current {
		if def, ok := model.FindRoleDefinition(r); ok && !def.Assignable {
			want[r] = struct{}{}
		}
	}

	var toRemove, toAdd []string
	for _, r := range current {
		if _, ok := want[r]; !ok {
			toRemove = append(toRemove, r)
		}
	}
	have := toSet(current)
	for _, r := range roles {
		if _, ok := have[r]; !ok {
			toAdd = append(toAdd, r)
		}
	}

	if len(toRemove) > 0 {
		if _, err := e.RemoveGroupingPolicies(groupingRules(sub, toRemove)); err != nil {
			logger.WithContext(ctx).Error("SetUserRoles: remove", zap.String("user_id", userID), zap.Error(err))
			return model.UserRoles{}, common.ErrDBFailed
		}
	}
	if len(toAdd) > 0 {
		if _, err := e.AddGroupingPoliciesEx(groupingRules(sub, toAdd)); err != nil {
			logger.WithContext(ctx).Error("SetUserRoles: add", zap.String("user_id", userID), zap.Error(err))
			return model.UserRoles{}, common.ErrDBFailed
		}
	}
	return s.userRoles(e, userID)
}

func (s *permissionService) RemoveUser(ctx context.Context, userID string) error {
	if userID == "" {
		return common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return err
	}
	if _, err := e.DeleteRolesForUser(model.UserSubject(userID)); err != nil {
		logger.WithContext(ctx).Error("RemoveUser", zap.String("user_id", userID), zap.Error(err))
		return common.ErrDBFailed
	}
	return nil
}

// ─── queries ──────────────────────────────────────────────────────────────────

func (s *permissionService) GetUserRoles(ctx context.Context, userID string) (model.UserRoles, error) {
	if userID == "" {
		return model.UserRoles{}, common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return model.UserRoles{}, err
	}
	return s.userRoles(e, userID)
}

func (s *permissionService) GetUserPermissions(ctx context.Context, userID string) ([]model.Permission, error) {
	if userID == "" {
		return nil, common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := e.GetImplicitPermissionsForUser(model.UserSubject(userID))
	if err != nil {
		return nil, common.ErrEnforcerFailed
	}
	return rowsToPermissions(rows), nil
}

func (s *permissionService) CheckPermission(ctx context.Context, userID, obj, act string) (bool, error) {
	if userID == "" {
		return false, common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return false, err
	}
	ok, err := e.Enforce(model.UserSubject(userID), obj, act)
	if err != nil {
		logger.WithContext(ctx).Error("CheckPermission", zap.Error(err))
		return false, common.ErrEnforcerFailed
	}
	return ok, nil
}

// ListRoles returns the built-in catalog merged with custom roles found in
// casbin_rule, each with the permissions currently stored for it.
func (s *permissionService) ListRoles(ctx context.Context) ([]model.RoleDefinition, error) {
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return nil, err
	}
	names, err := s.knownRoles(e)
	if err != nil {
		return nil, err
	}

	out := make([]model.RoleDefinition, 0, len(names))
	for name := range names {
		def, ok := model.FindRoleDefinition(name)
		if !ok {
			def = model.RoleDefinition{Name: name, Assignable: true}
		}

		rows, err := e.GetFilteredPolicy(0, name)
		if err != nil {
			return nil, common.ErrEnforcerFailed
		}
		def.Permissions = rowsToPermissions(rows)

		parents, err := e.GetRolesForUser(name)
		if err != nil {
			return nil, common.ErrEnforcerFailed
		}
		def.Inherits = parents
		out = append(out, def)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Rank != out[j].Rank {
			return out[i].Rank < out[j].Rank
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// ─── seeding ──────────────────────────────────────────────────────────────────

// SeedDefaults only adds missing rules; it never deletes, so permissions an
// admin added or changed through the casbin CRUD API survive restarts.
func (s *permissionService) SeedDefaults(ctx context.Context, superAdminUserIDs []string) error {
	lg := logger.WithContext(ctx)

	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return err
	}

	var policies, groupings [][]string
	for _, role := range model.DefaultRoles {
		for _, p := range role.Permissions {
			policies = append(policies, []string{role.Name, p.Object, p.Action})
		}
		for _, parent := range role.Inherits {
			// g, child, parent: the child inherits every permission of parent.
			groupings = append(groupings, []string{role.Name, parent})
		}
	}
	for _, id := range dedupe(superAdminUserIDs) {
		groupings = append(groupings, []string{model.UserSubject(id), model.RoleSuperAdmin})
	}

	if _, err := e.AddPoliciesEx(policies); err != nil {
		lg.Error("SeedDefaults: policies", zap.Error(err))
		return fmt.Errorf("seed policies: %w", err)
	}
	if _, err := e.AddGroupingPoliciesEx(groupings); err != nil {
		lg.Error("SeedDefaults: groupings", zap.Error(err))
		return fmt.Errorf("seed groupings: %w", err)
	}

	lg.Info("SeedDefaults: role catalog ready",
		zap.Int("policies", len(policies)),
		zap.Int("groupings", len(groupings)))
	return nil
}

// ─── helpers ──────────────────────────────────────────────────────────────────

// prepare validates a role-assignment request coming from another service.
func (s *permissionService) prepare(ctx context.Context, userID string, roles []string) (*casbin.SyncedEnforcer, string, []string, error) {
	if userID == "" {
		return nil, "", nil, common.ErrInvalidUserID
	}
	e, err := s.enforcers.Get(ctx)
	if err != nil {
		return nil, "", nil, err
	}

	known, err := s.knownRoles(e)
	if err != nil {
		return nil, "", nil, err
	}

	roles = dedupe(roles)
	for _, r := range roles {
		if _, ok := known[r]; !ok {
			return nil, "", nil, fmt.Errorf("%w: %q", common.ErrUnknownRole, r)
		}
		if def, ok := model.FindRoleDefinition(r); ok && !def.Assignable {
			return nil, "", nil, fmt.Errorf("%w: %q", common.ErrRoleNotAssignable, r)
		}
	}
	return e, model.UserSubject(userID), roles, nil
}

// knownRoles = built-in roles ∪ every non-user subject of a p rule ∪ every
// role appearing on the right side of a g rule.
func (s *permissionService) knownRoles(e *casbin.SyncedEnforcer) (map[string]struct{}, error) {
	out := map[string]struct{}{}
	for _, r := range model.DefaultRoles {
		out[r.Name] = struct{}{}
	}

	subjects, err := e.GetAllSubjects()
	if err != nil {
		return nil, common.ErrEnforcerFailed
	}
	roles, err := e.GetAllRoles()
	if err != nil {
		return nil, common.ErrEnforcerFailed
	}
	for _, name := range append(subjects, roles...) {
		if name != "" && !model.IsUserSubject(name) {
			out[name] = struct{}{}
		}
	}
	return out, nil
}

func (s *permissionService) userRoles(e *casbin.SyncedEnforcer, userID string) (model.UserRoles, error) {
	sub := model.UserSubject(userID)

	direct, err := e.GetRolesForUser(sub)
	if err != nil {
		return model.UserRoles{}, common.ErrEnforcerFailed
	}
	effective, err := e.GetImplicitRolesForUser(sub)
	if err != nil {
		return model.UserRoles{}, common.ErrEnforcerFailed
	}

	sortByRank(direct)
	sortByRank(effective)

	res := model.UserRoles{UserID: userID, Roles: direct, EffectiveRoles: effective}
	if len(direct) > 0 {
		res.PrimaryRole = direct[0]
	}
	return res, nil
}

// sortByRank orders roles from most to least privileged.
func sortByRank(roles []string) {
	sort.SliceStable(roles, func(i, j int) bool {
		ri, rj := model.RoleRank(roles[i]), model.RoleRank(roles[j])
		if ri != rj {
			return ri > rj
		}
		return roles[i] < roles[j]
	})
}

func groupingRules(sub string, roles []string) [][]string {
	rules := make([][]string, 0, len(roles))
	for _, r := range roles {
		rules = append(rules, []string{sub, r})
	}
	return rules
}

func rowsToPermissions(rows [][]string) []model.Permission {
	out := make([]model.Permission, 0, len(rows))
	for _, row := range rows {
		if len(row) < 3 {
			continue
		}
		out = append(out, model.Permission{GrantedBy: row[0], Object: row[1], Action: row[2]})
	}
	return out
}

func dedupe(in []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(in))
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func toSet(in []string) map[string]struct{} {
	out := make(map[string]struct{}, len(in))
	for _, v := range in {
		out[v] = struct{}{}
	}
	return out
}
