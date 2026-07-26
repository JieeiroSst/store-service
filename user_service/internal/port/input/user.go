package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type UserService interface {
	SignUp(ctx context.Context, req dto.SignUpRequest) (dto.SignUpResponse, error)
	UpdateProfile(ctx context.Context, req dto.UpdateProfileRequest) (dto.UpdateProfileResponse, error)
	LockAccount(ctx context.Context, req dto.LockAccountRequest) (dto.LockAccountResponse, error)
	FindUser(ctx context.Context, req dto.FindUserRequest) (dto.FindUserResponse, error)
}
