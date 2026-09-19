package port

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrForbidden         = errors.New("caller does not own this resource")
	ErrAlreadyLiked      = errors.New("already liked")
	ErrAlreadyFollowing  = errors.New("already following this user")
	ErrCannotFollowSelf  = errors.New("cannot follow yourself")
	ErrAlreadyReposted   = errors.New("already reposted")
	ErrAlreadyBookmarked = errors.New("already bookmarked")
)
