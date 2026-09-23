package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
)

type Page[T any] struct {
	Items      []T    `json:"items"`
	NextCursor string `json:"next_cursor"`
	IsLastPage bool   `json:"is_last_page"`
}

type EventFilter struct {
	Status   model.EventStatus
	Category string
	Tag      string
	Query    string
	Featured bool
	Sort     string // volume (default), newest, ending
	Limit    int
	Cursor   string
}

type CategoryCount struct {
	Category string `json:"category"`
	Events   int64  `json:"events"`
}

type EventRepository interface {
	Create(ctx context.Context, event *model.Event) error
	GetByID(ctx context.Context, id int64) (*model.Event, error)
	GetBySlug(ctx context.Context, slug string) (*model.Event, error)
	GetByIDs(ctx context.Context, ids []int64) ([]model.Event, error)
	List(ctx context.Context, f EventFilter) (*Page[model.Event], error)
	Related(ctx context.Context, id int64, category string, limit int) ([]model.Event, error)
	Categories(ctx context.Context) ([]CategoryCount, error)
	AddVolume(ctx context.Context, id, delta int64) error
	Save(ctx context.Context, event *model.Event) error
}

type MarketFilter struct {
	Query  string
	Status model.MarketStatus
	Limit  int
	Offset int
}

type MarketRepository interface {
	Create(ctx context.Context, market *model.Market) error
	GetByID(ctx context.Context, id int64) (*model.Market, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*model.Market, error)
	GetBySlug(ctx context.Context, slug string) (*model.Market, error)
	GetByIDs(ctx context.Context, ids []int64) ([]model.Market, error)
	ListByEvents(ctx context.Context, eventIDs []int64) ([]model.Market, error)
	ListByEventForUpdate(ctx context.Context, eventID int64) ([]model.Market, error)
	ListRewarded(ctx context.Context) ([]model.Market, error)
	Search(ctx context.Context, f MarketFilter) ([]model.Market, int64, error)
	Save(ctx context.Context, market *model.Market) error
}

type OrderFilter struct {
	UserID   string
	MarketID int64
	Status   model.OrderStatus
	Limit    int
	Cursor   string
}

type BookLevel struct {
	YesPrice int64 `json:"price"`
	Size     int64 `json:"size"`
}

type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id int64) (*model.Order, error)
	GetByIDForUpdate(ctx context.Context, id int64) (*model.Order, error)
	GetByClientID(ctx context.Context, userID, clientOrderID string) (*model.Order, error)
	Save(ctx context.Context, order *model.Order) error
	ListCrossing(ctx context.Context, marketID int64, makerSide model.BookSide, yesPrice int64, limit int) ([]model.Order, error)
	ListOpenByMarketForUpdate(ctx context.Context, marketID int64) ([]model.Order, error)
	ListOpenByMarket(ctx context.Context, marketID int64) ([]model.Order, error)
	List(ctx context.Context, f OrderFilter) (*Page[model.Order], error)
	BookLevels(ctx context.Context, marketID int64, side model.BookSide) ([]BookLevel, error)
}

type TradeRepository interface {
	Create(ctx context.Context, trade *model.Trade) error
	ListByMarket(ctx context.Context, marketID int64, limit, offset int) ([]model.Trade, error)
	ListByMarketSince(ctx context.Context, marketID int64, since time.Time, limit int) ([]model.Trade, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.Trade, error)
}

type BalanceRepository interface {
	GetForUpdate(ctx context.Context, userID string) (*model.Balance, error)
	Get(ctx context.Context, userID string) (*model.Balance, error)
	Save(ctx context.Context, balance *model.Balance) error
}

type LedgerRepository interface {
	Create(ctx context.Context, entry *model.LedgerEntry) error
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.LedgerEntry, error)
}

type Holder struct {
	UserID string
	Shares int64
}

type PositionRepository interface {
	GetForUpdate(ctx context.Context, marketID int64, userID string, outcome model.Outcome) (*model.Position, error)
	Save(ctx context.Context, position *model.Position) error
	ListByUser(ctx context.Context, userID string, settled bool) ([]model.Position, error)
	ListByMarketForUpdate(ctx context.Context, marketID int64) ([]model.Position, error)
	TopHolders(ctx context.Context, marketID int64, outcome model.Outcome, limit int) ([]Holder, error)
}

