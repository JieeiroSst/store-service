package oauthprovider

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/url"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"golang.org/x/oauth2"
)

const facebookVersion = "v21.0"

type Facebook struct {
	base
	cfg Config
}

func NewFacebook(cfg Config) *Facebook {
	if cfg.AuthURL == "" {
		cfg.AuthURL = "https://www.facebook.com/" + facebookVersion + "/dialog/oauth"
	}
	if cfg.APIURL == "" {
		cfg.APIURL = "https://graph.facebook.com/" + facebookVersion
	}
	if cfg.TokenURL == "" {
		cfg.TokenURL = cfg.APIURL + "/oauth/access_token"
	}
	if len(cfg.Scopes) == 0 {
		cfg.Scopes = []string{"public_profile", "email"}
	}
	return &Facebook{base: newBase(cfg, oauth2.AuthStyleInParams, false), cfg: cfg}
}

func (f *Facebook) Exchange(ctx context.Context, code, verifier string) (domain.ExternalIdentity, error) {
	token, err := f.exchange(ctx, code, verifier)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	return f.VerifyToken(ctx, token.AccessToken)
}

func (f *Facebook) VerifyToken(ctx context.Context, token string) (domain.ExternalIdentity, error) {
	var debug struct {
		Data struct {
			AppID   string   `json:"app_id"`
			IsValid bool     `json:"is_valid"`
			UserID  string   `json:"user_id"`
			Scopes  []string `json:"scopes"`
		} `json:"data"`
	}
	q := url.Values{
		"input_token":  {token},
		"access_token": {f.cfg.ClientID + "|" + f.cfg.ClientSecret},
	}
	if _, err := f.getJSON(ctx, f.cfg.APIURL+"/debug_token?"+q.Encode(), "", &debug); err != nil {
		return domain.ExternalIdentity{}, err
	}
	if !debug.Data.IsValid || debug.Data.AppID != f.cfg.ClientID {
		return domain.ExternalIdentity{}, errors.New("facebook token was not issued to this app")
	}
	identity, err := f.me(ctx, token)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	if identity.Subject != debug.Data.UserID {
		return domain.ExternalIdentity{}, errors.New("facebook token user mismatch")
	}
	identity.Scopes = debug.Data.Scopes
	return identity, nil
}

func (f *Facebook) me(ctx context.Context, token string) (domain.ExternalIdentity, error) {
	mac := hmac.New(sha256.New, []byte(f.cfg.ClientSecret))
	mac.Write([]byte(token))
	q := url.Values{
		"fields":          {"id,name,email,picture.type(large)"},
		"access_token":    {token},
		"appsecret_proof": {hex.EncodeToString(mac.Sum(nil))},
	}
	var me struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Email   string `json:"email"`
		Picture struct {
			Data struct {
				URL string `json:"url"`
			} `json:"data"`
		} `json:"picture"`
	}
	raw, err := f.getJSON(ctx, f.cfg.APIURL+"/me?"+q.Encode(), "", &me)
	if err != nil {
		return domain.ExternalIdentity{}, err
	}
	return domain.ExternalIdentity{
		Subject:       me.ID,
		Email:         me.Email,
		EmailVerified: me.Email != "",
		Name:          me.Name,
		AvatarURL:     me.Picture.Data.URL,
		Profile:       raw,
	}, nil
}
