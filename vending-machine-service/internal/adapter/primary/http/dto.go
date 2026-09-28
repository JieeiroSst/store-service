package http

import (
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

type machineRequest struct {
	Location string `json:"location"`
	Model    string `json:"model"`
	Status   string `json:"status"`
}

type machinePatchRequest struct {
	Location *string `json:"location"`
	Model    *string `json:"model"`
}

type statusRequest struct {
	Status string `json:"status"`
}

type maintenanceRequest struct {
	TechnicianID    string `json:"technician_id"`
	MaintenanceType string `json:"maintenance_type"`
	Notes           string `json:"notes"`
}

type slotRequest struct {
	ProductID    string `json:"product_id"`
	Quantity     int    `json:"quantity"`
	MaxCapacity  int    `json:"max_capacity"`
	LowThreshold int    `json:"low_threshold"`
}

type restockRequest struct {
	Quantity int `json:"quantity"`
}

type categoryRequest struct {
	Name         string `json:"name"`
	Description  string `json:"description"`
	DisplayOrder int    `json:"display_order"`
}

type productRequest struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	PriceCents  int               `json:"price_cents"`
	CategoryID  string            `json:"category_id"`
	ImageURL    string            `json:"image_url"`
	Barcode     string            `json:"barcode"`
	IsActive    *bool             `json:"is_active"`
	Attributes  map[string]string `json:"attributes"`
}

type productPatchRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	PriceCents  *int    `json:"price_cents"`
	ImageURL    *string `json:"image_url"`
	Barcode     *string `json:"barcode"`
	IsActive    *bool   `json:"is_active"`
}

type sessionRequest struct {
	MachineID string `json:"machine_id"`
}

type reserveRequest struct {
	Slot string `json:"slot"`
}

type checkoutRequest struct {
	PaymentMethod string `json:"payment_method"`
	WalletID      string `json:"wallet_id"`
	CouponCode    string `json:"coupon_code"`
}

type refundRequest struct {
	Reason string `json:"reason"`
}

type dispenseRequest struct {
	Dispensed *bool `json:"dispensed"`
}

type machineResponse struct {
	ID              string     `json:"id"`
	Location        string     `json:"location"`
	Model           string     `json:"model"`
	Status          string     `json:"status"`
	LastMaintenance *time.Time `json:"last_maintenance_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func toMachine(m *domain.Machine) machineResponse {
	return machineResponse{
		ID:              m.ID,
		Location:        m.Location,
		Model:           m.Model,
		Status:          string(m.Status),
		LastMaintenance: m.LastMaintenance,
		CreatedAt:       m.CreatedAt,
		UpdatedAt:       m.UpdatedAt,
	}
}

type maintenanceResponse struct {
	ID              string    `json:"id"`
	MachineID       string    `json:"machine_id"`
	TechnicianID    string    `json:"technician_id"`
	MaintenanceType string    `json:"maintenance_type"`
	Notes           string    `json:"notes"`
	PerformedAt     time.Time `json:"performed_at"`
}

func toMaintenance(l *domain.MaintenanceLog) maintenanceResponse {
	return maintenanceResponse{
		ID:              l.ID,
		MachineID:       l.MachineID,
		TechnicianID:    l.TechnicianID,
		MaintenanceType: l.MaintenanceType,
		Notes:           l.Notes,
		PerformedAt:     l.PerformedAt,
	}
}

type eventResponse struct {
	ID            string         `json:"id"`
	EventType     string         `json:"event_type"`
	RelatedEntity string         `json:"related_entity"`
	EntityID      string         `json:"entity_id"`
	MachineID     string         `json:"machine_id"`
	Data          map[string]any `json:"data"`
	OccurredAt    time.Time      `json:"occurred_at"`
}

func toEvent(e *domain.Event) eventResponse {
	return eventResponse{
		ID:            e.ID,
		EventType:     e.EventType,
		RelatedEntity: e.RelatedEntity,
		EntityID:      e.EntityID,
		MachineID:     e.MachineID,
		Data:          e.Data,
		OccurredAt:    e.OccurredAt,
	}
}

type categoryResponse struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
}

func toCategory(c *domain.Category) categoryResponse {
	return categoryResponse{
		ID:           c.ID,
		Name:         c.Name,
		Description:  c.Description,
		DisplayOrder: c.DisplayOrder,
		CreatedAt:    c.CreatedAt,
	}
}

type productResponse struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	PriceCents  int               `json:"price_cents"`
	CategoryID  string            `json:"category_id"`
	ImageURL    string            `json:"image_url"`
	Barcode     string            `json:"barcode"`
	IsActive    bool              `json:"is_active"`
	Attributes  map[string]string `json:"attributes,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

func toProduct(p *domain.Product) productResponse {
	return productResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		PriceCents:  p.PriceCents,
		CategoryID:  p.CategoryID,
		ImageURL:    p.ImageURL,
		Barcode:     p.Barcode,
		IsActive:    p.IsActive,
		Attributes:  p.Attributes,
		CreatedAt:   p.CreatedAt,
	}
}

