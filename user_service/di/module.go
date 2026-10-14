package di

import (
	"context"

	"github.com/JIeeiroSst/user-service/config"
	"github.com/JIeeiroSst/user-service/internal/adapter/inbound/grpcadapter"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/authorizeclient"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/jwttoken"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/oauthprovider"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/oidcprovider"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/pg"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/pwhash"
	"github.com/JIeeiroSst/user-service/internal/adapter/outbound/sessionstore"
	"github.com/JIeeiroSst/user-service/internal/application/auth"
	"github.com/JIeeiroSst/user-service/internal/application/role"
	"github.com/JIeeiroSst/user-service/internal/application/roleitem"
	"github.com/JIeeiroSst/user-service/internal/application/session"
	"github.com/JIeeiroSst/user-service/internal/application/sso"
	"github.com/JIeeiroSst/user-service/internal/application/user"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/input"
	"github.com/JIeeiroSst/user-service/internal/port/output"
	"github.com/JIeeiroSst/utils/cache/expire"
	pgconn "github.com/JIeeiroSst/utils/postgres"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const ecosystem = ".env"

func loadConfig() (*config.Config, error) {
	return config.InitializeConfiguration(ecosystem)
}

func newDB(cfg *config.Config) (*gorm.DB, error) {
	db := pgconn.NewPostgresConn(pgconn.PostgresConfig{
		PostgresqlHost:     cfg.Postgres.PostgresqlHost,
		PostgresqlPort:     cfg.Postgres.PostgresqlPort,
		PostgresqlUser:     cfg.Postgres.PostgresqlUser,
		PostgresqlPassword: cfg.Postgres.PostgresqlPassword,
		PostgresqlDbname:   cfg.Postgres.PostgresqlDbname,
		PostgresqlSSLMode:  cfg.Postgres.PostgresqlSSLMode,
	})
	if err := db.AutoMigrate(&domain.User{}, &domain.Role{}, &domain.LoginSession{}, &domain.SocialLogin{}); err != nil {
		return nil, err
	}
	return db, nil
}

func newCache(cfg *config.Config) expire.CacheHelper {
	return expire.NewCacheHelper(cfg.Redis.Dns)
}

func newUserRepository(db *gorm.DB) output.UserRepository { return pg.NewUserRepository(db) }

func newRoleRepository(db *gorm.DB) output.RoleRepository { return pg.NewRoleRepository(db) }

func newSessionRepository(db *gorm.DB) output.SessionRepository { return pg.NewSessionRepository(db) }

func newSocialLoginRepository(db *gorm.DB) output.SocialLoginRepository {
	return pg.NewSocialLoginRepository(db)
}

func newAuthorizer(lc fx.Lifecycle, cfg *config.Config) (output.Authorizer, error) {
	client, err := authorizeclient.New(cfg.AuthorizeService.GrpcAddress, cfg.AuthorizeService.Timeout())
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error { return client.Close() }})
	return client, nil
}

func newHasher() output.Hasher { return pwhash.New() }

func newTokenGenerator(cfg *config.Config) output.TokenGenerator {
	return jwttoken.New(cfg.Secret.JwtSecretKey, cfg.Token.AccessTokenTTL())
}

func newTokenStore(cache expire.CacheHelper) output.TokenStore {
	return sessionstore.New(cache)
}

func newSessionService(cfg *config.Config, repo output.SessionRepository, store output.TokenStore, tokenGen output.TokenGenerator, authorizer output.Authorizer) *session.Service {
	return session.New(repo, store, tokenGen, authorizer, session.Config{
		AccessTTL:  cfg.Token.AccessTokenTTL(),
		RefreshTTL: cfg.Token.RefreshTokenTTL(),
		MaxPerUser: cfg.Token.MaxSessionsPerUser,
	})
}

func asSessionService(s *session.Service) input.SessionService { return s }

func newAuthService(userRepo output.UserRepository, hasher output.Hasher, tokenGen output.TokenGenerator, tokenStore output.TokenStore, sessions *session.Service) *auth.Service {
	return auth.New(userRepo, hasher, tokenGen, tokenStore, sessions)
}

func asAuthService(s *auth.Service) input.AuthService { return s }

type ssoDeps struct {
	fx.In
	UserRepo     output.UserRepository
	SocialLogins output.SocialLoginRepository
	Hasher       output.Hasher
	Authorizer   output.Authorizer
	Sessions     *session.Service
}

