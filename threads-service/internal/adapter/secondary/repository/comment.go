package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"gorm.io/gorm"
)

type commentRepository struct {
	db *gorm.DB
}

func NewCommentRepository(db *gorm.DB) port.CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) error {
	return r.db.WithContext(ctx).Create(comment).Error
}

func (r *commentRepository) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.WithContext(ctx).First(&comment, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) ListByPost(ctx context.Context, postID string, cursor string, limit int) ([]model.Comment, string, error) {
	limit = model.ClampLimit(limit)
	tx := applyAscCursor(r.db.WithContext(ctx).Where("post_id = ?", postID), cursor, "created_at", "id")

	var comments []model.Comment
	if err := tx.Order("created_at ASC, id ASC").Limit(limit + 1).Find(&comments).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(comments, limit)
	return page, next, nil
}

func (r *commentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Comment{}, "id = ?", id).Error
}
