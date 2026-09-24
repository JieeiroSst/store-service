package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
)

const maxBodyBytes = 1 << 20

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err error) {
	status, msg := http.StatusInternalServerError, "internal error"
	switch {
	case errors.Is(err, model.ErrNotFound):
		status, msg = http.StatusNotFound, "not found"
	case errors.Is(err, model.ErrConflict):
		status, msg = http.StatusConflict, "already exists or still in use"
	case errors.Is(err, model.ErrInvalid):
		status, msg = http.StatusBadRequest, err.Error()
	}
	if status == http.StatusInternalServerError {
		slog.Error("request failed", "error", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func decode(w http.ResponseWriter, r *http.Request, v any) bool {
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		writeError(w, model.Invalid("malformed JSON body: %v", err))
		return false
	}
	return true
}

func pathID(w http.ResponseWriter, r *http.Request, name string) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue(name), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, model.Invalid("%s must be a positive integer", name))
		return 0, false
	}
	return id, true
}

type Handler interface{ Register(mux *http.ServeMux) }
