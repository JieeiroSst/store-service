package port

import (
	"context"
	"strconv"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
)

// ---- catalog ----------------------------------------------------------

type CreateMarketInput struct {
	Slug            string
	Question        string
	GroupItemTitle  string
	Description     string
	EndTime         time.Time // defaults to the event's end date
	ShareValue      int64     // defaults to the configured value
	MinOrderSize    int64
	RewardPool      int64
	RewardMaxSpread int64
	RewardMinSize   int64
}

type CreateEventInput struct {
	Slug        string
	Title       string
	Description string
	Category    string
	Tags        []string
	ImageURL    string
	Featured    bool
	NegRisk     bool
	EndDate     time.Time
	Markets     []CreateMarketInput
}

type MarketList struct {
	Items []model.Market
	Total int64
}

type CatalogUsecase interface {
	CreateEvent(ctx context.Context, in CreateEventInput) (*model.Event, error)
	GetEvent(ctx context.Context, idOrSlug string) (*model.Event, error)
	ListEvents(ctx context.Context, f EventFilter) (*Page[model.Event], error)
	RelatedEvents(ctx context.Context, idOrSlug string, limit int) ([]model.Event, error)
	Categories(ctx context.Context) ([]CategoryCount, error)
	GetMarket(ctx context.Context, idOrSlug string) (*model.Market, error)
	SearchMarkets(ctx context.Context, f MarketFilter) (*MarketList, error)
}

// ---- resolution -------------------------------------------------------

type ResolutionUsecase interface {
	Propose(ctx context.Context, marketID int64, outcome model.Outcome) (*model.Market, error)
	// Dispute needs a bond, taken from the disputer's balance.
	Dispute(ctx context.Context, marketID int64, userID, reason string) (*model.Market, error)
	Finalize(ctx context.Context, marketID int64) (*model.Market, error)
	Resolve(ctx context.Context, marketID int64, outcome model.Outcome) (*model.Market, error)
	// ProposeEvent and ResolveEvent are the only way to resolve a neg-risk
	// event: exactly one market wins YES and all the others resolve NO.
	ProposeEvent(ctx context.Context, eventID, winnerMarketID int64) (*model.Event, error)
	ResolveEvent(ctx context.Context, eventID, winnerMarketID int64) (*model.Event, error)
}

// ---- exchange ---------------------------------------------------------

type PlaceOrderInput struct {
	MarketID      int64
	UserID        string
	ClientOrderID string
	Outcome       model.Outcome
	Side          model.Side
	Type          model.OrderType
	TimeInForce   model.TimeInForce
	Price         int64 // limit orders
	Size          int64 // limit orders and market sells, in shares
	Amount        int64 // market buys, cash to spend
	ExpiresAt     *time.Time
}

type PlaceOrderResult struct {
	Order  *model.Order
	Trades []model.Trade
}

type QuoteInput struct {
	MarketID int64
	Outcome  model.Outcome
	Side     model.Side
	Amount   int64
}

type Quote struct {
	Shares     int64
	Cash       int64
	AvgPrice   float64
	WorstPrice int64
	Fee        int64 // taker fee the trade would pay
	Fillable   bool
}

type OrderBook struct {
	MarketID  int64         `json:"market_id"`
	Outcome   model.Outcome `json:"outcome"`
	Bids      []BookLevel   `json:"bids"`
	Asks      []BookLevel   `json:"asks"`
	BestBid   int64         `json:"best_bid"`
	BestAsk   int64         `json:"best_ask"`
	Spread    int64         `json:"spread"`
	Midpoint  float64       `json:"midpoint"`
	LastTrade int64         `json:"last_trade_price"`
}

type Candle struct {
	Time   time.Time `json:"time"`
	Open   int64     `json:"open"`
	High   int64     `json:"high"`
	Low    int64     `json:"low"`
	Close  int64     `json:"close"`
	Volume int64     `json:"volume"`
}

type PriceHistoryInput struct {
	MarketID int64
	Outcome  model.Outcome
	Range    string // 1h, 6h, 1d, 1w, 1m, max
	Bucket   time.Duration
}

type ExchangeUsecase interface {
	PlaceOrder(ctx context.Context, in PlaceOrderInput) (*PlaceOrderResult, error)
	CancelOrder(ctx context.Context, orderID int64, userID string) (*model.Order, error)
	CancelAll(ctx context.Context, marketID int64, userID string) (int, error)
	GetOrder(ctx context.Context, orderID int64) (*model.Order, error)
	ListOrders(ctx context.Context, f OrderFilter) (*Page[model.Order], error)
	OrderBook(ctx context.Context, marketID int64, outcome model.Outcome) (*OrderBook, error)
	Quote(ctx context.Context, in QuoteInput) (*Quote, error)
	MarketTrades(ctx context.Context, marketID int64, limit, offset int) ([]model.Trade, error)
	PriceHistory(ctx context.Context, in PriceHistoryInput) ([]Candle, error)
	Convert(ctx context.Context, in ConvertInput) (*ConvertResult, error)
}

type ConvertInput struct {
	EventID   int64
	UserID    string
	MarketIDs []int64 // markets whose NO shares are given up
	Amount    int64   // shares of each
}

type ConvertResult struct {
	Cash       int64
	YesMarkets []int64
}

// ---- accounts ---------------------------------------------------------

type WalletBalance struct {
	WalletID string
	Balance  int64
	Currency string
	Status   string
}

