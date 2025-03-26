package dto

import "time"

type LinkStatus int32

const (
	LinkStatus_ACTIVE LinkStatus = iota
	LinkStatus_DISABLED
	LinkStatus_EXPIRED
)

type Link struct {
	ID          string     `json:"id" db:"id"`
	OriginalURL string     `json:"original_url" db:"original_url"`
	Shortlink   string     `json:"shortlink" db:"shortlink"`
	ShortCode   string     `json:"short_code" db:"short_code"`
	UserID      string     `json:"user_id" db:"user_id"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	ExpiredAt   time.Time  `json:"expired_at" db:"expired_at"`
	TotalClicks int32      `json:"total_clicks" db:"total_clicks"`
	Status      LinkStatus `json:"status" db:"status"`
}

type LinkClick struct {
	ID         string    `json:"id" db:"id"`
	LinkID     string    `json:"link_id" db:"link_id"`
	ClickedAt  time.Time `json:"clicked_at" db:"clicked_at"`
	IPAddress  string    `json:"ip_address" db:"ip_address"`
	Country    string    `json:"country" db:"country"`
	Browser    string    `json:"browser" db:"browser"`
	DeviceType string    `json:"device_type" db:"device_type"`
}

func (l LinkStatus) ToInt32() int32 {
	return int32(l)
}

func (l LinkStatus) String() string {
	return [...]string{"ACTIVE", "DISABLED", "EXPIRED"}[l]
}

func ToLinkStatus(i int32) LinkStatus {
	return LinkStatus(i)
}
