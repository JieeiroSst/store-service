package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/draw-image-service/internal/domain"
	"github.com/JIeeiroSst/draw-image-service/internal/port"
	"gorm.io/gorm"
)

type collageModel struct {
	ID          string    `gorm:"type:char(36);primaryKey"`
	Bucket      string    `gorm:"type:varchar(63);not null"`
	ObjectKey   string    `gorm:"type:varchar(255);not null"`
	ContentType string    `gorm:"type:varchar(64);not null"`
	Size        int64     `gorm:"not null"`
	Width       int       `gorm:"not null"`
	Height      int       `gorm:"not null"`
	ImageCount  int       `gorm:"not null"`
	Layout      string    `gorm:"type:varchar(16);not null"`
	Columns     int       `gorm:"column:columns_count;not null"`
	CreatedAt   time.Time `gorm:"type:datetime(3);not null;index"`
}

func (collageModel) TableName() string { return "collages" }

type collageRepository struct {
	db *gorm.DB
}

func NewCollageRepository(db *gorm.DB) (port.CollageRepository, error) {
	if err := db.AutoMigrate(&collageModel{}); err != nil {
		return nil, err
	}
	return &collageRepository{db: db}, nil
}

func (r *collageRepository) Save(ctx context.Context, c *domain.Collage) error {
	m := toModel(c)
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *collageRepository) FindByID(ctx context.Context, id string) (*domain.Collage, error) {
	var m collageModel
	err := r.db.WithContext(ctx).Where("id = ?", id).Take(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return toDomain(m), nil
}

func toModel(c *domain.Collage) collageModel {
	return collageModel{
		ID:          c.ID,
		Bucket:      c.Bucket,
		ObjectKey:   c.ObjectKey,
		ContentType: c.ContentType,
		Size:        c.Size,
		Width:       c.Width,
		Height:      c.Height,
		ImageCount:  c.ImageCount,
		Layout:      string(c.Layout),
		Columns:     c.Columns,
		CreatedAt:   c.CreatedAt,
	}
}

func toDomain(m collageModel) *domain.Collage {
	return &domain.Collage{
		ID:          m.ID,
		Bucket:      m.Bucket,
		ObjectKey:   m.ObjectKey,
		ContentType: m.ContentType,
		Size:        m.Size,
		Width:       m.Width,
		Height:      m.Height,
		ImageCount:  m.ImageCount,
		Layout:      domain.Layout(m.Layout),
		Columns:     m.Columns,
		CreatedAt:   m.CreatedAt,
	}
}
