package port

import "errors"

// Sentinel errors usecases can return, kept in this package (rather than
// the application package that implements them) so primary adapters can
// map them to the right HTTP status without depending on application's
// internals - only on the port contract.
var (
	ErrNotFound  = errors.New("resource not found")
	ErrForbidden = errors.New("caller does not own this resource")
)
