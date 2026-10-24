package application

import (
	"testing"

	"github.com/JIeeiroSst/payment-service/internal/domain/model"
)

func TestWebhookMayMove(t *testing.T) {
	cases := []struct {
		from, to model.PaymentStatus
		want     bool
	}{
		{model.PaymentStatusPending, model.PaymentStatusCaptured, true},
		{model.PaymentStatusFailed, model.PaymentStatusCaptured, true},
		{model.PaymentStatusCaptured, model.PaymentStatusFailed, false},
		{model.PaymentStatusCaptured, model.PaymentStatusPending, false},
		{model.PaymentStatusCaptured, model.PaymentStatusRefunded, true},
		{model.PaymentStatusRefunded, model.PaymentStatusCaptured, false},
		{model.PaymentStatusCaptured, model.PaymentStatusCaptured, false},
	}
	for _, c := range cases {
		if got := webhookMayMove(c.from, c.to); got != c.want {
			t.Errorf("webhookMayMove(%s, %s) = %v, want %v", c.from, c.to, got, c.want)
		}
	}
}
