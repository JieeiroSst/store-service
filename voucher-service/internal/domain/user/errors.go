package user

import "errors"

var (
	ErrInvalidUser        = errors.New("invalid user")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserInactive       = errors.New("user is inactive")
)
