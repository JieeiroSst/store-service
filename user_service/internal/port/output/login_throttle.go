package output

import "context"

type LoginThrottle interface {
	Blocked(ctx context.Context, username string) (bool, error)
	RecordFailure(ctx context.Context, username string) error
	Reset(ctx context.Context, username string) error
}
