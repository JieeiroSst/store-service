package repository

import (
	"context"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) port.MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Create(ctx context.Context, media *model.Media) error {
	return r.db.WithContext(ctx).Create(media).Error
}
