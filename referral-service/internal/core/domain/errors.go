package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrLinkNotActive = errors.New("link not active")
	ErrSelfReferral  = errors.New("self-referral not allowed")
)