func (d ssoDeps) service(provider string, idp output.IdentityProvider) input.SSOService {
	return sso.New(provider, idp, d.UserRepo, d.SocialLogins, d.Hasher, d.Authorizer, d.Sessions)
}

func newSSOService(cfg *config.Config, deps ssoDeps) input.SSOService {
	if !cfg.SSO.Enabled {
		return nil
	}
	return deps.service(domain.ProviderSSO, oidcprovider.New(oidcprovider.Config{
		Issuer:       cfg.SSO.Issuer,
		ClientID:     cfg.SSO.ClientID,
		ClientSecret: cfg.SSO.ClientSecret,
		RedirectURL:  cfg.SSO.RedirectURL,
		Scopes:       cfg.SSO.Scopes,
	}))
}

func newSocialLogins(cfg *config.Config, deps ssoDeps) []SocialLogin {
	var out []SocialLogin
	add := func(name string, c config.OAuthProviderConfig, idp func(config.OAuthProviderConfig) output.IdentityProvider) {
		if c.Enabled {
			out = append(out, SocialLogin{Name: name, Service: deps.service(name, idp(c))})
		}
	}
	add(domain.ProviderGoogle, cfg.Social.Google, func(c config.OAuthProviderConfig) output.IdentityProvider {
		return oidcprovider.New(oidcprovider.Config{
			Issuer:       oidcprovider.GoogleIssuer,
			ClientID:     c.ClientID,
			ClientSecret: c.ClientSecret,
			RedirectURL:  c.RedirectURL,
			Scopes:       c.Scopes,
			Audiences:    c.ExtraAudiences,
			PKCE:         true,
		})
	})
	add(domain.ProviderMicrosoft, cfg.Social.Microsoft, func(c config.OAuthProviderConfig) output.IdentityProvider {
		issuer, template := oidcprovider.MicrosoftIssuer(c.Tenant)
		return oidcprovider.New(oidcprovider.Config{
			Issuer:         issuer,
			IssuerTemplate: template,
			ClientID:       c.ClientID,
			ClientSecret:   c.ClientSecret,
			RedirectURL:    c.RedirectURL,
			Scopes:         c.Scopes,
			Audiences:      c.ExtraAudiences,
			PKCE:           true,
			VerifiedEmail:  oidcprovider.MicrosoftVerifiedEmail,
		})
	})
	add(domain.ProviderFacebook, cfg.Social.Facebook, func(c config.OAuthProviderConfig) output.IdentityProvider {
		return oauthprovider.NewFacebook(oauthConfig(c))
	})
	add(domain.ProviderTwitter, cfg.Social.Twitter, func(c config.OAuthProviderConfig) output.IdentityProvider {
		return oauthprovider.NewTwitter(oauthConfig(c))
	})
	return out
}

func oauthConfig(c config.OAuthProviderConfig) oauthprovider.Config {
	return oauthprovider.Config{ClientID: c.ClientID, ClientSecret: c.ClientSecret, RedirectURL: c.RedirectURL, Scopes: c.Scopes}
}

func newUserService(userRepo output.UserRepository, hasher output.Hasher, cache expire.CacheHelper, authorizer output.Authorizer, sessions *session.Service) input.UserService {
	return user.New(userRepo, hasher, cache, authorizer, sessions)
}

func newRoleService(roleRepo output.RoleRepository, authorizer output.Authorizer) input.RoleService {
	return role.New(roleRepo, authorizer)
}

func newRoleItemService(roleRepo output.RoleRepository, authorizer output.Authorizer) input.RoleItemService {
	return roleitem.New(roleRepo, authorizer)
}

func newGRPCHandler(authSvc input.AuthService, userSvc input.UserService, roleSvc input.RoleService, roleItemSvc input.RoleItemService) *grpcadapter.Handler {
	return grpcadapter.NewHandler(authSvc, userSvc, roleSvc, roleItemSvc)
}

var Module = fx.Options(
	fx.Provide(
		loadConfig,
		newDB,
		newCache,
		newUserRepository,
		newRoleRepository,
		newSessionRepository,
		newSocialLoginRepository,
		newAuthorizer,
		newHasher,
		newTokenGenerator,
		newTokenStore,
		newSessionService,
		asSessionService,
		newAuthService,
		asAuthService,
		newSSOService,
		newSocialLogins,
		newUserService,
		newRoleService,
		newRoleItemService,
		newGRPCHandler,
	),
	fx.Invoke(RegisterServer),
)
