package domain

import "time"

type SocialLogin struct {
	ID            int       `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID        int       `gorm:"index;not null" json:"user_id"`
	Provider      string    `gorm:"size:32;not null;uniqueIndex:idx_social_login_provider_subject" json:"provider"`
	Subject       string    `gorm:"size:255;not null;uniqueIndex:idx_social_login_provider_subject" json:"subject"`
	Email         string    `gorm:"size:320" json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Name          string    `gorm:"size:255" json:"name"`
	Username      string    `gorm:"size:255" json:"username"`
	AvatarURL     string    `gorm:"size:2048" json:"avatar_url"`
	Scopes        string    `gorm:"size:1024" json:"scopes"`
	Profile       string    `gorm:"type:jsonb;not null;default:'{}'" json:"profile"`
	ConsentIP     string    `gorm:"size:64" json:"consent_ip"`
	ConsentedAt   time.Time `json:"consented_at"`
}

func (SocialLogin) TableName() string { return "social_logins" }
