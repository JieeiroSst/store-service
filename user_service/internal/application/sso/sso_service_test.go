package sso

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type fakeIDP struct {
	identity domain.ExternalIdentity
	err      error
	verifier string
}

func (f *fakeIDP) AuthCodeURL(context.Context, string, string) (string, error) { return "", nil }

func (f *fakeIDP) Exchange(_ context.Context, _ string, verifier string) (domain.ExternalIdentity, error) {
	f.verifier = verifier
	return f.identity, f.err
}

func (f *fakeIDP) VerifyToken(context.Context, string) (domain.ExternalIdentity, error) {
	return f.identity, f.err
}

type fakeRepo struct {
	output.UserRepository
	users   []domain.User
	created []domain.User
	deleted []int
}

func (r *fakeRepo) SearchUsers(_ context.Context, f domain.UserFilter) ([]domain.User, int64, error) {
	for _, u := range r.users {
		if u.Email == f.Email {
			return []domain.User{u}, 1, nil
		}
	}
	return nil, 0, nil
}

func (r *fakeRepo) FindUser(_ context.Context, id int) (domain.User, error) {
	for _, u := range append(r.users, r.created...) {
		if u.Id == id {
			return u, nil
		}
	}
	return domain.User{}, domain.ErrUserNotExist
}

func (r *fakeRepo) CheckAccountExists(_ context.Context, user domain.User) error {
	for _, u := range r.users {
		if u.Username == user.Username {
			return domain.ErrUserExist
		}
	}
	return nil
}

func (r *fakeRepo) CreateAccount(_ context.Context, user domain.User) (domain.User, error) {
	r.created = append(r.created, user)
	return user, nil
}

func (r *fakeRepo) DeleteAccount(_ context.Context, id int) error {
	r.deleted = append(r.deleted, id)
	return nil
}

type fakeIdentities struct {
	links   []domain.SocialLogin
	linkErr error
}

func (f *fakeIdentities) Find(_ context.Context, provider, subject string) (domain.SocialLogin, error) {
	for _, l := range f.links {
		if l.Provider == provider && l.Subject == subject {
			return l, nil
		}
	}
	return domain.SocialLogin{}, domain.ErrUserNotExist
}

func (f *fakeIdentities) Create(_ context.Context, l domain.SocialLogin) error {
	if f.linkErr != nil {
		return f.linkErr
	}
	f.links = append(f.links, l)
	return nil
}

type fakeHasher struct{}

func (fakeHasher) HashPassword(p string) (string, error) { return "hashed:" + p, nil }
func (fakeHasher) CheckPassword(string, string) error    { return nil }

type fakeAuthorizer struct {
	output.Authorizer
	err error
}

func (a fakeAuthorizer) AssignRoles(context.Context, int, ...string) (domain.UserRoles, error) {
	return domain.UserRoles{}, a.err
}

type fakeSessions struct {
	userID   int
	provider string
}

func (s *fakeSessions) Start(_ context.Context, userID int, username, provider string) (dto.LoginResponse, error) {
	s.userID, s.provider = userID, provider
	return dto.LoginResponse{SessionToken: "token-" + username}, nil
}

var verified = domain.ExternalIdentity{Subject: "sub", Email: "quan@example.com", EmailVerified: true, Name: "Quan", Username: "quan"}

type fixture struct {
	idp        *fakeIDP
	repo       *fakeRepo
	identities *fakeIdentities
	sessions   *fakeSessions
	authz      fakeAuthorizer
}

func newFixture(identity domain.ExternalIdentity, users ...domain.User) *fixture {
	return &fixture{
		idp:        &fakeIDP{identity: identity},
		repo:       &fakeRepo{users: users},
		identities: &fakeIdentities{},
		sessions:   &fakeSessions{},
	}
}

func (f *fixture) svc() *Service {
	return New(domain.ProviderGoogle, f.idp, f.repo, f.identities, fakeHasher{}, f.authz, f.sessions)
}

