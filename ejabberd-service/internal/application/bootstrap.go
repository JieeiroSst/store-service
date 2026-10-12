package application

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"go.uber.org/fx"
)

const (
	bootstrapMinBackoff = 2 * time.Second
	bootstrapMaxBackoff = 30 * time.Second
)

type AdminBootstrap struct {
	cfg      *config.Config
	matcher  *domain.ACLMatcher
	external port.ExternalAuthenticator
	done     chan struct{}
}

type BootstrapParams struct {
	fx.In

	Config   *config.Config
	Matcher  *domain.ACLMatcher
	External port.ExternalAuthenticator `optional:"true"`
}

func NewAdminBootstrap(p BootstrapParams) *AdminBootstrap {
	return &AdminBootstrap{cfg: p.Config, matcher: p.Matcher, external: p.External, done: make(chan struct{})}
}

func (b *AdminBootstrap) Done() <-chan struct{} { return b.done }

func (b *AdminBootstrap) Run(ctx context.Context) {
	defer close(b.done)
	if !b.cfg.ExternalAuth() || b.external == nil || len(b.cfg.UserService.Admins) == 0 {
		return
	}
	var wg sync.WaitGroup
	for _, admin := range b.cfg.UserService.Admins {
		b.checkACL(admin)
		wg.Add(1)
		go func(a config.UserServiceAccount) {
			defer wg.Done()
			b.ensure(ctx, a)
		}(admin)
	}
	wg.Wait()
}

func (b *AdminBootstrap) checkACL(a config.UserServiceAccount) {
	host := b.cfg.MyName()
	jid, err := domain.NewJID(a.Username, host, "")
	if err != nil {
		slog.Error("bootstrap admin: username is not a valid XMPP localpart", "username", a.Username)
		return
	}
	isAdmin := b.matcher.MatchAccess(domain.AccessRule{{Allow: true, Specs: []domain.ACLSpec{{Kind: domain.ACLRef, Name: "admin"}}}}, domain.MatchJID(jid))
	if !isAdmin {
		slog.Warn("bootstrap admin is not in acl admin, it will not get admin API permissions", "jid", jid.String())
	}
}

func (b *AdminBootstrap) ensure(ctx context.Context, a config.UserServiceAccount) {
	account := domain.Account{Username: a.Username, Password: a.Password, Email: a.Email, Name: a.Name}
	if account.Name == "" {
		account.Name = a.Username
	}
	backoff := bootstrapMinBackoff
	for {
		created, err := b.external.EnsureUser(ctx, account)
		switch {
		case err == nil:
			b.verify(ctx, a, created)
			return
		case errors.Is(err, domain.ErrAccountRejected):
			slog.Error("bootstrap admin: user_service rejected the account", "username", a.Username, "err", err)
			return
		}
		slog.Warn("bootstrap admin: user_service not ready, retrying", "username", a.Username, "retry_in", backoff.String(), "err", err)
		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
		}
		backoff = min(backoff*2, bootstrapMaxBackoff)
	}
}

func (b *AdminBootstrap) verify(ctx context.Context, a config.UserServiceAccount, created bool) {
	lu, _ := domain.Nodeprep(a.Username)
	ok, err := b.external.CheckPassword(ctx, a.Username, lu, b.cfg.MyName(), a.Password)
	switch {
	case err != nil:
		slog.Warn("bootstrap admin: could not verify the password", "username", a.Username, "err", err)
	case created:
		slog.Info("bootstrap admin: created in user_service", "username", a.Username, "login_ok", ok)
	case ok:
		slog.Info("bootstrap admin: already exists in user_service", "username", a.Username)
	default:
		slog.Warn("bootstrap admin: already exists in user_service but the configured password does not log in; HTTP API Basic auth with it will fail",
			"username", a.Username, "allow_password_login", b.cfg.UserService.AllowPasswordLogin)
	}
}

func startAdminBootstrap(lc fx.Lifecycle, b *AdminBootstrap) {
	ctx, cancel := context.WithCancel(context.Background())
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go b.Run(ctx)
			return nil
		},
		OnStop: func(stopCtx context.Context) error {
			cancel()
			select {
			case <-b.done:
			case <-stopCtx.Done():
			}
			return nil
		},
	})
}
