package input

import (
	"context"

	"github.com/JIeeiroSst/user-service/dto"
)

type SessionService interface {
	List(ctx context.Context, accessToken string) (dto.ListSessionsResponse, error)
	Revoke(ctx context.Context, accessToken, sessionID string) error
	RevokeOthers(ctx context.Context, accessToken string, includeCurrent bool) (dto.RevokeSessionsResponse, error)
}
