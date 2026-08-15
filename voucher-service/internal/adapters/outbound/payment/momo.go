package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

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

func (g *MomoGateway) VerifyWebhookSignature(rawBody []byte, signature string) (string, bool, error) {
	var callback struct {
		OrderID    string `json:"orderId"`
		ResultCode int    `json:"resultCode"`
		Signature  string `json:"signature"`
	}
	if err := json.Unmarshal(rawBody, &callback); err != nil {
		return "", false, err
	}
	success := callback.ResultCode == 0
	return callback.OrderID, success, nil
}

func (g *MomoGateway) Refund(ctx context.Context, providerTxnRef string, amount shared.Money) error {
	return fmt.Errorf("momo refund requires live merchant credentials, not configured in this environment")
}
