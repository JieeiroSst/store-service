package ssohttp

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/clientinfo"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/oauth2"
)

const (
	SSOBasePath    = "/api/v1/sso"
	SocialBasePath = "/api/v1/auth/"
	stateCookie    = "sso_state"
	stateMaxAge    = 600
	maxBodyBytes   = 16 << 10
)

type Handler struct {
	sso      input.SSOService
	basePath string
}

func New(sso input.SSOService, basePath string) *Handler {
	return &Handler{sso: sso, basePath: basePath}
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	if err := mux.HandlePath(http.MethodGet, h.basePath+"/login", h.Login); err != nil {
		return err
	}
	if err := mux.HandlePath(http.MethodGet, h.basePath+"/callback", h.Callback); err != nil {
		return err
	}
	return mux.HandlePath(http.MethodPost, h.basePath+"/token", h.Token)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	state, err := newState()
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	verifier := oauth2.GenerateVerifier()
	url, err := h.sso.LoginURL(r.Context(), state, verifier)
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	h.setStateCookie(w, r, state+"."+verifier, stateMaxAge)
	http.Redirect(w, r, url, http.StatusFound)
}

func (h *Handler) Callback(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	q := r.URL.Query()
	if e := q.Get("error"); e != "" {
		log.Printf("sso: provider returned error %q: %s", e, q.Get("error_description"))
		grpcadapter.WriteError(w, r, domain.ErrSSODenied)
		return
	}

	var saved, verifier string
	if cookie, err := r.Cookie(stateCookie); err == nil {
		saved, verifier, _ = strings.Cut(cookie.Value, ".")
	}
	h.setStateCookie(w, r, "", -1)
	state := q.Get("state")
	if saved == "" || state == "" || subtle.ConstantTimeCompare([]byte(saved), []byte(state)) != 1 {
		grpcadapter.WriteError(w, r, domain.ErrSSOStateMismatch)
		return
	}

	resp, err := h.sso.Callback(clientinfo.FromHTTP(r), q.Get("code"), verifier)
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	writeJSON(w, resp)
}

func (h *Handler) Token(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	var req struct {
		IDToken     string `json:"id_token"`
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodyBytes)).Decode(&req); err != nil {
		grpcadapter.WriteError(w, r, domain.ErrInvalidRequest)
		return
	}
	token := req.IDToken
	if token == "" {
		token = req.AccessToken
	}
	resp, err := h.sso.LoginWithToken(clientinfo.FromHTTP(r), token)
	if err != nil {
		grpcadapter.WriteError(w, r, err)
		return
	}
	writeJSON(w, resp)
}

func writeJSON(w http.ResponseWriter, resp any) {
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

func (h *Handler) setStateCookie(w http.ResponseWriter, r *http.Request, value string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookie,
		Value:    value,
		Path:     h.basePath,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https",
		SameSite: http.SameSiteLaxMode,
	})
}
