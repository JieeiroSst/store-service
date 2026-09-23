package model

import "time"

type Side string

const (
	Buy  Side = "buy"
	Sell Side = "sell"
)

type OrderType string

const (
	LimitOrder  OrderType = "limit"
	MarketOrder OrderType = "market"
)

type TimeInForce string

const (
	GTC TimeInForce = "gtc" // rest on the book until filled or cancelled
	GTD TimeInForce = "gtd" // rest until ExpiresAt
	FOK TimeInForce = "fok" // fill completely right now or do nothing
	FAK TimeInForce = "fak" // fill what is possible right now, cancel the rest
)

type OrderStatus string

const (
	OrderOpen     OrderStatus = "open"
	OrderFilled   OrderStatus = "filled"
	OrderCanceled OrderStatus = "canceled"
	OrderExpired  OrderStatus = "expired"
)

type BookSide string

const (
	Bid BookSide = "bid"
	Ask BookSide = "ask"
)

type Order struct {
	ID            int64       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MarketID      int64       `gorm:"column:market_id" json:"market_id"`
	UserID        string      `gorm:"column:user_id" json:"user_id"`
	ClientOrderID string      `gorm:"column:client_order_id" json:"client_order_id,omitempty"`
	Outcome       Outcome     `gorm:"column:outcome" json:"outcome"`
	Side          Side        `gorm:"column:side" json:"side"`
	Type          OrderType   `gorm:"column:type" json:"type"`
	TimeInForce   TimeInForce `gorm:"column:time_in_force" json:"time_in_force"`
	Price         int64       `gorm:"column:price" json:"price"`
	Size          int64       `gorm:"column:size" json:"size"`
	Filled        int64       `gorm:"column:filled" json:"filled"`
	Budget        int64       `gorm:"column:budget" json:"budget,omitempty"`
	LockedCash    int64       `gorm:"column:locked_cash" json:"-"`
	FeeReserve    int64       `gorm:"column:fee_reserve" json:"-"`
	Fee           int64       `gorm:"column:fee" json:"fee"`
	FilledCash    int64       `gorm:"column:filled_cash" json:"filled_cash"`
	YesPrice      int64       `gorm:"column:yes_price" json:"-"`
	BookSide      BookSide    `gorm:"column:book_side" json:"-"`
	Status        OrderStatus `gorm:"column:status" json:"status"`
	ExpiresAt     *time.Time  `gorm:"column:expires_at" json:"expires_at,omitempty"`
	CreatedAt     time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time   `gorm:"column:updated_at" json:"updated_at"`
}

func (Order) TableName() string { return "orders" }

func (o *Order) Remaining() int64 { return o.Size - o.Filled }

func (o *Order) IsBudget() bool { return o.Budget > 0 }

func OwnPrice(o Outcome, yesPrice, shareValue int64) int64 {
	if o == OutcomeNo {
		return shareValue - yesPrice
	}
	return yesPrice
}

func BookPlacement(o Outcome, s Side, price, shareValue int64) (BookSide, int64) {
	yes := OwnPrice(o, price, shareValue)
	if (s == Buy) == (o == OutcomeYes) {
		return Bid, yes
	}
	return Ask, yes
}

type TradeKind string

const (
	KindSwap  TradeKind = "swap"  // one side hands existing shares to the other
	KindMint  TradeKind = "mint"  // two buyers fund a fresh YES/NO pair
	KindMerge TradeKind = "merge" // two sellers redeem a YES/NO pair
)

type Trade struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	MarketID     int64     `gorm:"column:market_id" json:"market_id"`
	YesPrice     int64     `gorm:"column:yes_price" json:"yes_price"`
	Size         int64     `gorm:"column:size" json:"size"`
	Kind         TradeKind `gorm:"column:kind" json:"kind"`
	MakerOrderID int64     `gorm:"column:maker_order_id" json:"maker_order_id"`
	TakerOrderID int64     `gorm:"column:taker_order_id" json:"taker_order_id"`
	MakerUserID  string    `gorm:"column:maker_user_id" json:"maker_user_id"`
	TakerUserID  string    `gorm:"column:taker_user_id" json:"taker_user_id"`
	MakerOutcome Outcome   `gorm:"column:maker_outcome" json:"maker_outcome"`
	MakerSide    Side      `gorm:"column:maker_side" json:"maker_side"`
	TakerOutcome Outcome   `gorm:"column:taker_outcome" json:"taker_outcome"`
	TakerSide    Side      `gorm:"column:taker_side" json:"taker_side"`

	TakerFee       int64  `gorm:"column:taker_fee" json:"taker_fee"`
	MakerRebate    int64  `gorm:"column:maker_rebate" json:"maker_rebate"`
	ReferralFee    int64  `gorm:"column:referral_fee" json:"referral_fee"`
	ReferrerUserID string `gorm:"column:referrer_user_id" json:"-"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Trade) TableName() string { return "trades" }
