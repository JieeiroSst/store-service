package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) port.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(ctx context.Context, branch *model.Branch) error {
	if err := r.db.WithContext(ctx).Create(branch).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *branchRepository) GetByID(ctx context.Context, id int) (*model.Branch, error) {
	var branch model.Branch
	err := r.db.WithContext(ctx).First(&branch, "branch_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &branch, nil
}

func (r *branchRepository) List(ctx context.Context) ([]model.Branch, error) {
	var branches []model.Branch
	if err := r.db.WithContext(ctx).Find(&branches).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return branches, nil
}

func (r *branchRepository) Update(ctx context.Context, branch *model.Branch) error {
	if err := r.db.WithContext(ctx).Model(&model.Branch{}).Where("branch_id = ?", branch.BranchID).Updates(branch).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *branchRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Branch{}, "branch_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
