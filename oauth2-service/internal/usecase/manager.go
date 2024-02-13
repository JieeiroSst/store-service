package usecase

import (
	"context"
	"time"

	"github.com/JIeeiroSst/oauth2-service/config"
	"github.com/JIeeiroSst/oauth2-service/internal/repository"
	"github.com/JIeeiroSst/oauth2-service/models"
	"github.com/JIeeiroSst/oauth2-service/pkg/token"
	"github.com/JIeeiroSst/oauth2-service/utils"
)

func NewDefaultManager() *Manager {
	m := NewManager()
	m.MapAuthorizeGenerate(token.NewAuthorizeGenerate())
	m.MapAccessGenerate(token.NewAccessGenerate())

	return m
}

func NewManager() *Manager {
	return &Manager{
		gtcfg:       make(map[config.GrantType]*config.ConfigToken),
		validateURI: utils.DefaultValidateURI,
	}
}

type Manager struct {
	codeExp           time.Duration
	gtcfg             map[config.GrantType]*config.ConfigToken
	rcfg              *config.RefreshingConfig
	validateURI       utils.ValidateURIHandler
	authorizeGenerate token.IAuthorizeGenerate
	accessGenerate    token.IAccessGenerate
	tokenStore        repository.ITokensStore
	clientStore       repository.IClientStore
}

func (m *Manager) grantConfig(gt config.GrantType) *config.ConfigToken {
	if c, ok := m.gtcfg[gt]; ok && c != nil {
		return c
	}
	switch gt {
	case config.AuthorizationCode:
		return config.DefaultAuthorizeCodeTokenCfg
	case config.Implicit:
		return config.DefaultImplicitTokenCfg
	case config.PasswordCredentials:
		return config.DefaultPasswordTokenCfg
	case config.ClientCredentials:
		return config.DefaultClientTokenCfg
	}
	return &config.ConfigToken{}
}

func (m *Manager) SetAuthorizeCodeExp(exp time.Duration) {
	m.codeExp = exp
}

func (m *Manager) SetAuthorizeCodeTokenCfg(cfg *config.ConfigToken) {
	m.gtcfg[config.AuthorizationCode] = cfg
}

func (m *Manager) SetImplicitTokenCfg(cfg *config.ConfigToken) {
	m.gtcfg[config.Implicit] = cfg
}

func (m *Manager) SetPasswordTokenCfg(cfg *config.ConfigToken) {
	m.gtcfg[config.PasswordCredentials] = cfg
}

func (m *Manager) SetClientTokenCfg(cfg *config.ConfigToken) {
	m.gtcfg[config.ClientCredentials] = cfg
}

func (m *Manager) SetRefreshTokenCfg(cfg *config.RefreshingConfig) {
	m.rcfg = cfg
}

func (m *Manager) SetValidateURIHandler(handler utils.ValidateURIHandler) {
	m.validateURI = handler
}

func (m *Manager) MapAuthorizeGenerate(gen token.IAuthorizeGenerate) {
	m.authorizeGenerate = gen
}

func (m *Manager) MapAccessGenerate(gen token.IAccessGenerate) {
	m.accessGenerate = gen
}

func (m *Manager) MapClientStorage(stor repository.IClientStore) {
	m.clientStore = stor
}

func (m *Manager) MustClientStorage(stor repository.IClientStore, err error) {
	if err != nil {
		panic(err.Error())
	}
	m.clientStore = stor
}

func (m *Manager) MapTokenStorage(stor repository.ITokensStore) {
	m.tokenStore = stor
}

func (m *Manager) MustTokenStorage(stor repository.ITokensStore, err error) {
	if err != nil {
		panic(err)
	}
	m.tokenStore = stor
}

func (m *Manager) GetClient(ctx context.Context, clientID string) (cli token.ClientInfo, err error) {
	cli, err = m.clientStore.GetByID(ctx, clientID)
	if err != nil {
		return
	} else if cli == nil {
		err = config.ErrInvalidClient
	}
	return
}

