package shared

import "errors"

type DomainError interface {
	error
	Code() string
}

type domainError struct {
	code string
	msg  string
}

func (e *domainError) Error() string { return e.msg }
func (e *domainError) Code() string  { return e.code }

func NewDomainError(code, msg string) DomainError {
	return &domainError{code: code, msg: msg}
}

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrValidation       = errors.New("validation error")
	ErrNotFound         = errors.New("not found")
	ErrVersionConflict  = errors.New("version conflict")
	ErrDuplicateRequest = errors.New("duplicate request")
)
