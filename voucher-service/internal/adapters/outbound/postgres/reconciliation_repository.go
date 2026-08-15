package postgres

import (
	"context"
	"encoding/json"
	"time"

	reconciliationapp "github.com/JIeeiroSst/voucher-service/internal/application/reconciliation"
	"gorm.io/gorm"
)

type reconciliationRunModel struct {
	ID               string     `gorm:"column:id;primaryKey"`
	RunType          string     `gorm:"column:run_type"`
	Status           string     `gorm:"column:status"`
	DiscrepancyCount int        `gorm:"column:discrepancy_count"`
	Summary          []byte     `gorm:"column:summary"`
	StartedAt        *time.Time `gorm:"column:started_at"`
	FinishedAt       *time.Time `gorm:"column:finished_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
}

func (reconciliationRunModel) TableName() string { return "reconciliation_runs" }

type RunRepository struct{ db *gorm.DB }

func NewRunRepository(db *gorm.DB) reconciliationapp.RunRepository { return &RunRepository{db: db} }

func (r *RunRepository) Create(ctx context.Context, run *reconciliationapp.Run) error {
	if run.ID == "" {
		run.ID = newUUID()
	}
	summary, err := json.Marshal(run.Discrepancies)
	if err != nil {
		return err
	}
	model := reconciliationRunModel{
		ID:               run.ID,
		RunType:          run.RunType,
		Status:           string(run.Status),
		DiscrepancyCount: run.DiscrepancyCount,
		Summary:          summary,
		StartedAt:        run.StartedAt,
		FinishedAt:       run.FinishedAt,
		CreatedAt:        time.Now().UTC(),
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *RunRepository) Save(ctx context.Context, run *reconciliationapp.Run) error {
	summary, err := json.Marshal(run.Discrepancies)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Model(&reconciliationRunModel{}).
		Where("id = ?", run.ID).
		Updates(map[string]any{
			"status":            string(run.Status),
			"discrepancy_count": run.DiscrepancyCount,
			"summary":           summary,
			"started_at":        run.StartedAt,
			"finished_at":       run.FinishedAt,
		}).Error
}

func (r *RunRepository) FindByID(ctx context.Context, id string) (*reconciliationapp.Run, error) {
	var m reconciliationRunModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	var discrepancies []reconciliationapp.Discrepancy
	_ = json.Unmarshal(m.Summary, &discrepancies)
	return &reconciliationapp.Run{
		ID:               m.ID,
		RunType:          m.RunType,
		Status:           reconciliationapp.RunStatus(m.Status),
		DiscrepancyCount: m.DiscrepancyCount,
		Discrepancies:    discrepancies,
		StartedAt:        m.StartedAt,
		FinishedAt:       m.FinishedAt,
	}, nil
}
