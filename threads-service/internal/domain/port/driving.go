package port

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
)

type PostUsecase interface {
	CreatePost(ctx context.Context, input model.CreatePostInput) (*model.Post, error)
	GetPost(ctx context.Context, id string) (*model.Post, error)
	// ListFeed returns every post when authorID is empty, or just that
	// author's posts otherwise (a profile's post grid). Cursor-paginated -
	// see port.PostRepository.ListFeed.
	ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error)
	// ListHomeFeed returns posts from users callerUserID follows - the
	// "following" timeline, as opposed to ListFeed's global/profile views.
	ListHomeFeed(ctx context.Context, callerUserID string, cursor string, limit int) ([]model.Post, string, error)
	DeletePost(ctx context.Context, id, callerUserID string) error
	// Repost creates a reshare of postID as a new post owned by userID.
	Repost(ctx context.Context, userID, postID string) (*model.Post, error)
	// Unrepost removes userID's repost of postID, if any (a no-op is an
	// error here, unlike UnlikePost, since there's no separate "toggle"
	// affordance in the UI - the client already knows whether it reposted).
	Unrepost(ctx context.Context, userID, postID string) error
}

type CommentUsecase interface {
	CreateComment(ctx context.Context, input model.CreateCommentInput) (*model.Comment, error)
	ListComments(ctx context.Context, postID string, cursor string, limit int) ([]model.Comment, string, error)
	DeleteComment(ctx context.Context, id, callerUserID string) error
}

type LikeUsecase interface {
	LikePost(ctx context.Context, userID, postID string) error
	UnlikePost(ctx context.Context, userID, postID string) error
	LikeComment(ctx context.Context, userID, commentID string) error
	UnlikeComment(ctx context.Context, userID, commentID string) error
}

type FollowUsecase interface {
	Follow(ctx context.Context, followerID, followedID string) error
	Unfollow(ctx context.Context, followerID, followedID string) error
	ListFollowers(ctx context.Context, userID string, cursor string, limit int) ([]model.Author, string, error)
	ListFollowing(ctx context.Context, userID string, cursor string, limit int) ([]model.Author, string, error)
}

type BookmarkUsecase interface {
	Bookmark(ctx context.Context, userID, postID string) error
	Unbookmark(ctx context.Context, userID, postID string) error
	ListBookmarks(ctx context.Context, userID string, cursor string, limit int) ([]model.Post, string, error)
}

type TagUsecase interface {
	ListTrending(ctx context.Context, limit int) ([]model.TagCount, error)
}
