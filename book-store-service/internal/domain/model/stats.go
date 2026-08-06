package model

// AuthorStat is a read-model aggregated across an author's books; it is not
// a persisted table.
type AuthorStat struct {
	AuthorID       int    `json:"author_id"`
	AuthorName     string `json:"author_name"`
	TotalReads     int    `json:"total_reads"`
	TotalPurchases int    `json:"total_purchases"`
}

// CategoryPriceStat is a read-model of price distribution per category; it
// is not a persisted table.
type CategoryPriceStat struct {
	CategoryID   int     `json:"category_id"`
	CategoryName string  `json:"category_name"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	AvgPrice     float64 `json:"avg_price"`
	BookCount    int     `json:"book_count"`
}
