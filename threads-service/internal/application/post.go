package application

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
	"github.com/google/uuid"
)

type postService struct {
	posts    port.PostRepository
	tags     port.TagRepository
	follows  port.FollowRepository
	trending port.TrendingTagsStore
	users    port.UserClient
}

func NewPostService(posts port.PostRepository, tags port.TagRepository, follows port.FollowRepository, trending port.TrendingTagsStore, users port.UserClient) port.PostUsecase {
	return &postService{posts: posts, tags: tags, follows: follows, trending: trending, users: users}
}

func (s *postService) CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error) {
	post := &model.Post{
		ID:        uuid.NewString(),
		UserID:    input.UserID,
		Content:   input.Content,
		MediaURLs: input.MediaURLs,
	}
	if err := s.posts.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	if len(input.Tags) > 0 {
		tags, err := s.tags.GetOrCreateByNames(ctx, input.Tags)
		if err != nil {
			return nil, fmt.Errorf("resolve tags: %w", err)
		}
		ids := make([]string, len(tags))
		names := make([]string, len(tags))
		for i, tag := range tags {
			ids[i] = tag.ID
			names[i] = tag.Name
		}
		if err := s.tags.AttachToPost(ctx, post.ID, ids); err != nil {
			return nil, fmt.Errorf("attach tags: %w", err)
		}
		post.Tags = names

		// Best-effort: trending rank is a nice-to-have signal, not core
		// data, so a Redis hiccup here must never fail post creation.
		_ = s.trending.IncrementTags(ctx, names)
	}

	post.Author = s.authorOrNil(ctx, post.UserID)
	return post, nil
}

func (s *postService) GetPost(ctx context.Context, id string) (*model.Post, error) {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	post.Tags, _ = s.tags.ListNamesByPost(ctx, id)
	post.Author = s.authorOrNil(ctx, post.UserID)
	s.enrichOriginal(ctx, post)
	return post, nil
}

func (s *postService) ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error) {
	posts, next, err := s.posts.ListFeed(ctx, authorID, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	resolveAuthors(ctx, s.users, posts)
	s.enrichOriginals(ctx, posts)
	return posts, next, nil
}

func (s *postService) ListHomeFeed(ctx context.Context, callerUserID string, cursor string, limit int) ([]model.Post, string, error) {
	followedIDs, err := s.follows.ListFollowedIDs(ctx, callerUserID)
	if err != nil {
		return nil, "", fmt.Errorf("list followed users: %w", err)
	}
	if len(followedIDs) == 0 {
		return []model.Post{}, "", nil
	}

	posts, next, err := s.posts.ListByAuthors(ctx, followedIDs, cursor, limit)
	if err != nil {
		return nil, "", err
	}
	resolveAuthors(ctx, s.users, posts)
	s.enrichOriginals(ctx, posts)
	return posts, next, nil
}

func (s *postService) DeletePost(ctx context.Context, id, callerUserID string) error {
	post, err := s.posts.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if post.UserID != callerUserID {
		return port.ErrForbidden
	}
	return s.posts.Delete(ctx, id)
}

func (s *postService) Repost(ctx context.Context, userID, postID string) (*model.Post, error) {
	original, err := s.posts.GetByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("get original post: %w", err)
	}

	if _, err := s.posts.FindRepostBy(ctx, userID, postID); err == nil {
		return nil, port.ErrAlreadyReposted
	}

	repost := &model.Post{ID: uuid.NewString(), UserID: userID, Content: "", RepostOfID: &postID}
	if err := s.posts.Create(ctx, repost); err != nil {
		return nil, fmt.Errorf("create repost: %w", err)
	}
	if err := s.posts.IncrementRepostCount(ctx, postID, 1); err != nil {
		return nil, fmt.Errorf("increment repost count: %w", err)
	}

	repost.Author = s.authorOrNil(ctx, userID)
	original.Author = s.authorOrNil(ctx, original.UserID)
	repost.OriginalPost = original
	return repost, nil
}

func (s *postService) Unrepost(ctx context.Context, userID, postID string) error {
	repost, err := s.posts.FindRepostBy(ctx, userID, postID)
	if err != nil {
		return err
	}
	if err := s.posts.Delete(ctx, repost.ID); err != nil {
		return err
	}
	return s.posts.IncrementRepostCount(ctx, postID, -1)
}

// authorOrNil degrades gracefully: a post is still returned even if
// user_service is unreachable, just without an author field, rather than
// failing the whole request over a dependency that isn't this service's
// own data.
func (s *postService) authorOrNil(ctx context.Context, userID string) *model.Author {
	author, err := s.users.GetUser(ctx, userID)
	if err != nil {
		return nil
	}
	return author
}

// enrichOriginal resolves the original post a repost points at (best
// effort - a dangling or failed lookup just leaves OriginalPost nil rather
// than failing the request). GetByID is Redis-cached (see
// internal/adapter/secondary/cache), so this rarely costs an extra
// Postgres round trip in practice.
func (s *postService) enrichOriginal(ctx context.Context, post *model.Post) {
	if post.RepostOfID == nil {
		return
	}
	original, err := s.posts.GetByID(ctx, *post.RepostOfID)
	if err != nil {
		return
	}
	original.Author = s.authorOrNil(ctx, original.UserID)
	post.OriginalPost = original
}

func (s *postService) enrichOriginals(ctx context.Context, posts []model.Post) {
	for i := range posts {
		s.enrichOriginal(ctx, &posts[i])
	}
}

// resolveAuthors batch-fetches and attaches an Author to every post in
// posts, deduplicating repeat authors into a single UserClient.GetUsers
// call. Shared by postService and bookmarkService so both list-style
// usecases enrich the same way.
func resolveAuthors(ctx context.Context, users port.UserClient, posts []model.Post) {
	ids := make([]string, 0, len(posts))
	seen := make(map[string]bool, len(posts))
	for _, p := range posts {
		if !seen[p.UserID] {
			seen[p.UserID] = true
			ids = append(ids, p.UserID)
		}
	}
	if len(ids) == 0 {
		return
	}

	authors, err := users.GetUsers(ctx, ids)
	if err != nil {
		return
	}
	for i := range posts {
		posts[i].Author = authors[posts[i].UserID]
	}
}