type inventoryResponse struct {
	ID            string           `json:"id"`
	MachineID     string           `json:"machine_id"`
	Slot          string           `json:"slot"`
	ProductID     string           `json:"product_id"`
	Product       *productResponse `json:"product,omitempty"`
	Quantity      int              `json:"quantity"`
	MaxCapacity   int              `json:"max_capacity"`
	LowThreshold  int              `json:"low_threshold"`
	IsLow         bool             `json:"is_low"`
	LastRestocked *time.Time       `json:"last_restocked_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func toInventory(i *domain.Inventory) inventoryResponse {
	resp := inventoryResponse{
		ID:            i.ID,
		MachineID:     i.MachineID,
		Slot:          i.SlotIdentifier,
		ProductID:     i.ProductID,
		Quantity:      i.Quantity,
		MaxCapacity:   i.MaxCapacity,
		LowThreshold:  i.LowThreshold,
		IsLow:         i.IsLow(),
		LastRestocked: i.LastRestocked,
		UpdatedAt:     i.UpdatedAt,
	}
	if i.Product != nil {
		p := toProduct(i.Product)
		resp.Product = &p
	}
	return resp
}

type sessionResponse struct {
	ID        string    `json:"id"`
	MachineID string    `json:"machine_id"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func toSession(s *domain.Session) sessionResponse {
	return sessionResponse{
		ID:        s.ID,
		MachineID: s.MachineID,
		Status:    string(s.Status),
		StartedAt: s.StartedAt,
		ExpiresAt: s.ExpiresAt,
	}
}

type reservationResponse struct {
	ID          string    `json:"id"`
	SessionID   string    `json:"session_id"`
	InventoryID string    `json:"inventory_id"`
	Status      string    `json:"status"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func toReservation(r *domain.Reservation) reservationResponse {
	return reservationResponse{
		ID:          r.ID,
		SessionID:   r.SessionID,
		InventoryID: r.InventoryID,
		Status:      string(r.Status),
		ExpiresAt:   r.ExpiresAt,
	}
}

type paymentResponse struct {
	ID            string     `json:"id"`
	AmountCents   int        `json:"amount_cents"`
	DiscountCents int        `json:"discount_cents"`
	CouponCode    string     `json:"coupon_code,omitempty"`
	Currency      string     `json:"currency"`
	Method        string     `json:"payment_method"`
	Status        string     `json:"status"`
	TransactionID string     `json:"transaction_id"`
	CompletedAt   *time.Time `json:"completed_at"`
}

type orderResponse struct {
	ID            string     `json:"id"`
	OrderNo       int64      `json:"order_no"`
	SessionID     string     `json:"session_id"`
	ReservationID string     `json:"reservation_id"`
	PaymentID     string     `json:"payment_id"`
	Status        string     `json:"status"`
	FulfilledAt   *time.Time `json:"fulfilled_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

func toOrder(o *domain.Order) orderResponse {
	return orderResponse{
		ID:            o.ID,
		OrderNo:       o.OrderNo,
		SessionID:     o.SessionID,
		ReservationID: o.ReservationID,
		PaymentID:     o.PaymentID,
		Status:        string(o.Status),
		FulfilledAt:   o.FulfilledAt,
		CreatedAt:     o.CreatedAt,
	}
}

type checkoutResponse struct {
	Order   orderResponse   `json:"order"`
	Payment paymentResponse `json:"payment"`
}

func toPayment(p *domain.Payment) paymentResponse {
	return paymentResponse{
		ID:            p.ID,
		AmountCents:   p.AmountCents,
		DiscountCents: p.DiscountCents,
		CouponCode:    p.CouponCode,
		Currency:      p.Currency,
		Method:        string(p.Method),
		Status:        string(p.Status),
		TransactionID: p.TransactionID,
		CompletedAt:   p.CompletedAt,
	}
}

func toCheckout(c *domain.Checkout) checkoutResponse {
	return checkoutResponse{Order: toOrder(c.Order), Payment: toPayment(c.Payment)}
}

type paymentStatusResponse struct {
	Payment paymentResponse `json:"payment"`
	Order   *orderResponse  `json:"order"`
}

func toPaymentStatus(c *domain.Checkout) paymentStatusResponse {
	resp := paymentStatusResponse{Payment: toPayment(c.Payment)}
	if c.Order != nil {
		o := toOrder(c.Order)
		resp.Order = &o
	}
	return resp
}

type productSalesResponse struct {
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	Units         int    `json:"units"`
	RevenueCents  int    `json:"revenue_cents"`
	DiscountCents int    `json:"discount_cents"`
}

type salesReportResponse struct {
	MachineID     string                 `json:"machine_id"`
	From          time.Time              `json:"from"`
	To            time.Time              `json:"to"`
	Orders        int                    `json:"orders"`
	RevenueCents  int                    `json:"revenue_cents"`
	DiscountCents int                    `json:"discount_cents"`
	Products      []productSalesResponse `json:"products"`
}

func toSalesReport(r *domain.SalesReport) salesReportResponse {
	return salesReportResponse{
		MachineID:     r.MachineID,
		From:          r.From,
		To:            r.To,
		Orders:        r.Orders,
		RevenueCents:  r.RevenueCents,
		DiscountCents: r.DiscountCents,
		Products: mapSlice(r.Products, func(p *domain.ProductSales) productSalesResponse {
			return productSalesResponse{
				ProductID:     p.ProductID,
				ProductName:   p.ProductName,
				Units:         p.Units,
				RevenueCents:  p.RevenueCents,
				DiscountCents: p.DiscountCents,
			}
		}),
	}
}

func mapSlice[T, R any](in []T, f func(*T) R) []R {
	out := make([]R, len(in))
	for i := range in {
		out[i] = f(&in[i])
	}
	return out
}
