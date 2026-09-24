package repository

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"gorm.io/gorm"
)

type fileRepository struct {
	port.Repository[model.ContractFile]
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) port.ContractFileRepository {
	return &fileRepository{Repository: New[model.ContractFile](db), db: db}
}

func (r *fileRepository) SetLatest(ctx context.Context, id uint, latest bool) (bool, error) {
	res := conn(ctx, r.db).Model(&model.ContractFile{}).
		Where("id = ? AND latest = ?", id, !latest).
		Update("latest", latest)
	if res.Error != nil {
		return false, common.ErrDBFailed
	}
	return res.RowsAffected > 0, nil
}

func (r *fileRepository) Lineage(ctx context.Context, root uint) ([]model.ContractFile, error) {
	var files []model.ContractFile
	err := conn(ctx, r.db).Unscoped().
		Where("id = ? OR lineage_id = ?", root, root).
		Order("version, id").Find(&files).Error
	if err != nil {
		return nil, common.ErrDBFailed
	}
	return files, nil
}
