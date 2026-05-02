package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type ChannelType string

const (
	ChannelEmail    ChannelType = "email"
	ChannelSMS      ChannelType = "sms"
	ChannelFirebase ChannelType = "firebase"
)

type JobStatus string

const (
	JobStatusActive    JobStatus = "active"
	JobStatusPaused    JobStatus = "paused"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
)

type ScheduleType string

const (
	ScheduleCron     ScheduleType = "cron"
	ScheduleOnce     ScheduleType = "once"
	ScheduleInterval ScheduleType = "interval"
)

type NotifyStatus string

const (
	NotifyStatusPending  NotifyStatus = "pending"
	NotifyStatusSent     NotifyStatus = "sent"
	NotifyStatusFailed   NotifyStatus = "failed"
	NotifyStatusRetrying NotifyStatus = "retrying"
)

type JSONMap map[string]interface{}

func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return "{}", nil
	}
	b, err := json.Marshal(j)
	return string(b), err
}

func (j *JSONMap) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T", value)
	}
	return json.Unmarshal(bytes, j)
}

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	b, err := json.Marshal(s)
	return string(b), err
}

func (s *StringSlice) Scan(value interface{}) error {
	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return fmt.Errorf("cannot scan type %T", value)
	}
	return json.Unmarshal(bytes, s)
}

type Channel struct {
	ID          string      `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string      `gorm:"uniqueIndex;not null;type:varchar(100)" json:"name"`
	Type        ChannelType `gorm:"not null;type:varchar(20)" json:"type"`
	Config      JSONMap     `gorm:"type:json" json:"config"`
	IsActive    bool        `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type DataSource struct {
	ID          string            `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string            `gorm:"uniqueIndex;not null;type:varchar(100)" json:"name"`
	URL         string            `gorm:"not null;type:varchar(500)" json:"url"`
	Method      string            `gorm:"not null;type:varchar(10);default:GET" json:"method"`
	Headers     JSONMap           `gorm:"type:json" json:"headers"`
	Body        string            `gorm:"type:text" json:"body"`
	AuthType    string            `gorm:"type:varchar(20)" json:"auth_type"` // none|bearer|basic|apikey
	AuthConfig  JSONMap           `gorm:"type:json" json:"auth_config"`
	JSONPath    string            `gorm:"type:varchar(200)" json:"json_path"` // e.g. $.data.users
	IsActive    bool              `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Template struct {
	ID          string      `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name        string      `gorm:"uniqueIndex;not null;type:varchar(100)" json:"name"`
	Channel     ChannelType `gorm:"not null;type:varchar(20)" json:"channel"`
	Subject     string      `gorm:"type:varchar(300)" json:"subject"` // for email
	Body        string      `gorm:"not null;type:longtext" json:"body"`
	Variables   StringSlice `gorm:"type:json" json:"variables"` // list of expected variable names
	IsActive    bool        `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
}

type NotifyJob struct {
	ID             string       `gorm:"primaryKey;type:varchar(36)" json:"id"`
	Name           string       `gorm:"not null;type:varchar(200)" json:"name"`
	Description    string       `gorm:"type:text" json:"description"`
	ChannelID      string       `gorm:"not null;type:varchar(36);index" json:"channel_id"`
	TemplateID     string       `gorm:"not null;type:varchar(36);index" json:"template_id"`
	DataSourceID   string       `gorm:"type:varchar(36);index" json:"data_source_id"` // optional
	ScheduleType   ScheduleType `gorm:"not null;type:varchar(20)" json:"schedule_type"`
	CronExpr       string       `gorm:"type:varchar(100)" json:"cron_expr"`   // for cron type
	RunAt          *time.Time   `json:"run_at"`                                // for once type
	IntervalSec    int          `gorm:"default:0" json:"interval_sec"`         // for interval type
	Recipients     StringSlice  `gorm:"type:json" json:"recipients"`           // emails / phone / FCM tokens
	StaticPayload  JSONMap      `gorm:"type:json" json:"static_payload"`       // static template vars
	Status         JobStatus    `gorm:"default:'active';type:varchar(20)" json:"status"`
	LastRunAt      *time.Time   `json:"last_run_at"`
	NextRunAt      *time.Time   `json:"next_run_at"`
	RunCount       int64        `gorm:"default:0" json:"run_count"`
	FailCount      int64        `gorm:"default:0" json:"fail_count"`
	MaxRuns        int64        `gorm:"default:0" json:"max_runs"` // 0 = unlimited
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`

	Channel    *Channel    `gorm:"foreignKey:ChannelID" json:"channel,omitempty"`
	Template   *Template   `gorm:"foreignKey:TemplateID" json:"template,omitempty"`
	DataSource *DataSource `gorm:"foreignKey:DataSourceID" json:"data_source,omitempty"`
}

type NotifyHistory struct {
	ID          string       `gorm:"primaryKey;type:varchar(36)" json:"id"`
	JobID       string       `gorm:"not null;type:varchar(36);index" json:"job_id"`
	ChannelType ChannelType  `gorm:"not null;type:varchar(20)" json:"channel_type"`
	Recipient   string       `gorm:"not null;type:varchar(300)" json:"recipient"`
	Subject     string       `gorm:"type:varchar(300)" json:"subject"`
	Body        string       `gorm:"type:longtext" json:"body"`
	Status      NotifyStatus `gorm:"default:'pending';type:varchar(20)" json:"status"`
	RetryCount  int          `gorm:"default:0" json:"retry_count"`
	Error       string       `gorm:"type:text" json:"error"`
	SentAt      *time.Time   `json:"sent_at"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}
