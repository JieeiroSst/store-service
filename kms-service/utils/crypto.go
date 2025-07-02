package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"

	"golang.org/x/crypto/pbkdf2"
)

var masterKey []byte

func InitMasterKey(keyPath string) error {
	if keyPath == "" {
		return errors.New("master key path is required")
	}

	if data, err := os.ReadFile(keyPath); err == nil {
		masterKey = data
		return nil
	}

	masterKey = make([]byte, 32)
	if _, err := rand.Read(masterKey); err != nil {
		return err
	}

	if err := os.WriteFile(keyPath, masterKey, 0600); err != nil {
		return err
	}

	return nil
}

func DeriveKey(password, salt []byte) []byte {
	return pbkdf2.Key(password, salt, 100000, 32, sha256.New)
}

func EncryptAES(plaintext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

func DecryptAES(ciphertext []byte, key []byte) ([]byte, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return aesGCM.Open(nil, nonce, ciphertext, nil)
}

func EncryptWithMasterKey(plainKey []byte) (string, error) {
	encrypted, err := EncryptAES(plainKey, masterKey)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(encrypted), nil
}

func DecryptWithMasterKey(encryptedHex string) ([]byte, error) {
	encrypted, err := hex.DecodeString(encryptedHex)
	if err != nil {
		return nil, err
	}
	return DecryptAES(encrypted, masterKey)
}

func GenerateSecureRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}
