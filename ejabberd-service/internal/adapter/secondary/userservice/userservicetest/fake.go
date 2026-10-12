package userservicetest

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

type FakeUser struct {
	ID       int
	Username string
	Password string
}

type Fake struct {
	mu       sync.Mutex
	users    []FakeUser
	sessions map[string]FakeUser
	Logouts  atomic.Int64
	Lookups  atomic.Int64
	SignUps  atomic.Int64
	Down     atomic.Int64
	Token    string
}

func NewFake(token string, users ...FakeUser) *Fake {
	return &Fake{users: users, sessions: map[string]FakeUser{}, Token: token}
}

func (f *Fake) Login(username, password string) (string, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Username == username && u.Password == password {
			tok := issue(u)
			f.sessions[tok] = u
			return tok, true
		}
	}
	return "", false
}

var (
	passwordRule = regexp.MustCompile(`([A-Z])\w+`)
	emailRule    = regexp.MustCompile(`^[a-z][a-z0-9_\.]{5,32}@[a-z0-9]{2,}(\.[a-z0-9]{2,4}){1,2}$`)
)

func (f *Fake) HasUser(username string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Username == username {
			return true
		}
	}
	return false
}

func (f *Fake) signUp(username, password, email string) (int, string) {
	switch {
	case username == "":
		return http.StatusBadRequest, "username is required"
	case !emailRule.MatchString(email):
		return http.StatusBadRequest, "email does not satisfy the condition"
	case !passwordRule.MatchString(password):
		return http.StatusBadRequest, "password does not satisfy the condition"
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, u := range f.users {
		if u.Username == username {
			return http.StatusConflict, "username already exists"
		}
	}
	f.users = append(f.users, FakeUser{ID: 1000 + len(f.users), Username: username, Password: password})
	return http.StatusOK, ""
}

func (f *Fake) Revoke(token string) {
	f.mu.Lock()
	delete(f.sessions, token)
	f.mu.Unlock()
}

func issue(u FakeUser) string {
	enc := base64.RawURLEncoding
	header := enc.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	claims, _ := json.Marshal(map[string]any{"sub": strconv.Itoa(u.ID), "username": u.Username, "role": "user"})
	sig := make([]byte, 16)
	_, _ = rand.Read(sig)
	return header + "." + enc.EncodeToString(claims) + "." + hex.EncodeToString(sig)
}

func (f *Fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeJSON := func(status int, v any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_ = json.NewEncoder(w).Encode(v)
	}
	var body map[string]string
	if r.Method == http.MethodPost {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	if f.Down.Load() > 0 {
		f.Down.Add(-1)
		writeJSON(http.StatusServiceUnavailable, map[string]any{"message": "unavailable"})
		return
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/user/sign-up":
		f.SignUps.Add(1)
		status, msg := f.signUp(body["username"], body["password"], body["email"])
		if status != http.StatusOK {
			writeJSON(status, map[string]any{"message": msg})
			return
		}
		writeJSON(http.StatusOK, map[string]any{"message": "success"})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/login":
		tok, ok := f.Login(body["username"], body["password"])
		if !ok {
			writeJSON(http.StatusUnauthorized, map[string]any{"code": 16, "message": "invalid username or password"})
			return
		}
		writeJSON(http.StatusOK, map[string]any{"sessionToken": tok, "refreshToken": "r", "expiryTime": "900"})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/validate":
		f.mu.Lock()
		u, ok := f.sessions[body["session_token"]]
		f.mu.Unlock()
		if !ok {
			writeJSON(http.StatusOK, map[string]any{})
			return
		}
		writeJSON(http.StatusOK, map[string]any{"valid": true, "userId": strconv.Itoa(u.ID)})
	case r.Method == http.MethodPost && r.URL.Path == "/api/v1/logout":
		f.Logouts.Add(1)
		f.Revoke(body["session_token"])
		writeJSON(http.StatusOK, map[string]any{"message": "true"})
	case r.Method == http.MethodGet && r.URL.Path == "/user":
		f.Lookups.Add(1)
		name := r.URL.Query().Get("username")
		var users []map[string]any
		f.mu.Lock()
		for _, u := range f.users {
			if u.Username == name {
				users = append(users, map[string]any{"id": u.ID, "username": u.Username})
			}
		}
		f.mu.Unlock()
		writeJSON(http.StatusOK, map[string]any{"users": users, "total": len(users)})
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/internal/v1/users/"):
		f.Lookups.Add(1)
		if f.Token != "" && r.Header.Get("Authorization") != "Bearer "+f.Token {
			writeJSON(http.StatusUnauthorized, map[string]any{"message": "missing or invalid internal token"})
			return
		}
		id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/internal/v1/users/"))
		f.mu.Lock()
		defer f.mu.Unlock()
		for _, u := range f.users {
			if u.ID == id {
				writeJSON(http.StatusOK, map[string]any{"id": u.ID, "username": u.Username})
				return
			}
		}
		writeJSON(http.StatusNotFound, map[string]any{"message": "user does not exist"})
	default:
		writeJSON(http.StatusNotFound, map[string]any{"message": "not found"})
	}
}
