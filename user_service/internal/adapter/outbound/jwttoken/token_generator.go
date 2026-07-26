package jwttoken

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/dgrijalva/jwt-go"
)

type Generator struct {
	secretKey      string
	accessTokenTTL time.Duration
}

func New(secretKey string, accessTokenTTL time.Duration) *Generator {
	return &Generator{secretKey: secretKey, accessTokenTTL: accessTokenTTL}
}

func (g *Generator) GenerateAccessToken(ctx context.Context, userID int, username string) (string, error) {
	claims := jwt.MapClaims{
		"authorized": true,
		"sub":        strconv.Itoa(userID),
		"username":   username,
		"exp":        time.Now().Add(g.accessTokenTTL).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(g.secretKey))
}

func (g *Generator) GenerateRefreshToken(ctx context.Context) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func (g *Generator) ParseAccessToken(ctx context.Context, tokenStr string) (domain.AccessClaims, error) {
	parsed, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(g.secretKey), nil
	})
	if err != nil {
		return domain.AccessClaims{}, err
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return domain.AccessClaims{}, domain.ErrFailedToken
	}

	sub, _ := claims["sub"].(string)
	username, _ := claims["username"].(string)
	userID, err := strconv.Atoi(sub)
	if err != nil {
		return domain.AccessClaims{}, domain.ErrFailedToken
	}

	return domain.AccessClaims{UserID: userID, Username: username}, nil
}
