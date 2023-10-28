package domain

type VoucherVoucherOffer struct {
	ID               string `json:"id" db:"id"`
	VoucherID        string `json:"voucher_id" db:"voucher_id"`
	ConditionofferID string `json:"conditionoffer_id" db:"conditionoffer_id"`
}

