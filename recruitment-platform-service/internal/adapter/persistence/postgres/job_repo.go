package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/job"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type jobRepo struct {
	db *sqlx.DB
}

func NewJobRepository(db *sqlx.DB) job.Repository {
	return &jobRepo{db: db}
}

func (r *jobRepo) Save(ctx context.Context, j *job.Job) error {
	const q = `
		INSERT INTO jobs (
			id, title, code, department_id, hiring_manager_id, recruiter_ids,
			description, requirements, nice_to_have, skills, job_type, work_mode,
			location, salary_min, salary_max, salary_visible, headcount,
			pipeline_stages, status, priority, tags,
			total_applications, open_applications, created_at, updated_at
		) VALUES (
			:id, :title, :code, :department_id, :hiring_manager_id, :recruiter_ids,
			:description, :requirements, :nice_to_have, :skills, :job_type, :work_mode,
			:location, :salary_min, :salary_max, :salary_visible, :headcount,
			:pipeline_stages, :status, :priority, :tags,
			:total_applications, :open_applications, :created_at, :updated_at
		)`
	row, err := jobToRow(j)
	if err != nil {
		return fmt.Errorf("jobRepo.Save: %w", err)
	}
	_, err = r.db.NamedExecContext(ctx, q, row)
	return err
}

func (r *jobRepo) Update(ctx context.Context, j *job.Job) error {
	const q = `
		UPDATE jobs SET
			title = :title, description = :description,
			requirements = :requirements, nice_to_have = :nice_to_have,
			skills = :skills, job_type = :job_type, work_mode = :work_mode,
			location = :location, salary_min = :salary_min, salary_max = :salary_max,
			salary_visible = :salary_visible, headcount = :headcount,
			pipeline_stages = :pipeline_stages, status = :status, priority = :priority,
			tags = :tags, recruiter_ids = :recruiter_ids,
			total_applications = :total_applications, open_applications = :open_applications,
			embedding = :embedding, updated_at = NOW()
		WHERE id = :id AND deleted_at IS NULL`
	row, err := jobToRow(j)
	if err != nil {
		return fmt.Errorf("jobRepo.Update: %w", err)
	}
	_, err = r.db.NamedExecContext(ctx, q, row)
	return err
}

func (r *jobRepo) FindByID(ctx context.Context, id uuid.UUID) (*job.Job, error) {
	var row jobRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM jobs WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		return nil, fmt.Errorf("jobRepo.FindByID: %w", err)
	}
	return jobFromRow(&row)
}

