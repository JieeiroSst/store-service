package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

const shippingCurrency = "VND"

type checkoutService struct {
	baskets     port.BasketRepository
	lines       port.BasketLineRepository
	catalogue   port.Catalogue
	wallet      port.Wallet
	fulfillment port.Fulfillment
	addresses   port.AddressBook
	rates       port.ShippingRates
	now         func() time.Time
}

func NewCheckoutService(baskets port.BasketRepository, lines port.BasketLineRepository, catalogue port.Catalogue, wallet port.Wallet, fulfillment port.Fulfillment, addresses port.AddressBook, rates port.ShippingRates) port.CheckoutUsecase {
	return &checkoutService{baskets: baskets, lines: lines, catalogue: catalogue, wallet: wallet, fulfillment: fulfillment, addresses: addresses, rates: rates, now: time.Now}
}

type pricedBasket struct {
	basket   *model.Basket
	lines    []model.BasketLine
	currency string
	subtotal float64
}

type shippingPlan struct {
	address  *port.ShippingAddress
	snapshot string
	quote    *port.ShippingQuote
}

func (p shippingPlan) fee() float64 {
	if p.quote == nil {
		return 0
	}
	return float64(p.quote.Fee)
}

func (s *checkoutService) openBasket(ctx context.Context, basketID, userID int) (*pricedBasket, error) {
	basket, err := s.baskets.GetByID(ctx, basketID)
	if err != nil {
		return nil, err
	}
	if basket.UserID != userID {
		return nil, common.ErrNotFound
	}
	if !basket.IsOpen() {
		return nil, fmt.Errorf("%w: basket %d is already %s", common.ErrConflict, basket.ID, basket.Status)
	}
	lines, err := s.lines.ListByBasket(ctx, basket.ID)
	if err != nil {
		return nil, err
	}
	if len(lines) == 0 {
		return nil, fmt.Errorf("%w: basket is empty", common.ErrInvalidRequest)
	}
	p := &pricedBasket{basket: basket, lines: lines}
	for i := range lines {
		record, err := s.catalogue.StockRecord(ctx, lines[i].StockrecordID)
		if err != nil {
			return nil, err
		}
		if p.currency == "" {
			p.currency = record.PriceCurrency
		} else if p.currency != record.PriceCurrency {
			return nil, fmt.Errorf("%w: basket mixes %s and %s", common.ErrInvalidRequest, p.currency, record.PriceCurrency)
		}
		lines[i].PriceCurrency = record.PriceCurrency
		lines[i].PriceExclTax = record.Price
		lines[i].PriceInclTax = record.Price
		lines[i].ProductID = record.ProductID
		p.subtotal += record.Price * float64(lines[i].Quantity)
	}
	p.subtotal = roundMoney(p.subtotal)
	return p, nil
}

func (s *checkoutService) planShipping(ctx context.Context, p *pricedBasket, addressID int, token string) (shippingPlan, error) {
	var plan shippingPlan
	if addressID <= 0 || s.addresses == nil {
		return plan, nil
	}
	addr, err := s.addresses.ShippingAddress(ctx, token, addressID)
	switch {
	case errors.Is(err, common.ErrNotFound):
		return plan, fmt.Errorf("%w: shipping address %d not found", common.ErrInvalidRequest, addressID)
	case err != nil && s.rates != nil:
		return plan, err
	case err != nil:
		log.Printf("basket %d: read shipping address %d: %v", p.basket.ID, addressID, err)
		return plan, nil
	}
	raw, err := json.Marshal(addr)
	if err != nil {
		return plan, err
	}
	plan.address, plan.snapshot = addr, string(raw)
	if s.rates == nil {
		return plan, nil
	}
	if addr.WardCode == "" || addr.DistrictID <= 0 {
		return plan, fmt.Errorf("%w: địa chỉ giao hàng chưa có Tỉnh/Quận/Phường theo GHN, hãy chọn lại địa chỉ", common.ErrInvalidRequest)
	}
	if p.currency != shippingCurrency {
		return plan, fmt.Errorf("%w: phí giao hàng chỉ tính bằng %s", common.ErrInvalidRequest, shippingCurrency)
	}
	items := make([]port.ShippingItem, 0, len(p.lines))
	for _, l := range p.lines {
		items = append(items, port.ShippingItem{
			Name:     fmt.Sprintf("Sản phẩm #%d", l.ProductID),
			Code:     strconv.Itoa(l.ProductID),
			Quantity: l.Quantity,
			Price:    int64(math.Round(l.PriceInclTax)),
		})
	}
	quote, err := s.rates.Quote(ctx, *addr, items, int64(math.Round(p.subtotal)))
	if err != nil {
		return plan, err
	}
	plan.quote = quote
	return plan, nil
}

