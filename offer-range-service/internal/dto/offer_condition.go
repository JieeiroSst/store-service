package dto

type OfferCondition struct {
	Id         int    `json:"id"`
	Type       string `json:"type"`
	Value      int    `json:"value"`
	ProxyClass string `json:"proxy_class"`
	RangeId    int    `json:"range_id"`
}
