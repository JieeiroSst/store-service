package ollamaapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"aiagentsystem/internal/chat"
)

type Handlers struct {
	service   *chat.Service
	modelName string
}

func NewHandlers(service *chat.Service, modelName string) *Handlers {
	return &Handlers{service: service, modelName: modelName}
}

const userIDHeader = "X-User-Id"

func userIDFromRequest(r *http.Request) string {
	return r.Header.Get(userIDHeader)
}

func writeOllamaError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(OllamaErrorBody{Error: message})
}

func ollamaErrorStatusAndMessage(err error) (int, string) {
	var cErr *chat.Error
	if errors.As(err, &cErr) {
		switch cErr.Kind {
		case chat.ErrKindRateLimited:
			return http.StatusTooManyRequests, "rate limit exceeded, please try again"
		case chat.ErrKindUpstreamUnavailable:
			return http.StatusServiceUnavailable, "model temporarily unavailable"
		case chat.ErrKindBadRequest:
			return http.StatusBadRequest, cErr.Message
		}
	}
	return http.StatusInternalServerError, "internal error"
}

// Generate handles POST /api/generate.
func (h *Handlers) Generate(w http.ResponseWriter, r *http.Request) {
	var req GenerateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOllamaError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Prompt == "" {
		writeOllamaError(w, http.StatusBadRequest, "prompt must not be empty")
		return
	}

	model := h.resolveModel(req.Model)
	start := time.Now()

	if wantsStream(req.Stream) {
		h.streamGenerate(w, r, model, req, start)
		return
	}

	answerReq := generateRequestToAnswerRequest(req)
	answerReq.UserID = userIDFromRequest(r)

	result, err := h.service.AnswerQuestion(r.Context(), answerReq)
	if err != nil {
		status, msg := ollamaErrorStatusAndMessage(err)
		writeOllamaError(w, status, msg)
		return
	}

	writeJSONOK(w, generateFinalChunk(model, result, time.Since(start), true))
}

func (h *Handlers) streamGenerate(w http.ResponseWriter, r *http.Request, model string, req GenerateRequest, start time.Time) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeOllamaError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)

	answerReq := generateRequestToAnswerRequest(req)
	answerReq.UserID = userIDFromRequest(r)

	result, err := h.service.AnswerQuestionStream(r.Context(), answerReq, func(delta string) {
		writeNDJSONLine(w, generateDeltaChunk(model, delta))
		flusher.Flush()
	})
	if err != nil {
		// Status is already 200 and streaming has begun; Ollama clients
		// expect errors as a final NDJSON line, not a status-code change.
		_, msg := ollamaErrorStatusAndMessage(err)
		writeNDJSONLine(w, OllamaErrorBody{Error: msg})
		flusher.Flush()
		return
	}

	writeNDJSONLine(w, generateFinalChunk(model, result, time.Since(start), false))
	flusher.Flush()
}

// Chat handles POST /api/chat.
func (h *Handlers) Chat(w http.ResponseWriter, r *http.Request) {
	var req ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOllamaError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(req.Messages) == 0 {
		writeOllamaError(w, http.StatusBadRequest, "messages must not be empty")
		return
	}

	model := h.resolveModel(req.Model)
	start := time.Now()

	if wantsStream(req.Stream) {
		h.streamChat(w, r, model, req, start)
		return
	}

	answerReq := chatRequestToAnswerRequest(req)
	answerReq.UserID = userIDFromRequest(r)

	result, err := h.service.AnswerQuestion(r.Context(), answerReq)
	if err != nil {
		status, msg := ollamaErrorStatusAndMessage(err)
		writeOllamaError(w, status, msg)
		return
	}

	writeJSONOK(w, chatFinalChunk(model, result, time.Since(start), true))
}

func (h *Handlers) streamChat(w http.ResponseWriter, r *http.Request, model string, req ChatRequest, start time.Time) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeOllamaError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)

	answerReq := chatRequestToAnswerRequest(req)
	answerReq.UserID = userIDFromRequest(r)

	result, err := h.service.AnswerQuestionStream(r.Context(), answerReq, func(delta string) {
		writeNDJSONLine(w, chatDeltaChunk(model, delta))
		flusher.Flush()
	})
	if err != nil {
		_, msg := ollamaErrorStatusAndMessage(err)
		writeNDJSONLine(w, OllamaErrorBody{Error: msg})
		flusher.Flush()
		return
	}

	writeNDJSONLine(w, chatFinalChunk(model, result, time.Since(start), false))
	flusher.Flush()
}

// Tags handles GET /api/tags — lists the one model this server relays to.
func (h *Handlers) Tags(w http.ResponseWriter, r *http.Request) {
	writeJSONOK(w, TagsResponse{
		Models: []tagsModel{
			{
				Name:       h.modelName,
				Model:      h.modelName,
				ModifiedAt: startupTime.UTC().Format(time.RFC3339),
				Size:       0,
				Digest:     syntheticDigest(h.modelName),
				Details:    relayModelDetails(),
			},
		},
	})
}

// Show handles POST /api/show.
func (h *Handlers) Show(w http.ResponseWriter, r *http.Request) {
	var req ShowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeOllamaError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Model != h.modelName {
		writeOllamaError(w, http.StatusNotFound, "model '"+req.Model+"' not found")
		return
	}

	writeJSONOK(w, ShowResponse{
		Modelfile:  "# Relayed to a real Ollama server (model: " + h.modelName + ")",
		Parameters: "",
		Template:   "{{ .Prompt }}",
		Details:    relayModelDetails(),
		ModelInfo: map[string]any{
			"general.architecture": "ollama",
		},
		Capabilities: []string{"completion"},
	})
}

func (h *Handlers) resolveModel(requested string) string {
	return h.modelName
}

func writeJSONOK(w http.ResponseWriter, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(body)
}

func writeNDJSONLine(w http.ResponseWriter, body any) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return
	}
	w.Write(encoded)
	w.Write([]byte("\n"))
}

var startupTime = time.Now()