func (s *checkoutService) QuoteShipping(ctx context.Context, cmd port.ShippingQuoteCommand) (*port.ShippingQuoteResult, error) {
	p, err := s.openBasket(ctx, cmd.BasketID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	plan, err := s.planShipping(ctx, p, cmd.ShippingAddressID, cmd.AccessToken)
	if err != nil {
		return nil, err
	}
	out := &port.ShippingQuoteResult{
		Currency:        p.currency,
		Subtotal:        p.subtotal,
		ShippingFee:     plan.fee(),
		Total:           roundMoney(p.subtotal + plan.fee()),
		ShippingCharged: plan.quote != nil,
	}
	if plan.quote != nil {
		out.ServiceName, out.ExpectedDeliveryAt = plan.quote.ServiceName, plan.quote.ExpectedDeliveryAt
	}
	return out, nil
}

func (s *checkoutService) Checkout(ctx context.Context, cmd port.CheckoutCommand) (*model.Basket, error) {
	if strings.TrimSpace(cmd.WalletID) == "" || strings.TrimSpace(cmd.Pin) == "" {
		return nil, fmt.Errorf("%w: wallet_id and pin are required", common.ErrInvalidRequest)
	}
	p, err := s.openBasket(ctx, cmd.BasketID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	basket, lines := p.basket, p.lines
	plan, err := s.planShipping(ctx, p, cmd.ShippingAddressID, cmd.AccessToken)
	if err != nil {
		return nil, err
	}
	fee := plan.fee()
	if plan.quote != nil && (cmd.ShippingFee == nil || fee > *cmd.ShippingFee) {
		return nil, fmt.Errorf("%w: phí giao hàng là %.0f %s, hãy xác nhận lại trước khi thanh toán", common.ErrConflict, fee, shippingCurrency)
	}
	total := roundMoney(p.subtotal + fee)

	allocated := make([]model.BasketLine, 0, len(lines))
	release := func() {
		for _, l := range allocated {
			if err := s.catalogue.Release(context.WithoutCancel(ctx), l.StockrecordID, l.Quantity); err != nil {
				log.Printf("basket %d: release stock record %d: %v", basket.ID, l.StockrecordID, err)
			}
		}
	}
	for _, l := range lines {
		if err := s.catalogue.Allocate(ctx, l.StockrecordID, l.Quantity); err != nil {
			release()
			if errors.Is(err, common.ErrOutOfStock) {
				return nil, fmt.Errorf("%w: product %d", common.ErrOutOfStock, l.ProductID)
			}
			return nil, err
		}
		allocated = append(allocated, l)
	}

	metadata := map[string]string{"basket_id": fmt.Sprint(basket.ID)}
	if fee > 0 {
		metadata["shipping_fee"] = strconv.FormatFloat(fee, 'f', 0, 64)
	}
	txID, err := s.wallet.Pay(ctx, port.WalletPayment{
		UserID:         cmd.UserID,
		WalletID:       cmd.WalletID,
		Amount:         total,
		Pin:            cmd.Pin,
		Description:    fmt.Sprintf("Keiki Shop basket #%d", basket.ID),
		IdempotencyKey: checkoutKey(basket),
		Metadata:       metadata,
	})
	if err != nil {
		release()
		return nil, err
	}

	for i := range lines {
		if err := s.lines.Update(ctx, &lines[i]); err != nil {
			log.Printf("basket %d: persist final price of line %d: %v", basket.ID, lines[i].ID, err)
		}
	}
	now := s.now().UTC().Format(time.RFC3339)
	basket.Status = model.BasketStatusSubmitted
	basket.DateSubmitted = now
	basket.DatePlaced = now
	basket.Currency = p.currency
	basket.TotalInclTax = total
	basket.ShippingInclTax = fee
	if plan.quote != nil {
		basket.ShippingWarehouse, basket.ShippingService = plan.quote.WarehouseID, plan.quote.ServiceID
	}
	basket.PaymentReference = txID
	if cmd.ShippingAddressID > 0 {
		basket.ShippingAddressID = cmd.ShippingAddressID
		basket.ShippingSnapshot = plan.snapshot
	}
	if err := s.baskets.Update(ctx, basket); err != nil {
		log.Printf("basket %d: paid with %s but failed to mark submitted: %v", basket.ID, txID, err)
		return nil, err
	}
	s.startFulfillment(ctx, basket, lines)
	return basket, nil
}

func (s *checkoutService) startFulfillment(ctx context.Context, basket *model.Basket, lines []model.BasketLine) bool {
	if s.fulfillment == nil {
		return false
	}
	order := port.PaidOrder{
		BasketID:          basket.ID,
		UserID:            basket.UserID,
		Currency:          basket.Currency,
		Total:             basket.TotalInclTax,
		ShippingFee:       basket.ShippingInclTax,
		PaymentReference:  basket.PaymentReference,
		ShippingAddressID: basket.ShippingAddressID,
	}
	if basket.ShippingWarehouse > 0 && basket.ShippingService > 0 {
		order.ShippingRoute = &port.ShippingRoute{WarehouseID: basket.ShippingWarehouse, ServiceID: basket.ShippingService, Fee: basket.ShippingInclTax}
	}
	if basket.ShippingSnapshot != "" {
		var addr port.ShippingAddress
		if err := json.Unmarshal([]byte(basket.ShippingSnapshot), &addr); err == nil {
			order.ShippingAddress = &addr
		} else {
			log.Printf("basket %d: decode shipping address snapshot: %v", basket.ID, err)
		}
	}
	for _, l := range lines {
		order.Lines = append(order.Lines, port.PaidOrderLine{
			ProductID: l.ProductID, StockrecordID: l.StockrecordID, Quantity: l.Quantity, UnitPrice: l.PriceInclTax,
		})
	}
	id, err := s.fulfillment.StartPaidOrder(context.WithoutCancel(ctx), order)
	if err != nil {
		log.Printf("basket %d: start order workflow: %v", basket.ID, err)
		return false
	}
	basket.FulfillmentID = id
	if err := s.baskets.Update(ctx, basket); err != nil {
		log.Printf("basket %d: record fulfillment %s: %v", basket.ID, id, err)
		return false
	}
	return true
}

func (s *checkoutService) RetryFulfillment(ctx context.Context, limit int) (int, error) {
	baskets, err := s.baskets.ListAwaitingFulfillment(ctx, limit)
	if err != nil {
		return 0, err
	}
	started := 0
	for i := range baskets {
		lines, err := s.lines.ListByBasket(ctx, baskets[i].ID)
		if err != nil {
			log.Printf("basket %d: load lines for fulfillment: %v", baskets[i].ID, err)
			continue
		}
		if s.startFulfillment(ctx, &baskets[i], lines) {
			started++
		}
	}
	return started, nil
}

func roundMoney(v float64) float64 { return math.Round(v*100) / 100 }

func checkoutKey(b *model.Basket) string {
	created := strings.NewReplacer(" ", "", ":", "", "-", "", "T", "", "Z", "").Replace(b.DateCreated)
	return fmt.Sprintf("keiki-basket-%d-%s", b.ID, created)
}
