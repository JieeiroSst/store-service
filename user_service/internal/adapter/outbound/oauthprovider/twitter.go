package oauthprovider

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"golang.org/x/oauth2"
)

type Twitter struct {
	base
	cfg Config
}

func NewTwitter(cfg Config) *Twitter {
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://x.com/i/oauth2/authorize"
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "https://api.x.com/2"
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = cfg.APIURL + "/oauth2/token"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"users.read", "tweet.read", "users.email"}
	}
	return &Twitter{base: newBase(cfg, oauth2.AuthStyleInHeader, true), cfg: cfg}
}

func (t *Twitter) Exchange(ctx context.Context, code, verifier string) (domain.ExternalIdentity, error) {
	token, err := t.exchange(ctx, code, verifier)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	var me struct {
		Data struct {
			ID              string `json:"id"`
			Name            string `json:"name"`
			Username        string `json:"username"`
			ConfirmedEmail  string `json:"confirmed_email"`
			ProfileImageURL string `json:"profile_image_url"`
		} `json:"data"`
	}
	raw, err := t.getJSON(ctx, t.cfg.APIURL+"/users/me?user.fields=confirmed_email,name,username,profile_image_url", token.AccessToken, &me)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	var scopes []string
	if scope, ok := token.Extra("scope").(string); ok {
		scopes = strings.Fields(scope)
	}
	return domain.ExternalIdentity{
		Subject:       me.Data.ID,
		Email:         me.Data.ConfirmedEmail,
		EmailVerified: me.Data.ConfirmedEmail != "",
		Name:          me.Data.Name,
		Username:      me.Data.Username,
		AvatarURL:     me.Data.ProfileImageURL,
		Scopes:        scopes,
		Profile:       raw,
	}, nil
}

func (t *Twitter) VerifyToken(context.Context, string) (domain.ExternalIdentity, error) {
	return domain.ExternalIdentity{}, domain.ErrSSOTokenUnsupported
}
