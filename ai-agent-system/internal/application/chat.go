package application

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"aiagentsystem/internal/domain/model"
	"aiagentsystem/internal/domain/port"
)

const saveHistoryTimeout = 5 * time.Second

type ChatService struct {
	llm     port.LLMClient
	history port.HistoryRepository
	logger  *slog.Logger
	pending sync.WaitGroup
}

var _ port.ChatService = (*ChatService)(nil)

func NewChatService(llm port.LLMClient, history port.HistoryRepository, logger *slog.Logger) *ChatService {
	if logger == nil {
		logger = slog.Default()
	}
	return &ChatService{llm: llm, history: history, logger: logger}
}

func (s *ChatService) Answer(ctx context.Context, q model.Question) (model.Answer, error) {
	question := q.LastUserQuestion()
	if question == "" {
		return model.Answer{}, model.BadRequest("question must not be empty")
	}

	completion, err := s.llm.Complete(ctx, q)
	if err != nil {
		return model.Answer{}, model.AsError(err)
	}

	answer := model.Answer{Completion: completion, Source: model.SourceOllama}
	s.record(q.UserID, question, answer)
	return answer, nil
}

func (s *ChatService) AnswerStream(ctx context.Context, q model.Question, onDelta func(delta string)) (model.Answer, error) {
	question := q.LastUserQuestion()
	if question == "" {
		return model.Answer{}, model.BadRequest("question must not be empty")
	}

	stream, err := s.llm.Stream(ctx, q)
	if err != nil {
		return model.Answer{}, model.AsError(err)
	}
	defer stream.Close()

	for stream.Next() {
		if delta := stream.Current(); delta != "" {
			onDelta(delta)
		}
	}
	if err := stream.Err(); err != nil {
		return model.Answer{}, model.AsError(err)
	}

	answer := model.Answer{Completion: stream.Result(), Source: model.SourceOllama}
	s.record(q.UserID, question, answer)
	return answer, nil
}

func (s *ChatService) Wait() {
	s.pending.Wait()
}

func (s *ChatService) record(userID, question string, answer model.Answer) {
	if s.history == nil {
		return
	}
	if userID == "" {
		userID = model.AnonymousUserID
	}
	turn := model.Turn{UserID: userID, Question: question, Answer: answer.Text, Source: answer.Source}

	s.pending.Add(1)
	go func() {
		defer s.pending.Done()
		ctx, cancel := context.WithTimeout(context.Background(), saveHistoryTimeout)
		defer cancel()
		if err := s.history.SaveTurn(ctx, turn); err != nil {
			s.logger.Error("failed to save chat history", "error", err, "user_id", userID)
		}
	}()
}
