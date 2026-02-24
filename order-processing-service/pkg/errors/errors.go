package errors

import "fmt"

type OrderNotFoundError struct {
	OrderID string
}

func (e *OrderNotFoundError) Error() string {
	return fmt.Sprintf("order not found: %s", e.OrderID)
}

type PaymentFailedError struct {
	OrderID string
	Reason  string
}

func (e *PaymentFailedError) Error() string {
	return fmt.Sprintf("payment failed for order %s: %s", e.OrderID, e.Reason)
}

type InventoryError struct {
	ProductID string
	Reason    string
}

func (e *InventoryError) Error() string {
	return fmt.Sprintf("inventory error for product %s: %s", e.ProductID, e.Reason)
}

type ShippingError struct {
	OrderID string
	Reason  string
}

func (e *ShippingError) Error() string {
	return fmt.Sprintf("shipping error for order %s: %s", e.OrderID, e.Reason)
}

type ProxyError struct {
	Service string
	Method  string
	Err     error
}

func (e *ProxyError) Error() string {
	return fmt.Sprintf("proxy error calling %s.%s: %v", e.Service, e.Method, e.Err)
}

func NewProxyError(service, method string, err error) *ProxyError {
	return &ProxyError{Service: service, Method: method, Err: err}
}
