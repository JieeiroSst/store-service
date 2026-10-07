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

type searchRepo struct {
	output.UserRepository
	users  map[int]domain.User
	filter domain.UserFilter
}

func (r *searchRepo) FindUser(_ context.Context, id int) (domain.User, error) {
	u, ok := r.users[id]
	if !ok {
		return domain.User{}, domain.ErrUserNotExist
	}
	return u, nil
}

func (r *searchRepo) SearchUsers(_ context.Context, f domain.UserFilter) ([]domain.User, int64, error) {
	r.filter = f
	var out []domain.User
	for id := 1; id <= len(r.users); id++ {
		u := r.users[id]
		if (f.Username == "" || u.Username == f.Username) && (f.Email == "" || u.Email == f.Email) {
			out = append(out, u)
		}
	}
	return out, int64(len(out)), nil
}

func TestFindUser(t *testing.T) {
	repo := &searchRepo{users: map[int]domain.User{
		1: {Id: 1, Username: "alice", Email: "alice@mail.com", Password: "hash", Checked: true},
		2: {Id: 2, Username: "bob", Email: "bob@mail.com"},
	}}
	svc := New(repo, fakeHasher{}, nil, &fakeAuthorizer{})
	ctx := context.Background()

	byID, err := svc.FindUser(ctx, dto.FindUserRequest{UserId: 2})
	if err != nil || byID.Total != 1 || len(byID.Users) != 1 || byID.Users[0].Username != "bob" {
		t.Fatalf("by id: %v %+v", err, byID)
	}
	if _, err := svc.FindUser(ctx, dto.FindUserRequest{UserId: 9}); !errors.Is(err, domain.ErrUserNotExist) {
		t.Fatalf("missing id: %v", err)
	}

	name := "alice"
	byName, err := svc.FindUser(ctx, dto.FindUserRequest{Username: &name})
	if err != nil || byName.Total != 1 || byName.Users[0].Id != 1 || byName.Users[0].Password != "" {
		t.Fatalf("by username: %v %+v", err, byName)
	}
	if repo.filter.Page != domain.DefaultUserPage || repo.filter.Limit != domain.DefaultUserLimit {
		t.Fatalf("filter not normalized: %+v", repo.filter)
	}

	limit := int32(500)
	all, err := svc.FindUser(ctx, dto.FindUserRequest{Limit: &limit})
	if err != nil || all.Total != 2 || repo.filter.Limit != domain.MaxUserLimit {
		t.Fatalf("list: %v %+v %+v", err, all, repo.filter)
	}
}
