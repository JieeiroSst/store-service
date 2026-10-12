package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/application"
	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

const maxBody = 1 << 20

type Handler struct {
	cfg      *config.Config
	matcher  *domain.ACLMatcher
	auth     *application.AuthService
	health   port.HealthChecker
	files    *application.HTTPUploadService
	commands map[string]command
}

func NewHandler(cfg *config.Config, matcher *domain.ACLMatcher, auth *application.AuthService, admin *application.AdminService, health port.HealthChecker, files *application.HTTPUploadService) *Handler {
	cs := &commandSet{admin: admin, auth: auth, started: time.Now()}
	return &Handler{cfg: cfg, matcher: matcher, auth: auth, health: health, files: files, commands: cs.table()}
}

func (h *Handler) Router(l config.Listener) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /health", h.healthz)
	for path, module := range l.RequestHandlers {
		path = "/" + strings.Trim(path, "/")
		switch module {
		case "mod_http_api":
			if !h.cfg.Modules.HTTPAPI {
				slog.Warn("mod_http_api is not enabled in modules, skipping request handler", "path", path)
				continue
			}
			mux.Handle(path+"/", http.StripPrefix(path, http.HandlerFunc(h.api)))
		case "mod_http_upload":
			if !h.files.Enabled() {
				slog.Warn("mod_http_upload is not enabled in modules, skipping request handler", "path", path)
				continue
			}
			mux.Handle(path+"/", http.StripPrefix(path, http.HandlerFunc(h.upload)))
		default:
			slog.Warn("unsupported http request handler, skipping", "path", path, "module", module, "port", l.Port)
		}
	}
	return mux
}

func (h *Handler) healthz(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func clientIP(r *http.Request) net.IP {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	return net.ParseIP(host)
}

func (h *Handler) extractAuth(r *http.Request) (domain.Match, bool) {
	m := domain.Match{IP: clientIP(r)}
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return m, true
	}
	user, pass, ok := r.BasicAuth()
	if !ok {
		return m, false
	}
	j, err := domain.ParseJID(user)
	if err != nil || j.LUser == "" {
		return m, false
	}
	if !h.auth.CheckPassword(r.Context(), j.LUser, "", j.LServer, pass) {
		return m, false
	}
	bare := j.Bare()
	l := bare.LJID()
	m.USR = &l
	return m, true
}

func (h *Handler) api(w http.ResponseWriter, r *http.Request) {
	call := strings.Trim(r.URL.Path, "/")
	if i := strings.IndexByte(call, '/'); i >= 0 {
		call = call[:i]
	}
	switch r.Method {
	case http.MethodOptions:
		w.Header().Set("Allow", "POST, GET, OPTIONS")
		w.WriteHeader(http.StatusOK)
		return
	case http.MethodPost, http.MethodGet:
	default:
		writeError(w, http.StatusNotFound, 40, "Endpoint not found.")
		return
	}
	match, ok := h.extractAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, 10, "You are not authorized to call this command.")
		return
	}
	a := args{}
	if r.Method == http.MethodPost {
		body, err := io.ReadAll(io.LimitReader(r.Body, maxBody))
		if err != nil || len(bytes.TrimSpace(body)) == 0 {
			writeJSON(w, http.StatusBadRequest, "400 Bad Request")
			return
		}
		if err := json.Unmarshal(body, &a); err != nil {
			writeJSON(w, http.StatusBadRequest, "400 Bad Request")
			return
		}
	} else {
		for k, v := range r.URL.Query() {
			if len(v) > 0 {
				a[k] = v[0]
			}
		}
	}
	cmd, exists := h.commands[call]
	if !exists {
		writeError(w, http.StatusNotFound, 44, "Command not found.")
		return
	}
	if !canAccess(h.cfg, h.matcher, call, caller{match: match, module: "mod_http_api"}) {
		writeError(w, http.StatusForbidden, 32, "AccessRules: Account does not have the right to perform the operation.")
		return
	}
	slog.Info("API call", "call", call, "ip", match.IP.String())
	res, err := cmd(r.Context(), a)
	if err != nil {
		if ce, ok := asCommandError(err); ok {
			writeError(w, ce.Status, ce.Code, ce.Message)
			return
		}
		writeJSON(w, http.StatusBadRequest, err.Error())
		return
	}
	switch v := res.(type) {
	case rescode:
		if v {
			writeJSON(w, http.StatusOK, 0)
		} else {
			writeJSON(w, http.StatusOK, 1)
		}
	case restuple:
		writeJSON(w, http.StatusOK, string(v))
	default:
		writeJSON(w, http.StatusOK, v)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status, code int, msg string) {
	writeJSON(w, status, map[string]any{"status": "error", "code": code, "message": msg})
}
