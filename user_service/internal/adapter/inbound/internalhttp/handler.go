package internalhttp

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

const (
	UserPath    = "/internal/v1/users/{id}"
	ProfilePath = "/api/v1/users/{id}/profile"
)

type SessionValidator interface {
	ValidateSession(ctx context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error)
}

type Profile struct {
	ID       int32  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
}

type User struct {
	ID         int32      `json:"id"`
	Username   string     `json:"username"`
	Email      string     `json:"email"`
	Name       string     `json:"name"`
	Phone      string     `json:"phone"`
	Address    string     `json:"address"`
	Sex        string     `json:"sex"`
	Active     bool       `json:"active"`
	CreateTime *time.Time `json:"create_time,omitempty"`
	UpdateTime *time.Time `json:"update_time,omitempty"`
}

type Handler struct {
	users    input.UserService
	sessions SessionValidator
	tokens   [][]byte
}

func New(users input.UserService, sessions SessionValidator, tokens []string) *Handler {
	h := &Handler{users: users, sessions: sessions}
	for _, t := range tokens {
		h.tokens = append(h.tokens, []byte(t))
	}
	if len(h.tokens) == 0 {
		log.Printf("internal API %s is open: set internal.api_tokens or INTERNAL_API_TOKENS to protect it", UserPath)
	}
	return h
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	if err := mux.HandlePath(http.MethodGet, UserPath, h.GetUser); err != nil {
		return err
	}
	return mux.HandlePath(http.MethodGet, ProfilePath, h.GetProfile)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if !h.authorized(r) {
		writeError(w, http.StatusUnauthorized, domain.CodeTokenInvalid, "missing or invalid internal token")
		return
	}
	u, ok := h.lookup(w, r, params["id"])
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, toUser(u))
}

func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if !h.authorizedCaller(r) {
		writeError(w, http.StatusUnauthorized, domain.CodeTokenInvalid, "missing or invalid access token")
		return
	}
	u, ok := h.lookup(w, r, params["id"])
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, Profile{ID: u.Id, Username: u.Username, Name: u.Name, Active: u.Checked})
}

func (h *Handler) lookup(w http.ResponseWriter, r *http.Request, raw string) (*dto.User, bool) {
	id, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, domain.CodeInvalidRequest, "id must be a positive integer")
		return nil, false
	}
	resp, err := h.users.FindUser(r.Context(), dto.FindUserRequest{UserId: int32(id)})
	if err != nil {
		var de *domain.Error
		if errors.As(err, &de) && de.Code == domain.CodeUserNotFound {
			writeError(w, http.StatusNotFound, de.Code, de.Message)
			return nil, false
		}
		log.Printf("internal get user %d: %v", id, err)
		writeError(w, http.StatusServiceUnavailable, domain.CodeServiceUnavailable, "user lookup failed")
		return nil, false
	}
	if len(resp.Users) == 0 {
		writeError(w, http.StatusNotFound, domain.CodeUserNotFound, "user does not exist")
		return nil, false
	}
	return resp.Users[0], true
}

func (h *Handler) authorizedCaller(r *http.Request) bool {
	if t := strings.TrimSpace(r.Header.Get("X-Internal-Token")); t != "" {
		return h.matchesInternal(t)
	}
	t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	t = strings.TrimSpace(t)
	if !ok || t == "" {
		return false
	}
	if h.matchesInternal(t) {
		return true
	}
	if h.sessions == nil {
		return false
	}
	resp, err := h.sessions.ValidateSession(r.Context(), dto.ValidateRequest{SessionToken: t})
	if err != nil {
		log.Printf("profile: validate session: %v", err)
		return false
	}
	return resp.Valid && resp.UserId != ""
}

func (h *Handler) matchesInternal(t string) bool {
	match := 0
	for _, want := range h.tokens {
		match |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return match == 1
}

func (h *Handler) authorized(r *http.Request) bool {
	if len(h.tokens) == 0 {
		return true
	}
	t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
	if !ok {
		t = r.Header.Get("X-Internal-Token")
	}
	t = strings.TrimSpace(t)
	if t == "" {
		return false
	}
	return h.matchesInternal(t)
}

func toUser(u *dto.User) User {
	out := User{
		ID:       u.Id,
		Username: u.Username,
		Email:    u.Email,
		Name:     u.Name,
		Phone:    u.Phone,
		Address:  u.Address,
		Sex:      u.Sex,
		Active:   u.Checked,
	}
	if u.CreateTime != nil {
		t := u.CreateTime.AsTime()
		out.CreateTime = &t
	}
	if u.UpdateTime != nil {
		t := u.UpdateTime.AsTime()
		out.UpdateTime = &t
	}
	return out
}

func writeError(w http.ResponseWriter, status int, code domain.ErrorCode, msg string) {
	writeJSON(w, status, map[string]any{"code": code, "message": msg, "status": status})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
