package rbac

import (
	"context"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	memberships port.MembershipRepository
	roles       port.RoleRepository
	users       port.UserRepository
}

func NewService(memberships port.MembershipRepository, roles port.RoleRepository, users port.UserRepository) port.RBACService {
	return &service{memberships: memberships, roles: roles, users: users}
}

func (s *service) HasPermission(ctx context.Context, userID, projectID uuid.UUID, permission string) (bool, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return false, err
	}
	if user == nil {
		return false, apperr.ErrUnauthorized
	}
	if user.IsAdmin {
		return true, nil
	}

	membership, err := s.memberships.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return false, err
	}
	if membership == nil {
		return false, nil
	}

	permissions, err := s.roles.ListPermissions(ctx, membership.RoleID)
	if err != nil {
		return false, err
	}
	for _, p := range permissions {
		if p.Name == permission {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) AddMember(ctx context.Context, projectID, userID, roleID uuid.UUID) (*model.ProjectMembership, error) {
	existing, err := s.memberships.GetByProjectAndUser(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, apperr.ErrConflict
	}
	m := &model.ProjectMembership{ProjectID: projectID, UserID: userID, RoleID: roleID}
	if err := s.memberships.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *service) UpdateMemberRole(ctx context.Context, membershipID, roleID uuid.UUID) error {
	return s.memberships.UpdateRole(ctx, membershipID, roleID)
}

func (s *service) RemoveMember(ctx context.Context, membershipID uuid.UUID) error {
	return s.memberships.Delete(ctx, membershipID)
}

func (s *service) ListMembers(ctx context.Context, projectID uuid.UUID) ([]model.ProjectMembership, error) {
	return s.memberships.ListByProject(ctx, projectID)
}

func (s *service) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.roles.List(ctx)
}
