package card

import (
	"context"
)

type Activities struct{}

func (a *Activities) CreateStripeCharge(_ context.Context, cart CartState) error {

	return nil
}

func (a *Activities) SendAbandonedCartEmail(_ context.Context, email string) error {

	return nil
}
