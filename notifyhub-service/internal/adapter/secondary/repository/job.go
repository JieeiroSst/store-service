package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type jobRepository struct {
	db *gorm.DB
}

func NewJobRepository(db *gorm.DB) port.JobRepository {
	return &jobRepository{db: db}
}

func (r *jobRepository) withRelations(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Preload("Channel").
		Preload("Template").
		Preload("DataSource")
}

func (r *jobRepository) Create(ctx context.Context, j *model.NotifyJob) error {
	return translate(r.db.WithContext(ctx).Omit(clause.Associations).Create(j).Error, "job "+j.Name)
}

func (r *jobRepository) Get(ctx context.Context, id string) (*model.NotifyJob, error) {
	var j model.NotifyJob
	if err := r.withRelations(ctx).First(&j, "id = ?", id).Error; err != nil {
		return nil, translate(err, "job "+id)
	}
	return &j, nil
}

func (r *jobRepository) ListActive(ctx context.Context) ([]*model.NotifyJob, error) {
	var jobs []*model.NotifyJob
	return jobs, r.withRelations(ctx).
		Where("status = ?", model.JobStatusActive).
		Find(&jobs).Error
}

func (r *jobRepository) List(ctx context.Context, f port.JobFilter) ([]*model.NotifyJob, int64, error) {
	filter := func(q *gorm.DB) *gorm.DB {
		if f.Status != "" {
			q = q.Where("status = ?", f.Status)
		}
		return q
	}

	var total int64
	if err := filter(r.db.WithContext(ctx).Model(&model.NotifyJob{})).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var jobs []*model.NotifyJob
	err := filter(r.db.WithContext(ctx)).
		Preload("Channel").
		Preload("Template").
		Order("created_at desc").
		Offset(offset(f.Page, f.PageSize)).
		Limit(f.PageSize).
		Find(&jobs).Error
	return jobs, total, err
}

func (r *jobRepository) Update(ctx context.Context, j *model.NotifyJob) error {
	res := r.db.WithContext(ctx).
		Model(j).
		Select("*").
		Omit(clause.Associations, "created_at").
		Updates(j)
	return affected(res, "job "+j.ID)
}

func (r *jobRepository) UpdateStatus(ctx context.Context, id string, status model.JobStatus) error {
	res := r.db.WithContext(ctx).
		Model(&model.NotifyJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		})
	return affected(res, "job "+id)
}

func (r *jobRepository) RecordRun(ctx context.Context, id string, next *time.Time, failed bool) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_run_at": now,
		"next_run_at": next,
		"updated_at":  now,
		"run_count":   gorm.Expr("run_count + 1"),
	}
	if failed {
		updates["fail_count"] = gorm.Expr("fail_count + 1")
	}
	return r.db.WithContext(ctx).Model(&model.NotifyJob{}).Where("id = ?", id).Updates(updates).Error
}

func (r *jobRepository) Delete(ctx context.Context, id string) error {
	return affected(r.db.WithContext(ctx).Delete(&model.NotifyJob{}, "id = ?", id), "job "+id)
}

func (r *jobRepository) CountByChannel(ctx context.Context, channelID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&model.NotifyJob{}).Where("channel_id = ?", channelID).Count(&n).Error
	return n, err
}
