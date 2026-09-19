package port

import (
	"context"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) error
	GetByID(ctx context.Context, id string) (*model.Post, error)
	// ListFeed and every other List* method below is cursor-paginated
	// (see model.Cursor's doc comment for why): cursor is an opaque token
	// from a previous page's response (empty for the first page), limit
	// is clamped via model.ClampLimit, and the string return is the next
	// page's cursor (empty when there isn't one).
	ListFeed(ctx context.Context, authorID string, cursor string, limit int) ([]model.Post, string, error)
	// ListByAuthors is the home/following-feed query: posts by any of the
	// given author IDs, newest first.
	ListByAuthors(ctx context.Context, authorIDs []string, cursor string, limit int) ([]model.Post, string, error)
	Delete(ctx context.Context, id string) error
	IncrementLikeCount(ctx context.Context, id string, delta int) error
	IncrementCommentCount(ctx context.Context, id string, delta int) error
	IncrementRepostCount(ctx context.Context, id string, delta int) error
	// FindRepostBy returns userID's repost row of originalPostID, or
	// ErrNotFound if they haven't reposted it.
	FindRepostBy(ctx context.Context, userID, originalPostID string) (*model.Post, error)
}

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) error
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	// ListByPost is oldest-first (see model.Comment.CursorKey's doc
	// comment), unlike every other cursor-paginated list here.
	ListByPost(ctx context.Context, postID string, cursor string, limit int) ([]model.Comment, string, error)
	Delete(ctx context.Context, id string) error
}

type LikeRepository interface {
	Create(ctx context.Context, like *model.Like) error
	// Delete reports whether a row was actually removed, so callers don't
	// decrement a denormalized counter for a like that never existed.
	Delete(ctx context.Context, userID string, postID, commentID *string) (bool, error)
	Exists(ctx context.Context, userID string, postID, commentID *string) (bool, error)
}

type FollowRepository interface {
	Create(ctx context.Context, follow *model.Follow) error
	Delete(ctx context.Context, followerID, followedID string) error
	Exists(ctx context.Context, followerID, followedID string) (bool, error)
	ListFollowers(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error)
	ListFollowing(ctx context.Context, userID string, cursor string, limit int) ([]model.Follow, string, error)
	// ListFollowedIDs returns every user ID followerID follows, unpaginated
	// - used internally to build the home feed's author filter, not
	// exposed as an endpoint of its own, so it isn't cursor-paginated like
	// the client-facing lists above.
	ListFollowedIDs(ctx context.Context, followerID string) ([]string, error)
}

type TagRepository interface {
	// GetOrCreateByNames resolves each name to its (possibly newly
	// created) tag row, deduplicated by the tags.name UNIQUE constraint.
	GetOrCreateByNames(ctx context.Context, names []string) ([]model.Tag, error)
	AttachToPost(ctx context.Context, postID string, tagIDs []string) error
	ListNamesByPost(ctx context.Context, postID string) ([]string, error)
}

type BookmarkRepository interface {
	Create(ctx context.Context, bookmark *model.Bookmark) error
	Delete(ctx context.Context, userID, postID string) (bool, error)
	Exists(ctx context.Context, userID, postID string) (bool, error)
	// ListPostsByUser returns the bookmarked posts themselves (not the
	// join rows), ordered newest-post-first (see model.Post.CursorKey) -
	// not by when each was bookmarked, so it can share the same cursor
	// scheme as every other post list instead of a bespoke one.
	ListPostsByUser(ctx context.Context, userID string, cursor string, limit int) ([]model.Post, string, error)
}

// TrendingTagsStore tracks tag usage in Redis sorted sets - a purely
// Redis-native feature, not a Postgres cache. Trending rank is never
// computed from Postgres at read time, only accumulated incrementally as
// posts are created; falling back to an aggregate COUNT/GROUP BY query on
// every hit of a public, high-traffic endpoint would be the expensive
// path this whole feature exists to avoid. It's a bounded top-N read
// (see model.ClampLimit), not cursor-paginated - "browse all tags ever"
// isn't a thing this endpoint offers.
type TrendingTagsStore interface {
	IncrementTags(ctx context.Context, names []string) error
	TopTags(ctx context.Context, limit int) ([]model.TagCount, error)
}

// UserClient fetches author/profile info from user_service over HTTP.
// threads-service owns the social graph (posts, comments, likes, follows,
// tags) but never user identity/auth - see database.sql's header comment.
type UserClient interface {
	GetUser(ctx context.Context, userID string) (*model.Author, error)
	// GetUsers batch-resolves several IDs for feed/list enrichment. A
	// missing entry in the result map means that user's info couldn't be
	// fetched, not that the user doesn't exist - callers degrade to
	// omitting the author rather than failing outright.
	GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error)
}
