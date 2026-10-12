package domain

import (
	"bytes"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strconv"
	"strings"
)

var (
	ErrSASLNotAuthorized   = errors.New("not-authorized")
	ErrSASLMalformed       = errors.New("malformed-request")
	ErrSASLBadProtocol     = errors.New("bad-protocol")
	ErrSASLInvalidMech     = errors.New("invalid-mechanism")
	ErrSASLAccountDisabled = errors.New("account-disabled")
)

type SASLCredentials interface {
	GetPassword(user string) (Password, bool)
	CheckPassword(user, authzid, password string) bool
}

type SASLStep struct {
	Challenge []byte
	Success   bool
	Data      []byte
	User      string
}

type SASLMechanism interface {
	Step(clientIn []byte) (SASLStep, error)
}

func NewSASLMechanism(name string, creds SASLCredentials) (SASLMechanism, error) {
	switch name {
	case "PLAIN":
		return &plainMech{creds: creds}, nil
	case "SCRAM-SHA-1":
		return &scramMech{creds: creds, hash: ScramSHA1}, nil
	case "SCRAM-SHA-256":
		return &scramMech{creds: creds, hash: ScramSHA256}, nil
	case "SCRAM-SHA-512":
		return &scramMech{creds: creds, hash: ScramSHA512}, nil
	}
	return nil, ErrSASLInvalidMech
}

func SASLMechanisms(format PasswordFormat, h ScramHash) []string {
	plain := format == PasswordPlain
	var out []string
	if plain || h == ScramSHA1 {
		out = append(out, "SCRAM-SHA-1")
	}
	if plain || h == ScramSHA256 {
		out = append(out, "SCRAM-SHA-256")
	}
	if plain || h == ScramSHA512 {
		out = append(out, "SCRAM-SHA-512")
	}
	return append(out, "PLAIN")
}

type plainMech struct {
	creds SASLCredentials
}

func (m *plainMech) Step(in []byte) (SASLStep, error) {
	parts := bytes.Split(in, []byte{0})
	if len(parts) != 3 {
		return SASLStep{}, ErrSASLMalformed
	}
	authzid, user, pass := string(parts[0]), string(parts[1]), string(parts[2])
	if authzid != "" {
		if j, err := ParseJID(authzid); err == nil && j.LUser != "" {
			authzid = j.User
		}
	}
	if user == "" {
		return SASLStep{}, ErrSASLNotAuthorized
	}
	if strings.Contains(user, "@") {
		if j, err := ParseJID(user); err == nil {
			user = j.User
		}
	}
	if authzid == "" {
		authzid = user
	}
	if !m.creds.CheckPassword(user, authzid, pass) {
		return SASLStep{User: user}, ErrSASLNotAuthorized
	}
	return SASLStep{Success: true, User: user}, nil
}

type scramMech struct {
	creds       SASLCredentials
	hash        ScramHash
	step        int
	user        string
	clientFirst string
	serverFirst string
	nonce       string
	gs2Header   string
	storedKey   []byte
	serverKey   []byte
}

func (m *scramMech) Step(in []byte) (SASLStep, error) {
	m.step++
	switch m.step {
	case 1:
		return m.first(string(in))
	case 2:
		return m.final(string(in))
	}
	return SASLStep{}, ErrSASLBadProtocol
}

