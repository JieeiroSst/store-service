package model

import (
	"regexp"
	"time"
)

type CampaignChannel string

const (
	CampaignChannelPush  CampaignChannel = "push"
	CampaignChannelEmail CampaignChannel = "email"
)

type CampaignAudience string

const (
	AudienceAllDevices CampaignAudience = "all_devices"
	AudienceUsers      CampaignAudience = "users"
	AudienceTopic      CampaignAudience = "topic"
	AudienceEmails     CampaignAudience = "emails"
)

type CampaignStatus string

const (
	CampaignDraft     CampaignStatus = "DRAFT"
	CampaignPending   CampaignStatus = "PENDING"
	CampaignPlanning  CampaignStatus = "PLANNING"
	CampaignSending   CampaignStatus = "SENDING"
	CampaignCompleted CampaignStatus = "COMPLETED"
	CampaignCancelled CampaignStatus = "CANCELLED"
	CampaignFailed    CampaignStatus = "FAILED"
)

func (s CampaignStatus) Finished() bool {
	return s == CampaignCompleted || s == CampaignCancelled || s == CampaignFailed
}

var topicName = regexp.MustCompile(`^[a-zA-Z0-9\-_.~%]{1,900}$`)

func ValidTopic(topic string) bool { return topicName.MatchString(topic) }

type Campaign struct {
	ID                uint              `json:"id" gorm:"column:id;primaryKey"`
	Channel           CampaignChannel   `json:"channel" gorm:"column:channel;size:16"`
	Audience          CampaignAudience  `json:"audience" gorm:"column:audience;size:16"`
	Topic             string            `json:"topic,omitempty" gorm:"column:topic;size:900"`
	Title             string            `json:"title" gorm:"column:title"`
	Message           string            `json:"message" gorm:"column:message;type:text"`
	Data              map[string]string `json:"data,omitempty" gorm:"column:data;serializer:json;type:text"`
	TemplateType      string            `json:"template_type,omitempty" gorm:"column:template_type"`
	TemplateData      map[string]string `json:"template_data,omitempty" gorm:"column:template_data;serializer:json;type:text"`
	Status            CampaignStatus    `json:"status" gorm:"column:status;size:16;index"`
	Recipients        int               `json:"recipients" gorm:"column:recipients"`
	Cursor            uint              `json:"-" gorm:"column:cursor_id"`
	BatchesPlanned    int               `json:"batches_planned" gorm:"column:batches_planned"`
	PlannerAttempts   int               `json:"-" gorm:"column:planner_attempts"`
	PlannerLeaseUntil *time.Time        `json:"-" gorm:"column:planner_lease_until"`
	LastError         string            `json:"last_error,omitempty" gorm:"column:last_error;type:text"`
	RequestedBy       string            `json:"requested_by,omitempty" gorm:"column:requested_by;size:128"`
	CreatedAt         time.Time         `json:"created_at" gorm:"column:created_at"`
	UpdatedAt         time.Time         `json:"updated_at" gorm:"column:updated_at"`
	StartedAt         *time.Time        `json:"started_at" gorm:"column:started_at"`
	CompletedAt       *time.Time        `json:"completed_at" gorm:"column:completed_at"`
}

func (Campaign) TableName() string { return "notification_campaign" }

type CampaignRecipient struct {
	ID           uint   `gorm:"column:id;primaryKey"`
	CampaignID   uint   `gorm:"column:campaign_id;uniqueIndex:uk_campaign_recipient,priority:1;index:idx_campaign_recipient_user,priority:1"`
	RecipientKey string `gorm:"column:recipient_key;size:320;uniqueIndex:uk_campaign_recipient,priority:2"`
	UserID       uint   `gorm:"column:user_id;index:idx_campaign_recipient_user,priority:2"`
	Email        string `gorm:"column:email;size:320"`
}

func (CampaignRecipient) TableName() string { return "notification_campaign_recipient" }

type BatchState string

const (
	BatchQueued    BatchState = "QUEUED"
	BatchSending   BatchState = "SENDING"
	BatchSent      BatchState = "SENT"
	BatchFailed    BatchState = "FAILED"
	BatchCancelled BatchState = "CANCELLED"
)

type CampaignBatch struct {
	ID             uint       `json:"id" gorm:"column:id;primaryKey"`
	CampaignID     uint       `json:"campaign_id" gorm:"column:campaign_id;uniqueIndex:uk_campaign_batch_seq,priority:1;index:idx_campaign_batch_state,priority:1"`
	Seq            int        `json:"seq" gorm:"column:seq;uniqueIndex:uk_campaign_batch_seq,priority:2"`
	AfterID        uint       `json:"after_id" gorm:"column:after_id"`
	UntilID        uint       `json:"until_id" gorm:"column:until_id"`
	State          BatchState `json:"state" gorm:"column:state;size:16;index:idx_campaign_batch_state,priority:2;index:idx_batch_republish,priority:1"`
	Attempts       int        `json:"attempts" gorm:"column:attempts"`
	Targets        int        `json:"targets" gorm:"column:targets"`
	Sent           int        `json:"sent" gorm:"column:sent"`
	Failed         int        `json:"failed" gorm:"column:failed"`
	Invalid        int        `json:"invalid" gorm:"column:invalid"`
	LeaseUntil     *time.Time `json:"-" gorm:"column:lease_until"`
	RepublishAfter time.Time  `json:"-" gorm:"column:republish_after;index:idx_batch_republish,priority:2"`
	LastError      string     `json:"last_error,omitempty" gorm:"column:last_error;type:text"`
	CreatedAt      time.Time  `json:"created_at" gorm:"column:created_at"`
	UpdatedAt      time.Time  `json:"updated_at" gorm:"column:updated_at"`
}

func (CampaignBatch) TableName() string { return "notification_campaign_batch" }

type CampaignProgress struct {
	BatchesQueued    int `json:"batches_queued"`
	BatchesSending   int `json:"batches_sending"`
	BatchesSent      int `json:"batches_sent"`
	BatchesFailed    int `json:"batches_failed"`
	BatchesCancelled int `json:"batches_cancelled"`
	Targets          int `json:"targets"`
	Sent             int `json:"sent"`
	Failed           int `json:"failed"`
	InvalidTokens    int `json:"invalid_tokens"`
}

type BatchTarget struct {
	ID         uint
	UserID     uint
	DeviceType string
	Token      string
	Email      string
}