func TestCallbackLinksExistingUserByVerifiedEmail(t *testing.T) {
	f := newFixture(verified, domain.User{Id: 7, Username: "quanluu", Email: "quan@example.com", Checked: true})
	resp, err := f.svc().Callback(context.Background(), "c", "v")
	if err != nil || resp.SessionToken != "token-quanluu" || f.sessions.userID != 7 || len(f.repo.created) != 0 {
		t.Fatalf("%v %+v %+v", err, resp, f.repo.created)
	}
	if f.idp.verifier != "v" || f.sessions.provider != domain.ProviderGoogle {
		t.Fatalf("verifier=%q provider=%q", f.idp.verifier, f.sessions.provider)
	}
	if len(f.identities.links) != 1 || f.identities.links[0].UserID != 7 || f.identities.links[0].Subject != "sub" {
		t.Fatalf("%+v", f.identities.links)
	}
}

func TestCallbackUsesExistingLink(t *testing.T) {
	changedEmail := verified
	changedEmail.Email = "new@example.com"
	f := newFixture(changedEmail, domain.User{Id: 7, Username: "quanluu", Email: "quan@example.com", Checked: true})
	f.identities.links = []domain.SocialLogin{{Provider: domain.ProviderGoogle, Subject: "sub", UserID: 7}}
	if _, err := f.svc().Callback(context.Background(), "c", ""); err != nil || f.sessions.userID != 7 || len(f.repo.created) != 0 {
		t.Fatalf("%v %d %+v", err, f.sessions.userID, f.repo.created)
	}
}

func TestCallbackCreatesUser(t *testing.T) {
	f := newFixture(verified, domain.User{Id: 1, Username: "quan", Email: "other@example.com"})
	resp, err := f.svc().Callback(context.Background(), "c", "")
	if err != nil || len(f.repo.created) != 1 {
		t.Fatalf("%v %+v", err, f.repo.created)
	}
	u := f.repo.created[0]
	if u.Username != "quan@example.com" || u.Email != verified.Email || !u.Checked || u.Password == "" || resp.SessionToken != "token-quan@example.com" {
		t.Fatalf("%+v %+v", u, resp)
	}
	if len(f.identities.links) != 1 || f.identities.links[0].UserID != u.Id {
		t.Fatalf("%+v", f.identities.links)
	}
}

func TestUnverifiedEmailNeverLinksExistingAccount(t *testing.T) {
	unverified := verified
	unverified.EmailVerified = false
	unverified.Username = ""
	f := newFixture(unverified, domain.User{Id: 7, Username: "victim", Email: verified.Email, Checked: true})
	if _, err := f.svc().Callback(context.Background(), "c", ""); err != nil {
		t.Fatal(err)
	}
	if f.sessions.userID == 7 || len(f.repo.created) != 1 {
		t.Fatalf("linked to victim: %d %+v", f.sessions.userID, f.repo.created)
	}
	u := f.repo.created[0]
	if u.Email != "" || u.Username != "google_sub" {
		t.Fatalf("%+v", u)
	}
}

func TestCallbackRollsBackWhenRoleAssignFails(t *testing.T) {
	f := newFixture(verified)
	f.authz = fakeAuthorizer{err: errors.New("down")}
	_, err := f.svc().Callback(context.Background(), "c", "")
	if !errors.Is(err, domain.ErrAssignRoleFailed) || len(f.repo.deleted) != 1 {
		t.Fatalf("%v %v", err, f.repo.deleted)
	}
}

func TestCallbackRollsBackWhenLinkFails(t *testing.T) {
	f := newFixture(verified)
	f.identities.linkErr = errors.New("duplicate")
	if _, err := f.svc().Callback(context.Background(), "c", ""); err == nil || len(f.repo.deleted) != 1 {
		t.Fatalf("%v %v", err, f.repo.deleted)
	}
}

