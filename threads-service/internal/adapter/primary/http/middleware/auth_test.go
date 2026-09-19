package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret"

func signToken(t *testing.T, secret, subject string, expiresIn time.Duration) string {
	t.Helper()
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return token
}

func newAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/protected", RequireAuth(testSecret), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"userId": UserID(c)})
	})
	return r
}

func doRequest(r *gin.Engine, bearer string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRequireAuthRejectsMissingToken(t *testing.T) {
	w := doRequest(newAuthTestRouter(), "")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsGarbageToken(t *testing.T) {
	w := doRequest(newAuthTestRouter(), "not-a-real-jwt")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsExpiredToken(t *testing.T) {
	token := signToken(t, testSecret, "user-1", -time.Hour)
	w := doRequest(newAuthTestRouter(), token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthRejectsWrongSecret(t *testing.T) {
	token := signToken(t, "a-different-secret", "user-1", time.Hour)
	w := doRequest(newAuthTestRouter(), token)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestRequireAuthAcceptsValidTokenAndExposesUserID(t *testing.T) {
	token := signToken(t, testSecret, "user-1", time.Hour)
	w := doRequest(newAuthTestRouter(), token)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), `"userId":"user-1"`) {
		t.Errorf("expected body to expose userId=user-1, got %s", w.Body.String())
	}
}
