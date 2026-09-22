package model

import "time"

type FaceBiometric struct {
	ID           string    `json:"id" gorm:"type:uuid;primaryKey"`
	UserID       string    `json:"user_id" gorm:"uniqueIndex;not null"`
	Template     []byte    `json:"-" gorm:"type:bytea;not null"`
	FaceImageKey string    `json:"-"`
	QualityScore float64   `json:"quality_score"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (FaceBiometric) TableName() string { return "face_biometrics" }

type FaceTemplate struct {
	Template     []byte
	AlignedImage []byte
	QualityScore float64
}
