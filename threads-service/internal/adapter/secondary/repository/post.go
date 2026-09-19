package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) port.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) error {
	return r.db.WithContext(ctx).Create(post).Error
}

func (r *postRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	if err := r.db.WithContext(ctx).First(&post, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error) {
	limit = model.ClampLimit(limit)
	tx := r.db.WithContext(ctx)
	if authorID != "" {
		tx = tx.Where("user_id = ?", authorID)
	}
	tx = applyDescCursor(tx, cursor, "created_at", "id")

	var posts []model.Post
	// Over-fetch by one row so PageFromRows can tell whether there's a
	// next page without a separate COUNT(*) query.
	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&posts).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(posts, limit)
	return page, next, nil
}

func (r *postRepository) ListByAuthors(ctx context.Context, authorIDs []string, cursor string, limit int) ([]model.Post, string, error) {
	if len(authorIDs) == 0 {
		return []model.Post{}, "", nil
	}
	limit = model.ClampLimit(limit)
	tx := applyDescCursor(r.db.WithContext(ctx).Where("user_id IN ?", authorIDs), cursor, "created_at", "id")

	var posts []model.Post
	if err := tx.Order("created_at DESC, id DESC").Limit(limit + 1).Find(&posts).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(posts, limit)
	return page, next, nil
}

func (r *postRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&model.Post{}, "id = ?", id).Error
}

func (r *postRepository) IncrementLikeCount(ctx context.Context, id string, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *postRepository) IncrementCommentCount(ctx context.Context, id string, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		UpdateColumn("comment_count", gorm.Expr("comment_count + ?", delta)).Error
}

func (r *postRepository) IncrementRepostCount(ctx context.Context, id string, delta int) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		UpdateColumn("repost_count", gorm.Expr("repost_count + ?", delta)).Error
}

func (r *postRepository) FindRepostBy(ctx context.Context, userID, originalPostID string) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND repost_of_id = ?", userID, originalPostID).
		First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}
