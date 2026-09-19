package repository

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"gorm.io/gorm"
)

type followRepository struct {
	db *gorm.DB
}

func NewFollowRepository(db *gorm.DB) port.FollowRepository {
	return &followRepository{db: db}
}

func (r *followRepository) Create(ctx context.Context, follow *model.Follow) error {
	return r.db.WithContext(ctx).Create(follow).Error
}

func (r *followRepository) Delete(ctx context.Context, followerID, followedID string) error {
	return r.db.WithContext(ctx).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Delete(&model.Follow{}).Error
}

func (r *followRepository) Exists(ctx context.Context, followerID, followedID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ? AND followed_id = ?", followerID, followedID).
		Count(&count).Error
	return count > 0, err
}

func (r *followRepository) ListFollowers(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error) {
	limit = model.ClampLimit(limit)
	tx := applyDescCursor(r.db.WithContext(ctx).Where("followed_id = ?", userID), cursor, "created_at", "id")

	var follows []model.Follow
	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&follows).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(follows, limit)
	return page, next, nil
}

func (r *followRepository) ListFollowing(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error) {
	limit = model.ClampLimit(limit)
	tx := applyDescCursor(r.db.WithContext(ctx).Where("follower_id = ?", userID), cursor, "created_at", "id")

	var follows []model.Follow
	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&follows).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(follows, limit)
	return page, next, nil
}

func (r *followRepository) ListFollowedIDs(ctx context.Context, followerID string) ([]string, error) {
	var ids []string
	err := r.db.WithContext(ctx).Model(&model.Follow{}).
		Where("follower_id = ?", followerID).
		Pluck("followed_id", &ids).Error
	return ids, err
}
