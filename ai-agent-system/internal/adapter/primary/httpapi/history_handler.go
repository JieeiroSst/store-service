package httpapi

import (
	"net/http"
	"strconv"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

const historyTimeFormat = "2006-01-02T15:04:05.000Z07:00"

type historyHandler struct {
	history port.HistoryService
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

func (h *historyHandler) Get(w http.ResponseWriter, r *http.Request) {
	userID := userIDFromRequest(r)
	if userID == "" {
		writeError(w, model.BadRequest(userIDHeader+" header is required"))
		return
	}

	limit := 0
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed <= 0 {
			writeError(w, model.BadRequest("limit must be a positive integer"))
			return
		}
		limit = parsed
	}

	entries, err := h.history.List(r.Context(), userID, limit)
	if err != nil {
		writeError(w, err)
		return
	}

	resp := historyResponse{Messages: make([]historyMessage, 0, len(entries))}
	for _, e := range entries {
		resp.Messages = append(resp.Messages, historyMessage{
			Role:      e.Role,
			Content:   e.Content,
			Source:    e.Source,
			CreatedAt: e.CreatedAt.UTC().Format(historyTimeFormat),
		})
	}
	writeJSON(w, http.StatusOK, resp)
}
