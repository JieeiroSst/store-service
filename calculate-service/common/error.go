package common

import "errors"

var (
	ErrInvalidCoordinate   = errors.New("invalid coordinate")
	ErrLocationNotTracked  = errors.New("location not tracked")
	ErrUpstreamUnavailable = errors.New("upstream weather provider unavailable")
	ErrMarketUnavailable   = errors.New("upstream market provider unavailable")
)