type LeaderboardRow struct {
	UserID string
	Profit int64
	Volume int64
}

type StatsRepository interface {
	Leaderboard(ctx context.Context, metric string, since time.Time, limit int) ([]LeaderboardRow, error)
}

type CommentSort string

type SocialRepository interface {
	GetProfile(ctx context.Context, userID string) (*model.Profile, error)
	GetProfiles(ctx context.Context, userIDs []string) (map[string]model.Profile, error)
	SaveProfile(ctx context.Context, profile *model.Profile) error

	CreateComment(ctx context.Context, c *model.Comment) error
	GetComment(ctx context.Context, id int64) (*model.Comment, error)
	DeleteComment(ctx context.Context, id int64) error
	ListComments(ctx context.Context, eventID int64, sort string, limit int, cursor string) (*Page[model.Comment], error)
	AddLike(ctx context.Context, commentID int64, userID string) error
	RemoveLike(ctx context.Context, commentID int64, userID string) error

	AddBookmark(ctx context.Context, userID string, eventID int64) error
	RemoveBookmark(ctx context.Context, userID string, eventID int64) error
	ListBookmarks(ctx context.Context, userID string) ([]int64, error)
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

type Wallet struct {
	ID       string
	UserID   string
	Balance  int64
	Currency string
	Status   string
}

type TransferInput struct {
	SenderWalletID   string
	ReceiverWalletID string
	Amount           int64
	ReferenceID      string
	Description      string
}

type WalletGateway interface {
	GetWalletByUser(ctx context.Context, userID string) (*Wallet, error)
	Transfer(ctx context.Context, in TransferInput) (transferID string, err error)
	ReverseTransfer(ctx context.Context, transferID, reason string) error
}

type Notification struct {
	UserID  string
	Title   string
	Message string
}

type Notifier interface {
	Notify(ctx context.Context, n Notification) error
}

type Publisher interface {
	Publish(topic string, payload any)
}

type PnLRepository interface {
	CreateMany(ctx context.Context, entries []model.PnLEntry) error
}

type ExchangeAccountRepository interface {
	Add(ctx context.Context, entries ...model.ExchangeEntry) error
	Balance(ctx context.Context, bucket model.ExchangeBucket) (int64, error)
}

type RewardRepository interface {
	ClaimEpoch(ctx context.Context, start time.Time) (bool, error)
	CreatePayouts(ctx context.Context, payouts []model.RewardPayout) error
	ListPayouts(ctx context.Context, userID string, limit int, cursor string) (*Page[model.RewardPayout], error)
	TotalPaid(ctx context.Context, userID string) (int64, error)
}

type ReferralRepository interface {
	GetCode(ctx context.Context, userID string) (*model.ReferralCode, error)
	SaveCode(ctx context.Context, code *model.ReferralCode) error
	GetReferral(ctx context.Context, refereeUserID string) (*model.Referral, error)
	SaveReferral(ctx context.Context, r *model.Referral) error
	CountReferees(ctx context.Context, referrerUserID string) (int64, error)
	Earnings(ctx context.Context, referrerUserID string) (int64, error)
}

type UserDirectory interface {
	Validate(ctx context.Context, token string) (userID string, err error)
}

type ReferralLink struct {
	RefCode   string
	DeepLink  string
	ExpiresAt time.Time
}

type ReferralStats struct {
	TotalInvited   int64   `json:"total_invited"`
	TotalInstalled int64   `json:"total_installed"`
	TotalRewarded  int64   `json:"total_rewarded"`
	TotalRewardAmt float64 `json:"total_reward_amt"`
}

type ReferralGateway interface {
	GenerateLink(ctx context.Context, ownerUserID string) (*ReferralLink, error)
	Activate(ctx context.Context, refCode, userID string) (ownerUserID string, err error)
	Stats(ctx context.Context, userID string) (*ReferralStats, error)
}
