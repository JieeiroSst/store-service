package ports

import "context"

type DBPing interface {
	Ping(ctx context.Context) error
}
