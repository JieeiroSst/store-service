package mysql

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Store struct {
	db *gorm.DB
}

func New(dsn string, maxOpen, maxIdle, maxLifetimeSec int) (*Store, error) {
	db, err := gorm.Open(mysqldriver.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		PrepareStmt:            true, 
		SkipDefaultTransaction: true, 
	})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get sql.DB: %w", err)
	}
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetimeSec) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(maxLifetimeSec/2) * time.Second)

	if err := db.AutoMigrate(
		&model.Channel{},
		&model.DataSource{},
		&model.Template{},
		&model.NotifyJob{},
		&model.NotifyHistory{},
	); err != nil {
		return nil, fmt.Errorf("auto-migrate: %w", err)
	}

	return &Store{db: db}, nil
}

func (s *Store) DB() *gorm.DB { return s.db }

// Ping checks database connectivity.
func (s *Store) Ping() error {
	db, err := s.db.DB()
	if err != nil {
		return err
	}
	return db.Ping()
}


func (s *Store) CreateChannel(ctx context.Context, c *model.Channel) error {
	return s.db.WithContext(ctx).Create(c).Error
}

func (s *Store) GetChannel(ctx context.Context, id string) (*model.Channel, error) {
	var c model.Channel
	if err := s.db.WithContext(ctx).First(&c, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *Store) ListChannels(ctx context.Context, channelType string, active *bool) ([]*model.Channel, error) {
	q := s.db.WithContext(ctx).Model(&model.Channel{})
	if channelType != "" {
		q = q.Where("type = ?", channelType)
	}
	if active != nil {
		q = q.Where("is_active = ?", *active)
	}
	var channels []*model.Channel
	return channels, q.Order("created_at desc").Find(&channels).Error
}

func (s *Store) UpdateChannel(ctx context.Context, c *model.Channel) error {
	return s.db.WithContext(ctx).Save(c).Error
}

func (s *Store) DeleteChannel(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.Channel{}, "id = ?", id).Error
}


func (s *Store) CreateDataSource(ctx context.Context, ds *model.DataSource) error {
	return s.db.WithContext(ctx).Create(ds).Error
}

