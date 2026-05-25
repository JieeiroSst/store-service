package job

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusOpen      Status = "open"
	StatusPaused    Status = "paused"
	StatusClosed    Status = "closed"
	StatusCancelled Status = "cancelled"
)

type JobType string

const (
	JobTypeFullTime   JobType = "full_time"
	JobTypePartTime   JobType = "part_time"
	JobTypeContract   JobType = "contract"
	JobTypeFreelance  JobType = "freelance"
	JobTypeInternship JobType = "internship"
)

type WorkMode string

const (
	WorkModeOnsite WorkMode = "onsite"
	WorkModeRemote WorkMode = "remote"
	WorkModeHybrid WorkMode = "hybrid"
)

type Job struct {
	shared.BaseEntity

	Title           string      `db:"title"             json:"title"`
	Code            string      `db:"code"              json:"code"`
	DepartmentID    uuid.UUID   `db:"department_id"     json:"department_id"`
	HiringManagerID uuid.UUID   `db:"hiring_manager_id" json:"hiring_manager_id"`
	RecruiterIDs    []uuid.UUID `db:"recruiter_ids"     json:"recruiter_ids"`

	Description   string         `db:"description"    json:"description"`
	Requirements  []string       `db:"requirements"   json:"requirements"`
	NiceToHave    []string       `db:"nice_to_have"   json:"nice_to_have"`
	Skills        []string       `db:"skills"         json:"skills"`
	JobType       JobType        `db:"job_type"       json:"job_type"`
	WorkMode      WorkMode       `db:"work_mode"      json:"work_mode"`
	Location      shared.Address `db:"location"       json:"location"`
	SalaryMin     *shared.Money  `db:"salary_min"     json:"salary_min,omitempty"`
	SalaryMax     *shared.Money  `db:"salary_max"     json:"salary_max,omitempty"`
	SalaryVisible bool           `db:"salary_visible" json:"salary_visible"`

	Headcount      int             `db:"headcount"       json:"headcount"`
	PipelineStages []PipelineStage `db:"-"               json:"pipeline_stages"`

	Status   Status   `db:"status"   json:"status"`
	Priority int      `db:"priority" json:"priority"`
	Tags     []string `db:"tags"     json:"tags"`

	TotalApplications int       `db:"total_applications" json:"total_applications"`
	OpenApplications  int       `db:"open_applications"  json:"open_applications"`
	Embedding         []float32 `db:"embedding"          json:"-"`

	events []shared.DomainEvent
}

type PipelineStage struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Order    int       `json:"order"`
	IsSystem bool      `json:"is_system"`
}

func New(title, code string, deptID, hiringManagerID uuid.UUID) (*Job, error) {
	if title == "" {
		return nil, errors.New("job: title required")
	}
	if code == "" {
		return nil, errors.New("job: code required")
	}
	j := &Job{
		BaseEntity:      shared.NewBaseEntity(),
		Title:           title,
		Code:            code,
		DepartmentID:    deptID,
		HiringManagerID: hiringManagerID,
		Status:          StatusDraft,
		PipelineStages:  defaultPipelineStages(),
	}
	return j, nil
}

func defaultPipelineStages() []PipelineStage {
	return []PipelineStage{
		{ID: uuid.New(), Name: "Applied", Order: 1, IsSystem: true},
		{ID: uuid.New(), Name: "CV Review", Order: 2, IsSystem: true},
		{ID: uuid.New(), Name: "Phone Screen", Order: 3, IsSystem: false},
		{ID: uuid.New(), Name: "Technical Round", Order: 4, IsSystem: false},
		{ID: uuid.New(), Name: "Final Interview", Order: 5, IsSystem: false},
		{ID: uuid.New(), Name: "Offer", Order: 6, IsSystem: true},
		{ID: uuid.New(), Name: "Hired", Order: 7, IsSystem: true},
	}
}

func (j *Job) Publish() error {
	if j.Status != StatusDraft {
		return errors.New("job: only draft jobs can be published")
	}
	if j.Title == "" || j.Description == "" {
		return errors.New("job: title and description required before publishing")
	}
	j.Status = StatusOpen
	j.record("JobPublished", map[string]any{"job_id": j.ID})
	return nil
}

func (j *Job) Pause() error {
	if j.Status != StatusOpen {
		return errors.New("job: only open jobs can be paused")
	}
	j.Status = StatusPaused
	j.record("JobPaused", nil)
	return nil
}

func (j *Job) Close(reason string) {
	j.Status = StatusClosed
	j.record("JobClosed", map[string]any{"reason": reason})
}

func (j *Job) IncrementApplications() {
	j.TotalApplications++
	j.OpenApplications++
}

func (j *Job) DomainEvents() []shared.DomainEvent { return j.events }
func (j *Job) ClearEvents()                       { j.events = nil }
func (j *Job) record(t string, p interface{}) {
	j.events = append(j.events, shared.NewDomainEvent(t, p))
}

type Filter struct {
	Status       *Status    `form:"status"`
	DepartmentID *uuid.UUID `form:"department_id"`
	RecruiterID  *uuid.UUID `form:"recruiter_id"`
	WorkMode     *WorkMode  `form:"work_mode"`
	Skills       []string   `form:"skills"`
	Search       string     `form:"search"`
	shared.PaginationParams
}


type Repository interface {
	Save(ctx context.Context, j *Job) error
	Update(ctx context.Context, j *Job) error
	FindByID(ctx context.Context, id uuid.UUID) (*Job, error)
	FindAll(ctx context.Context, filter Filter) (shared.PaginatedResult[*Job], error)
	FindByRecruiter(ctx context.Context, recruiterID uuid.UUID) ([]*Job, error)
}