func (m *scramMech) first(in string) (SASLStep, error) {
	parts := strings.SplitN(in, ",", 3)
	if len(parts) != 3 {
		return SASLStep{}, ErrSASLMalformed
	}
	switch {
	case parts[0] == "n" || parts[0] == "y":
	case strings.HasPrefix(parts[0], "p="):
		return SASLStep{}, ErrSASLBadProtocol
	default:
		return SASLStep{}, ErrSASLMalformed
	}
	m.gs2Header = parts[0] + "," + parts[1] + ","
	m.clientFirst = parts[2]
	attrs := parseScramAttrs(m.clientFirst)
	user, ok := attrs["n"]
	if !ok || user == "" {
		return SASLStep{}, ErrSASLMalformed
	}
	user = strings.NewReplacer("=2C", ",", "=3D", "=").Replace(user)
	cnonce := attrs["r"]
	if cnonce == "" {
		return SASLStep{}, ErrSASLMalformed
	}
	m.user = user
	pass, ok := m.creds.GetPassword(user)
	if !ok {
		return SASLStep{User: user}, ErrSASLNotAuthorized
	}
	var salt []byte
	var iter int
	if pass.IsScram() {
		if pass.Scram.Hash != m.hash {
			return SASLStep{User: user}, ErrSASLNotAuthorized
		}
		var err error
		if salt, err = base64.StdEncoding.DecodeString(pass.Scram.Salt); err != nil {
			return SASLStep{User: user}, ErrSASLNotAuthorized
		}
		if m.storedKey, err = base64.StdEncoding.DecodeString(pass.Scram.StoredKey); err != nil {
			return SASLStep{User: user}, ErrSASLNotAuthorized
		}
		if m.serverKey, err = base64.StdEncoding.DecodeString(pass.Scram.ServerKey); err != nil {
			return SASLStep{User: user}, ErrSASLNotAuthorized
		}
		iter = pass.Scram.IterationCount
	} else {
		salt = make([]byte, SaltLength)
		if _, err := rand.Read(salt); err != nil {
			return SASLStep{}, err
		}
		iter = ScramDefaultIterationCount
		salted := SaltedPassword(m.hash, pass.Plain, salt, iter)
		m.storedKey = StoredKey(m.hash, ClientKey(m.hash, salted))
		m.serverKey = ServerKey(m.hash, salted)
	}
	snonce := make([]byte, 18)
	if _, err := rand.Read(snonce); err != nil {
		return SASLStep{}, err
	}
	m.nonce = cnonce + base64.StdEncoding.EncodeToString(snonce)
	m.serverFirst = "r=" + m.nonce + ",s=" + base64.StdEncoding.EncodeToString(salt) + ",i=" + strconv.Itoa(iter)
	return SASLStep{Challenge: []byte(m.serverFirst), User: user}, nil
}

func (m *scramMech) final(in string) (SASLStep, error) {
	idx := strings.LastIndex(in, ",p=")
	if idx < 0 {
		return SASLStep{}, ErrSASLMalformed
	}
	withoutProof := in[:idx]
	attrs := parseScramAttrs(in)
	cb, err := base64.StdEncoding.DecodeString(attrs["c"])
	if err != nil || string(cb) != m.gs2Header {
		return SASLStep{User: m.user}, ErrSASLBadProtocol
	}
	if attrs["r"] != m.nonce {
		return SASLStep{User: m.user}, ErrSASLBadProtocol
	}
	proof, err := base64.StdEncoding.DecodeString(attrs["p"])
	if err != nil || len(proof) != len(m.storedKey) {
		return SASLStep{User: m.user}, ErrSASLNotAuthorized
	}
	authMessage := []byte(m.clientFirst + "," + m.serverFirst + "," + withoutProof)
	clientSig := HMAC(m.hash, m.storedKey, authMessage)
	clientKey := make([]byte, len(proof))
	for i := range proof {
		clientKey[i] = proof[i] ^ clientSig[i]
	}
	if subtle.ConstantTimeCompare(StoredKey(m.hash, clientKey), m.storedKey) != 1 {
		return SASLStep{User: m.user}, ErrSASLNotAuthorized
	}
	serverSig := HMAC(m.hash, m.serverKey, authMessage)
	return SASLStep{Success: true, User: m.user, Data: []byte("v=" + base64.StdEncoding.EncodeToString(serverSig))}, nil
}

func parseScramAttrs(s string) map[string]string {
	out := map[string]string{}
	for _, kv := range strings.Split(s, ",") {
		if len(kv) >= 2 && kv[1] == '=' {
			out[kv[:1]] = kv[2:]
		}
	}
	return out
}
