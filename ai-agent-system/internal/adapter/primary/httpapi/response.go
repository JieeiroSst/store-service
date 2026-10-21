package httpapi

import (
	"encoding/json"
	"net/http"

	"aiagentsystem/internal/domain/model"
)

const userIDHeader = "X-User-Id"

type errorBody struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func errorStatus(err error) (status int, errType string, message string) {
	e := model.AsError(err)
	switch e.Kind {
	case model.KindRateLimited:
		return http.StatusTooManyRequests, "rate_limited", "rate limit exceeded, please try again shortly"
	case model.KindUpstreamUnavailable:
		return http.StatusServiceUnavailable, "upstream_unavailable", "the AI service is temporarily unavailable"
	case model.KindBadRequest:
		return http.StatusBadRequest, "bad_request", e.Message
	}
	return http.StatusInternalServerError, "internal", "an internal error occurred"
}

func writeError(w http.ResponseWriter, err error) {
	status, errType, message := errorStatus(err)
	writeJSON(w, status, errorResponse{Error: errorBody{Message: message, Type: errType}})
}

func userIDFromRequest(r *http.Request) string {
	return r.Header.Get(userIDHeader)
}
