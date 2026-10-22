package application

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/internal/domain/model"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
)

type checkoutService struct {
	baskets     port.BasketRepository
	lines       port.BasketLineRepository
	catalogue   port.Catalogue
	wallet      port.Wallet
	fulfillment port.Fulfillment
	now         func() time.Time
}

func NewCheckoutService(baskets port.BasketRepository, lines port.BasketLineRepository, catalogue port.Catalogue, wallet port.Wallet, fulfillment port.Fulfillment) port.CheckoutUsecase {
	return &checkoutService{baskets: baskets, lines: lines, catalogue: catalogue, wallet: wallet, fulfillment: fulfillment, now: time.Now}
}

func (s *checkoutService) Checkout(ctx context.Context, cmd port.CheckoutCommand) (*model.Basket, error) {
	if strings.TrimSpace(cmd.WalletID) == "" || strings.TrimSpace(cmd.Pin) == "" {
		return nil, fmt.Errorf("%w: wallet_id and pin are required", common.ErrInvalidRequest)
	}
	basket, err := s.baskets.GetByID(ctx, cmd.BasketID)
	if err != nil {
		return nil, err
	}
	if basket.UserID != cmd.UserID {
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

	currency := ""
	total := 0.0
	for i := range lines {
		record, err := s.catalogue.StockRecord(ctx, lines[i].StockrecordID)
		if err != nil {
			return nil, err
		}
		if currency == "" {
			currency = record.PriceCurrency
		} else if currency != record.PriceCurrency {
			return nil, fmt.Errorf("%w: basket mixes %s and %s", common.ErrInvalidRequest, currency, record.PriceCurrency)
		}
		lines[i].PriceCurrency = record.PriceCurrency
		lines[i].PriceExclTax = record.Price
		lines[i].PriceInclTax = record.Price
		lines[i].ProductID = record.ProductID
		total += record.Price * float64(lines[i].Quantity)
	}
	total = math.Round(total*100) / 100

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

	txID, err := s.wallet.Pay(ctx, port.WalletPayment{
		UserID:         cmd.UserID,
		WalletID:       cmd.WalletID,
		Amount:         total,
		Pin:            cmd.Pin,
		Description:    fmt.Sprintf("Keiki Shop basket #%d", basket.ID),
		IdempotencyKey: checkoutKey(basket),
		Metadata:       map[string]string{"basket_id": fmt.Sprint(basket.ID)},
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
	basket.Currency = currency
	basket.TotalInclTax = total
	basket.PaymentReference = txID
	if cmd.ShippingAddressID > 0 {
		basket.ShippingAddressID = cmd.ShippingAddressID
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
		PaymentReference:  basket.PaymentReference,
		ShippingAddressID: basket.ShippingAddressID,
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

func checkoutKey(b *model.Basket) string {
	created := strings.NewReplacer(" ", "", ":", "", "-", "", "T", "", "Z", "").Replace(b.DateCreated)
	return fmt.Sprintf("keiki-basket-%d-%s", b.ID, created)
}
