package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"gorm.io/gorm"
)

type driverAssignmentRepository struct {
	db *gorm.DB
}

func NewDriverAssignmentRepository(db *gorm.DB) port.DriverAssignmentRepository {
	return &driverAssignmentRepository{db: db}
}

func (r *driverAssignmentRepository) Create(ctx context.Context, assignment *model.DriverAssignment) (*model.DriverAssignment, error) {
	if err := r.db.WithContext(ctx).Create(assignment).Error; err != nil {
		return nil, err
	}
	return assignment, nil
}

func (r *driverAssignmentRepository) GetByID(ctx context.Context, id int64) (*model.DriverAssignment, error) {
	var assignment model.DriverAssignment
	if err := r.db.WithContext(ctx).First(&assignment, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *driverAssignmentRepository) Update(ctx context.Context, assignment *model.DriverAssignment) (*model.DriverAssignment, error) {
	if err := r.db.WithContext(ctx).Save(assignment).Error; err != nil {
		return nil, err
	}
	return assignment, nil
}

func (r *driverAssignmentRepository) ListByDriver(ctx context.Context, driverID string) ([]model.DriverAssignment, error) {
	var assignments []model.DriverAssignment
	err := r.db.WithContext(ctx).Where("driver_id = ?", driverID).Order("assigned_at desc").Find(&assignments).Error
	return assignments, err
}
