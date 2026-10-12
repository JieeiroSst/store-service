package httpapi

import (
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

const transferTimeout = 10 * time.Minute

func (h *Handler) upload(w http.ResponseWriter, r *http.Request) {
	segments := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	rc := http.NewResponseController(w)
	switch r.Method {
	case http.MethodPut:
		_ = rc.SetReadDeadline(time.Now().Add(transferTimeout))
		h.uploadPut(w, r, segments)
	case http.MethodGet, http.MethodHead:
		_ = rc.SetWriteDeadline(time.Now().Add(transferTimeout))
		h.uploadGet(w, r, segments)
	case http.MethodOptions:
		w.Header().Set("Allow", "OPTIONS, HEAD, GET, PUT")
		w.WriteHeader(http.StatusOK)
	default:
		w.Header().Set("Allow", "OPTIONS, HEAD, GET, PUT")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (h *Handler) uploadPut(w http.ResponseWriter, r *http.Request, segments []string) {
	if len(segments) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	err := h.files.Put(r.Context(), segments, r.URL.Query(), r.ContentLength, r.Body)
	if err != nil {
		status := uploadStatus(err)
		if status >= 500 {
			slog.Error("mod_http_upload: cannot store file", "path", r.URL.Path, "ip", clientIP(r).String(), "err", err)
		} else {
			slog.Warn("mod_http_upload: rejecting file", "path", r.URL.Path, "ip", clientIP(r).String(), "err", err)
		}
		w.WriteHeader(status)
		return
	}
	slog.Info("mod_http_upload: stored file", "path", r.URL.Path, "size", r.ContentLength, "ip", clientIP(r).String())
	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) uploadGet(w http.ResponseWriter, r *http.Request, segments []string) {
	f, err := h.files.Get(r.Context(), segments)
	if err != nil {
		status := uploadStatus(err)
		if status >= 500 {
			slog.Error("mod_http_upload: cannot read file", "path", r.URL.Path, "err", err)
		}
		w.WriteHeader(status)
		return
	}
	defer f.Body.Close()
	ct := f.ContentType
	if ct == "" {
		ct = "application/octet-stream"
	}
	disposition := "attachment"
	if strings.HasPrefix(ct, "image/") || strings.HasPrefix(ct, "video/") || strings.HasPrefix(ct, "audio/") {
		disposition = "inline"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": f.Name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	if f.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	}
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	if _, err := io.Copy(w, f.Body); err != nil {
		slog.Warn("mod_http_upload: download interrupted", "path", r.URL.Path, "err", err)
	}
}

func uploadStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalidSlot):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrSlotSizeMismatch), errors.Is(err, domain.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, domain.ErrUnsupportedFile):
		return http.StatusUnsupportedMediaType
	case errors.Is(err, domain.ErrFileNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrStorageUnavailable):
		return http.StatusServiceUnavailable
	}
	return http.StatusInternalServerError
}
