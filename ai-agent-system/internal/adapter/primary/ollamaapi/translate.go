package ollamaapi

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"aiagentsystem/internal/domain/model"
)

func generateRequestToQuestion(req GenerateRequest, userID string) model.Question {
	return model.Question{
		Messages: []model.Message{{Role: model.RoleUser, Content: req.Prompt}},
		UserID:   userID,
	}
}

func chatRequestToQuestion(req ChatRequest, userID string) model.Question {
	messages := make([]model.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, model.Message{Role: m.Role, Content: m.Content})
	}
	return model.Question{Messages: messages, UserID: userID}
}

func wantsStream(stream *bool) bool {
	return stream == nil || *stream
}

func generateDeltaChunk(modelName, text string) GenerateResponse {
	return GenerateResponse{
		Model:     modelName,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Response:  text,
		Done:      false,
	}
}

func generateFinalChunk(modelName string, answer model.Answer, elapsed time.Duration, includeText bool) GenerateResponse {
	resp := GenerateResponse{
		Model:           modelName,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Done:            true,
		Context:         []int{},
		TotalDuration:   elapsed.Nanoseconds(),
		PromptEvalCount: answer.InputTokens,
		EvalCount:       answer.OutputTokens,
	}
	if includeText {
		resp.Response = answer.Text
	}
	return resp
}

func chatDeltaChunk(modelName, content string) ChatResponse {
	return ChatResponse{
		Model:     modelName,
		CreatedAt: time.Now().UTC().Format(time.RFC3339Nano),
		Message:   OllamaMessage{Role: model.RoleAssistant, Content: content},
		Done:      false,
	}
}

func chatFinalChunk(modelName string, answer model.Answer, elapsed time.Duration, includeText bool) ChatResponse {
	content := ""
	if includeText {
		content = answer.Text
	}
	return ChatResponse{
		Model:           modelName,
		CreatedAt:       time.Now().UTC().Format(time.RFC3339Nano),
		Message:         OllamaMessage{Role: model.RoleAssistant, Content: content},
		Done:            true,
		TotalDuration:   elapsed.Nanoseconds(),
		PromptEvalCount: answer.InputTokens,
		EvalCount:       answer.OutputTokens,
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
