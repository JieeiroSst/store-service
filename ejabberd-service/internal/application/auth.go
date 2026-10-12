package application

import (
	"context"
	"errors"
	"log/slog"
	"sync"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
	"go.uber.org/fx"
)

type RemoveUserHook func(ctx context.Context, luser, lserver string)

type AuthService struct {
	cfg      *config.Config
	users    port.UserRepository
	external port.ExternalAuthenticator
	mu       sync.RWMutex
	hooks    []RemoveUserHook
}

type AuthParams struct {
	fx.In

	Config   *config.Config
	Users    port.UserRepository
	External port.ExternalAuthenticator `optional:"true"`
}

func NewAuthService(p AuthParams) *AuthService {
	s := &AuthService{cfg: p.Config, users: p.Users}
	if p.Config.ExternalAuth() {
		s.external = p.External
	}
	return s
}

func (s *AuthService) isExternal() bool {
	return s.external != nil
}

func (s *AuthService) OnRemoveUser(h RemoveUserHook) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.hooks = append(s.hooks, h)
}

func (s *AuthService) StoreType() domain.PasswordFormat {
	return s.cfg.AuthPasswordFormat
}

func (s *AuthService) Mechanisms() []string {
	if s.isExternal() {
		return []string{"PLAIN"}
	}
	return domain.SASLMechanisms(s.cfg.AuthPasswordFormat, s.cfg.AuthScramHash)
}

func validateCredentials(user, server string) (string, string, error) {
	lu, ok := domain.Nodeprep(user)
	if !ok || lu == "" {
		return "", "", domain.ErrInvalidJID
	}
	ls, ok := domain.Nameprep(server)
	if !ok {
		return "", "", domain.ErrInvalidJID
	}
	return lu, ls, nil
}

func (s *AuthService) CheckPassword(ctx context.Context, user, authzid, server, password string) bool {
	lu, ls, err := validateCredentials(user, server)
	if err != nil {
		return false
	}
	if authzid != "" {
		la, ok := domain.Nodeprep(authzid)
		if !ok || (la != "" && la != lu) {
			return false
		}
	}
	if s.isExternal() {
		ok, err := s.external.CheckPassword(ctx, user, lu, ls, password)
		if err != nil {
			slog.Error("user_service check_password failed", "user", lu, "server", ls, "err", err)
			return false
		}
		return ok
	}
	stored, err := s.users.GetPassword(ctx, lu, ls)
	if err != nil {
		return false
	}
	return domain.MatchPassword(password, stored)
}

func (s *AuthService) GetPassword(ctx context.Context, user, server string) (domain.Password, bool) {
	if s.isExternal() {
		return domain.Password{}, false
	}
	lu, ls, err := validateCredentials(user, server)
	if err != nil {
		return domain.Password{}, false
	}
	p, err := s.users.GetPassword(ctx, lu, ls)
	if err != nil {
		return domain.Password{}, false
	}
	return p, true
}

func (s *AuthService) encode(password string) (domain.Password, error) {
	if s.cfg.AuthPasswordFormat == domain.PasswordScram {
		sc, err := domain.PasswordToScram(s.cfg.AuthScramHash, password, domain.ScramDefaultIterationCount)
		if err != nil {
			return domain.Password{}, err
		}
		return domain.Password{Scram: sc}, nil
	}
	return domain.Password{Plain: password}, nil
}

func (s *AuthService) TryRegister(ctx context.Context, user, server, password string) error {
	if s.isExternal() {
		return domain.ErrNotAllowed
	}
	if password == "" {
		return domain.ErrInvalidPassword
	}
	lu, ls, err := validateCredentials(user, server)
	if err != nil {
		return err
	}
	if s.UserExists(ctx, lu, ls) {
		return domain.ErrUserExists
	}
	if !s.cfg.IsMyHost(ls) {
		return domain.ErrNotMyHost
	}
	p, err := s.encode(password)
	if err != nil {
		return err
	}
	if err := s.users.TryRegister(ctx, lu, ls, p); err != nil {
		return err
	}
	slog.Info("registered user", "jid", lu+"@"+ls)
	return nil
}

