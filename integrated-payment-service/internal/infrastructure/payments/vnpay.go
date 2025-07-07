package payments

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/entities"
	"github.com/JIeeiroSst/integrated-payment-service/internal/domain/interfaces"
)

type VNPayConfig struct {
	TmnCode    string
	HashSecret string
	Endpoint   string
	ReturnURL  string
}

type VNPayProcessor struct {
	config VNPayConfig
}

func NewVNPayProcessor(config VNPayConfig) *VNPayProcessor {
	return &VNPayProcessor{config: config}
}

func (v *VNPayProcessor) CreatePayment(ctx context.Context, payment *entities.Payment) (*interfaces.PaymentResponse, error) {
	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    v.config.TmnCode,
		"vnp_Amount":     fmt.Sprintf("%.0f", payment.Amount*100), // VNPay uses cents
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     payment.ID,
		"vnp_OrderInfo":  payment.Description,
		"vnp_OrderType":  "other",
		"vnp_ReturnUrl":  v.config.ReturnURL,
		"vnp_IpAddr":     "127.0.0.1",
		"vnp_CreateDate": time.Now().Format("20060102150405"),
	}

	signature := v.generateSignature(params)
	params["vnp_SecureHash"] = signature

	paymentURL := v.buildPaymentURL(params)

	return &interfaces.PaymentResponse{
		ID:            payment.ID,
		Status:        entities.PaymentStatusPending,
		Amount:        payment.Amount,
		Currency:      payment.Currency,
		PaymentURL:    paymentURL,
		TransactionID: payment.ID,
		ProcessorData: map[string]interface{}{"params": params},
	}, nil
}

func (v *VNPayProcessor) ProcessPayment(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (v *VNPayProcessor) RefundPayment(ctx context.Context, paymentID string, amount float64) (*interfaces.PaymentResponse, error) {
	// Implement refund logic
	return nil, nil
}

func (v *VNPayProcessor) GetPaymentStatus(ctx context.Context, paymentID string) (*interfaces.PaymentResponse, error) {
	return nil, nil
}

func (v *VNPayProcessor) ValidateWebhook(ctx context.Context, payload []byte, signature string) error {
	return nil
}

func (v *VNPayProcessor) generateSignature(params map[string]string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%s", key, params[key]))
	}
	query := strings.Join(parts, "&")

	h := hmac.New(sha512.New, []byte(v.config.HashSecret))
	h.Write([]byte(query))
	return hex.EncodeToString(h.Sum(nil))
}

func (v *VNPayProcessor) buildPaymentURL(params map[string]string) string {
	u, _ := url.Parse(v.config.Endpoint)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}
