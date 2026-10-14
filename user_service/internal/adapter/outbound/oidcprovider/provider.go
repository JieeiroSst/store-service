package oidcprovider

import (
	"context"
	"encoding/json"
	"errors"
	"slices"
	"strings"
	"sync"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const (
	GoogleIssuer             = "https://accounts.google.com"
	MicrosoftIssuerTemplate  = "https://login.microsoftonline.com/{tenantid}/v2.0"
	microsoftConsumersTenant = "9188040d-6c67-4c5b-b112-36a304b66dad"
)

type Claims struct {
	Email             string `json:"email"`
	EmailVerified     any    `json:"email_verified"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Picture           string `json:"picture"`
	TenantID          string `json:"tid"`
	EmailDomainOwner  any    `json:"xms_edov"`
}

type Config struct {
	Issuer         string
	ClientID       string
	ClientSecret   string
	RedirectURL    string
	Scopes         []string
	Audiences      []string
	PKCE           bool
	IssuerTemplate string
	VerifiedEmail  func(Claims) bool
}

type Provider struct {
	cfg      Config
	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(cfg Config) *Provider {
	if cfg.VerifiedEmail == nil {
		cfg.VerifiedEmail = func(c Claims) bool { return truthy(c.EmailVerified) }
	}
	return &Provider{cfg: cfg}
}

func MicrosoftIssuer(tenant string) (issuer, template string) {
	if tenant == "" {
		tenant = "common"
	}
	issuer = strings.Replace(MicrosoftIssuerTemplate, "{tenantid}", tenant, 1)
	switch tenant {
	case "common", "organizations", "consumers":
		return issuer, MicrosoftIssuerTemplate
	}
	return issuer, ""
}

func MicrosoftVerifiedEmail(c Claims) bool {
	return c.TenantID == microsoftConsumersTenant || truthy(c.EmailDomainOwner)
}

func (p *Provider) AuthCodeURL(ctx context.Context, state, verifier string) (string, error) {
	oauth, _, err := p.discover(ctx)
	if err != nil {
		return "", err
	}
	var opts []oauth2.AuthCodeOption
	if p.cfg.PKCE && verifier != "" {
		opts = append(opts, oauth2.S256ChallengeOption(verifier))
	}
	return oauth.AuthCodeURL(state, opts...), nil
}

func (p *Provider) Exchange(ctx context.Context, code, verifier string) (domain.ExternalIdentity, error) {
	oauth, verifierFn, err := p.discover(ctx)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	var opts []oauth2.AuthCodeOption
	if p.cfg.PKCE && verifier != "" {
		opts = append(opts, oauth2.VerifierOption(verifier))
	}
	token, err := oauth.Exchange(ctx, code, opts...)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return domain.ExternalIdentity{}, errors.New("token response has no id_token")
	}
	identity, err := p.identity(ctx, verifierFn, rawIDToken)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	if scope, ok := token.Extra("scope").(string); ok {
		identity.Scopes = strings.Fields(scope)
	}
	return identity, nil
}

func (p *Provider) VerifyToken(ctx context.Context, rawIDToken string) (domain.ExternalIdentity, error) {
	_, verifier, err := p.discover(ctx)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	return p.identity(ctx, verifier, rawIDToken)
}

func (p *Provider) identity(ctx context.Context, verifier *oidc.IDTokenVerifier, rawIDToken string) (domain.ExternalIdentity, error) {
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	if !slices.ContainsFunc(idToken.Audience, p.allowedAudience) {
		return domain.ExternalIdentity{}, errors.New("id_token audience is not allowed")
	}

	var claims Claims
	if err := idToken.Claims(&claims); err != nil {
		return domain.ExternalIdentity{}, err
	}
	var profile json.RawMessage
	if err := idToken.Claims(&profile); err != nil {
		return domain.ExternalIdentity{}, err
	}
	if p.cfg.IssuerTemplate != "" {
		want := strings.Replace(p.cfg.IssuerTemplate, "{tenantid}", claims.TenantID, 1)
		if claims.TenantID == "" || idToken.Issuer != want {
			return domain.ExternalIdentity{}, errors.New("id_token issuer does not match its tenant")
		}
	}
	return domain.ExternalIdentity{
		Subject:       idToken.Subject,
		Email:         claims.Email,
		EmailVerified: claims.Email != "" && p.cfg.VerifiedEmail(claims),
		Name:          claims.Name,
		Username:      claims.PreferredUsername,
		AvatarURL:     claims.Picture,
		Profile:       profile,
	}, nil
}

func (p *Provider) discover(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.oauth != nil {
		return p.oauth, p.verifier, nil
	}

	discoveryCtx := ctx
	if p.cfg.IssuerTemplate != "" {
		discoveryCtx = oidc.InsecureIssuerURLContext(ctx, p.cfg.IssuerTemplate)
	}
	provider, err := oidc.NewProvider(discoveryCtx, p.cfg.Issuer)
	if err != nil {
		return nil, nil, err
	}
	scopes := p.cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"profile", "email"}
	}
	if !slices.Contains(scopes, oidc.ScopeOpenID) {
		scopes = append([]string{oidc.ScopeOpenID}, scopes...)
	}

	p.oauth = &oauth2.Config{
		ClientID:     p.cfg.ClientID,
		ClientSecret: p.cfg.ClientSecret,
		RedirectURL:  p.cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       scopes,
	}
	p.verifier = provider.Verifier(&oidc.Config{SkipClientIDCheck: true, SkipIssuerCheck: p.cfg.IssuerTemplate != ""})
	return p.oauth, p.verifier, nil
}

func (p *Provider) allowedAudience(aud string) bool {
	return aud != "" && (aud == p.cfg.ClientID || slices.Contains(p.cfg.Audiences, aud))
}

func truthy(v any) bool {
	return v == true || v == "true" || v == "1"
}
