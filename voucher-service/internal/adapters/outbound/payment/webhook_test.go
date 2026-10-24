package payment

import (
	"encoding/json"
	"fmt"
	"net/url"
	"testing"
)

func signedMomoBody(t *testing.T, g *MomoGateway, amount string, resultCode string, tamper func(map[string]any)) []byte {
	t.Helper()
	cb := map[string]any{
		"partnerCode": g.partnerCode, "orderId": "pay-1", "requestId": "pay-1", "amount": json.Number(amount),
		"orderInfo": "Voucher order pay-1", "orderType": "momo_wallet", "transId": json.Number("4088878653"),
		"resultCode": json.Number(resultCode), "message": "Successful.", "payType": "qr",
		"responseTime": json.Number("1721720663942"), "extraData": "",
	}
	raw := fmt.Sprintf("accessKey=%s&amount=%s&extraData=%s&message=%s&orderId=%s&orderInfo=%s&orderType=%s&partnerCode=%s&payType=%s&requestId=%s&responseTime=%s&resultCode=%s&transId=%s",
		g.accessKey, cb["amount"], cb["extraData"], cb["message"], cb["orderId"], cb["orderInfo"], cb["orderType"], cb["partnerCode"], cb["payType"], cb["requestId"], cb["responseTime"], cb["resultCode"], cb["transId"])
	cb["signature"] = g.sign(raw)
	if tamper != nil {
		tamper(cb)
	}
	b, _ := json.Marshal(cb)
	return b
}

func TestMomoWebhookSignature(t *testing.T) {
	g := &MomoGateway{partnerCode: "MOMO", accessKey: "ak", secretKey: "sk"}

	res, err := g.VerifyWebhookSignature(signedMomoBody(t, g, "50000", "0", nil), "")
	if err != nil || !res.Success || res.PaymentID != "pay-1" || res.Amount != 50000 {
		t.Fatalf("valid callback: res=%+v err=%v", res, err)
	}

	forged := []byte(`{"orderId":"pay-1","resultCode":0}`)
	if _, err := g.VerifyWebhookSignature(forged, ""); err == nil {
		t.Fatal("unsigned callback accepted")
	}

	tampered := signedMomoBody(t, g, "50000", "0", func(m map[string]any) { m["amount"] = json.Number("1") })
	if _, err := g.VerifyWebhookSignature(tampered, ""); err == nil {
		t.Fatal("callback with tampered amount accepted")
	}

	unconfigured := &MomoGateway{partnerCode: "MOMO", accessKey: "ak"}
	if _, err := unconfigured.VerifyWebhookSignature(signedMomoBody(t, unconfigured, "50000", "0", nil), ""); err == nil {
		t.Fatal("callback accepted without a configured secret")
	}
}

func TestVNPayWebhookSignature(t *testing.T) {
	g := &VNPayGateway{tmnCode: "TMN", hashSecret: "secret"}
	params := map[string]string{"vnp_TmnCode": "TMN", "vnp_TxnRef": "pay-1", "vnp_Amount": "5000000", "vnp_ResponseCode": "00", "vnp_TransactionStatus": "00"}
	query, hash := g.sign(params)

	res, err := g.VerifyWebhookSignature([]byte(query+"&vnp_SecureHash="+hash), "")
	if err != nil || !res.Success || res.Amount != 50000 || res.PaymentID != "pay-1" {
		t.Fatalf("valid callback: res=%+v err=%v", res, err)
	}

	empty := &VNPayGateway{tmnCode: "TMN"}
	q2, h2 := empty.sign(params)
	if _, err := empty.VerifyWebhookSignature([]byte(q2+"&vnp_SecureHash="+h2), ""); err == nil {
		t.Fatal("callback accepted with an empty hash secret")
	}

	v, _ := url.ParseQuery(query)
	v.Set("vnp_Amount", "100")
	v.Set("vnp_SecureHash", hash)
	if _, err := g.VerifyWebhookSignature([]byte(v.Encode()), ""); err == nil {
		t.Fatal("callback with tampered amount accepted")
	}
}
