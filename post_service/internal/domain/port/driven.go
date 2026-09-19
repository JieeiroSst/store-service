package port

import (
	"context"
	"io"

	"github.com/JIeeiroSst/post-service/model"
)

type PostRepository interface {
	// Create persists post and, when categoryID is non-empty, links it to
	// that category - both in one transaction, so a post is never left
	// half-created if the category link fails.
	Create(ctx context.Context, post *model.Post, categoryID string) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	// List is cursor-paginated - see model.Cursor's doc comment for why
	// (not offset/page based).
	List(ctx context.Context, cursor string, limit int) ([]model.Post, string, error)
	Update(ctx context.Context, id string, post *model.Post) error
}

type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	Update(ctx context.Context, id string, category *model.Category) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*model.Category, error)
	List(ctx context.Context, cursor string, limit int) ([]model.Category, string, error)
}

type MediaRepository interface {
	Create(ctx context.Context, media *model.Media) error
}

// UploadFileInput is a framework-agnostic view of an uploaded file - the
// HTTP adapter extracts this from a multipart.FileHeader so the
// application/domain layers never import net/http or mime/multipart.
type UploadFileInput struct {
	FileName    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

type UploadFileResult struct {
	URL string
}

// ObjectStorage is the S3/MinIO-compatible sink for uploaded post media.
type ObjectStorage interface {
	UploadFile(ctx context.Context, input UploadFileInput) (*UploadFileResult, error)
	RemoveObject(ctx context.Context, fileName string) error
}

// IDGenerator issues the snowflake IDs this service's rows use (see
// internal/adapter/secondary/idgen) - unchanged from before this refactor,
// so existing IDs/links stay valid.
type IDGenerator interface {
	NewID() string
}
