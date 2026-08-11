package application

import "errors"

var (
	ErrStreamKeyNotFound = errors.New("stream key not found")
	ErrNodeAtCapacity    = errors.New("node at capacity")
	ErrNoNodeAvailable   = errors.New("no transcode node available")
)
