package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	merchantapp "github.com/JIeeiroSst/voucher-service/internal/application/merchant"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
)

const codeSeparator = "::"

type APIProvider struct {
	merchants  merchantapp.MerchantService
	httpClient *http.Client
}

func NewAPIProvider(merchants merchantapp.MerchantService) voucherapp.MerchantProvider {
	return &APIProvider{
		merchants:  merchants,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}
}

func (p *APIProvider) Type() shared.ProviderType { return shared.ProviderTypeAPI }

func encodeCode(merchantID shared.MerchantID, code string) string {
	return merchantID.String() + codeSeparator + code
}

func decodeCode(encoded string) (shared.MerchantID, string, error) {
	parts := strings.SplitN(encoded, codeSeparator, 2)
	if len(parts) != 2 {
		return shared.MerchantID{}, "", fmt.Errorf("code %q is not a valid api-provider code", encoded)
	}
	merchantID, err := shared.ParseMerchantID(parts[0])
	if err != nil {
		return shared.MerchantID{}, "", err
	}
	return merchantID, parts[1], nil
}

func (p *APIProvider) merchantEndpoint(ctx context.Context, merchantID shared.MerchantID) (baseURL, apiKey string, err error) {
	m, err := p.merchants.GetMerchant(ctx, merchantID)
	if err != nil {
		return "", "", err
	}
	baseURL, _ = m.Config["base_url"].(string)
	apiKey, _ = m.Config["api_key"].(string)
	if baseURL == "" {
		return "", "", fmt.Errorf("merchant %s has no api provider base_url configured", merchantID)
	}
	return baseURL, apiKey, nil
}

func (p *APIProvider) doJSON(ctx context.Context, baseURL, apiKey, path string, reqBody, respBody any) error {
	buf, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(buf))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("merchant api %s returned status %d", path, resp.StatusCode)
	}
	if respBody != nil {
		return json.NewDecoder(resp.Body).Decode(respBody)
	}
	return nil
}

type apiIssueRequest struct {
	SKU      string `json:"sku"`
	Quantity int    `json:"quantity"`
}

type apiIssueResponse struct {
	Codes []shared.VoucherCode `json:"codes"`
}

func (p *APIProvider) Issue(ctx context.Context, ref shared.ProductRef, qty int) ([]shared.VoucherCode, error) {
	baseURL, apiKey, err := p.merchantEndpoint(ctx, ref.MerchantID)
	if err != nil {
		return nil, err
	}
	var resp apiIssueResponse
	if err := p.doJSON(ctx, baseURL, apiKey, "/issue", apiIssueRequest{SKU: ref.SKU, Quantity: qty}, &resp); err != nil {
		return nil, err
	}
	codes := make([]shared.VoucherCode, len(resp.Codes))
	for i, c := range resp.Codes {
		codes[i] = shared.VoucherCode{Code: encodeCode(ref.MerchantID, c.Code), PIN: c.PIN}
	}
	return codes, nil
}

type apiValidateRequest struct {
	Code string `json:"code"`
	PIN  string `json:"pin"`
}

type apiValidateResponse struct {
	Valid   bool   `json:"valid"`
	Balance int64  `json:"balance"`
	Reason  string `json:"reason"`
}

func (p *APIProvider) Validate(ctx context.Context, code, pin string) (shared.ValidationResult, error) {
	merchantID, rawCode, err := decodeCode(code)
	if err != nil {
		return shared.ValidationResult{}, err
	}
	baseURL, apiKey, err := p.merchantEndpoint(ctx, merchantID)
	if err != nil {
		return shared.ValidationResult{}, err
	}
	var resp apiValidateResponse
	if err := p.doJSON(ctx, baseURL, apiKey, "/validate", apiValidateRequest{Code: rawCode, PIN: pin}, &resp); err != nil {
		return shared.ValidationResult{}, err
	}
	return shared.ValidationResult{Valid: resp.Valid, Balance: shared.NewMoney(resp.Balance, "VND"), Reason: resp.Reason}, nil
}

type apiRedeemRequest struct {
	Code   string `json:"code"`
	PIN    string `json:"pin"`
	Amount int64  `json:"amount"`
}

type apiRedeemResponse struct {
	Success        bool   `json:"success"`
	RedeemedAmount int64  `json:"redeemed_amount"`
	ProviderTxnRef string `json:"provider_txn_ref"`
}

func (p *APIProvider) Redeem(ctx context.Context, code, pin string, amount shared.Money) (shared.RedeemResult, error) {
	merchantID, rawCode, err := decodeCode(code)
	if err != nil {
		return shared.RedeemResult{}, err
	}
	baseURL, apiKey, err := p.merchantEndpoint(ctx, merchantID)
	if err != nil {
		return shared.RedeemResult{}, err
	}
	var resp apiRedeemResponse
	req := apiRedeemRequest{Code: rawCode, PIN: pin, Amount: amount.Amount}
	if err := p.doJSON(ctx, baseURL, apiKey, "/redeem", req, &resp); err != nil {
		return shared.RedeemResult{}, err
	}
	return shared.RedeemResult{
		Success:        resp.Success,
		RedeemedAmount: shared.NewMoney(resp.RedeemedAmount, amount.Currency),
		ProviderTxnRef: resp.ProviderTxnRef,
	}, nil
}
