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
}

func (f fakeIDP) AuthCodeURL(context.Context, string) (string, error) { return "", nil }

func (f fakeIDP) Exchange(context.Context, string) (domain.ExternalIdentity, error) {
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

type fakeSessions struct{ userID int }

func (s *fakeSessions) IssueSession(_ context.Context, userID int, username string) (dto.LoginResponse, error) {
	s.userID = userID
	return dto.LoginResponse{SessionToken: "token-" + username}, nil
}

var verified = domain.ExternalIdentity{Subject: "sub", Email: "quan@example.com", EmailVerified: true, Name: "Quan", Username: "quan"}

func TestCallbackLinksExistingUser(t *testing.T) {
	repo := &fakeRepo{users: []domain.User{{Id: 7, Username: "quanluu", Email: "quan@example.com", Checked: true}}}
	sessions := &fakeSessions{}
	resp, err := New(fakeIDP{identity: verified}, repo, fakeHasher{}, fakeAuthorizer{}, sessions).Callback(context.Background(), "c")
	if err != nil || resp.SessionToken != "token-quanluu" || sessions.userID != 7 || len(repo.created) != 0 {
		t.Fatalf("%v %+v %+v", err, resp, repo.created)
	}
}

func TestCallbackCreatesUser(t *testing.T) {
	repo := &fakeRepo{users: []domain.User{{Id: 1, Username: "quan", Email: "other@example.com"}}}
	resp, err := New(fakeIDP{identity: verified}, repo, fakeHasher{}, fakeAuthorizer{}, &fakeSessions{}).Callback(context.Background(), "c")
	if err != nil || len(repo.created) != 1 {
		t.Fatalf("%v %+v", err, repo.created)
	}
	u := repo.created[0]
	if u.Username != "quan@example.com" || u.Email != verified.Email || !u.Checked || u.Password == "" || resp.SessionToken != "token-quan@example.com" {
		t.Fatalf("%+v %+v", u, resp)
	}
}

func TestCallbackRollsBackWhenRoleAssignFails(t *testing.T) {
	repo := &fakeRepo{}
	_, err := New(fakeIDP{identity: verified}, repo, fakeHasher{}, fakeAuthorizer{err: errors.New("down")}, &fakeSessions{}).Callback(context.Background(), "c")
	if !errors.Is(err, domain.ErrAssignRoleFailed) || len(repo.deleted) != 1 {
		t.Fatalf("%v %v", err, repo.deleted)
	}
}

func TestCallbackRejections(t *testing.T) {
	unverified := verified
	unverified.EmailVerified = false
	locked := &fakeRepo{users: []domain.User{{Id: 7, Username: "quan", Email: verified.Email, Checked: false}}}
	cases := []struct {
		name string
		idp  fakeIDP
		repo *fakeRepo
		code string
		want error
	}{
		{"no code", fakeIDP{identity: verified}, &fakeRepo{}, "", domain.ErrInvalidRequest},
		{"exchange fails", fakeIDP{err: errors.New("bad")}, &fakeRepo{}, "c", domain.ErrSSOExchangeFailed},
		{"unverified email", fakeIDP{identity: unverified}, &fakeRepo{}, "c", domain.ErrSSOEmailUnverified},
		{"locked", fakeIDP{identity: verified}, locked, "c", domain.ErrAccountLocked},
	}
	for _, c := range cases {
		_, err := New(c.idp, c.repo, fakeHasher{}, fakeAuthorizer{}, &fakeSessions{}).Callback(context.Background(), c.code)
		if !errors.Is(err, c.want) {
			t.Fatalf("%s: got %v want %v", c.name, err, c.want)
		}
	}
}
