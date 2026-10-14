package sessionhttp

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	SessionsPath = "/api/v1/sessions"
	SessionPath  = "/api/v1/sessions/{id}"
)

type Handler struct {
	sessions input.SessionService
}

func New(sessions input.SessionService) *Handler {
	return &Handler{sessions: sessions}
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	if err := mux.HandlePath(http.MethodGet, SessionsPath, h.List); err != nil {
		return err
	}
	if err := mux.HandlePath(http.MethodDelete, SessionsPath, h.RevokeAll); err != nil {
		return err
	}
	return mux.HandlePath(http.MethodDelete, SessionPath, h.Revoke)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	resp, err := h.sessions.List(r.Context(), bearer(r))
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	writeJSON(w, resp)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if err := h.sessions.Revoke(r.Context(), bearer(r), params["id"]); err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) RevokeAll(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	includeCurrent, _ := strconv.ParseBool(r.URL.Query().Get("include_current"))
	resp, err := h.sessions.RevokeOthers(r.Context(), bearer(r), includeCurrent)
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	writeJSON(w, resp)
}

func bearer(r *http.Request) string {
	scheme, token, ok := strings.Cut(r.Header.Get("Authorization"), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}
