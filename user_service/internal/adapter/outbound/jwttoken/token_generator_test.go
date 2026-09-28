package jwttoken

import (
	"context"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func TestAccessTokenCarriesRoles(t *testing.T) {
	g := New("secret", time.Minute)
	tok, err := g.GenerateAccessToken(context.Background(), 42, "alice", "operator", []string{"operator", "crm-manager", "user"})
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := jwt.Parse(tok, func(*jwt.Token) (interface{}, error) { return []byte("secret"), nil })
	if err != nil {
		t.Fatal(err)
	}
	claims := parsed.Claims.(jwt.MapClaims)
	if claims["sub"] != "42" || claims["role"] != "operator" {
		t.Errorf("claims = %v", claims)
	}
	roles, _ := claims["roles"].([]interface{})
	if len(roles) != 3 || roles[1] != "crm-manager" {
		t.Errorf("roles = %v", claims["roles"])
	}

	// No roles (authorize-service unreachable at login) is an empty list.
	tok, _ = g.GenerateAccessToken(context.Background(), 42, "alice", "", nil)
	parsed, _ = jwt.Parse(tok, func(*jwt.Token) (interface{}, error) { return []byte("secret"), nil })
	if r, ok := parsed.Claims.(jwt.MapClaims)["roles"].([]interface{}); !ok || len(r) != 0 {
		t.Errorf("roles = %v, want []", parsed.Claims.(jwt.MapClaims)["roles"])
	}
}
