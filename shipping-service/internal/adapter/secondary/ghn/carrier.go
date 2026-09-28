package ghn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

const (
	pathServices     = "/v2/shipping-order/available-services"
	pathFee          = "/v2/shipping-order/fee"
	pathLeadtime     = "/v2/shipping-order/leadtime"
	pathCreate       = "/v2/shipping-order/create"
	pathDetail       = "/v2/shipping-order/detail"
	pathDetailClient = "/v2/shipping-order/detail-by-client-code"
	pathCancel       = "/v2/switch-status/cancel"

	paymentShop  = 1
	paymentBuyer = 2
)

type item struct {
	Name     string `json:"name"`
	Code     string `json:"code,omitempty"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price,omitempty"`
	Weight   int    `json:"weight"`
	Length   int    `json:"length,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

func items(p model.Parcel) []item {
	out := make([]item, len(p.Items))
	for i, it := range p.Items {
		out[i] = item{Name: it.Name, Code: it.Code, Quantity: it.Quantity, Price: it.Price, Weight: it.WeightGram}
	}
	return out
}

func (c *Client) AvailableServices(ctx context.Context, shopID int64, fromDistrictID, toDistrictID int) ([]port.CarrierService, error) {
	var data []struct {
		ServiceID     int    `json:"service_id"`
		ShortName     string `json:"short_name"`
		ServiceTypeID int    `json:"service_type_id"`
	}
	body := map[string]any{"shop_id": shopID, "from_district": fromDistrictID, "to_district": toDistrictID}
	if err := c.do(ctx, shopID, pathServices, body, &data); err != nil {
		return nil, err
	}
	out := make([]port.CarrierService, 0, len(data))
	for _, d := range data {
		out = append(out, port.CarrierService{ID: d.ServiceID, TypeID: d.ServiceTypeID, Name: d.ShortName})
	}
	return out, nil
}

func (c *Client) Quote(ctx context.Context, req port.CarrierQuoteRequest) (*port.CarrierQuote, error) {
	var fee struct {
		Total int64 `json:"total"`
	}
	err := c.do(ctx, req.ShopID, pathFee, map[string]any{
		"service_id":       req.ServiceID,
		"from_district_id": req.FromDistrictID,
		"from_ward_code":   req.FromWardCode,
		"to_district_id":   req.ToDistrictID,
		"to_ward_code":     req.ToWardCode,
		"weight":           req.Parcel.WeightGram,
		"length":           req.Parcel.LengthCm,
		"width":            req.Parcel.WidthCm,
		"height":           req.Parcel.HeightCm,
		"insurance_value":  req.InsuranceValue,
		"cod_value":        req.CODAmount,
		"items":            items(req.Parcel),
	}, &fee)
	if err != nil {
		return nil, err
	}

	var lead struct {
		Leadtime int64 `json:"leadtime"`
	}
	err = c.do(ctx, req.ShopID, pathLeadtime, map[string]any{
		"from_district_id": req.FromDistrictID,
		"from_ward_code":   req.FromWardCode,
		"to_district_id":   req.ToDistrictID,
		"to_ward_code":     req.ToWardCode,
		"service_id":       req.ServiceID,
	}, &lead)
	if err != nil {
		return nil, err
	}
	if lead.Leadtime <= 0 {
		return nil, fmt.Errorf("%w: ghn returned no leadtime for service %d", port.ErrCarrierRejected, req.ServiceID)
	}
	return &port.CarrierQuote{Fee: fee.Total, ExpectedDeliveryAt: time.Unix(lead.Leadtime, 0).UTC()}, nil
}

func (c *Client) CreateOrder(ctx context.Context, req port.CarrierOrderRequest) (*port.CarrierOrder, error) {
	payment := paymentShop
	if req.PaymentType == model.PaymentByRecipient {
		payment = paymentBuyer
	}
	var data struct {
		OrderCode            string `json:"order_code"`
		TotalFee             int64  `json:"total_fee"`
		ExpectedDeliveryTime string `json:"expected_delivery_time"`
	}
	err := c.do(ctx, req.ShopID, pathCreate, map[string]any{
		"client_order_code": req.ClientOrderCode,
		"to_name":           req.Recipient.Name,
		"to_phone":          req.Recipient.Phone,
		"to_address":        req.Recipient.Street,
		"to_ward_code":      req.Recipient.WardCode,
		"to_district_id":    req.Recipient.DistrictID,
		"cod_amount":        req.CODAmount,
		"content":           req.Parcel.Content,
		"weight":            req.Parcel.WeightGram,
		"length":            req.Parcel.LengthCm,
		"width":             req.Parcel.WidthCm,
		"height":            req.Parcel.HeightCm,
		"insurance_value":   req.InsuranceValue,
		"service_id":        req.ServiceID,
		"service_type_id":   req.ServiceTypeID,
		"payment_type_id":   payment,
		"required_note":     string(req.RequiredNote),
		"note":              req.Note,
		"items":             items(req.Parcel),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.OrderCode == "" {
		return nil, fmt.Errorf("%w: ghn create returned no order_code", port.ErrCarrierUnavailable)
	}
	return &port.CarrierOrder{
		OrderCode:          data.OrderCode,
		Status:             "ready_to_pick",
		Fee:                data.TotalFee,
		ExpectedDeliveryAt: parseTime(data.ExpectedDeliveryTime),
	}, nil
}

type orderDetail struct {
	OrderCode   string `json:"order_code"`
	Status      string `json:"status"`
	Leadtime    string `json:"leadtime"`
	UpdatedDate string `json:"updated_date"`
}

func (d orderDetail) toOrder() *port.CarrierOrder {
	return &port.CarrierOrder{
		OrderCode:          d.OrderCode,
		Status:             d.Status,
		ExpectedDeliveryAt: parseTime(d.Leadtime),
		UpdatedAt:          parseTime(d.UpdatedDate),
	}
}

func (c *Client) FindByClientOrderCode(ctx context.Context, shopID int64, clientOrderCode string) (*port.CarrierOrder, error) {
	var d orderDetail
	err := c.do(ctx, shopID, pathDetailClient, map[string]any{"client_order_code": clientOrderCode}, &d)
	if errors.Is(err, port.ErrCarrierRejected) || (err == nil && d.OrderCode == "") {
		return nil, fmt.Errorf("%w: ghn order with client code %s", port.ErrNotFound, clientOrderCode)
	}
	if err != nil {
		return nil, err
	}
	return d.toOrder(), nil
}

func (c *Client) GetOrder(ctx context.Context, shopID int64, orderCode string) (*port.CarrierOrder, error) {
	var d orderDetail
	if err := c.do(ctx, shopID, pathDetail, map[string]any{"order_code": orderCode}, &d); err != nil {
		return nil, err
	}
	return d.toOrder(), nil
}

func (c *Client) CancelOrder(ctx context.Context, shopID int64, orderCode string) error {
	var results []struct {
		OrderCode string `json:"order_code"`
		Result    bool   `json:"result"`
		Message   string `json:"message"`
	}
	if err := c.do(ctx, shopID, pathCancel, map[string]any{"order_codes": []string{orderCode}}, &results); err != nil {
		return err
	}
	for _, r := range results {
		if r.OrderCode == orderCode && !r.Result {
			return fmt.Errorf("%w: ghn refused to cancel %s: %s", port.ErrCarrierRejected, orderCode, r.Message)
		}
	}
	return nil
}