func (m *Manager) GenerateAuthToken(ctx context.Context, rt config.ResponseType, tgr *token.TokenGenerateRequest) (token.TokenInfo, error) {
	cli, err := m.GetClient(ctx, tgr.ClientID)
	if err != nil {
		return nil, err
	} else if tgr.RedirectURI != "" {
		if err := m.validateURI(cli.GetDomain(), tgr.RedirectURI); err != nil {
			return nil, err
		}
	}

	ti := models.NewToken()
	ti.SetClientID(tgr.ClientID)
	ti.SetUserID(tgr.UserID)
	ti.SetRedirectURI(tgr.RedirectURI)
	ti.SetScope(tgr.Scope)

	createAt := time.Now()
	td := &token.GenerateBasic{
		Client:    cli,
		UserID:    tgr.UserID,
		CreateAt:  createAt,
		TokenInfo: ti,
		Request:   tgr.Request,
	}
	switch rt {
	case config.Code:
		codeExp := m.codeExp
		if codeExp == 0 {
			codeExp = config.DefaultCodeExp
		}
		ti.SetCodeCreateAt(createAt)
		ti.SetCodeExpiresIn(codeExp)
		if exp := tgr.AccessTokenExp; exp > 0 {
			ti.SetAccessExpiresIn(exp)
		}
		if tgr.CodeChallenge != "" {
			ti.SetCodeChallenge(tgr.CodeChallenge)
			ti.SetCodeChallengeMethod(tgr.CodeChallengeMethod)
		}

		tv, err := m.authorizeGenerate.Token(ctx, td)
		if err != nil {
			return nil, err
		}
		ti.SetCode(tv)
	case config.Token:
		icfg := m.grantConfig(config.Implicit)
		aexp := icfg.AccessTokenExp
		if exp := tgr.AccessTokenExp; exp > 0 {
			aexp = exp
		}
		ti.SetAccessCreateAt(createAt)
		ti.SetAccessExpiresIn(aexp)

		if icfg.IsGenerateRefresh {
			ti.SetRefreshCreateAt(createAt)
			ti.SetRefreshExpiresIn(icfg.RefreshTokenExp)
		}

		tv, rv, err := m.accessGenerate.Token(ctx, td, icfg.IsGenerateRefresh)
		if err != nil {
			return nil, err
		}
		ti.SetAccess(tv)

		if rv != "" {
			ti.SetRefresh(rv)
		}
	}

	err = m.tokenStore.Create(ctx, ti)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func (m *Manager) getAuthorizationCode(ctx context.Context, code string) (token.TokenInfo, error) {
	ti, err := m.tokenStore.GetByCode(ctx, code)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetCode() != code || ti.GetCodeCreateAt().Add(ti.GetCodeExpiresIn()).Before(time.Now()) {
		err = config.ErrInvalidAuthorizeCode
		return nil, config.ErrInvalidAuthorizeCode
	}
	return ti, nil
}

func (m *Manager) delAuthorizationCode(ctx context.Context, code string) error {
	return m.tokenStore.RemoveByCode(ctx, code)
}

func (m *Manager) getAndDelAuthorizationCode(ctx context.Context, tgr *token.TokenGenerateRequest) (token.TokenInfo, error) {
	code := tgr.Code
	ti, err := m.getAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	} else if ti.GetClientID() != tgr.ClientID {
		return nil, config.ErrInvalidAuthorizeCode
	} else if codeURI := ti.GetRedirectURI(); codeURI != "" && codeURI != tgr.RedirectURI {
		return nil, config.ErrInvalidAuthorizeCode
	}

	err = m.delAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return ti, nil
}

func (m *Manager) validateCodeChallenge(ti token.TokenInfo, ver string) error {
	cc := ti.GetCodeChallenge()
	if cc == "" && ver == "" {
		return nil
	}
	if cc == "" {
		return config.ErrMissingCodeVerifier
	}
	if ver == "" {
		return config.ErrMissingCodeVerifier
	}
	ccm := ti.GetCodeChallengeMethod()
	if ccm.String() == "" {
		ccm = config.CodeChallengePlain
	}
	if !ccm.Validate(cc, ver) {
		return config.ErrInvalidCodeChallenge
	}
	return nil
}

func (m *Manager) GenerateAccessToken(ctx context.Context, gt config.GrantType, tgr *token.TokenGenerateRequest) (token.TokenInfo, error) {
	cli, err := m.GetClient(ctx, tgr.ClientID)
	if err != nil {
		return nil, err
	}
	if cliPass, ok := cli.(token.ClientPasswordVerifier); ok {
		if !cliPass.VerifyPassword(tgr.ClientSecret) {
			return nil, config.ErrInvalidClient
		}
	} else if len(cli.GetSecret()) > 0 && tgr.ClientSecret != cli.GetSecret() {
		return nil, config.ErrInvalidClient
	}
	if tgr.RedirectURI != "" {
		if err := m.validateURI(cli.GetDomain(), tgr.RedirectURI); err != nil {
			return nil, err
		}
	}

	if gt == config.ClientCredentials && cli.IsPublic() == true {
		return nil, config.ErrInvalidClient
	}

	if gt == config.AuthorizationCode {
		ti, err := m.getAndDelAuthorizationCode(ctx, tgr)
		if err != nil {
			return nil, err
		}
		if err := m.validateCodeChallenge(ti, tgr.CodeVerifier); err != nil {
			return nil, err
		}
		tgr.UserID = ti.GetUserID()
		tgr.Scope = ti.GetScope()
		if exp := ti.GetAccessExpiresIn(); exp > 0 {
			tgr.AccessTokenExp = exp
		}
	}

	ti := models.NewToken()
	ti.SetClientID(tgr.ClientID)
	ti.SetUserID(tgr.UserID)
	ti.SetRedirectURI(tgr.RedirectURI)
	ti.SetScope(tgr.Scope)

	createAt := time.Now()
	ti.SetAccessCreateAt(createAt)

	gcfg := m.grantConfig(gt)
	aexp := gcfg.AccessTokenExp
	if exp := tgr.AccessTokenExp; exp > 0 {
		aexp = exp
	}
	ti.SetAccessExpiresIn(aexp)
	if gcfg.IsGenerateRefresh {
		ti.SetRefreshCreateAt(createAt)
		ti.SetRefreshExpiresIn(gcfg.RefreshTokenExp)
	}

	td := &token.GenerateBasic{
		Client:    cli,
		UserID:    tgr.UserID,
		CreateAt:  createAt,
		TokenInfo: ti,
		Request:   tgr.Request,
	}

	av, rv, err := m.accessGenerate.Token(ctx, td, gcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}
	ti.SetAccess(av)

	if rv != "" {
		ti.SetRefresh(rv)
	}

	err = m.tokenStore.Create(ctx, ti)
	if err != nil {
		return nil, err
	}

	return ti, nil
}

