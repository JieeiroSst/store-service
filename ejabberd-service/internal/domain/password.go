package domain

import (
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base64"
	"hash"
)

const (
	ScramDefaultIterationCount = 4096
	SaltLength                 = 16
)

type ScramHash string

const (
	ScramSHA1   ScramHash = "sha"
	ScramSHA256 ScramHash = "sha256"
	ScramSHA512 ScramHash = "sha512"
)

func (h ScramHash) New() func() hash.Hash {
	switch h {
	case ScramSHA256:
		return sha256.New
	case ScramSHA512:
		return sha512.New
	}
	return sha1.New
}

func (h ScramHash) Size() int {
	return h.New()().Size()
}

type PasswordFormat string

const (
	PasswordPlain PasswordFormat = "plain"
	PasswordScram PasswordFormat = "scram"
)

type Scram struct {
	StoredKey      string
	ServerKey      string
	Salt           string
	Hash           ScramHash
	IterationCount int
}

type Password struct {
	Plain string
	Scram *Scram
}

func (p Password) IsScram() bool { return p.Scram != nil }

func SaltedPassword(h ScramHash, password string, salt []byte, iter int) []byte {
	out, err := pbkdf2.Key(h.New(), password, salt, iter, h.Size())
	if err != nil {
		return nil
	}
	return out
}

func HMAC(h ScramHash, key, data []byte) []byte {
	m := hmac.New(h.New(), key)
	m.Write(data)
	return m.Sum(nil)
}

func ClientKey(h ScramHash, salted []byte) []byte { return HMAC(h, salted, []byte("Client Key")) }

func ServerKey(h ScramHash, salted []byte) []byte { return HMAC(h, salted, []byte("Server Key")) }

func StoredKey(h ScramHash, clientKey []byte) []byte {
	d := h.New()()
	d.Write(clientKey)
	return d.Sum(nil)
}

func PasswordToScram(h ScramHash, password string, iter int) (*Scram, error) {
	salt := make([]byte, SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return nil, err
	}
	return PasswordToScramWithSalt(h, password, salt, iter), nil
}

func PasswordToScramWithSalt(h ScramHash, password string, salt []byte, iter int) *Scram {
	salted := SaltedPassword(h, password, salt, iter)
	return &Scram{
		StoredKey:      base64.StdEncoding.EncodeToString(StoredKey(h, ClientKey(h, salted))),
		ServerKey:      base64.StdEncoding.EncodeToString(ServerKey(h, salted)),
		Salt:           base64.StdEncoding.EncodeToString(salt),
		Hash:           h,
		IterationCount: iter,
	}
}

func IsPasswordScramValid(password string, s *Scram) bool {
	if password == "" {
		return false
	}
	salt, err := base64.StdEncoding.DecodeString(s.Salt)
	if err != nil {
		return false
	}
	salted := SaltedPassword(s.Hash, password, salt, s.IterationCount)
	stored := base64.StdEncoding.EncodeToString(StoredKey(s.Hash, ClientKey(s.Hash, salted)))
	return subtle.ConstantTimeCompare([]byte(stored), []byte(s.StoredKey)) == 1
}

func MatchPassword(provided string, stored Password) bool {
	if stored.IsScram() {
		return IsPasswordScramValid(provided, stored.Scram)
	}
	return provided != "" && subtle.ConstantTimeCompare([]byte(provided), []byte(stored.Plain)) == 1
}
