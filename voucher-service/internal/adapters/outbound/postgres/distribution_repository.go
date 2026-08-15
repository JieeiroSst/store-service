package postgres

import (
	"context"
	"time"

	distributionapp "github.com/JIeeiroSst/voucher-service/internal/application/distribution"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"gorm.io/gorm"
)

type distributionJobModel struct {
	ID              string    `gorm:"column:id;primaryKey"`
	CorporateID     string    `gorm:"column:corporate_id"`
	OrderID         *string   `gorm:"column:order_id"`
	Status          string    `gorm:"column:status"`
	TotalRecipients int       `gorm:"column:total_recipients"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`
}

func (distributionJobModel) TableName() string { return "distribution_jobs" }

type distributionClaimModel struct {
	ID                  string     `gorm:"column:id;primaryKey"`
	DistributionJobID   string     `gorm:"column:distribution_job_id"`
	VoucherID           *string    `gorm:"column:voucher_id"`
	RecipientIdentifier string     `gorm:"column:recipient_identifier"`
	ClaimToken          string     `gorm:"column:claim_token"`
	ClaimTokenExpiresAt time.Time  `gorm:"column:claim_token_expires_at"`
	Status              string     `gorm:"column:status"`
	ClaimedAt           *time.Time `gorm:"column:claimed_at"`
	CreatedAt           time.Time  `gorm:"column:created_at"`
	UpdatedAt           time.Time  `gorm:"column:updated_at"`
}

func (distributionClaimModel) TableName() string { return "distribution_claims" }

type JobRepository struct{ db *gorm.DB }

func NewJobRepository(db *gorm.DB) distributionapp.JobRepository { return &JobRepository{db: db} }

func (r *JobRepository) Create(ctx context.Context, j *distributionapp.Job) error {
	if j.ID == "" {
		j.ID = newUUID()
	}
	model := distributionJobModel{
		ID:              j.ID,
		CorporateID:     j.CorporateID.String(),
		Status:          string(j.Status),
		TotalRecipients: j.TotalRecipients,
		CreatedAt:       time.Now().UTC(),
		UpdatedAt:       time.Now().UTC(),
	}
	if j.OrderID != nil {
		s := j.OrderID.String()
		model.OrderID = &s
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

func (r *JobRepository) FindByID(ctx context.Context, id string) (*distributionapp.Job, error) {
	var m distributionJobModel
	if err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error; err != nil {
		return nil, err
	}
	corpID, err := shared.ParseCorporateID(m.CorporateID)
	if err != nil {
		return nil, err
	}
	return &distributionapp.Job{
		ID:              m.ID,
		CorporateID:     corpID,
		Status:          distributionapp.Status(m.Status),
		TotalRecipients: m.TotalRecipients,
		CreatedAt:       m.CreatedAt,
	}, nil
}

func (r *JobRepository) Save(ctx context.Context, j *distributionapp.Job) error {
	return r.db.WithContext(ctx).Model(&distributionJobModel{}).
		Where("id = ?", j.ID).
		Updates(map[string]any{"status": string(j.Status), "updated_at": time.Now().UTC()}).Error
}

type ClaimRepository struct{ db *gorm.DB }

func NewClaimRepository(db *gorm.DB) distributionapp.ClaimRepository { return &ClaimRepository{db: db} }

func (r *ClaimRepository) CreateBatch(ctx context.Context, claims []*distributionapp.Claim) error {
	if len(claims) == 0 {
		return nil
	}
	now := time.Now().UTC()
	models := make([]distributionClaimModel, 0, len(claims))
	for _, c := range claims {
		if c.ID == "" {
			c.ID = newUUID()
		}
		models = append(models, distributionClaimModel{
			ID:                  c.ID,
			DistributionJobID:   c.DistributionJobID,
			RecipientIdentifier: c.RecipientIdentifier,
			ClaimToken:          c.ClaimToken,
			ClaimTokenExpiresAt: c.ClaimTokenExpiresAt,
			Status:              string(c.Status),
			CreatedAt:           now,
			UpdatedAt:           now,
		})
	}
	return r.db.WithContext(ctx).Create(&models).Error
}

func (r *ClaimRepository) FindByToken(ctx context.Context, token string) (*distributionapp.Claim, error) {
	var m distributionClaimModel
	if err := r.db.WithContext(ctx).First(&m, "claim_token = ?", token).Error; err != nil {
		return nil, err
	}
	claim := &distributionapp.Claim{
		ID:                   m.ID,
		DistributionJobID:    m.DistributionJobID,
		RecipientIdentifier:  m.RecipientIdentifier,
		ClaimToken:           m.ClaimToken,
		ClaimTokenExpiresAt:  m.ClaimTokenExpiresAt,
		Status:               distributionapp.ClaimStatus(m.Status),
	}
	if m.VoucherID != nil {
		vid, err := shared.ParseVoucherID(*m.VoucherID)
		if err != nil {
			return nil, err
		}
		claim.VoucherID = &vid
	}
	return claim, nil
}

func (r *ClaimRepository) Save(ctx context.Context, c *distributionapp.Claim) error {
	updates := map[string]any{"status": string(c.Status), "updated_at": time.Now().UTC(), "claimed_at": c.ClaimedAt}
	if c.VoucherID != nil {
		vid := c.VoucherID.String()
		updates["voucher_id"] = vid
	}
	return r.db.WithContext(ctx).Model(&distributionClaimModel{}).Where("id = ?", c.ID).Updates(updates).Error
}
