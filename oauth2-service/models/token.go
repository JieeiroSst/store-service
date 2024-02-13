package models

import (
	"time"

	"github.com/JIeeiroSst/oauth2-service/pkg/token"
)

func NewToken() *Token {
	return &Token{}
}

type Token struct {
	ClientID            string        `bson:"ClientID"`
	UserID              string        `bson:"UserID"`
	RedirectURI         string        `bson:"RedirectURI"`
	Scope               string        `bson:"Scope"`
	Code                string        `bson:"Code"`
	CodeChallenge       string        `bson:"CodeChallenge"`
	CodeChallengeMethod string        `bson:"CodeChallengeMethod"`
	CodeCreateAt        time.Time     `bson:"CodeCreateAt"`
	CodeExpiresIn       time.Duration `bson:"CodeExpiresIn"`
	Access              string        `bson:"Access"`
	AccessCreateAt      time.Time     `bson:"AccessCreateAt"`
	AccessExpiresIn     time.Duration `bson:"AccessExpiresIn"`
	Refresh             string        `bson:"Refresh"`
	RefreshCreateAt     time.Time     `bson:"RefreshCreateAt"`
	RefreshExpiresIn    time.Duration `bson:"RefreshExpiresIn"`
}

func (t *Token) New() token.TokenInfo {
	return NewToken()
}

func (t *Token) GetClientID() string {
	return t.ClientID
}

func (t *Token) SetClientID(clientID string) {
	t.ClientID = clientID
}

func (t *Token) GetUserID() string {
	return t.UserID
}

func (t *Token) SetUserID(userID string) {
	t.UserID = userID
}

func (t *Token) GetRedirectURI() string {
	return t.RedirectURI
}

func (t *Token) SetRedirectURI(redirectURI string) {
	t.RedirectURI = redirectURI
}

func (t *Token) GetScope() string {
	return t.Scope
}

func (t *Token) SetScope(scope string) {
	t.Scope = scope
}

func (t *Token) GetCode() string {
	return t.Code
}

func (t *Token) SetCode(code string) {
	t.Code = code
}

func (t *Token) GetCodeCreateAt() time.Time {
	return t.CodeCreateAt
}

func (t *Token) SetCodeCreateAt(createAt time.Time) {
	t.CodeCreateAt = createAt
}

func (t *Token) GetCodeExpiresIn() time.Duration {
	return t.CodeExpiresIn
}

func (t *Token) SetCodeExpiresIn(exp time.Duration) {
	t.CodeExpiresIn = exp
}

func (t *Token) GetCodeChallenge() string {
	return t.CodeChallenge
}

func (t *Token) SetCodeChallenge(code string) {
	t.CodeChallenge = code
}

func (t *Token) GetCodeChallengeMethod() token.CodeChallengeMethod {
	return token.CodeChallengeMethod(t.CodeChallengeMethod)
}

func (t *Token) SetCodeChallengeMethod(method token.CodeChallengeMethod) {
	t.CodeChallengeMethod = string(method)
}

func (t *Token) GetAccess() string {
	return t.Access
}

func (t *Token) SetAccess(access string) {
	t.Access = access
}

func (t *Token) GetAccessCreateAt() time.Time {
	return t.AccessCreateAt
}

func (t *Token) SetAccessCreateAt(createAt time.Time) {
	t.AccessCreateAt = createAt
}

func (t *Token) GetAccessExpiresIn() time.Duration {
	return t.AccessExpiresIn
}

func (t *Token) SetAccessExpiresIn(exp time.Duration) {
	t.AccessExpiresIn = exp
}

func (t *Token) GetRefresh() string {
	return t.Refresh
}

func (t *Token) SetRefresh(refresh string) {
	t.Refresh = refresh
}

func (t *Token) GetRefreshCreateAt() time.Time {
	return t.RefreshCreateAt
}

func (t *Token) SetRefreshCreateAt(createAt time.Time) {
	t.RefreshCreateAt = createAt
}

func (t *Token) GetRefreshExpiresIn() time.Duration {
	return t.RefreshExpiresIn
}

func (t *Token) SetRefreshExpiresIn(exp time.Duration) {
	t.RefreshExpiresIn = exp
}
