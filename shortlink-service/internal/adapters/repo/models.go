package repo

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/datatypes"
)

type OrganizationModel struct {
	ID          string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        *string        `gorm:"column:name"`
	Settings    datatypes.JSON `gorm:"column:settings"`
	SuspendedAt *time.Time     `gorm:"column:suspended_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (OrganizationModel) TableName() string { return "organizations" }

type LinkTemplateModel struct {
	ID          string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      *string        `gorm:"column:user_id;type:uuid"`
	Name        string         `gorm:"column:name"`
	Slug        string         `gorm:"column:slug"`
	Description *string        `gorm:"column:description"`
	Settings    datatypes.JSON `gorm:"column:settings"`
	IsDefault   bool           `gorm:"column:is_default"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
}

func (LinkTemplateModel) TableName() string { return "link_templates" }

type LinkModel struct {
	ID                     string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID                 *string        `gorm:"column:user_id;type:uuid"`
	OrganizationID         *string        `gorm:"column:organization_id;type:uuid"`
	TemplateID             *string        `gorm:"column:template_id;type:uuid"`
	ShortCode              string         `gorm:"column:short_code"`
	OriginalURL            string         `gorm:"column:original_url"`
	Title                  *string        `gorm:"column:title"`
	Description            *string        `gorm:"column:description"`
	IOSAppStoreURL         *string        `gorm:"column:ios_app_store_url"`
	AndroidAppStoreURL     *string        `gorm:"column:android_app_store_url"`
	WebFallbackURL         *string        `gorm:"column:web_fallback_url"`
	AppScheme              *string        `gorm:"column:app_scheme"`
	IOSUniversalLink       *string        `gorm:"column:ios_universal_link"`
	AndroidAppLink         *string        `gorm:"column:android_app_link"`
	DeepLinkPath           *string        `gorm:"column:deep_link_path"`
	DeepLinkParameters     datatypes.JSON `gorm:"column:deep_link_parameters"`
	UTMParameters          datatypes.JSON `gorm:"column:utm_parameters"`
	TargetingRules         datatypes.JSON `gorm:"column:targeting_rules"`
	OGTitle                *string        `gorm:"column:og_title"`
	OGDescription          *string        `gorm:"column:og_description"`
	OGImageURL             *string        `gorm:"column:og_image_url"`
	OGType                 *string        `gorm:"column:og_type"`
	AttributionWindowHours int            `gorm:"column:attribution_window_hours"`
	IsActive               bool           `gorm:"column:is_active"`
	ExpiresAt              *time.Time     `gorm:"column:expires_at"`
	WarnAt                 *time.Time     `gorm:"column:warn_at"`
	DisabledAt             *time.Time     `gorm:"column:disabled_at"`
	DisabledReason         *string        `gorm:"column:disabled_reason"`
	AppendClickID          bool           `gorm:"column:append_click_id"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
	UpdatedAt              time.Time      `gorm:"column:updated_at"`
}

func (LinkModel) TableName() string { return "links" }

type ClickEventModel struct {
	ID          string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	LinkID      string    `gorm:"column:link_id;type:uuid"`
	ClickedAt   time.Time `gorm:"column:clicked_at"`
	IPAddress   *string   `gorm:"column:ip_address"`
	UserAgent   *string   `gorm:"column:user_agent"`
	DeviceType  *string   `gorm:"column:device_type"`
	Platform    *string   `gorm:"column:platform"`
	CountryCode *string   `gorm:"column:country_code"`
	CountryName *string   `gorm:"column:country_name"`
	Region      *string   `gorm:"column:region"`
	City        *string   `gorm:"column:city"`
	Latitude    *float64  `gorm:"column:latitude"`
	Longitude   *float64  `gorm:"column:longitude"`
	Timezone    *string   `gorm:"column:timezone"`
	UTMSource   *string   `gorm:"column:utm_source"`
	UTMMedium   *string   `gorm:"column:utm_medium"`
	UTMCampaign *string   `gorm:"column:utm_campaign"`
	Referrer    *string   `gorm:"column:referrer"`
	IsBot       bool      `gorm:"column:is_bot"`
	BotReason   *string   `gorm:"column:bot_reason"`
}

func (ClickEventModel) TableName() string { return "click_events" }

type DeviceFingerprintModel struct {
	ID              string    `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	ClickID         string    `gorm:"column:click_id;type:uuid"`
	FingerprintHash string    `gorm:"column:fingerprint_hash"`
	IPAddress       *string   `gorm:"column:ip_address"`
	UserAgent       *string   `gorm:"column:user_agent"`
	Timezone        *string   `gorm:"column:timezone"`
	Language        *string   `gorm:"column:language"`
	ScreenWidth     *int      `gorm:"column:screen_width"`
	ScreenHeight    *int      `gorm:"column:screen_height"`
	Platform        *string   `gorm:"column:platform"`
	PlatformVersion *string   `gorm:"column:platform_version"`
	CreatedAt       time.Time `gorm:"column:created_at"`
}

