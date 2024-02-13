package token

import (
	"time"

	"github.com/JIeeiroSst/oauth2-service/config"
)

type (
	ClientInfo interface {
		GetID() string
		GetSecret() string
		GetDomain() string
		IsPublic() bool
		GetUserID() string
	}

	ClientPasswordVerifier interface {
		VerifyPassword(string) bool
	}

	TokenInfo interface {
		New() TokenInfo

		GetClientID() string
		SetClientID(string)
		GetUserID() string
		SetUserID(string)
		GetRedirectURI() string
		SetRedirectURI(string)
		GetScope() string
		SetScope(string)

		GetCode() string
		SetCode(string)
		GetCodeCreateAt() time.Time
		SetCodeCreateAt(time.Time)
		GetCodeExpiresIn() time.Duration
		SetCodeExpiresIn(time.Duration)
		GetCodeChallenge() string
		SetCodeChallenge(string)
		GetCodeChallengeMethod() config.CodeChallengeMethod
		SetCodeChallengeMethod(config.CodeChallengeMethod)

		GetAccess() string
		SetAccess(string)
		GetAccessCreateAt() time.Time
		SetAccessCreateAt(time.Time)
		GetAccessExpiresIn() time.Duration
		SetAccessExpiresIn(time.Duration)

		GetRefresh() string
		SetRefresh(string)
		GetRefreshCreateAt() time.Time
		SetRefreshCreateAt(time.Time)
		GetRefreshExpiresIn() time.Duration
		SetRefreshExpiresIn(time.Duration)
	}
)
