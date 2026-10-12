package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type SSOService interface {
	LoginURL(ctx context.Context, state string) (string, error)
	Callback(ctx context.Context, code string) (dto.LoginResponse, error)
}
