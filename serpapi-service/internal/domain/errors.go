package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrUnauthorized = errors.New("unauthorized")
	ErrExpired      = errors.New("search expired and was deleted from the archive")

	ErrRateLimited = errors.New("serpapi rate limit or plan quota exceeded")
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

type UpstreamError struct {
	Status int
	Msg    string
}

func (e *UpstreamError) Error() string {
	if e.Status == 0 {
		return "serpapi: " + e.Msg
	}
	return fmt.Sprintf("serpapi: %d %s", e.Status, e.Msg)
}

func IsUpstream(err error) bool {
	var ue *UpstreamError
	return errors.As(err, &ue)
}
