package internalhttp

import (
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

const UserPath = "/internal/v1/users/{id}"

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
	users  input.UserService
	tokens [][]byte
}

func New(users input.UserService, tokens []string) *Handler {
	h := &Handler{users: users}
	for _, t := range tokens {
		h.tokens = append(h.tokens, []byte(t))
	}
	if len(h.tokens) == 0 {
		log.Printf("internal API %s is open: set internal.api_tokens or INTERNAL_API_TOKENS to protect it", UserPath)
	}
	return h
}

func (h *Handler) Register(mux *runtime.ServeMux) error {
	return mux.HandlePath(http.MethodGet, UserPath, h.GetUser)
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request, params map[string]string) {
	if !h.authorized(r) {
		writeError(w, http.StatusUnauthorized, domain.CodeTokenInvalid, "missing or invalid internal token")
		return
	}
	id, err := strconv.ParseInt(params["id"], 10, 32)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, domain.CodeInvalidRequest, "id must be a positive integer")
		return
	}
	resp, err := h.users.FindUser(r.Context(), dto.FindUserRequest{UserId: int32(id)})
	if err != nil {
		var de *domain.Error
		if errors.As(err, &de) && de.Code == domain.CodeUserNotFound {
			writeError(w, http.StatusNotFound, de.Code, de.Message)
			return
		}
		log.Printf("internal get user %d: %v", id, err)
		writeError(w, http.StatusServiceUnavailable, domain.CodeServiceUnavailable, "user lookup failed")
		return
	}
	if len(resp.Users) == 0 {
		writeError(w, http.StatusNotFound, domain.CodeUserNotFound, "user does not exist")
		return
	}
	writeJSON(w, http.StatusOK, toUser(resp.Users[0]))
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
	match := 0
	for _, want := range h.tokens {
		match |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return match == 1
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
