package mfa

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
)

var errMalformed = errors.New("secretbox: malformed ciphertext")

type SecretBox struct {
	aead cipher.AEAD
}

func NewSecretBox(key string) (*SecretBox, error) {
	if key == "" {
		return nil, errors.New("secretbox: empty key")
	}
	sum := sha256.Sum256([]byte(key))
	block, err := aes.NewCipher(sum[:])
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretBox{aead: aead}, nil
}

func (b *SecretBox) Seal(plaintext []byte) (string, error) {
	nonce := make([]byte, b.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(b.aead.Seal(nonce, nonce, plaintext, nil)), nil
}

func (b *SecretBox) Open(sealed string) ([]byte, error) {
	raw, err := base64.RawStdEncoding.DecodeString(sealed)
	if err != nil || len(raw) < b.aead.NonceSize() {
		return nil, errMalformed
	}
	n := b.aead.NonceSize()
	return b.aead.Open(nil, raw[:n], raw[n:], nil)
}
