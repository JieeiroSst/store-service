package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

const (
	AuditSourceNotification = "notification"
	AuditSourceCampaign     = "campaign"

	AuditSent         = "sent"
	AuditFailed       = "failed"
	AuditInvalidToken = "invalid_token"

	DefaultRequester = "unknown"
)

type AuditContent struct {
	ID           uint              `json:"id" gorm:"column:id;primaryKey"`
	Hash         string            `json:"-" gorm:"column:hash;size:64;uniqueIndex:uk_audit_content_hash"`
	Channel      string            `json:"channel" gorm:"column:channel;size:16"`
	Title        string            `json:"title" gorm:"column:title;type:text"`
	Body         string            `json:"body" gorm:"column:body;type:longtext"`
	Data         map[string]string `json:"data,omitempty" gorm:"column:data;serializer:json;type:text"`
	TemplateType string            `json:"template_type,omitempty" gorm:"column:template_type;size:128"`
	CreatedAt    time.Time         `json:"created_at" gorm:"column:created_at"`
}

func (AuditContent) TableName() string { return "notification_audit_content" }

func (c *AuditContent) ComputeHash() {
	data, _ := json.Marshal(c.Data)
	sum := sha256.Sum256([]byte(c.Channel + "\x00" + c.TemplateType + "\x00" + c.Title + "\x00" + c.Body + "\x00" + string(data)))
	c.Hash = hex.EncodeToString(sum[:])
}

type AuditDelivery struct {
	ID          uint      `json:"id" gorm:"column:id;primaryKey"`
	ContentID   uint      `json:"content_id" gorm:"column:content_id;index:idx_audit_content"`
	SourceType  string    `json:"source_type" gorm:"column:source_type;size:16;index:idx_audit_source,priority:1"`
	SourceID    uint      `json:"source_id" gorm:"column:source_id;index:idx_audit_source,priority:2"`
	BatchID     uint      `json:"batch_id,omitempty" gorm:"column:batch_id"`
	Attempt     int       `json:"attempt" gorm:"column:attempt"`
	Channel     string    `json:"channel" gorm:"column:channel;size:16"`
	UserID      uint      `json:"user_id,omitempty" gorm:"column:user_id;index:idx_audit_user,priority:1"`
	Recipient   string    `json:"recipient,omitempty" gorm:"column:recipient;size:320;index:idx_audit_recipient,priority:1"`
	DeviceID    uint      `json:"device_id,omitempty" gorm:"column:device_id"`
	DeviceType  string    `json:"device_type,omitempty" gorm:"column:device_type;size:16"`
	TokenHash   string    `json:"token_hash,omitempty" gorm:"column:token_hash;size:64"`
	Status      string    `json:"status" gorm:"column:status;size:16"`
	MessageID   string    `json:"message_id,omitempty" gorm:"column:message_id;size:255"`
	Error       string    `json:"error,omitempty" gorm:"column:error;type:text"`
	RequestedBy string    `json:"requested_by" gorm:"column:requested_by;size:128"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at;index:idx_audit_created;index:idx_audit_user,priority:2;index:idx_audit_recipient,priority:2"`
}

func (AuditDelivery) TableName() string { return "notification_audit_delivery" }
