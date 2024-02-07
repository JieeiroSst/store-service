package repository

import (
	"context"

	"github.com/JIeeiroSst/message-service/model"
	"gorm.io/gorm"
)

type Tracks interface {
	TrackProducer(ctx context.Context, model model.Track) error
	TrackConsume(ctx context.Context, model model.Track) error
}

type TrackRepo struct {
	db *gorm.DB
}

func NewTrackRepo(db *gorm.DB) *TrackRepo {
	return &TrackRepo{
		db: db,
	}
}

func (r *TrackRepo) TrackProducer(ctx context.Context, model model.Track) error {
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	return nil
}

func (r *TrackRepo) TrackConsume(ctx context.Context, model model.Track) error {
	if err := r.db.Create(model).Error; err != nil {
		return err
	}
	return nil
}
