package repository

import (
	"context"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
)

type channelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) port.ChannelRepository {
	return &channelRepository{db: db}
}

func (r *channelRepository) Create(ctx context.Context, c *model.Channel) error {
	return translate(r.db.WithContext(ctx).Create(c).Error, "channel "+c.Name)
}

func (r *channelRepository) Get(ctx context.Context, id string) (*model.Channel, error) {
	var c model.Channel
	if err := r.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, translate(err, "channel "+id)
	}
	return &c, nil
}

func (r *channelRepository) List(ctx context.Context, f port.ChannelFilter) ([]*model.Channel, error) {
	q := r.db.WithContext(ctx).Model(&model.Channel{})
	if f.Type != "" {
		q = q.Where("type = ?", f.Type)
	}
	if f.Active != nil {
		q = q.Where("is_active = ?", *f.Active)
	}
	var channels []*model.Channel
	return channels, q.Order("created_at desc").Find(&channels).Error
}

func (r *channelRepository) Update(ctx context.Context, c *model.Channel) error {
	return affected(r.db.WithContext(ctx).Model(c).Select("*").Omit("created_at").Updates(c), "channel "+c.ID)
}

func (r *channelRepository) Delete(ctx context.Context, id string) error {
	return affected(r.db.WithContext(ctx).Delete(&model.Channel{}, "id = ?", id), "channel "+id)
}