func (s *Store) GetDataSource(ctx context.Context, id string) (*model.DataSource, error) {
	var ds model.DataSource
	if err := s.db.WithContext(ctx).First(&ds, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &ds, nil
}

func (s *Store) ListDataSources(ctx context.Context) ([]*model.DataSource, error) {
	var dss []*model.DataSource
	return dss, s.db.WithContext(ctx).
		Where("is_active = true").
		Order("created_at desc").
		Find(&dss).Error
}

func (s *Store) UpdateDataSource(ctx context.Context, ds *model.DataSource) error {
	return s.db.WithContext(ctx).Save(ds).Error
}


func (s *Store) CreateTemplate(ctx context.Context, t *model.Template) error {
	return s.db.WithContext(ctx).Create(t).Error
}

func (s *Store) GetTemplate(ctx context.Context, id string) (*model.Template, error) {
	var t model.Template
	if err := s.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) ListTemplates(ctx context.Context, channelFilter string) ([]*model.Template, error) {
	q := s.db.WithContext(ctx).Where("is_active = true")
	if channelFilter != "" {
		q = q.Where("channel = ?", channelFilter)
	}
	var ts []*model.Template
	return ts, q.Order("created_at desc").Find(&ts).Error
}

func (s *Store) UpdateTemplate(ctx context.Context, t *model.Template) error {
	return s.db.WithContext(ctx).Save(t).Error
}

func (s *Store) CreateJob(ctx context.Context, j *model.NotifyJob) error {
	return s.db.WithContext(ctx).Create(j).Error
}

func (s *Store) GetJob(ctx context.Context, id string) (*model.NotifyJob, error) {
	var j model.NotifyJob
	err := s.db.WithContext(ctx).
		Preload("Channel").
		Preload("Template").
		Preload("DataSource").
		First(&j, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &j, nil
}

func (s *Store) ListActiveJobs(ctx context.Context) ([]*model.NotifyJob, error) {
	var jobs []*model.NotifyJob
	return jobs, s.db.WithContext(ctx).
		Preload("Channel").
		Preload("Template").
		Preload("DataSource").
		Where("status = ?", model.JobStatusActive).
		Find(&jobs).Error
}

func (s *Store) ListJobs(ctx context.Context, status string, page, pageSize int) ([]*model.NotifyJob, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.NotifyJob{})
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var jobs []*model.NotifyJob
	offset := (page - 1) * pageSize
	err := q.
		Preload("Channel").
		Preload("Template").
		Order("created_at desc").
		Offset(offset).
		Limit(pageSize).
		Find(&jobs).Error
	return jobs, total, err
}

func (s *Store) UpdateJob(ctx context.Context, j *model.NotifyJob) error {
	return s.db.WithContext(ctx).Save(j).Error
}

func (s *Store) UpdateJobStatus(ctx context.Context, id string, status model.JobStatus) error {
	return s.db.WithContext(ctx).
		Model(&model.NotifyJob{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

func (s *Store) UpdateJobRun(ctx context.Context, id string, next *time.Time, failed bool) error {
	now := time.Now()
	baseUpdates := map[string]interface{}{
		"last_run_at": now,
		"next_run_at": next,
		"updated_at":  now,
	}

	db := s.db.WithContext(ctx).Model(&model.NotifyJob{}).Where("id = ?", id)
	if err := db.Updates(baseUpdates).Error; err != nil {
		return err
	}

	if failed {
		return s.db.WithContext(ctx).Model(&model.NotifyJob{}).
			Where("id = ?", id).
			UpdateColumn("fail_count", gorm.Expr("fail_count + 1")).Error
	}
	return s.db.WithContext(ctx).Model(&model.NotifyJob{}).
		Where("id = ?", id).
		UpdateColumn("run_count", gorm.Expr("run_count + 1")).Error
}

func (s *Store) DeleteJob(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).Delete(&model.NotifyJob{}, "id = ?", id).Error
}

func (s *Store) CreateHistory(ctx context.Context, h *model.NotifyHistory) error {
	return s.db.WithContext(ctx).Create(h).Error
}

func (s *Store) UpdateHistoryStatus(ctx context.Context, id string, status model.NotifyStatus, errMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"error":      errMsg,
		"updated_at": time.Now(),
	}
	if status == model.NotifyStatusSent {
		now := time.Now()
		updates["sent_at"] = now
	}
	return s.db.WithContext(ctx).
		Model(&model.NotifyHistory{}).
		Where("id = ?", id).
		Updates(updates).Error
}

func (s *Store) IncrementHistoryRetry(ctx context.Context, id string) error {
	return s.db.WithContext(ctx).
		Model(&model.NotifyHistory{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":      model.NotifyStatusRetrying,
			"retry_count": gorm.Expr("retry_count + 1"),
			"updated_at":  time.Now(),
		}).Error
}

func (s *Store) ListHistory(ctx context.Context, jobID, status string, page, pageSize int) ([]*model.NotifyHistory, int64, error) {
	q := s.db.WithContext(ctx).Model(&model.NotifyHistory{})
	if jobID != "" {
		q = q.Where("job_id = ?", jobID)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var hist []*model.NotifyHistory
	offset := (page - 1) * pageSize
	err := q.Order("created_at desc").Offset(offset).Limit(pageSize).Find(&hist).Error
	return hist, total, err
}

type HistorySummary struct {
	ChannelType model.ChannelType  `json:"channel_type"`
	Status      model.NotifyStatus `json:"status"`
	Count       int64              `json:"count"`
}

func (s *Store) GetHistorySummary(ctx context.Context, jobID string) ([]*HistorySummary, error) {
	var result []*HistorySummary
	err := s.db.WithContext(ctx).
		Model(&model.NotifyHistory{}).
		Select("channel_type, status, COUNT(*) as count").
		Where("job_id = ?", jobID).
		Group("channel_type, status").
		Scan(&result).Error
	return result, err
}
