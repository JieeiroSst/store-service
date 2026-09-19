package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/post-service/internal/domain/port"
	"github.com/JIeeiroSst/post-service/model"
)

type postService struct {
	posts   port.PostRepository
	media   port.MediaRepository
	storage port.ObjectStorage
	ids     port.IDGenerator
}

func NewPostService(posts port.PostRepository, media port.MediaRepository, storage port.ObjectStorage, ids port.IDGenerator) port.PostUsecase {
	return &postService{posts: posts, media: media, storage: storage, ids: ids}
}

func (s *postService) CreatePost(ctx context.Context, input model.CreatePostInput, file *port.UploadFileInput) (*model.Post, error) {
	post := &model.Post{
		ID:          s.ids.NewID(),
		AuthorID:    input.AuthorID,
		Name:        input.Name,
		Content:     input.Content,
		Description: input.Description,
	}

	if file != nil {
		uploaded, err := s.storage.UploadFile(ctx, *file)
		if err != nil {
			return nil, fmt.Errorf("upload file: %w", err)
		}

		media := &model.Media{
			ID:          s.ids.NewID(),
			URL:         uploaded.URL,
			Description: file.FileName,
		}
		if err := s.media.Create(ctx, media); err != nil {
			return nil, fmt.Errorf("save media: %w", err)
		}
		// The post links to the media row this call just created, not
		// whatever media_id the caller happened to send - that was the
		// original bug (a freshly uploaded file's post never actually
		// pointed at it).
		post.MediaID = &media.ID
	}

	if err := s.posts.Create(ctx, post, input.CategoryID); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}
	return post, nil
}

func (s *postService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	return s.posts.GetByID(ctx, id)
}

func (s *postService) ListPosts(ctx context.Context, cursor string, limit int) ([]model.Post, string, error) {
	return s.posts.List(ctx, cursor, limit)
}

func (s *postService) UpdatePost(ctx context.Context, id, callerAuthorID string, input model.UpdatePostInput) error {
	existing, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if existing.AuthorID != callerAuthorID {
		return port.ErrForbidden
	}

	existing.Name = input.Name
	existing.Content = input.Content
	existing.Description = input.Description
	return s.posts.Update(ctx, id, existing)
}
