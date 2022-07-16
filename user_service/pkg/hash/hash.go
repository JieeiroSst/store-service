package hash

import (
	"golang.org/x/crypto/bcrypt"
)

type Hash interface {
	HashPassword(password string) (string, error)
	CheckPassowrd(password, hash string) error
}

type hash struct{}

func NewHash() Hash {
	return &hash{}
}

func (h *hash) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}

func (h *hash) CheckPassowrd(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err
}
