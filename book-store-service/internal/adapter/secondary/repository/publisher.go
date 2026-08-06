package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/bookStore-service/common"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
	"gorm.io/gorm"
)

type publisherRepository struct {
	db *gorm.DB
}

func NewPublisherRepository(db *gorm.DB) port.PublisherRepository {
	return &publisherRepository{db: db}
}

func (r *publisherRepository) Create(ctx context.Context, publisher *model.Publisher) error {
	if err := r.db.WithContext(ctx).Create(publisher).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *publisherRepository) GetByID(ctx context.Context, id int) (*model.Publisher, error) {
	var publisher model.Publisher
	if err := r.db.WithContext(ctx).First(&publisher, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &publisher, nil
}

func (r *publisherRepository) Update(ctx context.Context, publisher *model.Publisher) error {
	if err := r.db.WithContext(ctx).Model(&model.Publisher{}).Where("id = ?", publisher.ID).Updates(publisher).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *publisherRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Publisher{}, "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *publisherRepository) List(ctx context.Context) ([]model.Publisher, error) {
	var publishers []model.Publisher
	if err := r.db.WithContext(ctx).Find(&publishers).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return publishers, nil
}
