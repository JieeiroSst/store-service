package graph

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/JIeeiroSst/gateway-api/auth"
)

func currentUserID(ctx context.Context) (int, error) {
	user, err := currentUser(ctx)
	if err != nil {
		return 0, err
	}
	id, err := strconv.Atoi(user.ID)
	if err != nil {
		return 0, auth.NewError(ctx, auth.CodeUnauthenticated, "invalid session user")
	}
	return id, nil
}

func withField(input any, key string, value any) map[string]any {
	out := toMap(input)
	out[key] = value
	return out
}

func toMap(input any) map[string]any {
	out := map[string]any{}
	if input == nil {
		return out
	}
	b, err := json.Marshal(input)
	if err != nil {
		return out
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	_ = dec.Decode(&out)
	if out == nil {
		out = map[string]any{}
	}
	return out
}

func notOwned(ctx context.Context) error {
	return auth.NewError(ctx, auth.CodeNotFound, "not found")
}

func mergeBody(input any, fixed string) map[string]any {
	out := toMap(input)
	extra := map[string]any{}
	_ = json.Unmarshal([]byte(fixed), &extra)
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func userIDOf(ctx context.Context) string {
	user, _ := auth.UserFromContext(ctx)
	return user.ID
}

func userIDInt(ctx context.Context) int {
	id, _ := strconv.Atoi(userIDOf(ctx))
	return id
}