func (r *jobRepo) FindAll(ctx context.Context, f job.Filter) (shared.PaginatedResult[*job.Job], error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	idx := 1

	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, *f.Status)
		idx++
	}
	if f.DepartmentID != nil {
		conds = append(conds, fmt.Sprintf("department_id = $%d", idx))
		args = append(args, *f.DepartmentID)
		idx++
	}
	if f.RecruiterID != nil {
		conds = append(conds, fmt.Sprintf("recruiter_ids @> $%d", idx))
		args = append(args, fmt.Sprintf(`["%s"]`, f.RecruiterID))
		idx++
	}
	if f.WorkMode != nil {
		conds = append(conds, fmt.Sprintf("work_mode = $%d", idx))
		args = append(args, *f.WorkMode)
		idx++
	}
	if len(f.Skills) > 0 {
		conds = append(conds, fmt.Sprintf("skills && $%d", idx))
		args = append(args, f.Skills)
		idx++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf(
			"(title ILIKE $%d OR description ILIKE $%d)", idx, idx+1,
		))
		like := "%" + f.Search + "%"
		args = append(args, like, like)
		idx += 2
	}

	where := strings.Join(conds, " AND ")
	var total int64
	if err := r.db.GetContext(ctx, &total,
		fmt.Sprintf(`SELECT COUNT(*) FROM jobs WHERE %s`, where), args...,
	); err != nil {
		return shared.PaginatedResult[*job.Job]{}, err
	}

	offset := f.Offset()
	rows := []jobRow{}
	dataQ := fmt.Sprintf(
		`SELECT * FROM jobs WHERE %s ORDER BY priority DESC, created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.Limit, offset)
	if err := r.db.SelectContext(ctx, &rows, dataQ, args...); err != nil {
		return shared.PaginatedResult[*job.Job]{}, err
	}

	items := make([]*job.Job, 0, len(rows))
	for i := range rows {
		j, err := jobFromRow(&rows[i])
		if err != nil {
			return shared.PaginatedResult[*job.Job]{}, err
		}
		items = append(items, j)
	}

	totalPages := int(total) / f.Limit
	if int(total)%f.Limit > 0 {
		totalPages++
	}
	return shared.PaginatedResult[*job.Job]{
		Items: items, Total: total,
		Page: f.Page, Limit: f.Limit, TotalPages: totalPages,
	}, nil
}

func (r *jobRepo) FindByRecruiter(ctx context.Context, recruiterID uuid.UUID) ([]*job.Job, error) {
	rows := []jobRow{}
	q := fmt.Sprintf(`SELECT * FROM jobs WHERE recruiter_ids @> '["%s"]' AND deleted_at IS NULL ORDER BY created_at DESC`, recruiterID)
	if err := r.db.SelectContext(ctx, &rows, q); err != nil {
		return nil, err
	}
	result := make([]*job.Job, 0, len(rows))
	for i := range rows {
		j, err := jobFromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, j)
	}
	return result, nil
}

// ─── Row mapping ──────────────────────────────────────────────────────────────

type jobRow struct {
	ID                uuid.UUID `db:"id"`
	Title             string    `db:"title"`
	Code              string    `db:"code"`
	DepartmentID      uuid.UUID `db:"department_id"`
	HiringManagerID   uuid.UUID `db:"hiring_manager_id"`
	RecruiterIDsJSON  []byte    `db:"recruiter_ids"`
	Description       string    `db:"description"`
	RequirementsJSON  []byte    `db:"requirements"`
	NiceToHaveJSON    []byte    `db:"nice_to_have"`
	SkillsJSON        []byte    `db:"skills"`
	JobType           string    `db:"job_type"`
	WorkMode          string    `db:"work_mode"`
	LocationJSON      []byte    `db:"location"`
	SalaryMinJSON     []byte    `db:"salary_min"`
	SalaryMaxJSON     []byte    `db:"salary_max"`
	SalaryVisible     bool      `db:"salary_visible"`
	Headcount         int       `db:"headcount"`
	PipelineJSON      []byte    `db:"pipeline_stages"`
	Status            string    `db:"status"`
	Priority          int       `db:"priority"`
	TagsJSON          []byte    `db:"tags"`
	TotalApplications int       `db:"total_applications"`
	OpenApplications  int       `db:"open_applications"`
}

func jobToRow(j *job.Job) (map[string]interface{}, error) {
	recruiterJSON, _ := json.Marshal(j.RecruiterIDs)
	requirementsJSON, _ := json.Marshal(j.Requirements)
	niceToHaveJSON, _ := json.Marshal(j.NiceToHave)
	skillsJSON, _ := json.Marshal(j.Skills)
	locationJSON, _ := json.Marshal(j.Location)
	salaryMinJSON, _ := json.Marshal(j.SalaryMin)
	salaryMaxJSON, _ := json.Marshal(j.SalaryMax)
	pipelineJSON, _ := json.Marshal(j.PipelineStages)
	tagsJSON, _ := json.Marshal(j.Tags)

	return map[string]interface{}{
		"id":                 j.ID,
		"title":              j.Title,
		"code":               j.Code,
		"department_id":      j.DepartmentID,
		"hiring_manager_id":  j.HiringManagerID,
		"recruiter_ids":      recruiterJSON,
		"description":        j.Description,
		"requirements":       requirementsJSON,
		"nice_to_have":       niceToHaveJSON,
		"skills":             skillsJSON,
		"job_type":           j.JobType,
		"work_mode":          j.WorkMode,
		"location":           locationJSON,
		"salary_min":         salaryMinJSON,
		"salary_max":         salaryMaxJSON,
		"salary_visible":     j.SalaryVisible,
		"headcount":          j.Headcount,
		"pipeline_stages":    pipelineJSON,
		"status":             j.Status,
		"priority":           j.Priority,
		"tags":               tagsJSON,
		"total_applications": j.TotalApplications,
		"open_applications":  j.OpenApplications,
		"embedding":          j.Embedding,
		"created_at":         j.CreatedAt,
		"updated_at":         j.UpdatedAt,
	}, nil
}

func jobFromRow(row *jobRow) (*job.Job, error) {
	j := &job.Job{}
	j.ID = row.ID
	j.Title = row.Title
	j.Code = row.Code
	j.DepartmentID = row.DepartmentID
	j.HiringManagerID = row.HiringManagerID
	j.JobType = job.JobType(row.JobType)
	j.WorkMode = job.WorkMode(row.WorkMode)
	j.Status = job.Status(row.Status)
	j.Priority = row.Priority
	j.Headcount = row.Headcount
	j.TotalApplications = row.TotalApplications
	j.OpenApplications = row.OpenApplications
	j.SalaryVisible = row.SalaryVisible

	_ = json.Unmarshal(row.RecruiterIDsJSON, &j.RecruiterIDs)
	_ = json.Unmarshal(row.RequirementsJSON, &j.Requirements)
	_ = json.Unmarshal(row.NiceToHaveJSON, &j.NiceToHave)
	_ = json.Unmarshal(row.SkillsJSON, &j.Skills)
	_ = json.Unmarshal(row.LocationJSON, &j.Location)
	_ = json.Unmarshal(row.SalaryMinJSON, &j.SalaryMin)
	_ = json.Unmarshal(row.SalaryMaxJSON, &j.SalaryMax)
	_ = json.Unmarshal(row.PipelineJSON, &j.PipelineStages)
	_ = json.Unmarshal(row.TagsJSON, &j.Tags)
	return j, nil
}
