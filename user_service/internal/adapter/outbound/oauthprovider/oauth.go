package oauthprovider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/oauth2"
)

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AuthURL      string
	TokenURL     string
	APIURL       string
}

type base struct {
	oauth  *oauth2.Config
	pkce   bool
	client *http.Client
}

func newBase(cfg Config, style oauth2.AuthStyle, pkce bool) base {
	return base{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Scopes:       cfg.Scopes,
			Endpoint:     oauth2.Endpoint{AuthURL: cfg.AuthURL, TokenURL: cfg.TokenURL, AuthStyle: style},
		},
		pkce:   pkce,
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (b *base) AuthCodeURL(_ context.Context, state, verifier string) (string, error) {
	if !b.pkce {
		return b.oauth.AuthCodeURL(state), nil
	}
	if verifier == "" {
		return "", fmt.Errorf("pkce verifier is required")
	}
	return b.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), nil
}

func (b *base) exchange(ctx context.Context, code, verifier string) (*oauth2.Token, error) {
	var opts []oauth2.AuthCodeOption
	if b.pkce {
		if verifier == "" {
			return nil, fmt.Errorf("pkce verifier is required")
		}
		opts = append(opts, oauth2.VerifierOption(verifier))
	}
	return b.oauth.Exchange(context.WithValue(ctx, oauth2.HTTPClient, b.client), code, opts...)
}

func (b *base) getJSON(ctx context.Context, url, bearer string, out any) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	resp, err := b.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %d: %s", req.URL.Path, resp.StatusCode, body)
	}
	return body, json.Unmarshal(body, out)
}
