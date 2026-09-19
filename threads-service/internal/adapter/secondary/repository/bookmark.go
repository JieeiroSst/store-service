package repository

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"gorm.io/gorm"
)

type bookmarkRepository struct {
	db *gorm.DB
}

func NewBookmarkRepository(db *gorm.DB) port.BookmarkRepository {
	return &bookmarkRepository{db: db}
}

func (r *bookmarkRepository) Create(ctx context.Context, bookmark *model.Bookmark) error {
	return r.db.WithContext(ctx).Create(bookmark).Error
}

func (r *bookmarkRepository) Delete(ctx context.Context, userID, postID string) (bool, error) {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Delete(&model.Bookmark{})
	return result.RowsAffected > 0, result.Error
}

func (r *bookmarkRepository) Exists(ctx context.Context, userID, postID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Bookmark{}).
		Where("user_id = ? AND post_id = ?", userID, postID).
		Count(&count).Error
	return count > 0, err
}

func (r *bookmarkRepository) ListPostsByUser(ctx context.Context, userID string, cursor string, limit int) ([]model.Post, string, error) {
	limit = model.ClampLimit(limit)

	// posts.* is required, not cosmetic: without it, SELECT * across the
	// join would pull ambiguous columns (both tables have id/created_at),
	// and GORM would silently scan the wrong table's values into Post.
	tx := r.db.WithContext(ctx).
		Select("posts.*").
		Joins("JOIN bookmarks ON bookmarks.post_id = posts.id").
		Where("bookmarks.user_id = ?", userID)
	// Paginated by the post's own created_at (see BookmarkRepository's
	// doc comment) so this can reuse the same posts.id keyset cursor as
	// every other post list, instead of a bespoke bookmarked-at cursor.
	tx = applyDescCursor(tx, cursor, "posts.created_at", "posts.id")

	var posts []model.Post
	if err := tx.Order("posts.created_at DESC, posts.id DESC").Limit(limit + 1).Find(&posts).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(posts, limit)
	return page, next, nil
}
