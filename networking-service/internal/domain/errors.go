package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound         = errors.New("not found")
	ErrPermissionDenied = errors.New("Permission denied")
	ErrInvalidSession   = errors.New("invalid session")
	ErrLockDelay        = errors.New("lock delay in effect")
)

type InvalidError struct{ Msg string }

func (e *InvalidError) Error() string { return e.Msg }

func Invalid(format string, args ...any) error {
	return &InvalidError{Msg: fmt.Sprintf(format, args...)}
}

func IsInvalid(err error) bool {
	var ie *InvalidError
	return errors.As(err, &ie)
}
