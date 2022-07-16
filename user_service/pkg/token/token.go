package token

import (
	"time"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/dgrijalva/jwt-go"
)

type Tokens interface {
	GenerateToken(username string) (string, error)
	ParseToken(tokenStr string) (string, error)
}

type token struct {
	config *config.Config
}

func NewToken(config *config.Config) Tokens {
	return &token{
		config: config,
	}
}

func (t *token) GenerateToken(username string) (string, error) {
	atClaims := jwt.MapClaims{}
	atClaims["authorized"] = true
	atClaims["username"] = username
	atClaims["exp"] = time.Now().Add(time.Hour * 60 * 60 * 60).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(t.config.Secret.JwtSecretKey))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (t *token) ParseToken(tokenStr string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.config.Secret.JwtSecretKey), nil
	})

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username := claims["username"].(string)
		return username, nil
	} else {
		return "Missing Authentication Token", err
	}
}
