package repository

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
)

type dataSourceRepository struct {
	db *gorm.DB
}

func NewDataSourceRepository(db *gorm.DB) port.DataSourceRepository {
	return &dataSourceRepository{db: db}
}

func (r *dataSourceRepository) Create(ctx context.Context, ds *model.DataSource) error {
	return translate(r.db.WithContext(ctx).Create(ds).Error, "data source "+ds.Name)
}

func (r *dataSourceRepository) Get(ctx context.Context, id string) (*model.DataSource, error) {
	var ds model.DataSource
	if err := r.db.WithContext(ctx).First(&ds, "id = ?", id).Error; err != nil {
		return nil, translate(err, "data source "+id)
	}
	return &ds, nil
}

func (r *dataSourceRepository) ListActive(ctx context.Context) ([]*model.DataSource, error) {
	var dss []*model.DataSource
	return dss, r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Order("created_at desc").
		Find(&dss).Error
}

func (r *dataSourceRepository) Update(ctx context.Context, ds *model.DataSource) error {
	return affected(r.db.WithContext(ctx).Model(ds).Select("*").Omit("created_at").Updates(ds), "data source "+ds.ID)
}
