package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (dto.LoginResponse, error)
	Logout(ctx context.Context, req dto.LogoutRequest) (dto.LogoutResponse, error)
	ValidateSession(ctx context.Context, req dto.ValidateRequest) (dto.ValidateResponse, error)
	RefreshToken(ctx context.Context, req dto.RefreshRequest) (dto.RefreshResponse, error)
	Authentication(ctx context.Context, req dto.AuthenticationRequest) (dto.AuthenticationResponse, error)
}
