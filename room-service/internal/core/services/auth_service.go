package services

import (
	"time"

	"github.com/JIeeiroSst/room-service/internal/core/domain/models"
	"github.com/JIeeiroSst/room-service/internal/core/ports"
	"github.com/golang-jwt/jwt/v5"
)

type authService struct {
	jwtSecret []byte
}

func NewAuthService(jwtSecret []byte) ports.AuthService {
	return &authService{jwtSecret: jwtSecret}
}

func (s *authService) GenerateToken(roomID uint, username string) (string, error) {
	claims := models.TokenClaims{
		Username: username,
		RoomID:   roomID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func (s *authService) ValidateToken(tokenString string) (*models.TokenClaims, error) {
	claims := &models.TokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	return claims, nil
}
