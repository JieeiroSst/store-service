package http

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const testInternalToken = "internal-secret"

func testToken(userID, role string) string {
	b, _ := json.Marshal(map[string]any{"sub": userID, "username": "u" + userID, "role": role})
	return "h." + base64.RawURLEncoding.EncodeToString(b) + ".s"
}

func fakeUserService(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in struct {
			SessionToken string `json:"session_token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&in)
		parts := strings.Split(in.SessionToken, ".")
		if len(parts) != 3 {
			_, _ = w.Write([]byte(`{"valid":false}`))
			return
		}
		payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
		var c struct {
			Sub string `json:"sub"`
		}
		_ = json.Unmarshal(payload, &c)
		_ = json.NewEncoder(w).Encode(map[string]any{"valid": c.Sub != "", "user_id": c.Sub})
	}))
	t.Cleanup(srv.Close)
	t.Setenv("USER_SERVICE_URL", srv.URL)
	t.Setenv("INTERNAL_API_TOKENS", testInternalToken)
}

type authCall struct {
	method, path, body string
	bearer, internal    string
}

func doAuth(h http.Handler, c authCall) (int, string) {
	var body io.Reader
	if c.body != "" {
		body = strings.NewReader(c.body)
	}
	r := httptest.NewRequest(c.method, c.path, body)
	r.Header.Set("Content-Type", "application/json")
	if c.bearer != "" {
		r.Header.Set("Authorization", "Bearer "+c.bearer)
	}
	if c.internal != "" {
		r.Header.Set("X-Internal-Token", c.internal)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w.Code, w.Body.String()
}
