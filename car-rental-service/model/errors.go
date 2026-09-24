package model

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrInvalidArgument  = errors.New("invalid argument")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrPermissionDenied = errors.New("permission denied")
	ErrFailedPrecond    = errors.New("failed precondition")
)
