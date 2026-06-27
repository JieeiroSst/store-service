package ollamaapi

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"aiagentsystem/internal/chat"
)

func generateRequestToAnswerRequest(req GenerateRequest) chat.AnswerRequest {
	return chat.AnswerRequest{
		Messages: []chat.Message{{Role: "user", Content: req.Prompt}},
	}
}

func chatRequestToAnswerRequest(req ChatRequest) chat.AnswerRequest {
	messages := make([]chat.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, chat.Message{Role: m.Role, Content: m.Content})
	}
	return chat.AnswerRequest{Messages: messages}
}

// wantsStream reports whether the request asked for streaming. Ollama
// defaults stream to true when the field is omitted.
func wantsStream(stream *bool) bool {
	return stream == nil || *stream
}

// generateDeltaChunk builds one NDJSON line for an in-progress (done:false)
// streamed token chunk.
func generateDeltaChunk(model, text string) GenerateResponse {
	return GenerateResponse{
		Model:     model,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Response:  text,
		Done:      false,
	}
}

// generateFinalChunk builds the final line (done:true) for both streaming
// and non-streaming /api/generate responses, with stats derived from
// the backend's actual token usage.
func generateFinalChunk(model string, result chat.AnswerResult, elapsed time.Duration, includeText bool) GenerateResponse {
	resp := GenerateResponse{
		Model:           model,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Done:            true,
		Context:         []int{},
		TotalDuration:   elapsed.Nanoseconds(),
		PromptEvalCount: result.InputTokens,
		EvalCount:       result.OutputTokens,
	}
	if includeText {
		resp.Response = result.Text
	}
	return resp
}

// chatDeltaChunk builds one NDJSON line for an in-progress streamed message chunk.
func chatDeltaChunk(model, content string) ChatResponse {
	return ChatResponse{
		Model:     model,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Message:   OllamaMessage{Role: "assistant", Content: content},
		Done:      false,
	}
}

// chatFinalChunk builds the final line for both streaming and non-streaming
// /api/chat responses.
func chatFinalChunk(model string, result chat.AnswerResult, elapsed time.Duration, includeText bool) ChatResponse {
	content := ""
	if includeText {
		content = result.Text
	}
	return ChatResponse{
		Model:           model,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Message:         OllamaMessage{Role: "assistant", Content: content},
		Done:            true,
		TotalDuration:   elapsed.Nanoseconds(),
		PromptEvalCount: result.InputTokens,
		EvalCount:       result.OutputTokens,
	}
}

func syntheticDigest(modelName string) string {
	sum := sha256.Sum256([]byte("relay:" + modelName))
	return hex.EncodeToString(sum[:])
}

func relayModelDetails() modelDetails {
	return modelDetails{
		Format:            "gguf",
		Family:            "ollama",
		Families:          []string{"ollama"},
		ParameterSize:     "N/A",
		QuantizationLevel: "N/A",
	}
}
