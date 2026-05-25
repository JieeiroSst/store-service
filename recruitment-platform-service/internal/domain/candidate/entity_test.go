package candidate_test

import (
	"testing"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
)

func TestNew_ValidInput(t *testing.T) {
	c, err := candidate.New("Nguyen Van A", "a@example.com", "0901234567", candidate.SourceLinkedIn)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if c.Status != candidate.StatusNew {
		t.Errorf("expected status 'new', got %s", c.Status)
	}
	if len(c.DomainEvents()) != 1 {
		t.Errorf("expected 1 domain event, got %d", len(c.DomainEvents()))
	}
	if c.DomainEvents()[0].EventType != "CandidateCreated" {
		t.Errorf("expected CandidateCreated event, got %s", c.DomainEvents()[0].EventType)
	}
}

func TestNew_MissingName(t *testing.T) {
	_, err := candidate.New("", "a@example.com", "", candidate.SourceDirect)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestNew_MissingEmail(t *testing.T) {
	_, err := candidate.New("Nguyen Van A", "", "", candidate.SourceDirect)
	if err == nil {
		t.Fatal("expected error for empty email")
	}
}

func TestTransitionTo_ValidPath(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)

	if err := c.TransitionTo(candidate.StatusScreening, "cv looks good"); err != nil {
		t.Fatalf("screening transition failed: %v", err)
	}
	if c.Status != candidate.StatusScreening {
		t.Errorf("expected screening, got %s", c.Status)
	}
}

func TestTransitionTo_InvalidPath(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)

	// Cannot jump directly from new → offer
	if err := c.TransitionTo(candidate.StatusOffer, ""); err == nil {
		t.Fatal("expected error for invalid transition new→offer")
	}
}

func TestTransitionTo_TerminalStateBlocked(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)
	_ = c.TransitionTo(candidate.StatusScreening, "")
	_ = c.TransitionTo(candidate.StatusRejected, "")

	// Cannot transition out of rejected
	if err := c.TransitionTo(candidate.StatusScreening, ""); err == nil {
		t.Fatal("expected error transitioning out of rejected")
	}
}

func TestUpdateAIScore(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)
	score := shared.AIScore{
		Score:      87.5,
		Confidence: 0.9,
		Breakdown:  map[string]float64{"skills": 90, "experience": 85},
		ModelID:    "gpt-4o",
	}
	c.UpdateAIScore(score)

	if c.AIScore == nil {
		t.Fatal("expected AI score to be set")
	}
	if c.AIScore.Score != 87.5 {
		t.Errorf("expected score 87.5, got %f", c.AIScore.Score)
	}

	events := c.DomainEvents()
	found := false
	for _, e := range events {
		if e.EventType == "CandidateScoredByAI" {
			found = true
		}
	}
	if !found {
		t.Error("expected CandidateScoredByAI event")
	}
}

func TestAddTag_Dedup(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)
	c.AddTag("golang")
	c.AddTag("golang") // duplicate
	c.AddTag("postgres")

	if len(c.Tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(c.Tags))
	}
}

func TestClearEvents(t *testing.T) {
	c, _ := candidate.New("Test", "t@t.com", "", candidate.SourceDirect)
	_ = c.TransitionTo(candidate.StatusScreening, "")

	if len(c.DomainEvents()) == 0 {
		t.Fatal("expected events before clear")
	}
	c.ClearEvents()
	if len(c.DomainEvents()) != 0 {
		t.Errorf("expected 0 events after clear, got %d", len(c.DomainEvents()))
	}
}
