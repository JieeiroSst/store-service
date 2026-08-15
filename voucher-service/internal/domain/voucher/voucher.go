package voucher

import (
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"golang.org/x/crypto/bcrypt"
)

type OwnerType string

const (
	OwnerTypeUser      OwnerType = "user"
	OwnerTypeCorporate OwnerType = "corporate"
)

type Voucher struct {
	ID         shared.VoucherID
	MerchantID shared.MerchantID
	OrderID    *shared.OrderID
	OwnerType  OwnerType
	OwnerID    *string

	ProductRef shared.ProductRef
	Code       string
	PinHash    string

	Status           Status
	Version          int
	PersistedVersion int

	IdempotencyKey string

	RedeemedAmount *shared.Money
	ProviderTxnRef string

	IssuedAt    *time.Time
	ActivatedAt *time.Time
	RedeemedAt  *time.Time
	ExpiresAt   *time.Time
	RevokedAt   *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	events []shared.DomainEvent
}

func NewVoucher(merchantID shared.MerchantID, ref shared.ProductRef, now time.Time) (*Voucher, error) {
	if merchantID.IsZero() {
		return nil, ErrInvalidVoucher
	}
	return &Voucher{
		ID:         shared.NewVoucherID(),
		MerchantID: merchantID,
		ProductRef: ref,
		Status:     StatusCreated,
		Version:    1,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func (v *Voucher) PullEvents() []shared.DomainEvent {
	events := v.events
	v.events = nil
	return events
}

func (v *Voucher) transitionTo(target Status, now time.Time) error {
	if !v.Status.canTransitionTo(target) {
		return ErrInvalidTransition
	}
	v.Status = target
	v.Version++
	v.UpdatedAt = now
	return nil
}

func (v *Voucher) Issue(code shared.VoucherCode, expiresAt *time.Time, now time.Time) error {
	if err := v.transitionTo(StatusIssued, now); err != nil {
		return err
	}
	v.Code = code.Code
	if code.PIN != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(code.PIN), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		v.PinHash = string(hash)
	}
	v.ExpiresAt = expiresAt
	v.IssuedAt = &now
	v.events = append(v.events, VoucherIssuedEvent{
		BaseEvent:  newEvent(EventTypeVoucherIssued, v.ID.String(), now),
		VoucherID:  v.ID.String(),
		MerchantID: v.MerchantID.String(),
		Code:       v.Code,
	})
	return nil
}

// Activate assigns an owner and moves ISSUED -> ACTIVE.
func (v *Voucher) Activate(ownerType OwnerType, ownerID string, now time.Time) error {
	if err := v.transitionTo(StatusActive, now); err != nil {
		return err
	}
	v.OwnerType = ownerType
	v.OwnerID = &ownerID
	v.ActivatedAt = &now
	v.events = append(v.events, VoucherActivatedEvent{
		BaseEvent: newEvent(EventTypeVoucherActivated, v.ID.String(), now),
		VoucherID: v.ID.String(),
	})
	return nil
}

func (v *Voucher) ValidatePIN(pin string) error {
	if v.PinHash == "" {
		return nil
	}
	if err := bcrypt.CompareHashAndPassword([]byte(v.PinHash), []byte(pin)); err != nil {
		return ErrInvalidPIN
	}
	return nil
}

func (v *Voucher) IsExpired(now time.Time) bool {
	return v.ExpiresAt != nil && now.After(*v.ExpiresAt)
}

func (v *Voucher) CanRedeem(now time.Time) bool {
	return v.Status == StatusActive && !v.IsExpired(now)
}

// Redeem moves ACTIVE -> REDEEMED, recording the amount actually consumed
func (v *Voucher) Redeem(amount shared.Money, providerTxnRef string, now time.Time) error {
	if v.IsExpired(now) {
		return ErrVoucherExpired
	}
	if v.Status == StatusRedeemed {
		return ErrAlreadyRedeemed
	}
	if v.ProductRef.Denomination.Currency != "" && v.ProductRef.Denomination.LessThan(amount) {
		return ErrInsufficientValue
	}
	if err := v.transitionTo(StatusRedeemed, now); err != nil {
		return err
	}
	v.RedeemedAmount = &amount
	v.ProviderTxnRef = providerTxnRef
	v.RedeemedAt = &now
	v.events = append(v.events, VoucherRedeemedEvent{
		BaseEvent:      newEvent(EventTypeVoucherRedeemed, v.ID.String(), now),
		VoucherID:      v.ID.String(),
		RedeemedAmount: amount,
		ProviderTxnRef: providerTxnRef,
	})
	return nil
}

// Revoke moves CREATED/ISSUED/ACTIVE -> REVOKED
func (v *Voucher) Revoke(reason string, now time.Time) error {
	if err := v.transitionTo(StatusRevoked, now); err != nil {
		return err
	}
	v.RevokedAt = &now
	v.events = append(v.events, VoucherRevokedEvent{
		BaseEvent: newEvent(EventTypeVoucherRevoked, v.ID.String(), now),
		VoucherID: v.ID.String(),
		Reason:    reason,
	})
	return nil
}

// Expire moves ISSUED/ACTIVE -> EXPIRED
func (v *Voucher) Expire(now time.Time) error {
	if err := v.transitionTo(StatusExpired, now); err != nil {
		return err
	}
	v.events = append(v.events, VoucherExpiredEvent{
		BaseEvent: newEvent(EventTypeVoucherExpired, v.ID.String(), now),
		VoucherID: v.ID.String(),
	})
	return nil
}
