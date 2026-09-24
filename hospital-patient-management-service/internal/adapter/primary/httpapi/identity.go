// Package httpapi serves the REST endpoints that are not part of the gRPC
// contract: linking a patient to a user account and eKYC, whose photo uploads
// are multipart and so do not fit the protobuf services.
package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

const (
	maxUploadBytes = 20 << 20 // whole request
	maxFormMemory  = 8 << 20
)

type IdentityHandler struct{ uc port.IdentityUsecase }

func NewIdentityHandler(uc port.IdentityUsecase) *IdentityHandler { return &IdentityHandler{uc: uc} }

func (h *IdentityHandler) Register(mux *http.ServeMux) {
	mux.HandleFunc("PUT /v1/patients/{id}/user", h.link)
	mux.HandleFunc("GET /v1/patients/{id}/identity", h.status)
	mux.HandleFunc("POST /v1/patients/{id}/identity/citizen-card", h.citizenCard)
	mux.HandleFunc("POST /v1/patients/{id}/identity/face-scan", h.faceScan)
	mux.HandleFunc("POST /v1/patients/{id}/identity/verify", h.verify)
}

type identityJSON struct {
	PatientID      int32      `json:"patient_id"`
	Status         string     `json:"status"`
	MatchScore     float64    `json:"match_score"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	HasCitizenCard bool       `json:"has_citizen_card"`
	HasFaceScan    bool       `json:"has_face_scan"`
}

func (h *IdentityHandler) link(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	var body struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<16)).Decode(&body); err != nil {
		writeError(w, model.Invalid("body must be JSON like {\"user_id\":\"...\"}"))
		return
	}
	p, err := h.uc.LinkUser(r.Context(), id, body.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	userID := ""
	if p.UserID != nil {
		userID = *p.UserID
	}
	writeJSON(w, http.StatusOK, map[string]any{"patient_id": p.ID, "user_id": userID})
}

func (h *IdentityHandler) status(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	h.reply(w, id, func() (*model.IdentityStatus, error) { return h.uc.Status(r.Context(), id) }, http.StatusOK)
}

func (h *IdentityHandler) verify(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	h.reply(w, id, func() (*model.IdentityStatus, error) { return h.uc.Verify(r.Context(), id) }, http.StatusOK)
}

func (h *IdentityHandler) citizenCard(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	form, ok := parseForm(w, r)
	if !ok {
		return
	}
	front, err := formFile(form, "front")
	if err == nil {
		var back []byte
		if back, err = formFile(form, "back"); err == nil {
			h.reply(w, id, func() (*model.IdentityStatus, error) {
				return h.uc.SubmitCitizenCard(r.Context(), id, front, back)
			}, http.StatusCreated)
			return
		}
	}
	writeError(w, model.Invalid("%v", err))
}

func (h *IdentityHandler) faceScan(w http.ResponseWriter, r *http.Request) {
	id, ok := patientID(w, r)
	if !ok {
		return
	}
	form, ok := parseForm(w, r)
	if !ok {
		return
	}
	var frames [][]byte
	for _, fh := range form.File["frames"] {
		data, err := readFile(fh)
		if err != nil {
			writeError(w, model.Invalid("unreadable frame: %v", err))
			return
		}
		frames = append(frames, data)
	}
	h.reply(w, id, func() (*model.IdentityStatus, error) { return h.uc.SubmitFaceScan(r.Context(), id, frames) }, http.StatusCreated)
}

func (h *IdentityHandler) reply(w http.ResponseWriter, id int32, call func() (*model.IdentityStatus, error), code int) {
	st, err := call()
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, code, identityJSON{
		PatientID: id, Status: string(st.State), MatchScore: st.MatchScore, VerifiedAt: st.VerifiedAt,
		HasCitizenCard: st.HasCard, HasFaceScan: st.HasFace,
	})
}

func patientID(w http.ResponseWriter, r *http.Request) (int32, bool) {
	n, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil || n <= 0 {
		writeError(w, model.Invalid("patient id must be a positive integer"))
		return 0, false
	}
	return int32(n), true
}

func parseForm(w http.ResponseWriter, r *http.Request) (*multipart.Form, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxFormMemory); err != nil {
		writeError(w, model.Invalid("expected a multipart form of at most %d MB", maxUploadBytes>>20))
		return nil, false
	}
	return r.MultipartForm, true
}

func formFile(form *multipart.Form, field string) ([]byte, error) {
	fhs := form.File[field]
	if len(fhs) == 0 {
		return nil, errors.New("missing '" + field + "' image")
	}
	return readFile(fhs[0])
}

func readFile(fh *multipart.FileHeader) ([]byte, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError uses the same {"code","message"} shape and status mapping as the
// grpc-gateway routes.
func writeError(w http.ResponseWriter, err error) {
	code, grpcCode, msg := http.StatusInternalServerError, 13, "internal error"
	switch {
	case errors.Is(err, model.ErrNotFound):
		code, grpcCode, msg = http.StatusNotFound, 5, err.Error()
	case errors.Is(err, model.ErrInvalid):
		code, grpcCode, msg = http.StatusBadRequest, 3, err.Error()
	case errors.Is(err, model.ErrConflict):
		code, grpcCode, msg = http.StatusConflict, 6, err.Error()
	case errors.Is(err, model.ErrUpstream):
		logrus.WithError(err).Error("dependency unavailable")
		code, grpcCode, msg = http.StatusServiceUnavailable, 14, "a dependent service is unavailable"
	default:
		logrus.WithError(err).Error("request failed")
	}
	writeJSON(w, code, map[string]any{"code": grpcCode, "message": msg})
}

var Module = fx.Options(fx.Provide(NewIdentityHandler, NewDocumentHandler))
