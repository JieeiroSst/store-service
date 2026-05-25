package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/candidate"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type candidateRepo struct {
	db *sqlx.DB
}

func NewCandidateRepository(db *sqlx.DB) candidate.Repository {
	return &candidateRepo{db: db}
}

func (r *candidateRepo) Save(ctx context.Context, c *candidate.Candidate) error {
	const q = `
		INSERT INTO candidates (
			id, full_name, email, phone, avatar_url, linkedin_url, resume_url,
			location, current_title, current_company, years_of_experience, experience_level,
			skills, expected_salary, notice_period_days, status, source,
			referred_by_id, tags, notes, ai_score, embedding,
			created_at, updated_at
		) VALUES (
			:id, :full_name, :email, :phone, :avatar_url, :linkedin_url, :resume_url,
			:location, :current_title, :current_company, :years_of_experience, :experience_level,
			:skills, :expected_salary, :notice_period_days, :status, :source,
			:referred_by_id, :tags, :notes, :ai_score, :embedding,
			:created_at, :updated_at
		)`
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("candidateRepo.Save: %w", err)
	}
	_, err = r.db.NamedExecContext(ctx, q, row)
	return err
}

func (r *candidateRepo) Update(ctx context.Context, c *candidate.Candidate) error {
	const q = `
		UPDATE candidates SET
			full_name = :full_name, email = :email, phone = :phone,
			avatar_url = :avatar_url, linkedin_url = :linkedin_url, resume_url = :resume_url,
			location = :location, current_title = :current_title, current_company = :current_company,
			years_of_experience = :years_of_experience, experience_level = :experience_level,
			skills = :skills, expected_salary = :expected_salary, notice_period_days = :notice_period_days,
			status = :status, source = :source, referred_by_id = :referred_by_id,
			tags = :tags, notes = :notes, ai_score = :ai_score, embedding = :embedding,
			updated_at = NOW()
		WHERE id = :id AND deleted_at IS NULL`
	row, err := toRow(c)
	if err != nil {
		return fmt.Errorf("candidateRepo.Update: %w", err)
	}
	_, err = r.db.NamedExecContext(ctx, q, row)
	return err
}

func (r *candidateRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE candidates SET deleted_at = NOW() WHERE id = $1`, id)
	return err
}

func (r *candidateRepo) FindByID(ctx context.Context, id uuid.UUID) (*candidate.Candidate, error) {
	var row candidateRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM candidates WHERE id = $1 AND deleted_at IS NULL`, id,
	); err != nil {
		return nil, fmt.Errorf("candidateRepo.FindByID: %w", err)
	}
	return fromRow(&row)
}

func (r *candidateRepo) FindByEmail(ctx context.Context, email string) (*candidate.Candidate, error) {
	var row candidateRow
	if err := r.db.GetContext(ctx, &row,
		`SELECT * FROM candidates WHERE email = $1 AND deleted_at IS NULL`, email,
	); err != nil {
		return nil, fmt.Errorf("candidateRepo.FindByEmail: %w", err)
	}
	return fromRow(&row)
}

func (r *candidateRepo) FindAll(ctx context.Context, f candidate.Filter) (shared.PaginatedResult[*candidate.Candidate], error) {
	args := []interface{}{}
	conds := []string{"deleted_at IS NULL"}
	idx := 1

	if f.Status != nil {
		conds = append(conds, fmt.Sprintf("status = $%d", idx))
		args = append(args, *f.Status)
		idx++
	}
	if f.Source != nil {
		conds = append(conds, fmt.Sprintf("source = $%d", idx))
		args = append(args, *f.Source)
		idx++
	}
	if f.ExperienceLevel != nil {
		conds = append(conds, fmt.Sprintf("experience_level = $%d", idx))
		args = append(args, *f.ExperienceLevel)
		idx++
	}
	if len(f.Skills) > 0 {
		conds = append(conds, fmt.Sprintf("skills && $%d", idx)) // array overlap
		args = append(args, f.Skills)
		idx++
	}
	if f.MinScore != nil {
		conds = append(conds, fmt.Sprintf("(ai_score->>'score')::float >= $%d", idx))
		args = append(args, *f.MinScore)
		idx++
	}
	if f.Search != "" {
		conds = append(conds, fmt.Sprintf(
			"(full_name ILIKE $%d OR email ILIKE $%d)", idx, idx+1,
		))
		like := "%" + f.Search + "%"
		args = append(args, like, like)
		idx += 2
	}

	where := strings.Join(conds, " AND ")
	countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM candidates WHERE %s`, where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return shared.PaginatedResult[*candidate.Candidate]{}, err
	}

	offset := f.Offset()
	dataQuery := fmt.Sprintf(
		`SELECT * FROM candidates WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, idx, idx+1,
	)
	args = append(args, f.Limit, offset)

	var rows []candidateRow
	if err := r.db.SelectContext(ctx, &rows, dataQuery, args...); err != nil {
		return shared.PaginatedResult[*candidate.Candidate]{}, err
	}

	items := make([]*candidate.Candidate, 0, len(rows))
	for i := range rows {
		c, err := fromRow(&rows[i])
		if err != nil {
			return shared.PaginatedResult[*candidate.Candidate]{}, err
		}
		items = append(items, c)
	}

	totalPages := int(total) / f.Limit
	if int(total)%f.Limit > 0 {
		totalPages++
	}

	return shared.PaginatedResult[*candidate.Candidate]{
		Items:      items,
		Total:      total,
		Page:       f.Page,
		Limit:      f.Limit,
		TotalPages: totalPages,
	}, nil
}

