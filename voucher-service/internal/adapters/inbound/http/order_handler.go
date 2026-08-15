package http

import (
	stdhttp "net/http"

	orderapp "github.com/JIeeiroSst/voucher-service/internal/application/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	svc orderapp.OrderService
}

func NewOrderHandler(svc orderapp.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

type createOrderItemRequest struct {
	MerchantID string `json:"merchant_id" binding:"required"`
	ProductSKU string `json:"product_sku" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required"`
	UnitPrice  int64  `json:"unit_price" binding:"required"`
}

type createOrderRequest struct {
	BuyerType      string                    `json:"buyer_type" binding:"required"`
	BuyerID        string                    `json:"buyer_id" binding:"required"`
	Currency       string                    `json:"currency" binding:"required"`
	Items          []createOrderItemRequest  `json:"items" binding:"required,dive"`
	IdempotencyKey string                    `json:"idempotency_key" binding:"required"`
}

func (h *OrderHandler) Create(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}

	items := make([]orderapp.CreateOrderItemInput, 0, len(req.Items))
	for _, item := range req.Items {
		merchantID, err := shared.ParseMerchantID(item.MerchantID)
		if err != nil {
			c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid merchant_id"}})
			return
		}
		items = append(items, orderapp.CreateOrderItemInput{
			MerchantID: merchantID,
			ProductSKU: item.ProductSKU,
			Quantity:   item.Quantity,
			UnitPrice:  shared.NewMoney(item.UnitPrice, req.Currency),
		})
	}

	o, err := h.svc.CreateOrder(c.Request.Context(), orderapp.CreateOrderInput{
		BuyerType:      order.BuyerType(req.BuyerType),
		BuyerID:        req.BuyerID,
		Currency:       req.Currency,
		Items:          items,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, toOrderView(o))
}

type checkoutRequest struct {
	PaymentMethod  string `json:"payment_method" binding:"required"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

func (h *OrderHandler) Checkout(c *gin.Context) {
	orderID, err := shared.ParseOrderID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid order id"}})
		return
	}
	var req checkoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}

	out, err := h.svc.Checkout(c.Request.Context(), orderapp.CheckoutInput{
		OrderID:        orderID,
		PaymentMethod:  req.PaymentMethod,
		IdempotencyKey: req.IdempotencyKey,
	})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, out)
}

func (h *OrderHandler) Get(c *gin.Context) {
	orderID, err := shared.ParseOrderID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid order id"}})
		return
	}
	o, err := h.svc.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, toOrderView(o))
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	orderID, err := shared.ParseOrderID(c.Param("id"))
	if err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": "invalid order id"}})
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	_ = c.ShouldBindJSON(&body)

	if err := h.svc.CancelOrder(c.Request.Context(), orderID, body.Reason); err != nil {
		mapError(c, err)
		return
	}
	c.Status(stdhttp.StatusNoContent)
}

type orderView struct {
	ID          string  `json:"id"`
	Status      string  `json:"status"`
	TotalAmount int64   `json:"total_amount"`
	Currency    string  `json:"currency"`
}

func toOrderView(o *order.Order) orderView {
	return orderView{ID: o.ID.String(), Status: string(o.Status), TotalAmount: o.TotalAmount.Amount, Currency: o.TotalAmount.Currency}
}
