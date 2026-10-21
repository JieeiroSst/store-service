package httpapi

import (
	"context"
	"net/http"

	"aiagentsystem/internal/domain/port"
)

type Dependencies struct {
	Chat    port.ChatService
	History port.HistoryService
	Ready   func(ctx context.Context) error
}

func NewRouter(d Dependencies) http.Handler {
	ws := &wsHandler{chat: d.Chat}
	history := &historyHandler{history: d.History}
	health := &healthHandler{ready: d.Ready}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/chat/ws", ws.Chat)
	mux.HandleFunc("GET /v1/chat/history", history.Get)
	mux.HandleFunc("GET /healthz", health.Live)
	mux.HandleFunc("GET /readyz", health.Ready)
	return mux
}