func (s *AuthService) SetPassword(ctx context.Context, user, server, password string) error {
	if s.isExternal() {
		return domain.ErrNotAllowed
	}
	if password == "" {
		return domain.ErrInvalidPassword
	}
	lu, ls, err := validateCredentials(user, server)
	if err != nil {
		return err
	}
	if !s.UserExists(ctx, lu, ls) {
		return domain.ErrUserNotFound
	}
	p, err := s.encode(password)
	if err != nil {
		return err
	}
	return s.users.SetPassword(ctx, lu, ls, p)
}

func (s *AuthService) UserExists(ctx context.Context, luser, lserver string) bool {
	if lserver == "" {
		return false
	}
	if s.isExternal() {
		ok, err := s.external.UserExists(ctx, luser, lserver)
		if err != nil {
			slog.Warn("user_service user lookup failed, assuming the user exists", "user", luser, "server", lserver, "err", err)
			return true
		}
		return ok
	}
	ok, err := s.users.UserExists(ctx, luser, lserver)
	if err != nil {
		slog.Error("user_exists failed", "user", luser, "server", lserver, "err", err)
		return false
	}
	return ok
}

func (s *AuthService) RemoveUser(ctx context.Context, user, server string) error {
	if s.isExternal() {
		return domain.ErrNotAllowed
	}
	lu, ls, err := validateCredentials(user, server)
	if err != nil {
		return err
	}
	if !s.UserExists(ctx, lu, ls) {
		return domain.ErrUserNotFound
	}
	s.mu.RLock()
	hooks := append([]RemoveUserHook(nil), s.hooks...)
	s.mu.RUnlock()
	for _, h := range hooks {
		h(ctx, lu, ls)
	}
	if err := s.users.RemoveUser(ctx, lu, ls); err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return err
	}
	slog.Info("removed user", "jid", lu+"@"+ls)
	return nil
}

func (s *AuthService) RemoveUserWithPassword(ctx context.Context, user, server, password string) error {
	if !s.CheckPassword(ctx, user, "", server, password) {
		return domain.ErrNotAllowed
	}
	return s.RemoveUser(ctx, user, server)
}

func (s *AuthService) ListUsers(ctx context.Context, server string) ([]string, error) {
	if s.isExternal() {
		return nil, domain.ErrNotAllowed
	}
	ls, ok := domain.Nameprep(server)
	if !ok {
		return nil, domain.ErrInvalidJID
	}
	return s.users.ListUsers(ctx, ls)
}

func (s *AuthService) CountUsers(ctx context.Context, server string) (int, error) {
	if s.isExternal() {
		return 0, domain.ErrNotAllowed
	}
	ls, ok := domain.Nameprep(server)
	if !ok {
		return 0, domain.ErrInvalidJID
	}
	return s.users.CountUsers(ctx, ls)
}

type saslCreds struct {
	ctx    context.Context
	auth   *AuthService
	server string
}

func (c saslCreds) GetPassword(user string) (domain.Password, bool) {
	return c.auth.GetPassword(c.ctx, user, c.server)
}

func (c saslCreds) CheckPassword(user, authzid, password string) bool {
	return c.auth.CheckPassword(c.ctx, user, authzid, c.server, password)
}

func (s *AuthService) NewSASL(ctx context.Context, mech, server string) (domain.SASLMechanism, error) {
	allowed := false
	for _, m := range s.Mechanisms() {
		if m == mech {
			allowed = true
		}
	}
	if !allowed {
		return nil, domain.ErrSASLInvalidMech
	}
	return domain.NewSASLMechanism(mech, saslCreds{ctx: ctx, auth: s, server: server})
}
