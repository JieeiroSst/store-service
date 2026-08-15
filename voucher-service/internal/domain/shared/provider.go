package shared

type ProviderType string

const (
	ProviderTypeAPI   ProviderType = "api"
	ProviderTypeStock ProviderType = "stock"
	ProviderTypeSelf  ProviderType = "self"
)

func (t ProviderType) Valid() bool {
	switch t {
	case ProviderTypeAPI, ProviderTypeStock, ProviderTypeSelf:
		return true
	default:
		return false
	}
}

type ProductRef struct {
	MerchantID   MerchantID `json:"merchant_id"`
	SKU          string     `json:"sku"`
	Denomination Money      `json:"denomination"`
}

type VoucherCode struct {
	Code string `json:"code"`
	PIN  string `json:"pin,omitempty"`
}

type ValidationResult struct {
	Valid   bool   `json:"valid"`
	Balance Money  `json:"balance"`
	Reason  string `json:"reason,omitempty"`
}

type RedeemResult struct {
	Success        bool   `json:"success"`
	RedeemedAmount Money  `json:"redeemed_amount"`
	ProviderTxnRef string `json:"provider_txn_ref,omitempty"`
}
