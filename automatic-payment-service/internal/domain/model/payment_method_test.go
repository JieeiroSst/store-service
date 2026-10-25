package model

import "testing"

func TestPaymentMethodCardFields(t *testing.T) {
	for _, s := range []string{"4111111111111111", "4111 1111 1111 1111", "5555-5555-5555-4444"} {
		if !LooksLikePAN(s) {
			t.Fatalf("LooksLikePAN(%q) = false", s)
		}
	}
	for _, s := range []string{"", "pm_1NxYz", "tok_visa", "4111111111111112", "1234"} {
		if LooksLikePAN(s) {
			t.Fatalf("LooksLikePAN(%q) = true", s)
		}
	}
	if !ValidLastFour("1111") || !ValidLastFour("") || ValidLastFour("4111111111111111") || ValidLastFour("12a4") {
		t.Fatal("ValidLastFour")
	}
	if !ValidExpiryDate("09/2028") || !ValidExpiryDate("") || ValidExpiryDate("13/2028") || ValidExpiryDate("9/28") {
		t.Fatal("ValidExpiryDate")
	}
}
