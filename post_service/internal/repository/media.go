package repository

import (
	"context"

	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type Medias interface {
	SaveMedia(ctx context.Context, args model.Media) error
	UpdateMedia(ctx context.Context, id int, args model.Media) error
}

type MediaRepo struct {
	db *gorm.DB
}

func NewMediaRepo(db *gorm.DB) *MediaRepo {
	return &MediaRepo{
		db: db,
	}
}

func (r *MediaRepo) SaveMedia(ctx context.Context, args model.Media) error {
	if err := r.db.Create(&args).Error; err != nil {
		return err
	}
	return nil
}

func (r *MediaRepo) UpdateMedia(ctx context.Context, id int, args model.Media) error {
	if err := r.db.Where("id = ?", id).Updates(&args).Error; err != nil {
		return err
	}
	return nil
}
