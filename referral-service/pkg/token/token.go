package token

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

const AuthorizationHeader = "Authorization"
const CurrentUserId = "user_id"

type Token struct {
	Sub       string `json:"sub"`
	Iss       string `json:"iss"`
	UserId    string `json:"custom:user_id"`
	Username  string `json:"cognito:username"`
	PhoneNo   string `json:"custom:phone_no"`
	OriginJTI string `json:"origin_jti"`
	Aud       string `json:"aud"`
	EventID   string `json:"event_id"`
	TokenUse  string `json:"token_use"`
	AuthTime  int64  `json:"auth_time"`
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
	Jti       string `json:"jti"`
}

func ParseJWT(tokenString string, dest any) (any, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}

	if err := json.Unmarshal(payload, dest); err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	return dest, nil
}

func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenHeader := c.Request.Header.Get(AuthorizationHeader)
		if tokenHeader != "" {
			c.Set(AuthorizationHeader, tokenHeader)
			c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), AuthorizationHeader, tokenHeader))
			tokenRaw, err := ParseJWT(tokenHeader, &Token{})
			tokenData, ok := tokenRaw.(*Token)
			if err == nil && ok {
				c.Set(CurrentUserId, tokenData.UserId)
				c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), CurrentUserId, tokenData.UserId))
			}
		}
		// Continue the request chain
		c.Next()
	}
}
