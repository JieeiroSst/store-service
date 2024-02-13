package token

import (
	"crypto/sha256"
	"encoding/base64"
	"strings"
)

type ResponseType string

const (
	Code  ResponseType = "code"
	Token ResponseType = "token"
)

func (rt ResponseType) String() string {
	return string(rt)
}

type GrantType string

const (
	AuthorizationCode   GrantType = "authorization_code"
	PasswordCredentials GrantType = "password"
	ClientCredentials   GrantType = "client_credentials"
	Refreshing          GrantType = "refresh_token"
	Implicit            GrantType = "__implicit"
)

func (gt GrantType) String() string {
	if gt == AuthorizationCode ||
		gt == PasswordCredentials ||
		gt == ClientCredentials ||
		gt == Refreshing {
		return string(gt)
	}
	return ""
}

type CodeChallengeMethod string

const (
	CodeChallengePlain CodeChallengeMethod = "plain"
	CodeChallengeS256  CodeChallengeMethod = "S256"
)

func (ccm CodeChallengeMethod) String() string {
	if ccm == CodeChallengePlain ||
		ccm == CodeChallengeS256 {
		return string(ccm)
	}
	return ""
}

func (ccm CodeChallengeMethod) Validate(cc, ver string) bool {
	switch ccm {
	case CodeChallengePlain:
		return cc == ver
	case CodeChallengeS256:
		s256 := sha256.Sum256([]byte(ver))
		a := strings.TrimRight(base64.URLEncoding.EncodeToString(s256[:]), "=")
		b := strings.TrimRight(cc, "=")
		return a == b
	default:
		return false
	}
}
