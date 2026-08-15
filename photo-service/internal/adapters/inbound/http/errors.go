package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/photo-service/internal/domain"
)

func statusForError(err error) int {
	switch {
	case errors.Is(err, domain.ErrJobNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrNoSources),
		errors.Is(err, domain.ErrTooManySources),
		errors.Is(err, domain.ErrInvalidLayout),
		errors.Is(err, domain.ErrInvalidLayoutType),
		errors.Is(err, domain.ErrInvalidCellFit),
		errors.Is(err, domain.ErrInvalidCellSpec),
		errors.Is(err, domain.ErrSourceCellMismatch),
		errors.Is(err, domain.ErrUnsupportedFormat),
		errors.Is(err, domain.ErrInvalidImageData):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrFetchSourceFailed):
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
