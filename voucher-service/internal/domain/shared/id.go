package shared

import "github.com/google/uuid"

type VoucherID uuid.UUID

func NewVoucherID() VoucherID  { return VoucherID(uuid.New()) }
func (id VoucherID) String() string { return uuid.UUID(id).String() }
func (id VoucherID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseVoucherID(s string) (VoucherID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return VoucherID{}, err
	}
	return VoucherID(u), nil
}

type OrderID uuid.UUID

func NewOrderID() OrderID      { return OrderID(uuid.New()) }
func (id OrderID) String() string { return uuid.UUID(id).String() }
func (id OrderID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseOrderID(s string) (OrderID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return OrderID{}, err
	}
	return OrderID(u), nil
}

type MerchantID uuid.UUID

func NewMerchantID() MerchantID   { return MerchantID(uuid.New()) }
func (id MerchantID) String() string { return uuid.UUID(id).String() }
func (id MerchantID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseMerchantID(s string) (MerchantID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return MerchantID{}, err
	}
	return MerchantID(u), nil
}

type CorporateID uuid.UUID

func NewCorporateID() CorporateID   { return CorporateID(uuid.New()) }
func (id CorporateID) String() string { return uuid.UUID(id).String() }
func (id CorporateID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseCorporateID(s string) (CorporateID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return CorporateID{}, err
	}
	return CorporateID(u), nil
}

type WalletID uuid.UUID

func NewWalletID() WalletID    { return WalletID(uuid.New()) }
func (id WalletID) String() string { return uuid.UUID(id).String() }
func (id WalletID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseWalletID(s string) (WalletID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return WalletID{}, err
	}
	return WalletID(u), nil
}

type UserID uuid.UUID

func NewUserID() UserID     { return UserID(uuid.New()) }
func (id UserID) String() string { return uuid.UUID(id).String() }
func (id UserID) IsZero() bool   { return uuid.UUID(id) == uuid.Nil }

func ParseUserID(s string) (UserID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, err
	}
	return UserID(u), nil
}