// Vector similarity search using pgvector
func (r *candidateRepo) FindSimilar(ctx context.Context, embedding []float32, limit int) ([]*candidate.Candidate, error) {
	var rows []candidateRow
	var err error

	if len(embedding) > 0 {
		embJSON, _ := json.Marshal(embedding)
		const vecQ = `
			SELECT * FROM candidates
			WHERE deleted_at IS NULL AND embedding IS NOT NULL
			ORDER BY embedding <=> $1::vector
			LIMIT $2`
		err = r.db.SelectContext(ctx, &rows, vecQ, string(embJSON), limit)
	} else {
		const fallbackQ = `
			SELECT * FROM candidates
			WHERE deleted_at IS NULL
			ORDER BY created_at DESC
			LIMIT $1`
		err = r.db.SelectContext(ctx, &rows, fallbackQ, limit)
	}
	if err != nil {
		return nil, err
	}

	result := make([]*candidate.Candidate, 0, len(rows))
	for i := range rows {
		c, err := fromRow(&rows[i])
		if err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, nil
}

func (r *candidateRepo) BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status candidate.Status) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE candidates SET status = $1, updated_at = NOW() WHERE id = ANY($2)`,
		status, ids,
	)
	return err
}

type candidateRow struct {
	ID                 uuid.UUID   `db:"id"`
	FullName           string      `db:"full_name"`
	Email              string      `db:"email"`
	Phone              string      `db:"phone"`
	AvatarURL          string      `db:"avatar_url"`
	LinkedInURL        string      `db:"linkedin_url"`
	ResumeURL          string      `db:"resume_url"`
	LocationJSON       []byte      `db:"location"`
	CurrentTitle       string      `db:"current_title"`
	CurrentCompany     string      `db:"current_company"`
	YearsOfExperience  int         `db:"years_of_experience"`
	ExperienceLevel    string      `db:"experience_level"`
	SkillsJSON         []byte      `db:"skills"`
	ExpectedSalaryJSON []byte      `db:"expected_salary"`
	NoticePeriodDays   int         `db:"notice_period_days"`
	Status             string      `db:"status"`
	Source             string      `db:"source"`
	ReferredByID       *uuid.UUID  `db:"referred_by_id"`
	TagsJSON           []byte      `db:"tags"`
	Notes              string      `db:"notes"`
	AIScoreJSON        []byte      `db:"ai_score"`
	CreatedAt          interface{} `db:"created_at"`
	UpdatedAt          interface{} `db:"updated_at"`
	DeletedAt          interface{} `db:"deleted_at"`
}

func toRow(c *candidate.Candidate) (map[string]interface{}, error) {
	locationJSON, _ := json.Marshal(c.Location)
	skillsJSON, _ := json.Marshal(c.Skills)
	tagsJSON, _ := json.Marshal(c.Tags)
	aiScoreJSON, _ := json.Marshal(c.AIScore)
	expSalaryJSON, _ := json.Marshal(c.ExpectedSalary)

	return map[string]interface{}{
		"id":                  c.ID,
		"full_name":           c.FullName,
		"email":               c.Email,
		"phone":               c.Phone,
		"avatar_url":          c.AvatarURL,
		"linkedin_url":        c.LinkedInURL,
		"resume_url":          c.ResumeURL,
		"location":            locationJSON,
		"current_title":       c.CurrentTitle,
		"current_company":     c.CurrentCompany,
		"years_of_experience": c.YearsOfExperience,
		"experience_level":    c.ExperienceLevel,
		"skills":              skillsJSON,
		"expected_salary":     expSalaryJSON,
		"notice_period_days":  c.NoticePeriodDays,
		"status":              c.Status,
		"source":              c.Source,
		"referred_by_id":      c.ReferredByID,
		"tags":                tagsJSON,
		"notes":               c.Notes,
		"ai_score":            aiScoreJSON,
		"embedding":           c.Embedding,
		"created_at":          c.CreatedAt,
		"updated_at":          c.UpdatedAt,
	}, nil
}

func fromRow(row *candidateRow) (*candidate.Candidate, error) {
	c := &candidate.Candidate{}
	c.ID = row.ID
	c.FullName = row.FullName
	c.Email = row.Email
	c.Phone = row.Phone
	c.AvatarURL = row.AvatarURL
	c.LinkedInURL = row.LinkedInURL
	c.ResumeURL = row.ResumeURL
	c.CurrentTitle = row.CurrentTitle
	c.CurrentCompany = row.CurrentCompany
	c.YearsOfExperience = row.YearsOfExperience
	c.ExperienceLevel = candidate.ExperienceLevel(row.ExperienceLevel)
	c.Status = candidate.Status(row.Status)
	c.Source = candidate.SourceChannel(row.Source)
	c.ReferredByID = row.ReferredByID
	c.Notes = row.Notes
	c.NoticePeriodDays = row.NoticePeriodDays

	_ = json.Unmarshal(row.LocationJSON, &c.Location)
	_ = json.Unmarshal(row.SkillsJSON, &c.Skills)
	_ = json.Unmarshal(row.TagsJSON, &c.Tags)
	_ = json.Unmarshal(row.AIScoreJSON, &c.AIScore)
	_ = json.Unmarshal(row.ExpectedSalaryJSON, &c.ExpectedSalary)

	return c, nil
}
