package port

import (
	"context"
	"encoding/json"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
)

type PrizeInput struct {
	ID       int64           `json:"id"`
	Name     string          `json:"name"`
	Type     model.PrizeType `json:"type"`
	Points   int64           `json:"points"`
	Weight   int             `json:"weight"`
	Quantity *int64          `json:"quantity"`
	Position int             `json:"position"`
	ImageURL string          `json:"image_url"`
}

type CampaignInput struct {
	GameType       string           `json:"game_type"`
	Name           string           `json:"name"`
	Description    string           `json:"description"`
	StartAt        time.Time        `json:"start_at"`
	EndAt          time.Time        `json:"end_at"`
	Timezone       string           `json:"timezone"`
	WinRate        float64          `json:"win_rate"`
	DailySpins     *int             `json:"daily_spins"`
	MaxSpinsPerDay int              `json:"max_spins_per_day"`
	MaxPointsDay   int64            `json:"max_points_per_user_day"`
	BudgetPoints   int64            `json:"budget_points"`
	ReviewPoints   int64            `json:"review_points"`
	SpinRules      []model.SpinRule `json:"spin_rules"`
	CheckinRewards []int64          `json:"checkin_rewards"`
	Settings       json.RawMessage  `json:"settings"`
	Prizes         []PrizeInput     `json:"prizes"`
}

type ListCampaignsInput struct {
	GameType string
	Status   string
	Limit    int
	Offset   int
}

type CampaignPage struct {
	Items  []model.Campaign `json:"items"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type CampaignUsecase interface {
	Create(ctx context.Context, actor model.Actor, in CampaignInput) (*model.Campaign, error)
	Get(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error)
	List(ctx context.Context, actor model.Actor, in ListCampaignsInput) (*CampaignPage, error)
	Update(ctx context.Context, actor model.Actor, id int64, in CampaignInput) (*model.Campaign, error)
	Delete(ctx context.Context, actor model.Actor, id int64) error
	Activate(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error)
	Pause(ctx context.Context, actor model.Actor, id int64) (*model.Campaign, error)
	Current(ctx context.Context, game string) (*model.Campaign, error)
	Lobby(ctx context.Context) ([]model.Campaign, error)
}

type PlayerStatus struct {
	CampaignID int64          `json:"campaign_id"`
	GameType   model.GameType `json:"game_type"`
	Day        string         `json:"day"`
	DailySpins int            `json:"daily_spins"`
	Remaining  int            `json:"remaining"`
	Usage      model.Usage    `json:"usage"`
	EndsAt     time.Time      `json:"ends_at"`
}

type PlayInput struct {
	Game           string
	Choice         int
	IdempotencyKey string
	Client         model.ClientInfo
}

type PlayResult struct {
	Play     model.Play   `json:"play"`
	Won      bool         `json:"won"`
	Replayed bool         `json:"replayed"`
	Status   PlayerStatus `json:"status"`
}

type GrantInput struct {
	UserID      string `json:"user_id"`
	Event       string `json:"event"`
	ReferenceID string `json:"reference_id"`
}

type GrantResult struct {
	Grants []model.CreditGrant `json:"grants"`
	Skips  []GrantSkip         `json:"skipped"`
}

type GrantSkip struct {
	CampaignID int64          `json:"campaign_id"`
	GameType   model.GameType `json:"game_type"`
	Reason     string         `json:"reason"`
}

type PlayPage struct {
	Items  []model.Play `json:"items"`
	Total  int64        `json:"total"`
	Limit  int          `json:"limit"`
	Offset int          `json:"offset"`
}

type ListPlaysInput struct {
	CampaignID int64
	GameType   string
	UserID     string
	DeviceID   string
	IP         string
	Limit      int
	Offset     int
}

type GameUsecase interface {
	Status(ctx context.Context, actor model.Actor, game string) (*PlayerStatus, error)
	Play(ctx context.Context, actor model.Actor, in PlayInput) (*PlayResult, error)
	History(ctx context.Context, actor model.Actor, in ListPlaysInput) (*PlayPage, error)
	CampaignPlays(ctx context.Context, actor model.Actor, in ListPlaysInput) (*PlayPage, error)
	GrantPlays(ctx context.Context, actor model.Actor, in GrantInput) (*GrantResult, error)
	Winners(ctx context.Context, game string, limit int) ([]model.Winner, error)
}

type CheckinStatus struct {
	CampaignID     int64     `json:"campaign_id"`
	Day            string    `json:"day"`
	CheckedInToday bool      `json:"checked_in_today"`
	Streak         int       `json:"streak"`
	CycleDay       int       `json:"cycle_day"`
	NextReward     int64     `json:"next_reward"`
	Rewards        []int64   `json:"rewards"`
	EndsAt         time.Time `json:"ends_at"`
}

type CheckinResult struct {
	Checkin model.Checkin `json:"checkin"`
	Status  CheckinStatus `json:"status"`
}

type CheckinPage struct {
	Items  []model.Checkin `json:"items"`
	Total  int64           `json:"total"`
	Limit  int             `json:"limit"`
	Offset int             `json:"offset"`
}

type CheckinUsecase interface {
	Status(ctx context.Context, actor model.Actor) (*CheckinStatus, error)
	CheckIn(ctx context.Context, actor model.Actor, client model.ClientInfo) (*CheckinResult, error)
	History(ctx context.Context, actor model.Actor, limit, offset int) (*CheckinPage, error)
}

type BlockInput struct {
	UserID string `json:"user_id"`
	Reason string `json:"reason"`
}

type BlockResult struct {
	Block       model.UserBlock `json:"block"`
	HeldRewards int64           `json:"held_rewards"`
}

type BlockPage struct {
	Items  []model.UserBlock `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}

type ListRewardsInput struct {
	Status string
	UserID string
	Limit  int
	Offset int
}

type RewardPage struct {
	Items  []model.ReviewReward `json:"items"`
	Total  int64                `json:"total"`
	Limit  int                  `json:"limit"`
	Offset int                  `json:"offset"`
}

type FraudUsecase interface {
	Block(ctx context.Context, actor model.Actor, in BlockInput) (*BlockResult, error)
	Unblock(ctx context.Context, actor model.Actor, userID string) error
	Blocks(ctx context.Context, actor model.Actor, limit, offset int) (*BlockPage, error)
	Rewards(ctx context.Context, actor model.Actor, in ListRewardsInput) (*RewardPage, error)
	Approve(ctx context.Context, actor model.Actor, id int64) (*model.ReviewReward, error)
	Reject(ctx context.Context, actor model.Actor, id int64) (*model.ReviewReward, error)
	Plays(ctx context.Context, actor model.Actor, in ListPlaysInput) (*PlayPage, error)
}
