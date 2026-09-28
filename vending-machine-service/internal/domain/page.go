package domain

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type PageRequest struct {
	Limit  int
	Cursor string
}

func (p PageRequest) Size() int {
	switch {
	case p.Limit <= 0:
		return DefaultPageSize
	case p.Limit > MaxPageSize:
		return MaxPageSize
	}
	return p.Limit
}

type Page[T any] struct {
	Items      []T
	NextCursor string
	IsLastPage bool
}
