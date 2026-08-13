package ports

import (
	"context"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
)

type LinkFilter struct {
	UserID *string
}

type LinkRepository interface {
	Create(ctx context.Context, link *domain.Link) error
	GetByID(ctx context.Context, id string, filter LinkFilter) (*domain.Link, error)
	List(ctx context.Context, filter LinkFilter) ([]*domain.Link, error)
	Update(ctx context.Context, id string, filter LinkFilter, patch map[string]interface{}) (*domain.Link, error)
	Delete(ctx context.Context, id string, filter LinkFilter) (*domain.Link, error)
	ExistsByShortCode(ctx context.Context, shortCode string) (bool, error)
	ResolveForRedirect(ctx context.Context, shortCode, templateSlug string) (*domain.Link, error)
	TemplateSlugByID(ctx context.Context, templateID string) (string, error)
}

type ClickEventRepository interface {
	Insert(ctx context.Context, event *domain.ClickEvent) error
	InsertWithID(ctx context.Context, id string, event *domain.ClickEvent) error
	GetByID(ctx context.Context, id string) (*domain.ClickEvent, error)
	CountTotalAndUnique(ctx context.Context, filter AnalyticsFilter) (total, unique int64, err error)
	CountByDate(ctx context.Context, filter AnalyticsFilter) ([]DateCount, error)
	CountByCountry(ctx context.Context, filter AnalyticsFilter) ([]CountryCount, error)
	CountByDevice(ctx context.Context, filter AnalyticsFilter) ([]DeviceCount, error)
	CountByPlatform(ctx context.Context, filter AnalyticsFilter) ([]PlatformCount, error)
	TopLinks(ctx context.Context, filter AnalyticsFilter, limit int) ([]TopLink, error)
}

type AnalyticsFilter struct {
	UserID *string
	LinkID *string
	Days   int
}

type DateCount struct {
	Date   time.Time
	Clicks int64
}
type CountryCount struct {
	CountryCode string
	Country     string
	Clicks      int64
}
type DeviceCount struct {
	Device string
	Clicks int64
}
type PlatformCount struct {
	Platform string
	Clicks   int64
}
type TopLink struct {
	ID           string
	ShortCode    string
	Title        *string
	OriginalURL  string
	TotalClicks  int64
	UniqueClicks int64
}

type FingerprintRepository interface {
	StoreForClick(ctx context.Context, clickID string, data domain.FingerprintData) error
	CandidateClicks(ctx context.Context, sinceMax time.Time) ([]FingerprintCandidate, error)
}

type FingerprintCandidate struct {
	ClickID                string
	LinkID                 string
	ClickedAt              time.Time
	AttributionWindowHours int
	Fingerprint            domain.FingerprintData
}

type InstallEventRepository interface {
	Insert(ctx context.Context, event *domain.InstallEvent) (id string, err error)
	GetByID(ctx context.Context, id string) (*domain.InstallEvent, error)
	SetDeepLinkData(ctx context.Context, id string, data map[string]interface{}) error
	LatestByFingerprintHash(ctx context.Context, hash string) (*domain.InstallEvent, error)
}

type InAppEventRepository interface {
	Insert(ctx context.Context, event *domain.InAppEvent) (id string, err error)
}

type WebhookRepository interface {
	Create(ctx context.Context, webhook *domain.Webhook) error
	List(ctx context.Context, userID *string) ([]*domain.Webhook, error)
	GetByID(ctx context.Context, id string, userID *string) (*domain.Webhook, error)
	Update(ctx context.Context, id string, userID *string, patch map[string]interface{}) (*domain.Webhook, error)
	Delete(ctx context.Context, id string, userID *string) error
	ActiveForUser(ctx context.Context, userID string) ([]*domain.Webhook, error)
	ActiveForLinkOwner(ctx context.Context, linkID string) ([]*domain.Webhook, error)
}

type TemplateRepository interface {
	Create(ctx context.Context, tpl *domain.LinkTemplate) error
	List(ctx context.Context, userID *string) ([]*domain.LinkTemplate, error)
	GetByID(ctx context.Context, id string, userID *string) (*domain.LinkTemplate, error)
	Update(ctx context.Context, id string, userID *string, patch map[string]interface{}) (*domain.LinkTemplate, error)
	Delete(ctx context.Context, id string, userID *string) error
	ExistsBySlug(ctx context.Context, slug string) (bool, error)
	SlugByID(ctx context.Context, id string) (string, error)
	LinkCountByTemplate(ctx context.Context, templateID string) (int64, error)
	UnsetDefaults(ctx context.Context, userID *string, exceptID string) error
	SetDefault(ctx context.Context, id string) (*domain.LinkTemplate, error)
}
