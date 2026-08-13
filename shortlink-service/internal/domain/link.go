package domain

import "time"

type UTMParameters struct {
	Source   string `json:"source,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Campaign string `json:"campaign,omitempty"`
	Term     string `json:"term,omitempty"`
	Content  string `json:"content,omitempty"`
}

type TargetingRules struct {
	Countries []string `json:"countries,omitempty"`
	Devices   []string `json:"devices,omitempty"` // "ios" | "android" | "web"
	Languages []string `json:"languages,omitempty"`
}

type Link struct {
	ID                     string
	UserID                 *string
	OrganizationID         *string
	TemplateID             *string
	TemplateSlug           *string // joined, not a column
	ShortCode              string
	OriginalURL            string
	Title                  *string
	Description            *string
	IOSAppStoreURL         *string
	AndroidAppStoreURL     *string
	WebFallbackURL         *string
	AppScheme              *string
	IOSUniversalLink       *string
	AndroidAppLink         *string
	DeepLinkPath           *string
	DeepLinkParameters     map[string]interface{}
	UTMParameters          UTMParameters
	TargetingRules         TargetingRules
	OGTitle                *string
	OGDescription          *string
	OGImageURL             *string
	OGType                 string
	AttributionWindowHours int
	IsActive               bool
	ExpiresAt              *time.Time
	WarnAt                 *time.Time
	DisabledAt             *time.Time
	DisabledReason         *string
	AppendClickID          bool
	CreatedAt              time.Time
	UpdatedAt              time.Time
	ClickCount             int64
	TemplateSettings       *LinkTemplateSettings
	OrgSettings            *OrganizationSettings
	OwnerSuspendedAt       *time.Time
}

type LinkTemplateSettings struct {
	DefaultIosURL                 string          `json:"defaultIosUrl,omitempty"`
	DefaultAndroidURL             string          `json:"defaultAndroidUrl,omitempty"`
	DefaultWebFallbackURL         string          `json:"defaultWebFallbackUrl,omitempty"`
	DefaultAttributionWindowHours *int            `json:"defaultAttributionWindowHours,omitempty"`
	UTMParameters                 *UTMParameters  `json:"utmParameters,omitempty"`
	TargetingRules                *TargetingRules `json:"targetingRules,omitempty"`
	ExpiresAfterDays              *int            `json:"expiresAfterDays,omitempty"`
}

type OrganizationAppConfig struct {
	IosAppStoreURL     string `json:"iosAppStoreUrl,omitempty"`
	AndroidAppStoreURL string `json:"androidAppStoreUrl,omitempty"`
	WebFallbackURL     string `json:"webFallbackUrl,omitempty"`
}

type OrganizationSettings struct {
	AppConfig OrganizationAppConfig `json:"appConfig,omitempty"`
}

type LinkTemplate struct {
	ID          string
	UserID      *string
	Name        string
	Slug        string
	Description *string
	Settings    LinkTemplateSettings
	IsDefault   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Organization struct {
	ID          string
	Name        *string
	Settings    OrganizationSettings
	SuspendedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type ClickEvent struct {
	ID          string
	LinkID      string
	ClickedAt   time.Time
	IPAddress   *string
	UserAgent   *string
	DeviceType  *string
	Platform    *string
	CountryCode *string
	CountryName *string
	Region      *string
	City        *string
	Latitude    *float64
	Longitude   *float64
	Timezone    *string
	UTMSource   *string
	UTMMedium   *string
	UTMCampaign *string
	Referrer    *string
	IsBot       bool
	BotReason   *string
}

type DeviceFingerprint struct {
	ID              string
	ClickID         string
	FingerprintHash string
	IPAddress       *string
	UserAgent       *string
	Timezone        *string
	Language        *string
	ScreenWidth     *int
	ScreenHeight    *int
	Platform        *string
	PlatformVersion *string
	CreatedAt       time.Time
}

type InstallEvent struct {
	ID                     string
	LinkID                 *string
	ClickID                *string
	FingerprintHash        string
	ConfidenceScore        *float64
	AttributionMethod      *string
	MatchedFactors         []string
	InstalledAt            time.Time
	FirstOpenAt            *time.Time
	DeepLinkRetrieved      bool
	DeepLinkData           map[string]interface{}
	AttributionWindowHours int
	IPAddress              *string
	UserAgent              *string
	Timezone               *string
	Language               *string
	ScreenWidth            *int
	ScreenHeight           *int
	Platform               *string
	PlatformVersion        *string
	DeviceID               *string
	SDKName                *string
	SDKVersion             *string
	CreatedAt              time.Time
}

type InAppEvent struct {
	ID                string
	InstallID         string
	EventName         string
	EventData         map[string]interface{}
	EventTimestamp    time.Time
	AttributedLinkID  *string
	AttributedClickID *string
	AttributedAt      *time.Time
	SessionID         *string
	SDKName           *string
	SDKVersion        *string
	CreatedAt         time.Time
}
