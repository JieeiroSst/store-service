package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/cdn-service/internal/domain/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fileRecord struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;column:id"`
	ObjectKey   string    `gorm:"column:object_key;uniqueIndex"`
	FileName    string    `gorm:"column:file_name"`
	ContentType string    `gorm:"column:content_type"`
	SizeBytes   int64     `gorm:"column:size_bytes"`
	Status      string    `gorm:"column:status;index"`
	UploadedBy  string    `gorm:"column:uploaded_by"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (fileRecord) TableName() string { return "files" }

func toRecord(f *model.File) fileRecord {
	return fileRecord{
		ID:          f.ID,
		ObjectKey:   f.ObjectKey,
		FileName:    f.FileName,
		ContentType: f.ContentType,
		SizeBytes:   f.SizeBytes,
		Status:      string(f.Status),
		UploadedBy:  f.UploadedBy,
	}
}

func toModel(r fileRecord) model.File {
	return model.File{
		ID:          r.ID,
		ObjectKey:   r.ObjectKey,
		FileName:    r.FileName,
		ContentType: r.ContentType,
		SizeBytes:   r.SizeBytes,
		Status:      model.FileStatus(r.Status),
		UploadedBy:  r.UploadedBy,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

type FileRepository struct {
	db *gorm.DB
}

func NewFileRepository(db *gorm.DB) *FileRepository {
	return &FileRepository{db: db}
}

func (r *FileRepository) Create(ctx context.Context, file *model.File) error {
	rec := toRecord(file)
	if err := r.db.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	*file = toModel(rec)
	return nil
}

func (r *FileRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.File, error) {
	var rec fileRecord
	err := r.db.WithContext(ctx).First(&rec, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f := toModel(rec)
	return &f, nil
}

func (r *FileRepository) Update(ctx context.Context, file *model.File) error {
	rec := toRecord(file)
	return r.db.WithContext(ctx).Model(&fileRecord{}).Where("id = ?", file.ID).Updates(&rec).Error
}

func (r *FileRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&fileRecord{}).Error
}

func (r *FileRepository) List(ctx context.Context, limit, offset int) ([]model.File, error) {
	var recs []fileRecord
	if err := r.db.WithContext(ctx).Order("created_at desc").Limit(limit).Offset(offset).Find(&recs).Error; err != nil {
		return nil, err
	}
	files := make([]model.File, len(recs))
	for i, rec := range recs {
		files[i] = toModel(rec)
	}
	return files, nil
}
