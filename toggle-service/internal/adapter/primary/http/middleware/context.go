package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
)

type ctxKey int

const (
	ctxKeyUserID ctxKey = iota
	ctxKeyIsAdmin
	ctxKeyAPIToken
)

func WithUser(ctx context.Context, userID string, isAdmin bool) context.Context {
	ctx = context.WithValue(ctx, ctxKeyUserID, userID)
	return context.WithValue(ctx, ctxKeyIsAdmin, isAdmin)
}

func UserID(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(ctxKeyUserID).(string)
	return v, ok
}

func IsAdmin(ctx context.Context) bool {
	v, _ := ctx.Value(ctxKeyIsAdmin).(bool)
	return v
}

func WithAPIToken(ctx context.Context, tok *model.APIToken) context.Context {
	return context.WithValue(ctx, ctxKeyAPIToken, tok)
}

func TokenFromContext(ctx context.Context) *model.APIToken {
	v, _ := ctx.Value(ctxKeyAPIToken).(*model.APIToken)
	return v
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	return strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
}