func TestCallbackRejections(t *testing.T) {
	noSubject := verified
	noSubject.Subject = ""
	cases := []struct {
		name  string
		f     *fixture
		code  string
		idErr error
		want  error
	}{
		{"no code", newFixture(verified), "", nil, domain.ErrInvalidRequest},
		{"exchange fails", newFixture(verified), "c", errors.New("bad"), domain.ErrSSOExchangeFailed},
		{"no subject", newFixture(noSubject), "c", nil, domain.ErrSSONoIdentity},
		{"locked", newFixture(verified, domain.User{Id: 7, Username: "quan", Email: verified.Email, Checked: false}), "c", nil, domain.ErrAccountLocked},
	}
	for _, c := range cases {
		c.f.idp.err = c.idErr
		_, err := c.f.svc().Callback(context.Background(), c.code, "")
		if !errors.Is(err, c.want) {
			t.Fatalf("%s: got %v want %v", c.name, err, c.want)
		}
	}
}

func TestLoginWithToken(t *testing.T) {
	f := newFixture(domain.ExternalIdentity{Subject: "g-1", Email: "quan@gmail.com", EmailVerified: true, Name: "Quan"})
	resp, err := f.svc().LoginWithToken(context.Background(), "idt")
	if err != nil || len(f.repo.created) != 1 || resp.SessionToken != "token-quan@gmail.com" {
		t.Fatalf("%v %+v %+v", err, f.repo.created, resp)
	}

	cases := []struct {
		name  string
		token string
		err   error
		want  error
	}{
		{"no token", "", nil, domain.ErrInvalidRequest},
		{"invalid token", "idt", errors.New("bad"), domain.ErrSSOIDTokenInvalid},
		{"unsupported", "idt", domain.ErrSSOTokenUnsupported, domain.ErrSSOTokenUnsupported},
	}
	for _, c := range cases {
		f := newFixture(verified)
		f.idp.err = c.err
		if _, err := f.svc().LoginWithToken(context.Background(), c.token); !errors.Is(err, c.want) {
			t.Fatalf("%s: got %v want %v", c.name, err, c.want)
		}
	}
}

func TestFirstConsentStoresSocialLogin(t *testing.T) {
	rich := verified
	rich.AvatarURL = "https://img.example/q.png"
	rich.Scopes = []string{"openid", "email", "profile"}
	rich.Profile = []byte(`{"sub":"sub","locale":"vi"}`)
	f := newFixture(rich)
	ctx := domain.WithClientInfo(context.Background(), domain.ClientInfo{IP: "203.0.113.5"})
	if _, err := f.svc().Callback(ctx, "c", ""); err != nil {
		t.Fatal(err)
	}
	if len(f.identities.links) != 1 {
		t.Fatalf("%+v", f.identities.links)
	}
	l := f.identities.links[0]
	if l.Provider != domain.ProviderGoogle || l.Subject != "sub" || l.UserID != f.repo.created[0].Id ||
		l.Email != verified.Email || !l.EmailVerified || l.Name != "Quan" || l.Username != "quan" ||
		l.AvatarURL != rich.AvatarURL || l.Scopes != "openid email profile" || l.Profile != `{"sub":"sub","locale":"vi"}` ||
		l.ConsentedAt.IsZero() || l.ConsentIP != "203.0.113.5" {
		t.Fatalf("%+v", l)
	}
}

func TestInvalidProfileStoredAsEmptyObject(t *testing.T) {
	bad := verified
	bad.Profile = []byte("not json")
	f := newFixture(bad)
	if _, err := f.svc().Callback(context.Background(), "c", ""); err != nil || f.identities.links[0].Profile != "{}" {
		t.Fatalf("%v %+v", err, f.identities.links)
	}
}

func TestReturningLoginDoesNotWriteAgain(t *testing.T) {
	updated := verified
	updated.Name = "Quan Luu"
	original := domain.SocialLogin{Provider: domain.ProviderGoogle, Subject: "sub", UserID: 7, Name: "Quan"}
	f := newFixture(updated, domain.User{Id: 7, Username: "quanluu", Email: verified.Email, Checked: true})
	f.identities.links = []domain.SocialLogin{original}
	f.identities.linkErr = errors.New("must not write")
	if _, err := f.svc().Callback(context.Background(), "c", ""); err != nil || f.sessions.userID != 7 {
		t.Fatalf("%v %d", err, f.sessions.userID)
	}
	if len(f.identities.links) != 1 || f.identities.links[0] != original {
		t.Fatalf("%+v", f.identities.links)
	}
}
