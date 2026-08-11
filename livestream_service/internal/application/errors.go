package application

import "github.com/JIeeiroSst/livestream-service/internal/domain/port"

// Aliased from the port package (see driven.go/driving.go's sibling
// errors.go) so every file in this package can keep referencing the bare
// names, while primary adapters reach the same values via port.ErrXxx
// without depending on this package's internals.
var (
	ErrStreamKeyNotFound = port.ErrStreamKeyNotFound
	ErrNodeAtCapacity    = port.ErrNodeAtCapacity
	ErrNoNodeAvailable   = port.ErrNoNodeAvailable
	ErrForbidden         = port.ErrForbidden
	ErrBanned            = port.ErrBanned
)
