package application

import "github.com/JIeeiroSst/shopify-service/internal/domain/port"

const (
	defaultLimit = 20
	maxLimit     = 100
)

func normalizeList(in port.ListInput) (limit, offset int) {
	limit, offset = in.Limit, in.Offset
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
