package oidcprovider

import (
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

type Config struct {
	Issuer       string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type Provider struct {
	cfg      Config
	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *oidc.IDTokenVerifier
}

func New(cfg Config) *Provider {
	return &Provider{cfg: cfg}
}

func (p *Provider) AuthCodeURL(ctx context.Context, state string) (string, error) {
	oauth, _, err := p.discover(ctx)
	if err != nil {
		return "", err
	}
	return oauth.AuthCodeURL(state), nil
}

func (p *Provider) Exchange(ctx context.Context, code string) (domain.ExternalIdentity, error) {
	oauth, verifier, err := p.discover(ctx)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	token, err := oauth.Exchange(ctx, code)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok || rawIDToken == "" {
		return domain.ExternalIdentity{}, errors.New("token response has no id_token")
	}
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}

	var claims struct {
		Email             string `json:"email"`
		EmailVerified     any    `json:"email_verified"`
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return domain.ExternalIdentity{}, err
	}
	return domain.ExternalIdentity{
		Subject:       idToken.Subject,
		Email:         claims.Email,
		EmailVerified: claims.EmailVerified == true || claims.EmailVerified == "true",
		Name:          claims.Name,
		Username:      claims.PreferredUsername,
	}, nil
}

func (p *Provider) discover(ctx context.Context) (*oauth2.Config, *oidc.IDTokenVerifier, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.oauth != nil {
		return p.oauth, p.verifier, nil
	}

	provider, err := oidc.NewProvider(ctx, p.cfg.Issuer)
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
	p.verifier = provider.Verifier(&oidc.Config{ClientID: p.cfg.ClientID})
	return p.oauth, p.verifier, nil
}
