package application_test

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/application"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

func newApp(t *testing.T) *application.Application {
	t.Helper()
	app, err := application.New(uuid.New(), uuid.New(), uuid.New())
	if err != nil {
		t.Fatalf("application.New failed: %v", err)
	}
	return app
}

func TestNew_Success(t *testing.T) {
	app := newApp(t)
	if app.Status != application.StatusApplied {
		t.Errorf("expected 'applied', got %s", app.Status)
	}
	if len(app.DomainEvents()) != 1 || app.DomainEvents()[0].EventType != "ApplicationCreated" {
		t.Error("expected ApplicationCreated event")
	}
}

func TestNew_MissingIDs(t *testing.T) {
	_, err := application.New(uuid.Nil, uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected error for nil job_id")
	}
}

func TestMoveToStage_Success(t *testing.T) {
	app := newApp(t)
	stageID := uuid.New()

	if err := app.MoveToStage(stageID, application.StatusCVReview); err != nil {
		t.Fatalf("MoveToStage failed: %v", err)
	}
	if app.Status != application.StatusCVReview {
		t.Errorf("expected cv_review, got %s", app.Status)
	}
	if app.DaysInStage != 0 {
		t.Error("days_in_stage should reset on move")
	}
}

func TestMoveToStage_TerminalBlocked(t *testing.T) {
	app := newApp(t)
	_ = app.Reject(application.RejectionSkillMismatch, "")
	if err := app.MoveToStage(uuid.New(), application.StatusCVReview); err == nil {
		t.Fatal("expected error moving out of rejected terminal state")
	}
}

func TestReject_Success(t *testing.T) {
	app := newApp(t)
	if err := app.Reject(application.RejectionSkillMismatch, "skill gap"); err != nil {
		t.Fatalf("Reject failed: %v", err)
	}
	if app.Status != application.StatusRejected {
		t.Errorf("expected rejected, got %s", app.Status)
	}
	if app.RejectionReason == nil || *app.RejectionReason != application.RejectionSkillMismatch {
		t.Error("rejection reason not set correctly")
	}
}

func TestReject_HiredBlocked(t *testing.T) {
	app := newApp(t)
	// Simulate full happy path to hired
	_ = app.MoveToStage(uuid.New(), application.StatusFinalRound)
	offer := application.Offer{
		ID:        uuid.New(),
		Salary:    shared.Money{Amount: 30_000_000, Currency: "VND"},
		StartDate: time.Now().Add(30 * 24 * time.Hour),
		Title:     "Backend Engineer",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	_ = app.ExtendOffer(offer)
	_ = app.AcceptOffer()
	_ = app.MarkHired()

	if err := app.Reject(application.RejectionSkillMismatch, ""); err == nil {
		t.Fatal("expected error rejecting a hired application")
	}
}

func TestExtendOffer_WrongStage(t *testing.T) {
	app := newApp(t)
	offer := application.Offer{ID: uuid.New(), ExpiresAt: time.Now().Add(7 * 24 * time.Hour)}
	if err := app.ExtendOffer(offer); err == nil {
		t.Fatal("expected error extending offer when not in final_round")
	}
}

func TestExtendOffer_Success(t *testing.T) {
	app := newApp(t)
	_ = app.MoveToStage(uuid.New(), application.StatusFinalRound)

	offer := application.Offer{
		ID:        uuid.New(),
		Salary:    shared.Money{Amount: 30_000_000, Currency: "VND"},
		StartDate: time.Now().Add(30 * 24 * time.Hour),
		Title:     "Senior Engineer",
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	}
	if err := app.ExtendOffer(offer); err != nil {
		t.Fatalf("ExtendOffer failed: %v", err)
	}
	if app.Status != application.StatusOffer {
		t.Errorf("expected offer status, got %s", app.Status)
	}
	if app.Offer == nil {
		t.Fatal("offer should be set on application")
	}
	if app.Offer.SentAt == nil {
		t.Error("sent_at should be set when offer is extended")
	}
}

func TestFullHappyPath(t *testing.T) {
	app := newApp(t)
	stages := []struct {
		stageID uuid.UUID
		status  application.Status
	}{
		{uuid.New(), application.StatusCVReview},
		{uuid.New(), application.StatusPhoneScreen},
		{uuid.New(), application.StatusTechnical},
		{uuid.New(), application.StatusFinalRound},
	}
	for _, s := range stages {
		if err := app.MoveToStage(s.stageID, s.status); err != nil {
			t.Fatalf("MoveToStage to %s failed: %v", s.status, err)
		}
	}

	offer := application.Offer{
		ID:        uuid.New(),
		Salary:    shared.Money{Amount: 40_000_000, Currency: "VND"},
		StartDate: time.Now().Add(30 * 24 * time.Hour),
		Title:     "Lead Engineer",
		ExpiresAt: time.Now().Add(5 * 24 * time.Hour),
	}
	if err := app.ExtendOffer(offer); err != nil {
		t.Fatalf("ExtendOffer failed: %v", err)
	}
	if err := app.AcceptOffer(); err != nil {
		t.Fatalf("AcceptOffer failed: %v", err)
	}
	if err := app.MarkHired(); err != nil {
		t.Fatalf("MarkHired failed: %v", err)
	}
	if app.Status != application.StatusHired {
		t.Errorf("expected hired, got %s", app.Status)
	}

	// Verify CandidateHired event was emitted
	found := false
	for _, e := range app.DomainEvents() {
		if e.EventType == "CandidateHired" {
			found = true
		}
	}
	if !found {
		t.Error("expected CandidateHired domain event")
	}
}

func TestAddInterview_Success(t *testing.T) {
	app := newApp(t)
	interview := application.Interview{
		ID:          uuid.New(),
		Round:       1,
		Title:       "Phone Screen",
		ScheduledAt: time.Now().Add(48 * time.Hour),
		DurationMin: 30,
		Type:        "online",
	}
	app.AddInterview(interview)

	if len(app.Interviews) != 1 {
		t.Errorf("expected 1 interview, got %d", len(app.Interviews))
	}

	found := false
	for _, e := range app.DomainEvents() {
		if e.EventType == "InterviewScheduled" {
			found = true
		}
	}
	if !found {
		t.Error("expected InterviewScheduled event")
	}
}

func TestSubmitFeedback_Success(t *testing.T) {
	app := newApp(t)
	ivID := uuid.New()
	interview := application.Interview{
		ID:          ivID,
		Round:       1,
		ScheduledAt: time.Now().Add(48 * time.Hour),
	}
	app.AddInterview(interview)

	feedback := application.InterviewFeedback{
		SubmittedBy: uuid.New(),
		Decision:    "pass",
		Score:       4,
		Strengths:   "Strong Go skills",
	}
	if err := app.SubmitFeedback(ivID, feedback); err != nil {
		t.Fatalf("SubmitFeedback failed: %v", err)
	}
	if app.Interviews[0].Feedback == nil {
		t.Error("feedback should be stored on interview")
	}
}

func TestWithdraw(t *testing.T) {
	app := newApp(t)
	app.Withdraw("accepted another offer")
	if app.Status != application.StatusWithdrawn {
		t.Errorf("expected withdrawn, got %s", app.Status)
	}
	if app.WithdrawReason == "" {
		t.Error("withdraw reason should be set")
	}
}
