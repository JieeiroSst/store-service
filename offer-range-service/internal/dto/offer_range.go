package dto

import "time"

type OfferRange struct {
	ID                 int       `json:"page"`
	Name               string    `json:"name"`
	Slug               string    `json:"slug"`
	Description        string    `json:"description"`
	IsPublic           bool      `json:"is_public"`
	IncluedAllProducts bool      `json:"inclued_all_products"`
	ProxyClass         string    `json:"proxy_class"`
	DateCreated        time.Time `json:"date_created"`
}
