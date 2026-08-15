package corporate

import "errors"

var (
	ErrBudgetExceeded    = errors.New("corporate budget exceeded")
	ErrInvalidCorporate  = errors.New("invalid corporate")
	ErrVersionConflict   = errors.New("corporate version conflict")
	ErrCorporateNotFound = errors.New("corporate not found")
	ErrCorporateInactive = errors.New("corporate is inactive")
)
