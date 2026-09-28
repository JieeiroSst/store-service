package model

import (
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strings"
	"time"
)

type UserDevice struct {
	ID          uint      `json:"id" gorm:"column:id;primaryKey"`
	UserID      uint      `json:"user_id" gorm:"column:user_id;index:idx_device_user_active,priority:1"`
	DeviceToken string    `json:"-" gorm:"column:device_token"`
	TokenHash   *string   `json:"-" gorm:"column:token_hash;size:64;uniqueIndex:uk_device_token_hash"`
	DeviceID    string    `json:"device_id" gorm:"column:device_id;size:128;index:idx_device_install"`
	DeviceType  string    `json:"device_type" gorm:"column:device_type"`
	IsActive    bool      `json:"is_active" gorm:"column:is_active;index:idx_device_user_active,priority:2"`
	LastUsedAt  time.Time `json:"last_used_at" gorm:"column:last_used_at"`
	CreatedAt   time.Time `json:"created_at" gorm:"column:created_at"`
}

func (UserDevice) TableName() string {
	return "notification_user_device"
}

const (
	DeviceIOS     = "ios"
	DeviceAndroid = "android"
	DeviceWeb     = "web"
)

var apnsToken = regexp.MustCompile(`^[0-9a-fA-F]{64}$`)

func ValidDeviceType(t string) bool {
	return t == DeviceIOS || t == DeviceAndroid || t == DeviceWeb
}

func LooksLikeAPNsToken(token string) bool { return apnsToken.MatchString(token) }

func (d UserDevice) TokenPreview() string {
	t := strings.TrimSpace(d.DeviceToken)
	if len(t) <= 10 {
		return "…"
	}
	return t[:6] + "…" + t[len(t)-4:]
}

type UserContact struct {
	UserID    uint      `json:"user_id" gorm:"column:user_id;primaryKey;autoIncrement:false"`
	Email     *string   `json:"email" gorm:"column:email;size:320;uniqueIndex:uk_contact_email"`
	Phone     *string   `json:"phone" gorm:"column:phone;size:32;uniqueIndex:uk_contact_phone"`
	UpdatedAt time.Time `json:"updated_at" gorm:"column:updated_at"`
}

func (UserContact) TableName() string { return "notification_user_contact" }

var phoneDigits = regexp.MustCompile(`^\+?[0-9]{8,15}$`)

func NormalizeEmail(email string) string { return strings.ToLower(strings.TrimSpace(email)) }

func NormalizePhone(raw string) (string, bool) {
	p := strings.NewReplacer(" ", "", ".", "", "-", "", "(", "", ")", "").Replace(strings.TrimSpace(raw))
	if !phoneDigits.MatchString(p) {
		return "", false
	}
	switch {
	case strings.HasPrefix(p, "+84"):
		p = "0" + p[3:]
	case strings.HasPrefix(p, "84") && len(p) == 11:
		p = "0" + p[2:]
	}
	return p, true
}

func HashToken(token string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(token)))
	return hex.EncodeToString(sum[:])
}

type Notification struct {
	ID          uint              `json:"id" gorm:"column:id;primaryKey"`
	UserID      uint              `json:"user_id" gorm:"column:user_id"`
	Recipient   string            `json:"recipient" gorm:"column:recipient"`
	Email       string            `json:"email,omitempty" gorm:"column:email;size:320"`
	Phone       string            `json:"phone,omitempty" gorm:"column:phone;size:32"`
	Data        map[string]string `json:"data,omitempty" gorm:"column:data;serializer:json;type:text"`
	RawData     map[string]any    `json:"raw_data,omitempty" gorm:"column:raw_data;serializer:json;type:text"`
	Title       string            `json:"title" gorm:"column:title"`
	Message     string            `json:"message" gorm:"column:message"`
	Type        string            `json:"type" gorm:"column:type"`
	Status      string            `json:"status" gorm:"column:status"`
	Priority    int               `json:"priority" gorm:"column:priority"`
	CreatedAt   time.Time         `json:"created_at" gorm:"column:created_at"`
	SentAt      *time.Time        `json:"sent_at" gorm:"column:sent_at"`
	RetryCount  int               `json:"retry_count" gorm:"column:retry_count"`
	LastError   string            `json:"last_error,omitempty" gorm:"column:last_error;type:text"`
	RequestedBy string            `json:"requested_by,omitempty" gorm:"column:requested_by;size:128"`

	// TemplateType selects an HTML email template (see
	// internal/adapter/secondary/template/templates); TemplateData supplies
	// the {{.Key}} placeholder values for it. Only used when Type == "email".
	// When TemplateType is empty, Title/Message are sent as-is instead.
	TemplateType string            `json:"template_type,omitempty" gorm:"column:template_type"`
	TemplateData map[string]string `json:"template_data,omitempty" gorm:"column:template_data;serializer:json"`
}

func (Notification) TableName() string {
	return "notification_notification"
}
