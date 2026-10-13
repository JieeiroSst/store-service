package middlware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/gateway-api/auth"
	"github.com/JIeeiroSst/gateway-api/proxy/user_proxy"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gin-gonic/gin"
)

func Authenticate(users user_proxy.UserProxy) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := strings.TrimSpace(c.GetHeader("Authorization"))
		if header == "" {
			c.Next()
			return
		}

		scheme, token, ok := strings.Cut(header, " ")
		token = strings.TrimSpace(token)
		if !ok || !strings.EqualFold(scheme, "Bearer") || token == "" {
			abort(c, http.StatusUnauthorized, auth.CodeUnauthenticated, "authorization header must be: Bearer <token>")
			return
		}

		session, err := users.ValidateSession(c.Request.Context(), token)
		if err != nil {
			logger.Error(c.Request.Context(), "validate session: %v", err)
			abort(c, http.StatusServiceUnavailable, auth.CodeServiceUnavailable, "cannot verify token, try again later")
			return
		}
		if !session.Valid || session.UserID == "" {
			abort(c, http.StatusUnauthorized, auth.CodeUnauthenticated, "invalid or expired token")
			return
		}

		ctx := auth.WithUser(c.Request.Context(), auth.User{ID: session.UserID, Username: usernameClaim(token), Token: token})
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func abort(c *gin.Context, status int, code, message string) {
	c.AbortWithStatusJSON(status, gin.H{
		"errors": []gin.H{{
			"message":    message,
			"extensions": gin.H{"code": code},
		}},
	})
}

func usernameClaim(token string) string {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return ""
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}
	var claims struct {
		Username string `json:"username"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return ""
	}
	return claims.Username
}
