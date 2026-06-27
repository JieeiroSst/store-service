package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"aiagentsystem/internal/chat"
)

type ChatRequest struct {
	Question string `json:"question"`
}

type errorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type ErrorResponse struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errorTypeAndStatus(err error) (status int, errType string, message string) {
	var cErr *chat.Error
	if errors.As(err, &cErr) {
		switch cErr.Kind {
		case chat.ErrKindRateLimited:
			return http.StatusTooManyRequests, "rate_limited", "rate limit exceeded, please try again shortly"
		case chat.ErrKindUpstreamUnavailable:
			return http.StatusServiceUnavailable, "upstream_unavailable", "the AI service is temporarily unavailable"
		case chat.ErrKindBadRequest:
			return http.StatusBadRequest, "bad_request", cErr.Message
		}
	}
	return http.StatusInternalServerError, "internal", "an internal error occurred"
}

func writeError(w http.ResponseWriter, err error) {
	status, errType, message := errorTypeAndStatus(err)
	writeJSON(w, status, ErrorResponse{Error: errorBody{Message: message, Type: errType}})
}

const userIDHeader = "X-User-Id"

func userIDFromRequest(r *http.Request) string {
	return r.Header.Get(userIDHeader)
}
