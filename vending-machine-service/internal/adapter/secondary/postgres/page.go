package postgres

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

var errBadCursor = fmt.Errorf("%w: invalid cursor", domain.ErrInvalidInput)

func encodeCursor(keys ...string) string {
	b, _ := json.Marshal(keys)
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(cursor string, n int) ([]string, error) {
	if cursor == "" {
		return nil, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return nil, errBadCursor
	}
	var keys []string
	if err := json.Unmarshal(raw, &keys); err != nil || len(keys) != n {
		return nil, errBadCursor
	}
	return keys, nil
}

func cursorArgs(cursor string, n int) ([]any, error) {
	keys, err := decodeCursor(cursor, n)
	if err != nil {
		return nil, err
	}
	args := make([]any, n)
	for i := range args {
		if keys != nil {
			args[i] = keys[i]
		}
	}
	return args, nil
}

func timeCursorArgs(cursor string) ([]any, error) {
	args, err := cursorArgs(cursor, 2)
	if err != nil || args[0] == nil {
		return args, err
	}
	t, err := time.Parse(time.RFC3339Nano, args[0].(string))
	if err != nil {
		return nil, errBadCursor
	}
	args[0] = t
	return args, nil
}

func intCursorArgs(cursor string) ([]any, error) {
	args, err := cursorArgs(cursor, 2)
	if err != nil || args[0] == nil {
		return args, err
	}
	n, err := strconv.Atoi(args[0].(string))
	if err != nil {
		return nil, errBadCursor
	}
	args[0] = n
	return args, nil
}

func paginate[T any](items []T, size int, key func(*T) []string) domain.Page[T] {
	page := domain.Page[T]{Items: items, IsLastPage: true}
	if len(items) > size {
		page.Items = items[:size]
		page.IsLastPage = false
		page.NextCursor = encodeCursor(key(&page.Items[size-1])...)
	}
	return page
}

func timeKey(t time.Time) string { return t.UTC().Format(time.RFC3339Nano) }
