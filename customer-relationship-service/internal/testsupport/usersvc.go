package testsupport

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

type UserService struct {
	Server *httptest.Server
	Down   bool

	mu     sync.Mutex
	tokens map[string]string
}

func NewUserService(t testing.TB) *UserService {
	t.Helper()
	u := &UserService{tokens: map[string]string{}}
	u.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if u.Down {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		if r.URL.Path != "/api/v1/validate" {
			http.NotFound(w, r)
			return
		}
		var in struct {
			SessionToken string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		u.mu.Lock()
		id, ok := u.tokens[in.SessionToken]
		u.mu.Unlock()
		if !ok {
			_, _ = w.Write([]byte(`{}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"valid": true, "userId": id})
	}))
	t.Cleanup(u.Server.Close)
	return u
}

func (u *UserService) URL() string { return u.Server.URL }

func (u *UserService) Token(t testing.TB, userID, username, role string, roles ...string) string {
	t.Helper()
	if len(roles) == 0 {
		roles = []string{role}
	}
	payload, _ := json.Marshal(map[string]any{"sub": userID, "username": username, "role": role, "roles": roles})
	tok := "e30." + base64.RawURLEncoding.EncodeToString(payload) + ".sig"
	u.mu.Lock()
	u.tokens[tok] = userID
	u.mu.Unlock()
	return tok
}
