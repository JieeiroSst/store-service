package chat

import (
	"context"
	"log/slog"
	"time"

	"aiagentsystem/internal/faq"
	"aiagentsystem/internal/history"
	"aiagentsystem/internal/proxy"
)

const (
	defaultUserID      = "anonymous"
	saveHistoryTimeout = 5 * time.Second
)

type Service struct {
	ollama   proxy.OllamaClient
	embedder proxy.EmbeddingClient
	faqs     *faq.Store
	history  *history.Store
	logger   *slog.Logger
}

func NewService(ollama proxy.OllamaClient, embedder proxy.EmbeddingClient, faqs *faq.Store, hist *history.Store, logger *slog.Logger) *Service {
	return &Service{ollama: ollama, embedder: embedder, faqs: faqs, history: hist, logger: logger}
}

func (s *Service) AnswerQuestion(ctx context.Context, req AnswerRequest) (AnswerResult, error) {
	question := req.LastUserQuestion()
	if question == "" {
		return AnswerResult{}, NewBadRequestError("question must not be empty")
	}

	if answer, ok := s.matchFAQ(ctx, question); ok {
		s.saveHistory(req.UserID, question, answer)
		return answer, nil
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

	if answer, ok := s.matchFAQ(ctx, question); ok {
		onDelta(answer.Text)
		s.saveHistory(req.UserID, question, answer)
		return answer, nil
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

func (s *Service) matchFAQ(ctx context.Context, question string) (AnswerResult, bool) {
	entry, score, ok, err := s.faqs.Match(ctx, s.embedder, question)
	if err != nil {
		if s.logger != nil {
			s.logger.WarnContext(ctx, "faq match failed, falling through to ollama", "error", err)
		}
		return AnswerResult{}, false
	}
	if !ok {
		return AnswerResult{}, false
	}
	if s.logger != nil {
		s.logger.DebugContext(ctx, "faq match", "score", score, "question", entry.Question)
	}
	return AnswerResult{Text: entry.Answer, Source: SourceFAQ}, true
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
	return proxy.CompletionRequest{Messages: messages}
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
