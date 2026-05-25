package candidate

import (
	"errors"
	"time"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

type Status string

const (
	StatusNew       Status = "new"
	StatusScreening Status = "screening"
	StatusInterview Status = "interview"
	StatusOffer     Status = "offer"
	StatusHired     Status = "hired"
	StatusRejected  Status = "rejected"
	StatusWithdrawn Status = "withdrawn"
	StatusBlacklist Status = "blacklist"
)

func (s Status) IsValid() bool {
	switch s {
	case StatusNew, StatusScreening, StatusInterview,
		StatusOffer, StatusHired, StatusRejected, StatusWithdrawn, StatusBlacklist:
		return true
	}
	return false
}

// valid lifecycle transitions
var allowedTransitions = map[Status][]Status{
	StatusNew:       {StatusScreening, StatusRejected},
	StatusScreening: {StatusInterview, StatusRejected, StatusWithdrawn},
	StatusInterview: {StatusOffer, StatusRejected, StatusWithdrawn},
	StatusOffer:     {StatusHired, StatusRejected, StatusWithdrawn},
	StatusHired:     {StatusBlacklist},
	StatusRejected:  {},
	StatusWithdrawn: {},
}

type SourceChannel string

const (
	SourceLinkedIn SourceChannel = "linkedin"
	SourceReferral SourceChannel = "referral"
	SourceJobBoard SourceChannel = "job_board"
	SourceDirect   SourceChannel = "direct"
	SourceAgency   SourceChannel = "agency"
)

type ExperienceLevel string

const (
	ExperienceFresher  ExperienceLevel = "fresher"
	ExperienceJunior   ExperienceLevel = "junior"
	ExperienceMidLevel ExperienceLevel = "mid"
	ExperienceSenior   ExperienceLevel = "senior"
	ExperienceLead     ExperienceLevel = "lead"
	ExperienceManager  ExperienceLevel = "manager"
)

type WorkExperience struct {
	Company     string     `json:"company"`
	Title       string     `json:"title"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date"`
	Description string     `json:"description"`
}

type Education struct {
	Institution string     `json:"institution"`
	Degree      string     `json:"degree"`
	Major       string     `json:"major"`
	GraduatedAt *time.Time `json:"graduated_at"`
}

type Candidate struct {
	shared.BaseEntity

	// Personal info
	FullName    string         `db:"full_name"    json:"full_name"`
	Email       string         `db:"email"        json:"email"`
	Phone       string         `db:"phone"        json:"phone"`
	AvatarURL   string         `db:"avatar_url"   json:"avatar_url"`
	LinkedInURL string         `db:"linkedin_url" json:"linkedin_url"`
	ResumeURL   string         `db:"resume_url"   json:"resume_url"`
	Location    shared.Address `db:"location" json:"location"`

	// Professional info
	CurrentTitle      string          `db:"current_title"       json:"current_title"`
	CurrentCompany    string          `db:"current_company"     json:"current_company"`
	YearsOfExperience int             `db:"years_of_experience" json:"years_of_experience"`
	ExperienceLevel   ExperienceLevel `db:"experience_level"    json:"experience_level"`
	Skills            []string        `db:"skills"              json:"skills"`
	ExpectedSalary    *shared.Money   `db:"expected_salary"     json:"expected_salary,omitempty"`
	NoticePeriodDays  int             `db:"notice_period_days"  json:"notice_period_days"`

	// Tracking
	Status       Status        `db:"status"         json:"status"`
	Source       SourceChannel `db:"source"         json:"source"`
	ReferredByID *uuid.UUID    `db:"referred_by_id" json:"referred_by_id,omitempty"`
	Tags         []string      `db:"tags"           json:"tags"`
	Notes        string        `db:"notes"          json:"notes"`

	// AI
	AIScore   *shared.AIScore `db:"ai_score"    json:"ai_score,omitempty"`
	Embedding []float32       `db:"embedding"   json:"-"` // pgvector

	// Relations (populated on demand)
	Experiences []WorkExperience `db:"-" json:"experiences,omitempty"`
	Educations  []Education      `db:"-" json:"educations,omitempty"`

	events []shared.DomainEvent
}

func New(fullName, email, phone string, source SourceChannel) (*Candidate, error) {
	if fullName == "" {
		return nil, errors.New("candidate: full name required")
	}
	if email == "" {
		return nil, errors.New("candidate: email required")
	}
	c := &Candidate{
		BaseEntity: shared.NewBaseEntity(),
		FullName:   fullName,
		Email:      email,
		Phone:      phone,
		Source:     source,
		Status:     StatusNew,
	}
	c.recordEvent("CandidateCreated", map[string]any{"candidate_id": c.ID})
	return c, nil
}

func (c *Candidate) TransitionTo(newStatus Status, reason string) error {
	allowed, ok := allowedTransitions[c.Status]
	if !ok {
		return errors.New("candidate: current status has no transitions defined")
	}
	for _, s := range allowed {
		if s == newStatus {
			c.Status = newStatus
			c.recordEvent("CandidateStatusChanged", map[string]any{
				"candidate_id": c.ID,
				"from":         c.Status,
				"to":           newStatus,
				"reason":       reason,
			})
			return nil
		}
	}
	return errors.New("candidate: transition not allowed from " + string(c.Status) + " to " + string(newStatus))
}

func (c *Candidate) UpdateAIScore(score shared.AIScore) {
	c.AIScore = &score
	c.recordEvent("CandidateScoredByAI", map[string]any{
		"candidate_id": c.ID,
		"score":        score.Score,
	})
}

func (c *Candidate) AddTag(tag string) {
	for _, t := range c.Tags {
		if t == tag {
			return
		}
	}
	c.Tags = append(c.Tags, tag)
}

func (c *Candidate) DomainEvents() []shared.DomainEvent {
	return c.events
}

func (c *Candidate) ClearEvents() {
	c.events = nil
}

func (c *Candidate) recordEvent(eventType string, payload interface{}) {
	c.events = append(c.events, shared.NewDomainEvent(eventType, payload))
}

type Filter struct {
	Status          *Status          `form:"status"`
	Source          *SourceChannel   `form:"source"`
	ExperienceLevel *ExperienceLevel `form:"experience_level"`
	Skills          []string         `form:"skills"`
	MinScore        *float64         `form:"min_score"`
	Search          string           `form:"search"` // full-text on name, email
	shared.PaginationParams
}
