package application

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"reflect"
	"testing"

	"aiagentsystem/internal/domain/model"
)

func newTestChatService(llm *fakeLLM, hist *fakeHistory) *ChatService {
	return NewChatService(llm, hist, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func question(userID, content string) model.Question {
	return model.Question{UserID: userID, Messages: []model.Message{{Role: model.RoleUser, Content: content}}}
}

func TestAnswerRejectsEmptyQuestion(t *testing.T) {
	llm := &fakeLLM{}
	svc := newTestChatService(llm, &fakeHistory{})

	_, err := svc.Answer(context.Background(), model.Question{})

	if got := model.AsError(err).Kind; got != model.KindBadRequest {
		t.Fatalf("kind = %v, want bad request", got)
	}
	if len(llm.calls) != 0 {
		t.Fatalf("llm called %d times, want 0", len(llm.calls))
	}
}

func TestAnswerReturnsCompletionAndRecordsTurn(t *testing.T) {
	llm := &fakeLLM{completion: model.Completion{Text: "hello", StopReason: "stop", InputTokens: 3, OutputTokens: 5}}
	hist := &fakeHistory{}
	svc := newTestChatService(llm, hist)

	answer, err := svc.Answer(context.Background(), question("", "hi"))
	svc.Wait()

	if err != nil {
		t.Fatalf("Answer: %v", err)
	}
	if answer.Text != "hello" || answer.Source != model.SourceOllama || answer.OutputTokens != 5 {
		t.Fatalf("answer = %+v", answer)
	}
	want := []model.Turn{{UserID: model.AnonymousUserID, Question: "hi", Answer: "hello", Source: model.SourceOllama}}
	if got := hist.savedTurns(); !reflect.DeepEqual(got, want) {
		t.Fatalf("turns = %+v, want %+v", got, want)
	}
}

func TestAnswerKeepsDomainErrorKind(t *testing.T) {
	llm := &fakeLLM{err: model.NewError(model.KindRateLimited, "slow down", nil)}
	hist := &fakeHistory{}
	svc := newTestChatService(llm, hist)

	_, err := svc.Answer(context.Background(), question("alice", "hi"))
	svc.Wait()

	if got := model.AsError(err).Kind; got != model.KindRateLimited {
		t.Fatalf("kind = %v, want rate limited", got)
	}
	if len(hist.savedTurns()) != 0 {
		t.Fatal("failed turn must not be recorded")
	}
}

func TestAnswerWrapsUnknownErrorAsInternal(t *testing.T) {
	cause := errors.New("boom")
	svc := newTestChatService(&fakeLLM{err: cause}, &fakeHistory{})

	_, err := svc.Answer(context.Background(), question("alice", "hi"))

	e := model.AsError(err)
	if e.Kind != model.KindInternal || !errors.Is(err, cause) {
		t.Fatalf("err = %v (kind %v), want internal wrapping cause", err, e.Kind)
	}
}

func TestAnswerSurvivesHistoryFailure(t *testing.T) {
	llm := &fakeLLM{completion: model.Completion{Text: "ok"}}
	svc := newTestChatService(llm, &fakeHistory{saveErr: errors.New("db down")})

	answer, err := svc.Answer(context.Background(), question("alice", "hi"))
	svc.Wait()

	if err != nil || answer.Text != "ok" {
		t.Fatalf("answer = %+v, err = %v", answer, err)
	}
}

func TestAnswerStreamForwardsDeltas(t *testing.T) {
	llm := &fakeLLM{deltas: []string{"he", "", "llo"}, completion: model.Completion{Text: "hello", StopReason: "stop"}}
	hist := &fakeHistory{}
	svc := newTestChatService(llm, hist)

	var got []string
	answer, err := svc.AnswerStream(context.Background(), question("bob", "hi"), func(d string) { got = append(got, d) })
	svc.Wait()

	if err != nil {
		t.Fatalf("AnswerStream: %v", err)
	}
	if !reflect.DeepEqual(got, []string{"he", "llo"}) {
		t.Fatalf("deltas = %v", got)
	}
	if answer.Text != "hello" || answer.Source != model.SourceOllama {
		t.Fatalf("answer = %+v", answer)
	}
	if turns := hist.savedTurns(); len(turns) != 1 || turns[0].UserID != "bob" {
		t.Fatalf("turns = %+v", turns)
	}
}

func TestAnswerStreamReportsStreamError(t *testing.T) {
	llm := &fakeLLM{deltas: []string{"partial"}, streamErr: model.NewError(model.KindUpstreamUnavailable, "gone", nil)}
	hist := &fakeHistory{}
	svc := newTestChatService(llm, hist)

	_, err := svc.AnswerStream(context.Background(), question("bob", "hi"), func(string) {})
	svc.Wait()

	if got := model.AsError(err).Kind; got != model.KindUpstreamUnavailable {
		t.Fatalf("kind = %v, want upstream unavailable", got)
	}
	if len(hist.savedTurns()) != 0 {
		t.Fatal("failed stream must not be recorded")
	}
}
