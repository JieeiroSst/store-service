package model

type OfferRangeproduct struct {
	ID           int `json:"id"`
	DisplayOffer int `json:"display_offer"`
	ProductId    int `json:"product_id"`
	RangeId      int `json:"range_id"`
}
