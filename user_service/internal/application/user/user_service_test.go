package user

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/user-service/dto"
	"github.com/JIeeiroSst/user-service/internal/domain"
	"github.com/JIeeiroSst/user-service/internal/port/output"
)

type fakeUserRepo struct {
	output.UserRepository
	created, deleted []int
}

func (f *fakeUserRepo) CheckAccountExists(context.Context, domain.User) error { return nil }

func (f *fakeUserRepo) CreateAccount(_ context.Context, u domain.User) (domain.User, error) {
	f.created = append(f.created, u.Id)
	return u, nil
}

func (f *fakeUserRepo) DeleteAccount(_ context.Context, id int) error {
	f.deleted = append(f.deleted, id)
	return nil
}

type fakeHasher struct{}

func (fakeHasher) HashPassword(p string) (string, error) { return "hashed-" + p, nil }
func (fakeHasher) CheckPassword(string, string) error    { return nil }

type fakeAuthorizer struct {
	output.Authorizer
	err      error
	assigned map[int][]string
}

func (f *fakeAuthorizer) AssignRoles(_ context.Context, userID int, roles ...string) (domain.UserRoles, error) {
	if f.err != nil {
		return domain.UserRoles{}, f.err
	}
	if f.assigned == nil {
		f.assigned = map[int][]string{}
	}
	f.assigned[userID] = append(f.assigned[userID], roles...)
	return domain.UserRoles{Roles: roles, PrimaryRole: roles[0]}, nil
}

var signUpReq = dto.SignUpRequest{Username: "alice", Password: "Secret123", Email: "alice.example@mail.com"}

func TestSignUpAssignsDefaultRole(t *testing.T) {
	repo, authz := &fakeUserRepo{}, &fakeAuthorizer{}
	svc := New(repo, fakeHasher{}, nil, authz)

	if _, err := svc.SignUp(context.Background(), signUpReq); err != nil {
		t.Fatalf("SignUp: %v", err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created %d accounts, want 1", len(repo.created))
	}
	got := authz.assigned[repo.created[0]]
	if len(got) != 1 || got[0] != domain.DefaultRole {
		t.Errorf("assigned roles = %v, want [%s]", got, domain.DefaultRole)
	}
}

func TestSignUpRollsBackWhenAuthorizeFails(t *testing.T) {
	repo := &fakeUserRepo{}
	svc := New(repo, fakeHasher{}, nil, &fakeAuthorizer{err: errors.New("unavailable")})

	_, err := svc.SignUp(context.Background(), signUpReq)
	if !errors.Is(err, domain.ErrAssignRoleFailed) {
		t.Fatalf("SignUp error = %v, want ErrAssignRoleFailed", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0] != repo.created[0] {
		t.Errorf("account not rolled back: created=%v deleted=%v", repo.created, repo.deleted)
	}
}
