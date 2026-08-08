package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/bot-service/internal/domain/model"
	"github.com/JIeeiroSst/bot-service/internal/domain/port"
	"gorm.io/gorm"
)

type postRepo struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) port.PostRepository {
	return &postRepo{db: db}
}

func (r *postRepo) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepo) Update(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Save(post).Error
}

func (r *postRepo) GetByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func (r *postRepo) List(ctx context.Context, limit, offset int32, status model.PostStatus, campaign string) ([]model.Post, error) {
	q := r.db.WithContext(ctx).Order("created_at desc")
	if status != "" {
		q = q.Where("status = ?", status)
	}
	if campaign != "" {
		q = q.Where("campaign = ?", campaign)
	}
	if limit > 0 {
		q = q.Limit(int(limit))
	}
	if offset > 0 {
		q = q.Offset(int(offset))
	}

	var posts []model.Post
	if err := q.Find(&posts).Error; err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepo) DueForPublish(ctx context.Context, before time.Time) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("status = ? AND (cron_expr IS NULL OR cron_expr = '') AND scheduled_at <= ?", model.PostStatusScheduled, before).
		Order("scheduled_at asc").
		Limit(50).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func (r *postRepo) ActiveRecurring(ctx context.Context) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.WithContext(ctx).
		Where("status = ? AND cron_expr IS NOT NULL AND cron_expr <> ''", model.PostStatusActive).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}
