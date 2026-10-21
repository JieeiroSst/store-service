package model

import "errors"

type ErrorKind int

const (
	KindInternal ErrorKind = iota
	KindBadRequest
	KindRateLimited
	KindUpstreamUnavailable
)

const internalMessage = "an internal error occurred"

type Error struct {
	Kind    ErrorKind
	Message string
	Err     error
}

func (e *Error) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

func (e *Error) Unwrap() error { return e.Err }

func NewError(kind ErrorKind, message string, err error) *Error {
	if err == nil {
		err = errors.New(message)
	}
	return &Error{Kind: kind, Message: message, Err: err}
}

func BadRequest(message string) error {
	return NewError(KindBadRequest, message, nil)
}

func AsError(err error) *Error {
	var e *Error
	if errors.As(err, &e) {
		return e
	}
	return &Error{Kind: KindInternal, Message: internalMessage, Err: err}
}
