package ollamaapi

import (
	"encoding/json"
	"net/http"
	"time"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

const userIDHeader = "X-User-Id"

var startupTime = time.Now()

type Handlers struct {
	chat      port.ChatService
	modelName string
}

func NewHandlers(chat port.ChatService, modelName string) *Handlers {
	return &Handlers{chat: chat, modelName: modelName}
}

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

	start := time.Now()
	q := generateRequestToQuestion(req, r.Header.Get(userIDHeader))

	if wantsStream(req.Stream) {
		h.stream(w, r, q, func(delta string) any {
			return generateDeltaChunk(h.modelName, delta)
		}, func(answer model.Answer) any {
			return generateFinalChunk(h.modelName, answer, time.Since(start), false)
		})
		return
	}

	answer, err := h.chat.Answer(r.Context(), q)
	if err != nil {
		status, msg := errorStatus(err)
		writeOllamaError(w, status, msg)
		return
	}
	writeJSONOK(w, generateFinalChunk(h.modelName, answer, time.Since(start), true))
}

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

	start := time.Now()
	q := chatRequestToQuestion(req, r.Header.Get(userIDHeader))

	if wantsStream(req.Stream) {
		h.stream(w, r, q, func(delta string) any {
			return chatDeltaChunk(h.modelName, delta)
		}, func(answer model.Answer) any {
			return chatFinalChunk(h.modelName, answer, time.Since(start), false)
		})
		return
	}

	answer, err := h.chat.Answer(r.Context(), q)
	if err != nil {
		status, msg := errorStatus(err)
		writeOllamaError(w, status, msg)
		return
	}
	writeJSONOK(w, chatFinalChunk(h.modelName, answer, time.Since(start), true))
}

func (h *Handlers) stream(w http.ResponseWriter, r *http.Request, q model.Question, delta func(string) any, final func(model.Answer) any) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeOllamaError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	w.WriteHeader(http.StatusOK)

	answer, err := h.chat.AnswerStream(r.Context(), q, func(d string) {
		writeNDJSONLine(w, delta(d))
		flusher.Flush()
	})
	if err != nil {
		_, msg := errorStatus(err)
		writeNDJSONLine(w, OllamaErrorBody{Error: msg})
		flusher.Flush()
		return
	}

	writeNDJSONLine(w, final(answer))
	flusher.Flush()
}

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

func errorStatus(err error) (int, string) {
	e := model.AsError(err)
	switch e.Kind {
	case model.KindRateLimited:
		return http.StatusTooManyRequests, "rate limit exceeded, please try again"
	case model.KindUpstreamUnavailable:
		return http.StatusServiceUnavailable, "model temporarily unavailable"
	case model.KindBadRequest:
		return http.StatusBadRequest, e.Message
	}
	return http.StatusInternalServerError, "internal error"
}

func writeOllamaError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(OllamaErrorBody{Error: message})
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
