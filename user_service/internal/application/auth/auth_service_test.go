package auth

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
	users map[string]string
}

func (f *fakeUserRepo) CheckAccount(_ context.Context, u domain.User) (int, string, string, error) {
	h, ok := f.users[u.Username]
	if !ok {
		return 0, "", "", domain.ErrUserNotExist
	}
	return 1, h, "", nil
}

func (f *fakeUserRepo) FindUser(context.Context, int) (domain.User, error) {
	return domain.User{Checked: true}, nil
}

type fakeHasher struct{ checks int }

func (h *fakeHasher) HashPassword(p string) (string, error) { return "hash:" + p, nil }

func (h *fakeHasher) CheckPassword(p, hash string) error {
	h.checks++
	if hash != "hash:"+p {
		return errors.New("mismatch")
	}
	return nil
}

type fakeThrottle struct {
	max      int
	failures map[string]int
}

func (t *fakeThrottle) Blocked(_ context.Context, u string) (bool, error) {
	return t.failures[u] >= t.max, nil
}

func (t *fakeThrottle) RecordFailure(_ context.Context, u string) error {
	t.failures[u]++
	return nil
}

func (t *fakeThrottle) Reset(_ context.Context, u string) error {
	delete(t.failures, u)
	return nil
}

type fakeSessions struct{ Sessions }

func (fakeSessions) Start(context.Context, int, string, string) (dto.LoginResponse, error) {
	return dto.LoginResponse{SessionToken: "tok"}, nil
}

func newTestService() (*Service, *fakeHasher, *fakeThrottle) {
	h := &fakeHasher{}
	th := &fakeThrottle{max: 3, failures: map[string]int{}}
	repo := &fakeUserRepo{users: map[string]string{"alice": "hash:right"}}
	return New(repo, h, nil, nil, fakeSessions{}, th), h, th
}

func TestLoginBlocksAfterRepeatedFailures(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if _, err := svc.Login(ctx, dto.LoginRequest{Username: "alice", Password: "wrong"}); !errors.Is(err, domain.ErrInvalidCredentials) {
			t.Fatalf("attempt %d: err = %v, want ErrInvalidCredentials", i, err)
		}
	}
	if _, err := svc.Login(ctx, dto.LoginRequest{Username: "alice", Password: "right"}); !errors.Is(err, domain.ErrTooManyAttempts) {
		t.Fatalf("err = %v, want ErrTooManyAttempts even with the right password", err)
	}
}

func TestLoginSuccessResetsFailures(t *testing.T) {
	svc, _, th := newTestService()
	ctx := context.Background()
	_, _ = svc.Login(ctx, dto.LoginRequest{Username: "alice", Password: "wrong"})
	if _, err := svc.Login(ctx, dto.LoginRequest{Username: "alice", Password: "right"}); err != nil {
		t.Fatalf("Login() error = %v", err)
	}
	if n := th.failures["alice"]; n != 0 {
		t.Fatalf("failures = %d after success, want 0", n)
	}
}

func TestLoginUnknownUserStillChecksPassword(t *testing.T) {
	svc, h, th := newTestService()
	if _, err := svc.Login(context.Background(), dto.LoginRequest{Username: "ghost", Password: "x"}); !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("err = %v, want ErrInvalidCredentials", err)
	}
	if h.checks != 1 {
		t.Fatalf("password checks = %d, want 1 so unknown users cost the same as known ones", h.checks)
	}
	if th.failures["ghost"] != 1 {
		t.Fatalf("unknown user failure not counted")
	}
}
