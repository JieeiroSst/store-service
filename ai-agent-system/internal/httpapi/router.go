package httpapi

import (
	"log/slog"
	"net/http"

	"aiagentsystem/internal/chat"
	"aiagentsystem/internal/history"
)

func NewRouter(service *chat.Service, historyStore *history.Store, logger *slog.Logger) http.Handler {
	wsH := &wsHandlers{service: service}
	historyH := &historyHandlers{store: historyStore}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/chat/ws", wsH.Chat)
	mux.HandleFunc("GET /v1/chat/history", historyH.Get)

	return withRecover(logger, withLogging(logger, mux))
}
