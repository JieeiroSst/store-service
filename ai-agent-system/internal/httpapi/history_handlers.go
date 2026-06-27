package httpapi

import (
	"net/http"
	"strconv"

	"aiagentsystem/internal/chat"
	"aiagentsystem/internal/history"
)

const (
	defaultHistoryLimit = 50
	maxHistoryLimit     = 200
)

type historyHandlers struct {
	store *history.Store
}

type historyMessage struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Source    string `json:"source,omitempty"`
	CreatedAt string `json:"created_at"`
}

type historyResponse struct {
	Messages []historyMessage `json:"messages"`
}

// Get handles GET /v1/chat/history — returns the most recent logged
// question/answer turns for the caller's X-User-Id, newest first. This is a
// read-only audit log; it has no effect on how future questions are
// answered.
func (h *historyHandlers) Get(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, chat.NewBadRequestError(userIDHeader+" header is required"))
		return
	}

	limit := defaultHistoryLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, chat.NewBadRequestError("limit must be a positive integer"))
			return
		}
		limit = parsed
	}
	if limit > maxHistoryLimit {
		limit = maxHistoryLimit
	}

	messages, err := h.store.ListByUser(r.Context(), userID, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := historyResponse{Messages: make([]historyMessage, 0, len(messages))}
	for _, m := range messages {
		resp.Messages = append(resp.Messages, historyMessage{
			Role:      m.Role,
			Content:   m.Content,
			Source:    m.Source,
			CreatedAt: m.CreatedAt.UTC().Format("2006-01-02T15:04:05.000Z07:00"),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
