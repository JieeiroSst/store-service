package token

import "context"

type (
	ClientStore interface {
		GetByID(ctx context.Context, id string) (ClientInfo, error)
	}

	TokenStore interface {
		Create(ctx context.Context, info TokenInfo) error

		RemoveByCode(ctx context.Context, code string) error

		RemoveByAccess(ctx context.Context, access string) error

		RemoveByRefresh(ctx context.Context, refresh string) error

		GetByCode(ctx context.Context, code string) (TokenInfo, error)

		GetByAccess(ctx context.Context, access string) (TokenInfo, error)

		GetByRefresh(ctx context.Context, refresh string) (TokenInfo, error)
	}
)
