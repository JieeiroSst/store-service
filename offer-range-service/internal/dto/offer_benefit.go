package dto

type OfferBenefit struct {
	ID               int    `json:"id"`
	Type             string `json:"type"`
	Value            int    `json:"value"`
	MaxAffectedItems int    `json:"max_affected_items"`
	ProxyClass       string `json:"proxy_class"`
	RangeID          int    `json:"page"`
}
