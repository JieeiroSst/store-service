package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

type postRepository struct {
	db *gorm.DB
}

func NewPostRepository(db *gorm.DB) port.PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post, categoryID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Both writes go through tx, not r.db - the original code took a
		// tx from Transaction() and then never used it, so the post and
		// its category link weren't actually atomic with each other.
		if err := tx.Create(post).Error; err != nil {
			return err
		}
		if categoryID == "" {
			return nil
		}
		link := model.PostCategory{NewID: post.ID, CategoryID: categoryID}
		return tx.Create(&link).Error
	})
}

func (r *postRepository) GetByID(ctx context.Context, id string) (*model.Post, error) {
	var post model.Post
	err := r.db.WithContext(ctx).Where("id = ?", id).
		Preload("Categories").Preload("Media").
		First(&post).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, cursor string, limit int) ([]model.Post, string, error) {
	limit = model.ClampLimit(limit)
	tx := applyDescCursor(r.db.WithContext(ctx), cursor, "created_at", "id")

	var posts []model.Post
	// Over-fetch by one row so PageFromRows can tell whether there's a
	// next page without a separate COUNT(*) query.
	if err := tx.Preload("Categories").Preload("Media").
		Order("created_at DESC, id DESC").Limit(limit + 1).Find(&posts).Error; err != nil {
		return nil, "", err
	}
	page, next := model.PageFromRows(posts, limit)
	return page, next, nil
}

// Update doesn't turn "0 rows affected" into ErrNotFound: MySQL reports 0
// rows affected for an UPDATE that matched a row but changed no column
// values (a no-op edit - saving a post with identical content), which
// isn't a not-found case. Callers that need existence checked do it
// themselves first (see application.postService.UpdatePost, which already
// calls GetByID to verify ownership before this runs).
func (r *postRepository) Update(ctx context.Context, id string, post *model.Post) error {
	return r.db.WithContext(ctx).Model(&model.Post{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":        post.Name,
			"content":     post.Content,
			"description": post.Description,
		}).Error
}
