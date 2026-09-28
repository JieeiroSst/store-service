package domain

import "time"

type ProductSales struct {
	ProductID     string
	ProductName   string
	Units         int
	RevenueCents  int
	DiscountCents int
}

type SalesReport struct {
	MachineID     string
	From          time.Time
	To            time.Time
	Orders        int
	RevenueCents  int
	DiscountCents int
	Products      []ProductSales
}
