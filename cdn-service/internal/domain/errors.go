package domain

import "errors"

var (
	ErrFileNotFound     = errors.New("file not found")
	ErrInvalidStatus    = errors.New("file is not in the expected status")
	ErrFileTooLarge     = errors.New("file exceeds the maximum allowed size")
	ErrInvalidFileInput = errors.New("invalid file input")
)
