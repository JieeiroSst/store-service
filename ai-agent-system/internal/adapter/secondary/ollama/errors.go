package ollama

import (
	"fmt"
	"net/http"

	"aiagentsystem/internal/domain/model"
)

func unreachableError(err error) error {
	return model.NewError(model.KindUpstreamUnavailable, "failed to reach the local Ollama server", err)
}

func internalError(message string, err error) error {
	return model.NewError(model.KindInternal, message, err)
}

func statusError(statusCode int, body string) error {
	kind := model.KindInternal
	switch {
	case statusCode == http.StatusTooManyRequests:
		kind = model.KindRateLimited
	case statusCode == http.StatusBadRequest, statusCode == http.StatusUnprocessableEntity, statusCode == http.StatusNotFound:
		kind = model.KindBadRequest
	case statusCode >= 500:
		kind = model.KindUpstreamUnavailable
	}
	return model.NewError(kind,
		"the local Ollama server rejected or failed to process the request",
		fmt.Errorf("ollama: status %d: %s", statusCode, body))
}
