package inbound

import "context"

type Principal struct {
	UserID int64
	Email  string
	Admin  bool
}

func (p Principal) IsAdmin() bool { return p.Admin }

type AuthUseCase interface {
	Authenticate(ctx context.Context, token string) (Principal, error)
}
