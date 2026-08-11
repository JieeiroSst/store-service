package port

import "errors"

// Sentinel errors usecases can return, kept in this package (rather than
// the application package that implements them) so primary adapters can
// map them to the right HTTP status without depending on application's
// internals - only on the port contract.
var (
	ErrStreamKeyNotFound = errors.New("stream key not found")
	ErrNodeAtCapacity    = errors.New("node at capacity")
	ErrNoNodeAvailable   = errors.New("no transcode node available")
	ErrForbidden         = errors.New("caller does not own this room")
	ErrBanned            = errors.New("caller is banned from chat in this room")
)
