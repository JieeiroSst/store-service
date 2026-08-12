package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type membershipRepository struct {
	db *gorm.DB
}

func NewMembershipRepository(db *gorm.DB) port.MembershipRepository {
	return &membershipRepository{db: db}
}

func (r *membershipRepository) Create(ctx context.Context, m *model.ProjectMembership) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *membershipRepository) GetByProjectAndUser(ctx context.Context, projectID, userID uuid.UUID) (*model.ProjectMembership, error) {
	var m model.ProjectMembership
	if err := r.db.WithContext(ctx).First(&m, "project_id = ? AND user_id = ?", projectID, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &m, nil
}

func (r *membershipRepository) ListByProject(ctx context.Context, projectID uuid.UUID) ([]model.ProjectMembership, error) {
	var memberships []model.ProjectMembership
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *membershipRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.ProjectMembership, error) {
	var memberships []model.ProjectMembership
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&memberships).Error; err != nil {
		return nil, err
	}
	return memberships, nil
}

func (r *membershipRepository) UpdateRole(ctx context.Context, id, roleID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&model.ProjectMembership{}).
		Where("id = ?", id).
		Update("role_id", roleID).Error
}

func (r *membershipRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.ProjectMembership{}, "id = ?", id).Error
}
