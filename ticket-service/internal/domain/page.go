package domain

type Page[T any] struct {
	Items  []T
	NextID int64
}

func PageOf[T any](rows []T, limit int, id func(T) int64) Page[T] {
	if len(rows) <= limit {
		return Page[T]{Items: rows}
	}
	rows = rows[:limit]
	return Page[T]{Items: rows, NextID: id(rows[limit-1])}
}
