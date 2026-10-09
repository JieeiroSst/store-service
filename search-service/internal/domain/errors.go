package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrUnavailable  = errors.New("search backend unavailable")
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
