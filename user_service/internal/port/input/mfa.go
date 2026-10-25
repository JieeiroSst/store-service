package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type MFAService interface {
	Status(ctx context.Context, accessToken string) (dto.MFAStatusResponse, error)
	SetupTOTP(ctx context.Context, accessToken string) (dto.TOTPSetupResponse, error)
	EnableTOTP(ctx context.Context, accessToken, code string) (dto.RecoveryCodesResponse, error)
	DisableTOTP(ctx context.Context, accessToken, code string) error
	RegenerateRecoveryCodes(ctx context.Context, accessToken, code string) (dto.RecoveryCodesResponse, error)
	VerifyLogin(ctx context.Context, mfaToken, code string) (dto.LoginResponse, error)
}
