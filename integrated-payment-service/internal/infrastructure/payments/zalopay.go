package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
)

type ZaloPayConfig struct {
	AppID     string
	Key1      string
	Key2      string
	Endpoint  string
	ReturnURL string
}

type ZaloPayProcessor struct {
	config ZaloPayConfig
}

func NewZaloPayProcessor(config ZaloPayConfig) *ZaloPayProcessor {
	return &ZaloPayProcessor{config: config}
}

func (z *ZaloPayProcessor) CreatePayment(ctx context.Context, payment *entities.Payment) (*interfaces.PaymentResponse, error) {
	appTransID := fmt.Sprintf("%s_%s", time.Now().Format("060102"), payment.ID)

	embedData := map[string]string{
		"redirecturl": z.config.ReturnURL,
	}
	embedDataJSON, _ := json.Marshal(embedData)

	item := []map[string]interface{}{
		{
			"itemid":       payment.ID,
			"itemname":     payment.Description,
			"itemprice":    payment.Amount,
			"itemquantity": 1,
		},
	}
	itemJSON, _ := json.Marshal(item)

	// Create order data
	orderData := map[string]interface{}{
		"appid":       z.config.AppID,
		"apptransid":  appTransID,
		"appuser":     payment.UserID,
		"apptime":     time.Now().Unix() * 1000,
		"item":        string(itemJSON),
		"embeddata":   string(embedDataJSON),
		"amount":      payment.Amount,
		"description": payment.Description,
	}

	mac := z.generateMAC(orderData)
	orderData["mac"] = mac

	return &interfaces.PaymentResponse{
		ID:            payment.ID,
		Status:        entities.PaymentStatusPending,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentURL:    "", // URL from ZaloPay response
		TransactionID: appTransID,
		ProcessorData: orderData,
	}, nil
}

func (z *ZaloPayProcessor) ProcessPayment(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (z *ZaloPayProcessor) RefundPayment(ctx context.Context, paymentID string, amount float64) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (z *ZaloPayProcessor) GetPaymentStatus(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (z *ZaloPayProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) error {
	return nil
}

func (z *ZaloPayProcessor) generateMAC(data map[string]interface{}) string {
	macData := fmt.Sprintf("%s|%s|%s|%v|%v|%s|%s",
		data["appid"], data["apptransid"], data["appuser"],
		data["amount"], data["apptime"], data["embeddata"], data["item"])

	h := hmac.New(sha256.New, []byte(z.config.Key1))
	h.Write([]byte(macData))
	return hex.EncodeToString(h.Sum(nil))
}
