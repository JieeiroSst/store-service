package entity

import (
	"time"
)

type QRCodeType string

const (
	QRCodeTypeURL   QRCodeType = "url"
	QRCodeTypeText  QRCodeType = "text"
	QRCodeTypeVCard QRCodeType = "vcard"
	QRCodeTypeWifi  QRCodeType = "wifi"
	QRCodeTypeEmail QRCodeType = "email"
	QRCodeTypePhone QRCodeType = "phone"
)

type QRCodeStatus string

const (
	QRCodeStatusActive   QRCodeStatus = "active"
	QRCodeStatusInactive QRCodeStatus = "inactive"
	QRCodeStatusExpired  QRCodeStatus = "expired"
)

type QRCode struct {
	ID          string       `bson:"_id,omitempty" json:"id"`
	ShortCode   string       `bson:"short_code" json:"short_code"`
	Title       string       `bson:"title" json:"title"`
	Type        QRCodeType   `bson:"type" json:"type"`
	Content     string       `bson:"content" json:"content"`
	RedirectURL string       `bson:"redirect_url" json:"redirect_url"`
	ImageBase64 string       `bson:"image_base64,omitempty" json:"image_base64,omitempty"`
	ImageURL    string       `bson:"image_url,omitempty" json:"image_url,omitempty"`
	Status      QRCodeStatus `bson:"status" json:"status"`
	ScanCount   int64        `bson:"scan_count" json:"scan_count"`
	Size        int          `bson:"size" json:"size"`
	ForeColor   string       `bson:"fore_color" json:"fore_color"`
	BackColor   string       `bson:"back_color" json:"back_color"`
	ExpiresAt   *time.Time   `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	CreatedAt   time.Time    `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time    `bson:"updated_at" json:"updated_at"`
	CreatedBy   string       `bson:"created_by,omitempty" json:"created_by,omitempty"`
}

type ScanHistory struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	QRCodeID  string    `bson:"qr_code_id" json:"qr_code_id"`
	ShortCode string    `bson:"short_code" json:"short_code"`
	IPAddress string    `bson:"ip_address" json:"ip_address"`
	UserAgent string    `bson:"user_agent" json:"user_agent"`
	Referer   string    `bson:"referer" json:"referer"`
	Country   string    `bson:"country,omitempty" json:"country,omitempty"`
	City      string    `bson:"city,omitempty" json:"city,omitempty"`
	DeviceType string   `bson:"device_type" json:"device_type"`
	OS        string    `bson:"os" json:"os"`
	Browser   string    `bson:"browser" json:"browser"`
	ScannedAt time.Time `bson:"scanned_at" json:"scanned_at"`
}

func (q *QRCode) IsExpired() bool {
	if q.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*q.ExpiresAt)
}

func (q *QRCode) IsActive() bool {
	return q.Status == QRCodeStatusActive && !q.IsExpired()
}
