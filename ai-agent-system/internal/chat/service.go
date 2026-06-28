package chat

import (
	"context"
	"log/slog"
	"time"

	"aiagentsystem/internal/history"
	"aiagentsystem/internal/proxy"
)

const (
	defaultUserID      = "anonymous"
	saveHistoryTimeout = 5 * time.Second
)

type Service struct {
	ollama  proxy.OllamaClient
	history *history.Store
	logger  *slog.Logger
}

func NewService(ollama proxy.OllamaClient, hist *history.Store, logger *slog.Logger) *Service {
	return &Service{ollama: ollama, history: hist, logger: logger}
}

func (s *Service) AnswerQuestion(ctx context.Context, req AnswerRequest) (AnswerResult, error) {
	question := req.LastUserQuestion()
	if question == "" {
		return AnswerResult{}, NewBadRequestError("question must not be empty")
	}

	result, err := s.ollama.Complete(ctx, toCompletionRequest(req))
	if err != nil {
		return AnswerResult{}, wrapProxyError(err)
	}
	answer := toAnswerResult(result)
	s.saveHistory(req.UserID, question, answer)
	return answer, nil
}

func (s *Service) AnswerQuestionStream(ctx context.Context, req AnswerRequest, onDelta func(delta string)) (AnswerResult, error) {
	question := req.LastUserQuestion()
	if question == "" {
		return AnswerResult{}, NewBadRequestError("question must not be empty")
	}

	stream, err := s.ollama.Stream(ctx, toCompletionRequest(req))
	if err != nil {
		return AnswerResult{}, wrapProxyError(err)
	}
	defer stream.Close()

	for stream.Next() {
		if delta := stream.Current(); delta != "" {
			onDelta(delta)
		}
	}
	if err := stream.Err(); err != nil {
		return AnswerResult{}, wrapProxyError(err)
	}

	answer := toAnswerResult(stream.Result())
	s.saveHistory(req.UserID, question, answer)
	return answer, nil
}

func (s *Service) saveHistory(userID, question string, answer AnswerResult) {
	if s.history == nil {
		return
	}
	if userID == "" {
		userID = defaultUserID
	}

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), saveHistoryTimeout)
		defer cancel()
		if err := s.history.SaveTurn(ctx, userID, question, answer.Text, string(answer.Source)); err != nil {
			if s.logger != nil {
				s.logger.Error("failed to save chat history", "error", err, "user_id", userID)
			}
		}
	}()
}

func toCompletionRequest(req AnswerRequest) proxy.CompletionRequest {
	messages := make([]proxy.Message, 0, len(req.Messages))
	for _, m := range req.Messages {
		messages = append(messages, proxy.Message{Role: m.Role, Content: m.Content})
	}
	return proxy.CompletionRequest{Messages: messages, UserID: req.UserID}
}

func toAnswerResult(result proxy.CompletionResult) AnswerResult {
	return AnswerResult{
		Text:         result.Text,
		Source:       SourceOllama,
		StopReason:   result.StopReason,
		InputTokens:  result.InputTokens,
		OutputTokens: result.OutputTokens,
	}
}
