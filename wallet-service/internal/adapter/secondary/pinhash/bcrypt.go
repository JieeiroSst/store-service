package pinhash

import (
	"golang.org/x/crypto/bcrypt"

	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

type bcryptHasher struct {
	cost int
}

func NewBcrypt(cost int) port.PinHasher {
	if cost < bcrypt.MinCost {
		cost = bcrypt.DefaultCost
	}
	return &bcryptHasher{cost: cost}
}

func (h *bcryptHasher) Hash(pin string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pin), h.cost)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (h *bcryptHasher) Compare(hash, pin string) bool {
	return hash != "" && bcrypt.CompareHashAndPassword([]byte(hash), []byte(pin)) == nil
}
