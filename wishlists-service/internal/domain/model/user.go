package model

type Owner struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
}

type ProductStat struct {
	ProductID     int64  `json:"product_id"`
	Title         string `json:"title"`
	WishlistCount int64  `json:"wishlist_count"`
	TotalQuantity int64  `json:"total_quantity"`
}

type ProductMembership struct {
	WishlistID   int64  `json:"wishlist_id"`
	WishlistName string `json:"wishlist_name"`
	LineID       int64  `json:"line_id"`
	Quantity     int    `json:"quantity"`
}
