package order

import "github.com/JIeeiroSst/voucher-service/internal/domain/shared"

type OrderItem struct {
	ID               shared.OrderID
	MerchantID       shared.MerchantID
	ProductSKU       string
	Quantity         int
	UnitPrice        shared.Money
	LineTotal        shared.Money
	IssuedVoucherIDs []shared.VoucherID
}

func NewOrderItem(merchantID shared.MerchantID, sku string, quantity int, unitPrice shared.Money) (OrderItem, error) {
	if quantity <= 0 {
		return OrderItem{}, ErrInvalidOrder
	}
	return OrderItem{
		ID:         shared.NewOrderID(),
		MerchantID: merchantID,
		ProductSKU: sku,
		Quantity:   quantity,
		UnitPrice:  unitPrice,
		LineTotal:  shared.NewMoney(unitPrice.Amount*int64(quantity), unitPrice.Currency),
	}, nil
}
