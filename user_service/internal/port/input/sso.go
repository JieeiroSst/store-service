package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type SSOService interface {
	LoginURL(ctx context.Context, state, verifier string) (string, error)
	Callback(ctx context.Context, code, verifier string) (dto.LoginResponse, error)
	LoginWithToken(ctx context.Context, token string) (dto.LoginResponse, error)
}
