package port

import (
	"errors"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
)

var (
	ErrNotFound             = errors.New("resource not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrConflict             = errors.New("conflict")
	ErrInsufficientStock    = errors.New("insufficient unexpired stock")
	ErrPrescriptionRequired = errors.New("prescription required")
	ErrSafetyBlocked        = errors.New("dispensing blocked by safety check")
	ErrOverrideRequired     = errors.New("safety warnings require an override reason")
)

type SafetyError struct {
	Reason error
	Report model.SafetyReport
}

func (e *SafetyError) Error() string { return e.Reason.Error() }
func (e *SafetyError) Unwrap() error { return e.Reason }
