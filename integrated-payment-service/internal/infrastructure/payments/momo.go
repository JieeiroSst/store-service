package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"

	"time"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
)

type MoMoConfig struct {
	PartnerCode string
	AccessKey   string
	SecretKey   string
	Endpoint    string
	ReturnURL   string
	NotifyURL   string
}

type MoMoProcessor struct {
	config MoMoConfig
	client *http.Client
}

func NewMoMoProcessor(config MoMoConfig) *MoMoProcessor {
	return &MoMoProcessor{
		config: config,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (m *MoMoProcessor) CreatePayment(ctx context.Context, payment *entities.Payment) (*interfaces.PaymentResponse, error) {
	orderId := payment.ID
	requestId := fmt.Sprintf("%s_%d", orderId, time.Now().Unix())

	// Tạo raw signature
	rawSignature := fmt.Sprintf("accessKey=%s&amount=%.0f&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=captureWallet",
		m.config.AccessKey, payment.Amount, m.config.NotifyURL, orderId,
		payment.Description, m.config.PartnerCode, m.config.ReturnURL, requestId)

	// Tạo signature
	signature := m.generateSignature(rawSignature)

	// Tạo request body
	reqBody := map[string]interface{}{
		"partnerCode": m.config.PartnerCode,
		"accessKey":   m.config.AccessKey,
		"requestId":   requestId,
		"amount":      payment.Amount,
		"orderId":     orderId,
		"orderInfo":   payment.Description,
		"redirectUrl": m.config.ReturnURL,
		"ipnUrl":      m.config.NotifyURL,
		"requestType": "captureWallet",
		"extraData":   "",
		"signature":   signature,
	}

	// Gửi request
	resp, err := m.sendRequest(ctx, reqBody)
	if err != nil {
		return nil, err
	}

	return &interfaces.PaymentResponse{
		ID:            payment.ID,
		Status:        entities.PaymentStatusPending,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentURL:    resp["payUrl"].(string),
		TransactionID: requestId,
		ProcessorData: resp,
	}, nil
}

func (m *MoMoProcessor) ProcessPayment(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	// Implement payment processing logic
	return nil, nil
}

func (m *MoMoProcessor) RefundPayment(ctx context.Context, paymentID string, amount float64) (*interfaces.PaymentResponse, error) {
	// Implement refund logic
	return nil, nil
}

func (m *MoMoProcessor) GetPaymentStatus(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	// Implement status check logic
	return nil, nil
}

func (m *MoMoProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) error {
	// Implement webhook validation
	return nil
}

func (m *MoMoProcessor) generateSignature(rawSignature string) string {
	h := hmac.New(sha256.New, []byte(m.config.SecretKey))
	h.Write([]byte(rawSignature))
	return hex.EncodeToString(h.Sum(nil))
}

func (m *MoMoProcessor) sendRequest(ctx context.Context, body map[string]interface{}) (map[string]interface{}, error) {
	// Implement HTTP request sending
	return nil, nil
}
