package refcode

import (
	"crypto/rand"
	"math/big"
)

const (
	charset = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	base    = int64(len(charset)) // 36
	length  = 8                   // 36^8 ≈ 2.8 trillion combinations
)

// Generate returns an 8-character base-36 code drawn from crypto/rand.
func Generate() (string, error) {
	max := new(big.Int).Exp(big.NewInt(base), big.NewInt(length), nil)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}

	buf := make([]byte, length)
	for i := length - 1; i >= 0; i-- {
		buf[i] = charset[new(big.Int).Mod(n, big.NewInt(base)).Int64()]
		n.Div(n, big.NewInt(base))
	}
	return string(buf), nil
}
