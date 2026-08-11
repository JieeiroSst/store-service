package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/livestream-service/internal/domain/model"
	"github.com/JIeeiroSst/livestream-service/internal/domain/port"
	"gorm.io/gorm"
)

type streamRepository struct {
	db *gorm.DB
}

func NewStreamRepository(db *gorm.DB) port.StreamRepository {
	return &streamRepository{db: db}
}

func (r *streamRepository) Create(ctx context.Context, stream *model.Stream) error {
	return r.db.WithContext(ctx).Create(stream).Error
}

func (r *streamRepository) GetActiveByRoomID(ctx context.Context, roomID string) (*model.Stream, error) {
	var stream model.Stream
	err := r.db.WithContext(ctx).
		Where("room_id = ? AND status IN ?", roomID, []model.StreamStatus{model.StreamStatusPending, model.StreamStatusLive}).
		Order("created_at desc").
		First(&stream).Error
	if err != nil {
		return nil, err
	}
	return &stream, nil
}

func (r *streamRepository) Complete(ctx context.Context, streamID string, endedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.Stream{}).
		Where("id = ?", streamID).
		Updates(map[string]any{"status": model.StreamStatusEnded, "ended_at": endedAt}).Error
}
