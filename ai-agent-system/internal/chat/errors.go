package chat

import (
	"errors"

	"aiagentsystem/internal/proxy"
)

type ErrKind int

const (
	ErrKindInternal ErrKind = iota
	ErrKindRateLimited
	ErrKindUpstreamUnavailable
	ErrKindBadRequest
)

type Error struct {
	Kind    ErrKind
	Message string
	Err     error
}

func (e *Error) Error() string { return e.Err.Error() }
func (e *Error) Unwrap() error { return e.Err }

func wrapProxyError(err error) error {
	if err == nil {
		return nil
	}

	var pErr *proxy.Error
	if errors.As(err, &pErr) {
		kind := ErrKindInternal
		switch pErr.Kind {
		case proxy.ErrKindRateLimited:
			kind = ErrKindRateLimited
		case proxy.ErrKindUpstreamUnavailable, proxy.ErrKindTimeout:
			kind = ErrKindUpstreamUnavailable
		case proxy.ErrKindInvalidRequest:
			kind = ErrKindBadRequest
		case proxy.ErrKindAuth:
			kind = ErrKindUpstreamUnavailable // a misconfigured server-side key is not the client's fault
		}
		return &Error{Kind: kind, Message: pErr.Message, Err: err}
	}

	return &Error{Kind: ErrKindInternal, Message: "an internal error occurred", Err: err}
}

func NewBadRequestError(msg string) error {
	return &Error{Kind: ErrKindBadRequest, Message: msg, Err: errors.New(msg)}
}
