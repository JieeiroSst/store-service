package authtoken

import (
	"fmt"
	"time"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/JIeeiroSst/voucher-service/internal/platform/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTIssuer struct {
	secret     []byte
	expiration time.Duration
}

func NewJWTIssuer(cfg *config.Config) authapp.TokenIssuer {
	return &JWTIssuer{secret: []byte(cfg.JWTSecret), expiration: cfg.JWTExpiration}
}

type jwtClaims struct {
	UserID      string `json:"user_id"`
	Role        string `json:"role"`
	CorporateID string `json:"corporate_id,omitempty"`
	jwt.RegisteredClaims
}

func (i *JWTIssuer) Issue(claims authapp.Claims) (string, int64, error) {
	now := time.Now()
	c := jwtClaims{
		UserID:      claims.UserID,
		Role:        claims.Role,
		CorporateID: claims.CorporateID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(i.expiration)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	signed, err := token.SignedString(i.secret)
	if err != nil {
		return "", 0, err
	}
	return signed, int64(i.expiration.Seconds()), nil
}

func (i *JWTIssuer) Verify(tokenStr string) (*authapp.Claims, error) {
	var c jwtClaims
	token, err := jwt.ParseWithClaims(tokenStr, &c, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return i.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token: %w", err)
	}
	return &authapp.Claims{UserID: c.UserID, Role: c.Role, CorporateID: c.CorporateID}, nil
}