func (m *Manager) RefreshAccessToken(ctx context.Context, tgr *token.TokenGenerateRequest) (token.TokenInfo, error) {
	ti, err := m.LoadRefreshToken(ctx, tgr.Refresh)
	if err != nil {
		return nil, err
	}

	cli, err := m.GetClient(ctx, ti.GetClientID())
	if err != nil {
		return nil, err
	}

	oldAccess, oldRefresh := ti.GetAccess(), ti.GetRefresh()

	td := &token.GenerateBasic{
		Client:    cli,
		UserID:    ti.GetUserID(),
		CreateAt:  time.Now(),
		TokenInfo: ti,
		Request:   tgr.Request,
	}

	rcfg := config.DefaultRefreshTokenCfg
	if v := m.rcfg; v != nil {
		rcfg = v
	}

	ti.SetAccessCreateAt(td.CreateAt)
	if v := rcfg.AccessTokenExp; v > 0 {
		ti.SetAccessExpiresIn(v)
	}

	if v := rcfg.RefreshTokenExp; v > 0 {
		ti.SetRefreshExpiresIn(v)
	}

	if rcfg.IsResetRefreshTime {
		ti.SetRefreshCreateAt(td.CreateAt)
	}

	if scope := tgr.Scope; scope != "" {
		ti.SetScope(scope)
	}

	tv, rv, err := m.accessGenerate.Token(ctx, td, rcfg.IsGenerateRefresh)
	if err != nil {
		return nil, err
	}

	ti.SetAccess(tv)
	if rv != "" {
		ti.SetRefresh(rv)
	}

	if err := m.tokenStore.Create(ctx, ti); err != nil {
		return nil, err
	}

	if rcfg.IsRemoveAccess {
		// remove the old access token
		if err := m.tokenStore.RemoveByAccess(ctx, oldAccess); err != nil {
			return nil, err
		}
	}

	if rcfg.IsRemoveRefreshing && rv != "" {
		// remove the old refresh token
		if err := m.tokenStore.RemoveByRefresh(ctx, oldRefresh); err != nil {
			return nil, err
		}
	}

	if rv == "" {
		ti.SetRefresh("")
		ti.SetRefreshCreateAt(time.Now())
		ti.SetRefreshExpiresIn(0)
	}

	return ti, nil
}

func (m *Manager) RemoveAccessToken(ctx context.Context, access string) error {
	if access == "" {
		return config.ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByAccess(ctx, access)
}

func (m *Manager) RemoveRefreshToken(ctx context.Context, refresh string) error {
	if refresh == "" {
		return config.ErrInvalidAccessToken
	}
	return m.tokenStore.RemoveByRefresh(ctx, refresh)
}

func (m *Manager) LoadAccessToken(ctx context.Context, access string) (token.TokenInfo, error) {
	if access == "" {
		return nil, config.ErrInvalidAccessToken
	}

	ct := time.Now()
	ti, err := m.tokenStore.GetByAccess(ctx, access)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetAccess() != access {
		return nil, config.ErrInvalidAccessToken
	} else if ti.GetRefresh() != "" && ti.GetRefreshExpiresIn() != 0 &&
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(ct) {
		return nil, config.ErrExpiredRefreshToken
	} else if ti.GetAccessExpiresIn() != 0 &&
		ti.GetAccessCreateAt().Add(ti.GetAccessExpiresIn()).Before(ct) {
		return nil, config.ErrExpiredAccessToken
	}
	return ti, nil
}

func (m *Manager) LoadRefreshToken(ctx context.Context, refresh string) (token.TokenInfo, error) {
	if refresh == "" {
		return nil, config.ErrInvalidRefreshToken
	}

	ti, err := m.tokenStore.GetByRefresh(ctx, refresh)
	if err != nil {
		return nil, err
	} else if ti == nil || ti.GetRefresh() != refresh {
		return nil, config.ErrInvalidRefreshToken
	} else if ti.GetRefreshExpiresIn() != 0 && 
		ti.GetRefreshCreateAt().Add(ti.GetRefreshExpiresIn()).Before(time.Now()) {
		return nil, config.ErrExpiredRefreshToken
	}
	return ti, nil
}
