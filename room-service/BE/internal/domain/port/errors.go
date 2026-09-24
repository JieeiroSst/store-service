package port

import "errors"

var (
	ErrUnauthenticated = errors.New("unauthenticated")
	ErrInvalidLogin    = errors.New("invalid username or password")
	ErrNotFound        = errors.New("not found")
	ErrForbidden       = errors.New("forbidden")
	ErrInvalidInput    = errors.New("invalid input")
	ErrAlreadyMember   = errors.New("already a member")
	ErrUnavailable     = errors.New("upstream unavailable")
)
