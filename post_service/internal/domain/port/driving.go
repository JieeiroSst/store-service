package port

import (
	"context"

	"github.com/JIeeiroSst/post-service/model"
)

type PostUsecase interface {
	// CreatePost uploads file first (when non-nil) and links the
	// resulting media to the post it creates - a nil file is a valid,
	// text-only post.
	CreatePost(ctx context.Context, input model.CreatePostInput, file *UploadFileInput) (*model.Post, error)
	GetPost(ctx context.Context, id string) (*model.Post, error)
	ListPosts(ctx context.Context, cursor string, limit int) ([]model.Post, string, error)
	// UpdatePost is author-only, enforced here against callerAuthorID
	// (from the verified JWT) - the original handler never checked this.
	UpdatePost(ctx context.Context, id, callerAuthorID string, input model.UpdatePostInput) error
}

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, input model.CreateCategoryInput) (*model.Category, error)
	UpdateCategory(ctx context.Context, id string, input model.UpdateCategoryInput) error
	DeleteCategory(ctx context.Context, id string) error
	GetCategory(ctx context.Context, id string) (*model.Category, error)
	ListCategories(ctx context.Context, cursor string, limit int) ([]model.Category, string, error)
}
