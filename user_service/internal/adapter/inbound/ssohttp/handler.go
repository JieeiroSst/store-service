package ssohttp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"

	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	LoginPath    = "/api/v1/sso/login"
	CallbackPath = "/api/v1/sso/callback"
	stateCookie  = "sso_state"
	stateMaxAge  = 600
)

type Handler struct {
	sso input.SSOService
}

func New(sso input.SSOService) *Handler {
	return &Handler{sso: sso}
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	if err := mux.HandlePath(http.MethodGet, LoginPath, h.Login); err != nil {
		return err
	}
	return mux.HandlePath(http.MethodGet, CallbackPath, h.Callback)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	state, err := newState()
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	url, err := h.sso.LoginURL(r.Context(), state)
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	setStateCookie(w, r, state, stateMaxAge)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		log.Printf("sso: provider returned error %q: %s", e, q.Get("error_description"))
		grpcadapter.WriteError(w, r, domain.ErrSSODenied)
		return
	}

	cookie, err := r.Cookie(stateCookie)
	setStateCookie(w, r, "", -1)
	state := q.Get("state")
	if err != nil || state == "" || subtle.ConstantTimeCompare([]byte(cookie.Value), []byte(state)) != 1 {
		grpcadapter.WriteError(w, r, domain.ErrSSOStateMismatch)
		return
	}

	resp, err := h.sso.Callback(r.Context(), q.Get("code"))
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

func newState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setStateCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookie,
		Value:    value,
		Path:     "/api/v1/sso",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}
