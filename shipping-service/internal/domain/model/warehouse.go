package model

import "time"

type Warehouse struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Code          string    `json:"code" gorm:"uniqueIndex"`
	Name          string    `json:"name"`
	Phone         string    `json:"phone"`
	Street        string    `json:"street"`
	WardCode      string    `json:"ward_code"`
	DistrictID    int       `json:"district_id"`
	ProvinceID    int       `json:"province_id"`
	CarrierShopID int64     `json:"carrier_shop_id"`
	Active        bool      `json:"active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Warehouse) TableName() string { return "warehouses" }
