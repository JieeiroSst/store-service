package voucher

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

const (
	EventTypeVoucherIssued    = "voucher.issued"
	EventTypeVoucherActivated = "voucher.activated"
	EventTypeVoucherRedeemed  = "voucher.redeemed"
	EventTypeVoucherRevoked   = "voucher.revoked"
	EventTypeVoucherExpired   = "voucher.expired"
)

type VoucherIssuedEvent struct {
	shared.BaseEvent
	VoucherID  string
	MerchantID string
	Code       string
}

type VoucherActivatedEvent struct {
	shared.BaseEvent
	VoucherID string
}

type VoucherRedeemedEvent struct {
	shared.BaseEvent
	VoucherID      string
	RedeemedAmount shared.Money
	ProviderTxnRef string
}

type VoucherRevokedEvent struct {
	shared.BaseEvent
	VoucherID string
	Reason    string
}

type VoucherExpiredEvent struct {
	shared.BaseEvent
	VoucherID string
}

func newEvent(eventType, voucherID string, now time.Time) shared.BaseEvent {
	return shared.NewBaseEvent(eventType, voucherID, now)
}
