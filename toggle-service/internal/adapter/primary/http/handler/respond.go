package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"

	"github.com/JIeeiroSst/toggle-service/internal/application/apperr"
)

var validate = validator.New()

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, apperr.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, apperr.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, apperr.ErrForbidden):
		status = http.StatusForbidden
	case errors.Is(err, apperr.ErrUnauthorized), errors.Is(err, apperr.ErrInvalidCredentials):
		status = http.StatusUnauthorized
	case errors.Is(err, apperr.ErrValidation):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func decodeAndValidate(r *http.Request, v any) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return apperr.ErrValidation
	}
	if err := validate.Struct(v); err != nil {
		return apperr.ErrValidation
	}
	return nil
}
