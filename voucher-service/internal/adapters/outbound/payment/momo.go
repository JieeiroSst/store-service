package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type MomoGateway struct {
	partnerCode string
	accessKey   string
	secretKey   string
	endpoint    string
}

func NewMomoGateway() paymentapp.PaymentGateway {
	return &MomoGateway{
		partnerCode: os.Getenv("MOMO_PARTNER_CODE"),
		accessKey:   os.Getenv("MOMO_ACCESS_KEY"),
		secretKey:   os.Getenv("MOMO_SECRET_KEY"),
		endpoint:    envOrDefault("MOMO_ENDPOINT", "https://test-payment.momo.vn/v2/gateway/api/create"),
	}
}

func (g *MomoGateway) Provider() string { return "momo" }

type momoCreateRequest struct {
	PartnerCode string `json:"partnerCode"`
	AccessKey   string `json:"accessKey"`
	RequestID   string `json:"requestId"`
	Amount      string `json:"amount"`
	OrderID     string `json:"orderId"`
	OrderInfo   string `json:"orderInfo"`
	ReturnURL   string `json:"redirectUrl"`
	NotifyURL   string `json:"ipnUrl"`
	RequestType string `json:"requestType"`
	Signature   string `json:"signature"`
}

func (g *MomoGateway) sign(raw string) string {
	mac := hmac.New(sha256.New, []byte(g.secretKey))
	mac.Write([]byte(raw))
	return hex.EncodeToString(mac.Sum(nil))
}

func (g *MomoGateway) InitPayment(ctx context.Context, refID string, amount shared.Money, returnURL string) (string, error) {
	amountStr := fmt.Sprintf("%d", amount.Amount)
	raw := fmt.Sprintf("accessKey=%s&amount=%s&extraData=&ipnUrl=%s&orderId=%s&orderInfo=%s&partnerCode=%s&redirectUrl=%s&requestId=%s&requestType=captureWallet",
		g.accessKey, amountStr, "", refID, "Voucher order "+refID, g.partnerCode, returnURL, refID)
	signature := g.sign(raw)

	req := momoCreateRequest{
		PartnerCode: g.partnerCode,
		AccessKey:   g.accessKey,
		RequestID:   refID,
		Amount:      amountStr,
		OrderID:     refID,
		OrderInfo:   "Voucher order " + refID,
		ReturnURL:   returnURL,
		RequestType: "captureWallet",
		Signature:   signature,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", err
	}
	_ = body
	return g.endpoint + "?orderId=" + refID, nil
}

type momoIPN struct {
	PartnerCode  string      `json:"partnerCode"`
	OrderID      string      `json:"orderId"`
	RequestID    string      `json:"requestId"`
	Amount       json.Number `json:"amount"`
	OrderInfo    string      `json:"orderInfo"`
	OrderType    string      `json:"orderType"`
	TransID      json.Number `json:"transId"`
	ResultCode   json.Number `json:"resultCode"`
	Message      string      `json:"message"`
	PayType      string      `json:"payType"`
	ResponseTime json.Number `json:"responseTime"`
	ExtraData    string      `json:"extraData"`
	Signature    string      `json:"signature"`
}

func (g *MomoGateway) VerifyWebhookSignature(rawBody []byte, _ string) (paymentapp.WebhookResult, error) {
	if g.secretKey == "" || g.accessKey == "" {
		return paymentapp.WebhookResult{}, fmt.Errorf("momo: merchant credentials are not configured")
	}
	var cb momoIPN
	if err := json.Unmarshal(rawBody, &cb); err != nil {
		return paymentapp.WebhookResult{}, fmt.Errorf("momo: malformed callback: %w", err)
	}
	raw := fmt.Sprintf("accessKey=%s&amount=%s&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%s&resultCode=%s&transId=%s",
		g.accessKey, cb.Amount, cb.ExtraData, cb.Message, cb.OrderID, cb.OrderInfo, cb.OrderType, cb.PartnerCode, cb.PayType, cb.RequestID, cb.ResponseTime, cb.ResultCode, cb.TransID)
	if !hmac.Equal([]byte(g.sign(raw)), []byte(strings.ToLower(cb.Signature))) {
		return paymentapp.WebhookResult{}, fmt.Errorf("momo: signature mismatch")
	}
	if cb.PartnerCode != g.partnerCode {
		return paymentapp.WebhookResult{}, fmt.Errorf("momo: unexpected partner code")
	}
	amount, err := cb.Amount.Int64()
	if err != nil {
		return paymentapp.WebhookResult{}, fmt.Errorf("momo: invalid amount: %w", err)
	}
	return paymentapp.WebhookResult{PaymentID: cb.OrderID, Success: cb.ResultCode.String() == "0", Amount: amount}, nil
}

func (g *MomoGateway) Refund(ctx context.Context, providerTxnRef string, amount shared.Money) error {
	return fmt.Errorf("momo refund requires live merchant credentials, not configured in this environment")
}
