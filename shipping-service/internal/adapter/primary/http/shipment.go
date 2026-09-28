package http

import (
	"errors"
	"net/http"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
	"github.com/gin-gonic/gin"
)

type quoteRequest struct {
	Recipient      model.Address `json:"recipient"`
	Parcel         model.Parcel  `json:"parcel"`
	CODAmount      int64         `json:"cod_amount"`
	InsuranceValue int64         `json:"insurance_value"`
	WarehouseIDs   []int64       `json:"warehouse_ids"`
}

func (h *Handler) Quote(c *gin.Context) {
	var req quoteRequest
	if !bind(c, &req) {
		return
	}
	result, err := h.shipments.Quote(c.Request.Context(), port.QuoteInput(req))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

type createShipmentRequest struct {
	ClientOrderCode string             `json:"client_order_code"`
	Recipient       model.Address      `json:"recipient"`
	Parcel          model.Parcel       `json:"parcel"`
	CODAmount       int64              `json:"cod_amount"`
	InsuranceValue  int64              `json:"insurance_value"`
	PaymentType     model.PaymentType  `json:"payment_type"`
	RequiredNote    model.RequiredNote `json:"required_note"`
	Note            string             `json:"note"`
	Strategy        model.Strategy     `json:"strategy"`
	WarehouseIDs    []int64            `json:"warehouse_ids"`
	WarehouseID     int64              `json:"warehouse_id"`
	ServiceID       int                `json:"service_id"`
	Customer        model.Customer     `json:"customer"`
	CallbackURL     string             `json:"callback_url"`
}

func (h *Handler) CreateShipment(c *gin.Context) {
	var req createShipmentRequest
	if !bind(c, &req) {
		return
	}
	shipment, created, err := h.shipments.Create(c.Request.Context(), client(c), port.CreateShipmentInput(req))
	h.respondShipment(c, shipment, created, err)
}

func (h *Handler) respondShipment(c *gin.Context, shipment *model.Shipment, created bool, err error) {
	if err != nil {
		if shipment != nil && (errors.Is(err, port.ErrCarrierRejected) || errors.Is(err, port.ErrCarrierUnavailable)) {
			status := http.StatusUnprocessableEntity
			if errors.Is(err, port.ErrCarrierUnavailable) {
				status = http.StatusBadGateway
			}
			c.JSON(status, gin.H{"error": err.Error(), "shipment": shipment})
			return
		}
		writeError(c, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	c.JSON(status, shipment)
}

func (h *Handler) PlaceShipment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	shipment, err := h.shipments.Place(c.Request.Context(), client(c), id)
	h.respondShipment(c, shipment, false, err)
}

func (h *Handler) GetShipment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	detail, err := h.shipments.Get(c.Request.Context(), client(c), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) GetShipmentByClientOrderCode(c *gin.Context) {
	detail, err := h.shipments.GetByClientOrderCode(c.Request.Context(), client(c), c.Param("code"))
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, detail)
}

func (h *Handler) ListShipments(c *gin.Context) {
	list, err := h.shipments.List(c.Request.Context(), client(c), port.ShipmentFilter{
		Status: model.Status(c.Query("status")),
		Limit:  queryInt(c, "limit"),
		Offset: queryInt(c, "offset"),
	})
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, list)
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

func (h *Handler) CancelShipment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req cancelRequest
	_ = c.ShouldBindJSON(&req)
	shipment, err := h.shipments.Cancel(c.Request.Context(), client(c), id, req.Reason)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, shipment)
}

func (h *Handler) SyncShipment(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	shipment, err := h.shipments.Sync(c.Request.Context(), client(c), id)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, shipment)
}
