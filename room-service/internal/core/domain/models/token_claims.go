package models

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	Username string `json:"username"`
	RoomID   uint   `json:"room_id"`
	jwt.RegisteredClaims
}