func (DeviceFingerprintModel) TableName() string { return "device_fingerprints" }

type InstallEventModel struct {
	ID                     string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	LinkID                 *string        `gorm:"column:link_id;type:uuid"`
	ClickID                *string        `gorm:"column:click_id;type:uuid"`
	FingerprintHash        string         `gorm:"column:fingerprint_hash"`
	ConfidenceScore        *float64       `gorm:"column:confidence_score"`
	AttributionMethod      *string        `gorm:"column:attribution_method"`
	MatchedFactors         pq.StringArray `gorm:"column:matched_factors;type:text[]"`
	InstalledAt            time.Time      `gorm:"column:installed_at"`
	FirstOpenAt            *time.Time     `gorm:"column:first_open_at"`
	DeepLinkRetrieved      bool           `gorm:"column:deep_link_retrieved"`
	DeepLinkData           datatypes.JSON `gorm:"column:deep_link_data"`
	AttributionWindowHours int            `gorm:"column:attribution_window_hours"`
	IPAddress              *string        `gorm:"column:ip_address"`
	UserAgent              *string        `gorm:"column:user_agent"`
	Timezone               *string        `gorm:"column:timezone"`
	Language               *string        `gorm:"column:language"`
	ScreenWidth            *int           `gorm:"column:screen_width"`
	ScreenHeight           *int           `gorm:"column:screen_height"`
	Platform               *string        `gorm:"column:platform"`
	PlatformVersion        *string        `gorm:"column:platform_version"`
	DeviceID               *string        `gorm:"column:device_id"`
	SDKName                *string        `gorm:"column:sdk_name"`
	SDKVersion             *string        `gorm:"column:sdk_version"`
	CreatedAt              time.Time      `gorm:"column:created_at"`
}

func (InstallEventModel) TableName() string { return "install_events" }

type InAppEventModel struct {
	ID                string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	InstallID         string         `gorm:"column:install_id;type:uuid"`
	EventName         string         `gorm:"column:event_name"`
	EventData         datatypes.JSON `gorm:"column:event_data"`
	EventTimestamp    time.Time      `gorm:"column:event_timestamp"`
	AttributedLinkID  *string        `gorm:"column:attributed_link_id;type:uuid"`
	AttributedClickID *string        `gorm:"column:attributed_click_id;type:uuid"`
	AttributedAt      *time.Time     `gorm:"column:attributed_at"`
	SessionID         *string        `gorm:"column:session_id;type:uuid"`
	SDKName           *string        `gorm:"column:sdk_name"`
	SDKVersion        *string        `gorm:"column:sdk_version"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
}

func (InAppEventModel) TableName() string { return "in_app_events" }

type WebhookModel struct {
	ID         string         `gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     *string        `gorm:"column:user_id;type:uuid"`
	Name       string         `gorm:"column:name"`
	URL        string         `gorm:"column:url"`
	Secret     string         `gorm:"column:secret"`
	Events     pq.StringArray `gorm:"column:events;type:text[]"`
	IsActive   bool           `gorm:"column:is_active"`
	RetryCount int            `gorm:"column:retry_count"`
	TimeoutMs  int            `gorm:"column:timeout_ms"`
	Headers    datatypes.JSON `gorm:"column:headers"`
	CreatedAt  time.Time      `gorm:"column:created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at"`
}

func (WebhookModel) TableName() string { return "webhooks" }
