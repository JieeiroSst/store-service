package http

import (
	"net/http"
	"strconv"

	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type createOrderItemRequest struct {
	MenuItemID          string `json:"menu_item_id" binding:"required"`
	Quantity            int    `json:"quantity" binding:"required"`
	Customizations      string `json:"customizations"`
	SpecialInstructions string `json:"special_instructions"`
}

type createOrderRequest struct {
	CustomerID          string                   `json:"customer_id" binding:"required"`
	RestaurantID        string                   `json:"restaurant_id" binding:"required"`
	DeliveryAddressID   string                   `json:"delivery_address_id" binding:"required"`
	PaymentMethodID     string                   `json:"payment_method_id" binding:"required"`
	DeliveryFee         float64                  `json:"delivery_fee"`
	ServiceFee          float64                  `json:"service_fee"`
	Tax                 float64                  `json:"tax"`
	Tip                 float64                  `json:"tip"`
	SpecialInstructions string                   `json:"special_instructions"`
	Items               []createOrderItemRequest `json:"items" binding:"required,min=1"`
}

func (r createOrderRequest) toInput() port.CreateOrderInput {
	items := make([]port.CreateOrderItemInput, len(r.Items))
	for i, item := range r.Items {
		items[i] = port.CreateOrderItemInput{
			MenuItemID:          item.MenuItemID,
			Quantity:            item.Quantity,
			Customizations:      item.Customizations,
			SpecialInstructions: item.SpecialInstructions,
		}
	}
	return port.CreateOrderInput{
		CustomerID:          r.CustomerID,
		RestaurantID:        r.RestaurantID,
		DeliveryAddressID:   r.DeliveryAddressID,
		PaymentMethodID:     r.PaymentMethodID,
		DeliveryFee:         r.DeliveryFee,
		ServiceFee:          r.ServiceFee,
		Tax:                 r.Tax,
		Tip:                 r.Tip,
		SpecialInstructions: r.SpecialInstructions,
		Items:               items,
	}
}

func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orders.CreateOrder(c.Request.Context(), req.toInput())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, order)
}

func (h *Handler) GetOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	order, err := h.orders.GetOrder(c.Request.Context(), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

func (h *Handler) ListOrders(c *gin.Context) {
	if customerID := c.Query("customer_id"); customerID != "" {
		orders, err := h.orders.ListOrdersByCustomer(c.Request.Context(), customerID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
		return
	}
	if restaurantID := c.Query("restaurant_id"); restaurantID != "" {
		orders, err := h.orders.ListOrdersByRestaurant(c.Request.Context(), restaurantID)
		if err != nil {
			writeError(c, err)
			return
		}
		c.JSON(http.StatusOK, orders)
		return
	}
	c.JSON(http.StatusBadRequest, gin.H{"error": "customer_id or restaurant_id query param is required"})
}

type updateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

func (h *Handler) UpdateOrderStatus(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req updateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orders.UpdateOrderStatus(c.Request.Context(), id, model.OrderStatus(req.Status))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

type cancelOrderRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req cancelOrderRequest
	_ = c.ShouldBindJSON(&req)

	order, err := h.orders.CancelOrder(c.Request.Context(), id, req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, order)
}

type assignDriverRequest struct {
	DriverID string `json:"driver_id" binding:"required"`
}

func (h *Handler) AssignDriver(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var req assignDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	assignment, err := h.assignments.AssignDriver(c.Request.Context(), id, req.DriverID)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusCreated, assignment)
}
