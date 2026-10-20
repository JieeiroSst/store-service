package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
)

type CampaignFilter struct {
	GameType model.GameType
	Status   model.CampaignStatus
	Limit    int
	Offset   int
}

type PlayFilter struct {
	CampaignID int64
	GameType   model.GameType
	UserID     string
	DeviceID   string
	IP         string
	Limit      int
	Offset     int
}

type CheckinFilter struct {
	CampaignID int64
	UserID     string
	Limit      int
	Offset     int
}

type CampaignRepository interface {
	Create(ctx context.Context, c *model.Campaign) error
	Update(ctx context.Context, c *model.Campaign) error
	Get(ctx context.Context, id int64) (*model.Campaign, error)
	List(ctx context.Context, f CampaignFilter) ([]model.Campaign, int64, error)
	Delete(ctx context.Context, id int64) error
	SetStatus(ctx context.Context, id int64, status model.CampaignStatus) error
	Running(ctx context.Context, at time.Time) ([]model.Campaign, error)
}

type PlayerLedger interface {
	PlayByKey(ctx context.Context, key string) (*model.Play, error)
	Usage(ctx context.Context, day string) (model.Usage, error)
	ClaimPrize(ctx context.Context, prizeID int64, points int64) (bool, error)
	PointsWonOn(ctx context.Context, day string) (int64, error)
	SavePlay(ctx context.Context, p *model.Play) error
	Grant(ctx context.Context, event, referenceID string) (*model.CreditGrant, error)
	GrantsOn(ctx context.Context, event, day string) (int, error)
	SaveGrant(ctx context.Context, g *model.CreditGrant) error
	LastCheckin(ctx context.Context) (*model.Checkin, error)
	SaveCheckin(ctx context.Context, c *model.Checkin) error
}

type PlayRepository interface {
	WithPlayer(ctx context.Context, campaignID int64, userID string, fn func(PlayerLedger) error) error
	Usage(ctx context.Context, campaignID int64, userID, day string) (model.Usage, error)
	LastCheckin(ctx context.Context, campaignID int64, userID string) (*model.Checkin, error)
	List(ctx context.Context, f PlayFilter) ([]model.Play, int64, error)
	Checkins(ctx context.Context, f CheckinFilter) ([]model.Checkin, int64, error)
	RecentWins(ctx context.Context, campaignID int64, limit int) ([]model.Play, error)
}

type RewardOutbox interface {
	Pending(ctx context.Context, before time.Time, limit int) ([]model.Reward, error)
	MarkPublished(ctx context.Context, id int64, at time.Time) error
}

type RewardFilter struct {
	Status model.RewardStatus
	UserID string
	Limit  int
	Offset int
}

type FraudRepository interface {
	BlockOf(ctx context.Context, userID string) (*model.UserBlock, error)
	Block(ctx context.Context, b *model.UserBlock) (int64, error)
	Unblock(ctx context.Context, userID string) error
	Blocks(ctx context.Context, limit, offset int) ([]model.UserBlock, int64, error)
	Rewards(ctx context.Context, f RewardFilter) ([]model.ReviewReward, int64, error)
	Review(ctx context.Context, id int64, from []model.RewardStatus, to model.RewardStatus, by string) (*model.ReviewReward, error)
}

type FraudSignals interface {
	Observe(ctx context.Context, kind model.SignalKind, value, userID string) (int, error)
}

type RewardPublisher interface {
	Enabled() bool
	PublishReward(ctx context.Context, r model.Reward) error
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type WinnerFeed interface {
	Push(ctx context.Context, w model.Winner) error
	Recent(ctx context.Context, campaignID int64, limit int) ([]model.Winner, error)
}

type Clock interface {
	Now() time.Time
}

type Random interface {
	Intn(n int) int
}

type HealthChecker interface {
	Name() string
	Critical() bool
	Ping(ctx context.Context) error
}
