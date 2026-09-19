package repository

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"gorm.io/gorm"
)

type likeRepository struct {
	db *gorm.DB
}

func NewLikeRepository(db *gorm.DB) port.LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Create(ctx context.Context, like *model.Like) error {
	return r.db.WithContext(ctx).Create(like).Error
}

func (r *likeRepository) Delete(ctx context.Context, userID string, postID, commentID *string) (bool, error) {
	tx := whereTarget(r.db.WithContext(ctx).Where("user_id = ?", userID), postID, commentID)
	result := tx.Delete(&model.Like{})
	return result.RowsAffected > 0, result.Error
}

func (r *likeRepository) Exists(ctx context.Context, userID string, postID, commentID *string) (bool, error) {
	tx := whereTarget(r.db.WithContext(ctx).Model(&model.Like{}).Where("user_id = ?", userID), postID, commentID)
	var count int64
	if err := tx.Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func whereTarget(tx *gorm.DB, postID, commentID *string) *gorm.DB {
	if postID != nil {
		return tx.Where("post_id = ?", *postID)
	}
	return tx.Where("comment_id = ?", *commentID)
}
