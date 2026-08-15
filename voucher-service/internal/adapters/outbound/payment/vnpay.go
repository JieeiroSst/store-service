package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	paymentapp "github.com/JIeeiroSst/voucher-service/internal/application/payment"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

type VNPayGateway struct {
	tmnCode    string
	hashSecret string
	payURL     string
}

func NewVNPayGateway() paymentapp.PaymentGateway {
	return &VNPayGateway{
		tmnCode:    os.Getenv("VNPAY_TMN_CODE"),
		hashSecret: os.Getenv("VNPAY_HASH_SECRET"),
		payURL:     envOrDefault("VNPAY_URL", "https://sandbox.vnpayment.vn/paymentv2/vpcpay.html"),
	}
}

func (g *VNPayGateway) Provider() string { return "vnpay" }

func (g *VNPayGateway) InitPayment(ctx context.Context, refID string, amount shared.Money, returnURL string) (string, error) {
	params := map[string]string{
		"vnp_Version":    "2.1.0",
		"vnp_Command":    "pay",
		"vnp_TmnCode":    g.tmnCode,
		"vnp_Amount":     strconv.FormatInt(amount.Amount*100, 10), // VNPay expects amount * 100
		"vnp_CurrCode":   "VND",
		"vnp_TxnRef":     refID,
		"vnp_OrderInfo":  "Voucher order " + refID,
		"vnp_ReturnUrl":  returnURL,
		"vnp_IpAddr":     "127.0.0.1",
		"vnp_CreateDate": time.Now().Format("20060102150405"),
	}
	query, secureHash := g.sign(params)
	return g.payURL + "?" + query + "&vnp_SecureHash=" + secureHash, nil
}

func (g *VNPayGateway) sign(params map[string]string) (query, secureHash string) {
	keys := make([]string, 0, len(params))
	for k := range params {
		if params[k] != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var signBuilder, queryBuilder strings.Builder
	for i, k := range keys {
		if i > 0 {
			signBuilder.WriteByte('&')
			queryBuilder.WriteByte('&')
		}
		signBuilder.WriteString(k)
		signBuilder.WriteByte('=')
		signBuilder.WriteString(url.QueryEscape(params[k]))
		queryBuilder.WriteString(k)
		queryBuilder.WriteByte('=')
		queryBuilder.WriteString(url.QueryEscape(params[k]))
	}

	mac := hmac.New(sha512.New, []byte(g.hashSecret))
	mac.Write([]byte(signBuilder.String()))
	secureHash = hex.EncodeToString(mac.Sum(nil))
	return queryBuilder.String(), secureHash
}

func (g *VNPayGateway) VerifyWebhookSignature(rawBody []byte, signature string) (string, bool, error) {
	values, err := url.ParseQuery(string(rawBody))
	if err != nil {
		return "", false, err
	}
	params := make(map[string]string, len(values))
	for k := range values {
		if k == "vnp_SecureHash" || k == "vnp_SecureHashType" {
			continue
		}
		params[k] = values.Get(k)
	}
	_, expectedHash := g.sign(params)
	if !hmac.Equal([]byte(expectedHash), []byte(values.Get("vnp_SecureHash"))) {
		return "", false, fmt.Errorf("vnpay: signature mismatch")
	}
	success := values.Get("vnp_ResponseCode") == "00"
	return values.Get("vnp_TxnRef"), success, nil
}

func (g *VNPayGateway) Refund(ctx context.Context, providerTxnRef string, amount shared.Money) error {
	return fmt.Errorf("vnpay refund requires live merchant credentials, not configured in this environment")
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
