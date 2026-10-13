package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ThreadsCreatePost(ctx context.Context, content string, mediaUrls []string, tags []string) (*model.ThreadsPost, error) {
	return restCall[*model.ThreadsPost](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: "/api/v1/posts", Body: map[string]any{"content": content, "media_urls": mediaUrls, "tags": tags}}, "")
}

func (r *mutationResolver) ThreadsDeletePost(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/posts/%v", id)}, "")
}

func (r *mutationResolver) ThreadsLikePost(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/posts/%v/likes", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ThreadsLikeComment(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/comments/%v/likes", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ThreadsUnlikePost(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/posts/%v/likes", id)}, "")
}

func (r *mutationResolver) ThreadsUnlikeComment(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/comments/%v/likes", id)}, "")
}

func (r *mutationResolver) ThreadsCreateComment(ctx context.Context, id string, content string, mediaUrls []string, parentCommentID *string) (*model.ThreadsComment, error) {
	return restCall[*model.ThreadsComment](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/posts/%v/comments", id), Body: map[string]any{"content": content, "media_urls": mediaUrls, "parent_comment_id": parentCommentID}}, "")
}

func (r *mutationResolver) ThreadsRepost(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/posts/%v/repost", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ThreadsUnrepost(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/posts/%v/repost", id)}, "")
}

func (r *mutationResolver) ThreadsBookmark(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/posts/%v/bookmark", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ThreadsUnbookmark(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/posts/%v/bookmark", id)}, "")
}

func (r *mutationResolver) ThreadsDeleteComment(ctx context.Context, id string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/comments/%v", id)}, "")
}

func (r *mutationResolver) ThreadsFollowUser(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/users/%v/follow", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ThreadsUnfollowUser(ctx context.Context, id string) (any, error) {
	return restCall[any](ctx, r.Proxy, "threads-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/users/%v/follow", id)}, "")
}

func (r *queryResolver) ThreadsPosts(ctx context.Context, authorID *string, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return restCall[*model.ThreadsPostPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: "/api/v1/posts", Query: restQuery("author_id", authorID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsPost(ctx context.Context, id string) (*model.ThreadsPost, error) {
	return restCall[*model.ThreadsPost](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/posts/%v", id)}, "")
}

func (r *queryResolver) ThreadsComments(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsCommentPage, error) {
	return restCall[*model.ThreadsCommentPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/posts/%v/comments", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsFollowers(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	return restCall[*model.ThreadsAuthorPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/followers", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsFollowing(ctx context.Context, id string, cursor *string, limit *int) (*model.ThreadsAuthorPage, error) {
	return restCall[*model.ThreadsAuthorPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/following", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsHomeFeed(ctx context.Context, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return restCall[*model.ThreadsPostPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: "/api/v1/feed", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsBookmarks(ctx context.Context, cursor *string, limit *int) (*model.ThreadsPostPage, error) {
	return restCall[*model.ThreadsPostPage](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: "/api/v1/bookmarks", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ThreadsTrendingTags(ctx context.Context, limit *int) ([]*model.ThreadsTagCount, error) {
	return restCall[[]*model.ThreadsTagCount](ctx, r.Proxy, "threads-service", rest.Request{Method: "GET", Path: "/api/v1/tags/trending", Query: restQuery("limit", limit)}, "tags")
}
