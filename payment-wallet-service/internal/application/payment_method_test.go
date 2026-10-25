package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/payment-wallet-service/internal/domain/port"
)

type fakePaymentMethodRepo struct {
	port.PaymentMethodRepository
	created []*model.PaymentMethod
}

func (r *fakePaymentMethodRepo) ListByUser(context.Context, string) ([]model.PaymentMethod, error) {
	return nil, nil
}

func (r *fakePaymentMethodRepo) Create(_ context.Context, m *model.PaymentMethod) error {
	r.created = append(r.created, m)
	return nil
}

func TestAddPaymentMethodCard(t *testing.T) {
	t.Run("stores only the masked card number", func(t *testing.T) {
		repo := &fakePaymentMethodRepo{}
		m, err := NewPaymentMethodService(repo).AddPaymentMethod(context.Background(), "u1", model.Card, "visa", "4111 1111 1111 1111")
		if err != nil {
			t.Fatalf("add: %v", err)
		}
		if m.AccountNumber != "**** **** **** 1111" || repo.created[0].AccountNumber != "**** **** **** 1111" {
			t.Fatalf("account number = %q, want masked", repo.created[0].AccountNumber)
		}
	})

	t.Run("rejects an invalid card number", func(t *testing.T) {
		repo := &fakePaymentMethodRepo{}
		for _, n := range []string{"4111111111111112", "1234", "41111111111111x1"} {
			_, err := NewPaymentMethodService(repo).AddPaymentMethod(context.Background(), "u1", model.Card, "visa", n)
			if !errors.Is(err, port.ErrInvalidCardNumber) {
				t.Fatalf("%q: err = %v, want ErrInvalidCardNumber", n, err)
			}
		}
		if len(repo.created) != 0 {
			t.Fatal("invalid card was stored")
		}
	})

	t.Run("keeps bank account numbers", func(t *testing.T) {
		repo := &fakePaymentMethodRepo{}
		m, err := NewPaymentMethodService(repo).AddPaymentMethod(context.Background(), "u1", model.BankAccount, "vcb", "0011001234567")
		if err != nil || m.AccountNumber != "0011001234567" {
			t.Fatalf("bank account = %q, %v", m.AccountNumber, err)
		}
	})
}
