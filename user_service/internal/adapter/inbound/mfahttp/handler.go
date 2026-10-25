package mfahttp

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/clientinfo"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	TOTPPath     = "/api/v1/mfa/totp"
	maxBodyBytes = 4 << 10
)

type Handler struct {
	mfa input.MFAService
}

func New(mfa input.MFAService) *Handler {
	return &Handler{mfa: mfa}
}

type codeRequest struct {
	Code     string `json:"code"`
	MFAToken string `json:"mfa_token"`
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	routes := []struct {
		method, path string
		fn           runtime.HandlerFunc
	}{
		{http.MethodGet, TOTPPath, h.Status},
		{http.MethodPost, TOTPPath + "/setup", h.Setup},
		{http.MethodPost, TOTPPath + "/enable", h.Enable},
		{http.MethodPost, TOTPPath + "/disable", h.Disable},
		{http.MethodPost, TOTPPath + "/recovery-codes", h.RecoveryCodes},
		{http.MethodPost, TOTPPath + "/verify", h.Verify},
	}
	for _, rt := range routes {
		if err := mux.HandlePath(rt.method, rt.path, rt.fn); err != nil {
			return err
		}
	}
	return nil
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	resp, err := h.mfa.Status(r.Context(), bearer(r))
	respond(w, r, resp, err)
}

func (h *Handler) Setup(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	resp, err := h.mfa.SetupTOTP(r.Context(), bearer(r))
	respond(w, r, resp, err)
}

func (h *Handler) Enable(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	req, ok := decode(w, r)
	if !ok {
		return
	}
	resp, err := h.mfa.EnableTOTP(r.Context(), bearer(r), req.Code)
	respond(w, r, resp, err)
}

func (h *Handler) Disable(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	req, ok := decode(w, r)
	if !ok {
		return
	}
	if err := h.mfa.DisableTOTP(r.Context(), bearer(r), req.Code); err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RecoveryCodes(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	req, ok := decode(w, r)
	if !ok {
		return
	}
	resp, err := h.mfa.RegenerateRecoveryCodes(r.Context(), bearer(r), req.Code)
	respond(w, r, resp, err)
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	req, ok := decode(w, r)
	if !ok {
		return
	}
	resp, err := h.mfa.VerifyLogin(clientinfo.FromHTTP(r), req.MFAToken, req.Code)
	respond(w, r, resp, err)
}

func decode(w http.ResponseWriter, r *http.Request) (codeRequest, bool) {
	var req codeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		grpcadapter.WriteError(w, r, domain.ErrInvalidRequest)
		return codeRequest{}, false
	}
	return req, true
}

func respond(w http.ResponseWriter, r *http.Request, v any, err error) {
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func bearer(r *http.Request) string {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}
