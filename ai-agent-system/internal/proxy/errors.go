package proxy

import (
	"fmt"
	"net/http"
)

type ErrorKind int

const (
	ErrKindUnknown ErrorKind = iota
	ErrKindRateLimited
	ErrKindUpstreamUnavailable
	ErrKindInvalidRequest
	ErrKindAuth
	ErrKindTimeout
)

type Error struct {
	Kind       ErrorKind
	StatusCode int
	Message    string
	Err        error
}

func (e *Error) Error() string {
	return fmt.Sprintf("proxy: %s (kind=%d, status=%d): %v", e.Message, e.Kind, e.StatusCode, e.Err)
}

func (e *Error) Unwrap() error { return e.Err }

func mapOllamaError(err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Kind:    ErrKindUpstreamUnavailable,
		Message: "failed to reach the local Ollama server",
		Err:     err,
	}
}

func mapOllamaHTTPError(statusCode int, body string) *Error {
	kind := ErrKindUnknown
	switch {
	case statusCode == http.StatusTooManyRequests:
		kind = ErrKindRateLimited
	case statusCode == http.StatusBadRequest || statusCode == http.StatusUnprocessableEntity:
		kind = ErrKindInvalidRequest
	case statusCode == http.StatusNotFound:
		// Most common cause in practice: OLLAMA_MODEL hasn't been pulled
		// into the Ollama server yet.
		kind = ErrKindInvalidRequest
	case statusCode >= 500:
		kind = ErrKindUpstreamUnavailable
	}
	return &Error{
		Kind:       kind,
		StatusCode: statusCode,
		Message:    "the local Ollama server rejected or failed to process the request",
		Err:        fmt.Errorf("ollama: status %d: %s", statusCode, body),
	}
}
