package http

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/JIeeiroSst/admanagement-service/common"
)

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload != nil {
		_ = json.NewEncoder(w).Encode(payload)
	}
}

func decodeJSON(r *http.Request, dst interface{}) error {
	defer r.Body.Close()
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		return common.ErrInvalidInput
	}
	return nil
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, common.ErrNotFound), errors.Is(err, common.ErrNoAdAvailable):
		status = http.StatusNotFound
	case errors.Is(err, common.ErrInvalidInput), errors.Is(err, common.ErrInvalidStatus):
		status = http.StatusBadRequest
	}
	writeJSON(w, status, map[string]string{"error": err.Error()})
}
