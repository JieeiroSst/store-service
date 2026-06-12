package domain

import "errors"

var (
	ErrNotFound          = errors.New("not found")
	ErrLinkNotActive     = errors.New("link not active")
	ErrSelfReferral      = errors.New("self-referral not allowed")
	ErrAlreadyReferred   = errors.New("user already referred by this owner")
	ErrDuplicateRefCode  = errors.New("ref_code already exists")
	ErrRateLimitExceeded = errors.New("daily link generation limit exceeded")
)
