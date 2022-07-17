package token

import (
	"time"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/dgrijalva/jwt-go"
)

type Tokens interface {
	GenerateToken(username string) (string, error)
	ParseToken(tokenStr string) (*model.ParseToken, error)
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
	atClaims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	at := jwt.NewWithClaims(jwt.SigningMethodHS256, atClaims)
	token, err := at.SignedString([]byte(t.config.Secret.JwtSecretKey))
	if err != nil {
		return "", err
	}
	return token, nil
}

func (t *token) ParseToken(tokenStr string) (*model.ParseToken, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		return []byte(t.config.Secret.JwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok && !token.Valid {
		return nil, common.FailedToken
	}

	username := claims["username"].(string)
	return &model.ParseToken{
		Username: username,
	}, nil
}
