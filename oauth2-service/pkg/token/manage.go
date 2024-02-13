package token

import (
	"context"
	"net/http"
	"time"
)

type TokenGenerateRequest struct {
	ClientID            string
	ClientSecret        string
	UserID              string
	RedirectURI         string
	Scope               string
	Code                string
	CodeChallenge       string
	CodeChallengeMethod CodeChallengeMethod
	Refresh             string
	CodeVerifier        string
	AccessTokenExp      time.Duration
	Request             *http.Request
}

type Manager interface {
	GetClient(ctx context.Context, clientID string) (cli ClientInfo, err error)

	GenerateAuthToken(ctx context.Context, rt ResponseType, tgr *TokenGenerateRequest) (authToken TokenInfo, err error)

	GenerateAccessToken(ctx context.Context, gt GrantType, tgr *TokenGenerateRequest) (accessToken TokenInfo, err error)

	RefreshAccessToken(ctx context.Context, tgr *TokenGenerateRequest) (accessToken TokenInfo, err error)

	RemoveAccessToken(ctx context.Context, access string) (err error)

	RemoveRefreshToken(ctx context.Context, refresh string) (err error)

	LoadAccessToken(ctx context.Context, access string) (ti TokenInfo, err error)

	LoadRefreshToken(ctx context.Context, refresh string) (ti TokenInfo, err error)
}
