package output

import (
	"context"

	"github.com/JIeeiroSst/user-service/internal/domain"
)

type TokenGenerator interface {
	GenerateAccessToken(ctx context.Context, userID int, username string) (string, error)
	GenerateRefreshToken(ctx context.Context) (string, error)
	ParseAccessToken(ctx context.Context, token string) (domain.AccessClaims, error)
}