type PositionView struct {
	model.Position
	Question      string             `json:"question"`
	EventID       int64              `json:"event_id"`
	AvgPrice      float64            `json:"avg_price"`
	CurrentPrice  float64            `json:"current_price"`
	Value         float64            `json:"value"`
	UnrealizedPnL float64            `json:"unrealized_pnl"`
	MarketStatus  model.MarketStatus `json:"market_status"`
}

type Portfolio struct {
	Balance        model.Balance  `json:"balance"`
	Positions      []PositionView `json:"positions"`
	PositionsValue float64        `json:"positions_value"`
	UnrealizedPnL  float64        `json:"unrealized_pnl"`
	RealizedPnL    int64          `json:"realized_pnl"`
	TotalValue     float64        `json:"total_value"`
}

type Activity struct {
	Type      string        `json:"type"` // trade, deposit, withdraw, withdraw_reversal, settlement
	MarketID  int64         `json:"market_id,omitempty"`
	Outcome   model.Outcome `json:"outcome,omitempty"`
	Side      model.Side    `json:"side,omitempty"`
	Price     int64         `json:"price,omitempty"`
	Size      int64         `json:"size,omitempty"`
	Amount    int64         `json:"amount"`
	CreatedAt time.Time     `json:"created_at"`
}

type AccountUsecase interface {
	Deposit(ctx context.Context, userID string, amount int64) (*model.Balance, error)
	Withdraw(ctx context.Context, userID string, amount int64) (*model.Balance, error)
	Balance(ctx context.Context, userID string) (*model.Balance, error)
	Wallet(ctx context.Context, userID string) (*WalletBalance, error)
	Portfolio(ctx context.Context, userID string) (*Portfolio, error)
	ClosedPositions(ctx context.Context, userID string) ([]model.Position, error)
	Activity(ctx context.Context, userID string, limit int) ([]Activity, error)
	FundExchange(ctx context.Context, fromUserID string, amount int64) (*ExchangeSummary, error)
	ExchangeSummary(ctx context.Context) (*ExchangeSummary, error)
}

type ExchangeSummary struct {
	Revenue  int64 `json:"revenue"`
	BondHeld int64 `json:"bond_held"`
}

// ---- social -----------------------------------------------------------

type UpdateProfileInput struct {
	UserID    string
	Username  string
	Bio       string
	AvatarURL string
}

type LeaderboardEntry struct {
	Rank     int    `json:"rank"`
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
	Profit   int64  `json:"profit"`
	Volume   int64  `json:"volume"`
}

type HolderEntry struct {
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
	Shares   int64  `json:"shares"`
}

type SocialUsecase interface {
	GetProfile(ctx context.Context, userID string) (*model.Profile, error)
	UpdateProfile(ctx context.Context, in UpdateProfileInput) (*model.Profile, error)

	AddComment(ctx context.Context, eventID int64, userID, body string, parentID int64) (*model.Comment, error)
	ListComments(ctx context.Context, eventID int64, sort string, limit int, cursor string) (*Page[model.Comment], error)
	DeleteComment(ctx context.Context, commentID int64, userID string) error
	LikeComment(ctx context.Context, commentID int64, userID string) error
	UnlikeComment(ctx context.Context, commentID int64, userID string) error

	Watch(ctx context.Context, userID string, eventID int64) error
	Unwatch(ctx context.Context, userID string, eventID int64) error
	Watchlist(ctx context.Context, userID string) ([]model.Event, error)

	Leaderboard(ctx context.Context, metric, window string, limit int) ([]LeaderboardEntry, error)
	TopHolders(ctx context.Context, marketID int64, outcome model.Outcome, limit int) ([]HolderEntry, error)
}

// ---- streaming --------------------------------------------------------

type StreamMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type EventStream interface {
	Subscribe(topic string) (<-chan StreamMessage, func())
}

func MarketTopic(marketID int64) string { return "market:" + strconv.FormatInt(marketID, 10) }

// ---- identity, referral, rewards -----------------------------------

type IdentityUsecase interface {
	Authenticate(ctx context.Context, token string) (userID string, err error)
}

type ReferralSummary struct {
	RefCode       string         `json:"ref_code"`
	DeepLink      string         `json:"deep_link"`
	ReferredBy    string         `json:"referred_by,omitempty"`
	Referees      int64          `json:"referees"`
	Earnings      int64          `json:"earnings"`
	CommissionBps int64          `json:"commission_bps"`
	ServiceStats  *ReferralStats `json:"service_stats,omitempty"`
}

type ReferralUsecase interface {
	Summary(ctx context.Context, userID string) (*ReferralSummary, error)
	Redeem(ctx context.Context, userID, refCode string) (*model.Referral, error)
}

type MarketRewards struct {
	MarketID     int64 `json:"market_id"`
	DailyPool    int64 `json:"daily_pool"`
	MaxSpread    int64 `json:"max_spread"`
	MinSize      int64 `json:"min_size"`
	EpochSeconds int64 `json:"epoch_seconds"`
	EpochPool    int64 `json:"epoch_pool"`
}

type UserRewards struct {
	TotalPaid int64 `json:"total_paid"`
}

type RewardUsecase interface {
	SetMarketRewards(ctx context.Context, marketID, dailyPool, maxSpread, minSize int64) (*model.Market, error)
	MarketRewards(ctx context.Context, marketID int64) (*MarketRewards, error)
	UserRewards(ctx context.Context, userID string, limit int, cursor string) (*Page[model.RewardPayout], *UserRewards, error)
	RunEpoch(ctx context.Context, now time.Time) (int, error)
}
