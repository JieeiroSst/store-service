package dto

type OfferRangeExcludedProducts struct {
	ID        int`json:"id"`
	RangeID   int`json:"range_id"`
	ProductID int`json:"product_id"`
}
