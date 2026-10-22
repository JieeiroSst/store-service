package activity

import (
	"context"
	"fmt"
	"strconv"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	"github.com/JIeeiroSst/order-processing-service/internal/domain/repository"
	"github.com/JIeeiroSst/order-processing-service/internal/proxy"
	"go.temporal.io/sdk/temporal"
)

type PaidOrderActivities struct {
	Store    repository.PaidOrderStore
	Notifier proxy.NotificationProxy
}

func (a *PaidOrderActivities) RecordPaidOrder(ctx context.Context, in entity.PaidOrderInput) (*entity.PaidOrderRecord, error) {
	if err := in.Validate(); err != nil {
		return nil, temporal.NewNonRetryableApplicationError(err.Error(), "InvalidPaidOrder", err)
	}
	return a.Store.CreatePaidOrder(ctx, in)
}

func (a *PaidOrderActivities) NotifyOrderPlaced(ctx context.Context, in entity.PaidOrderInput, number string) (bool, error) {
	resp, err := a.Notifier.Send(ctx, entity.NotificationRequest{
		CustomerID: strconv.Itoa(in.UserID),
		OrderID:    number,
		Type:       "push",
		Message:    fmt.Sprintf("Đơn %s đã được thanh toán %.0f %s và đang được xử lý.", number, in.Total, in.Currency),
	})
	if err != nil {
		return false, err
	}
	return resp.Sent, nil
}

func (a *PaidOrderActivities) SetPaidOrderStatus(ctx context.Context, number, status string) error {
	return a.Store.SetStatus(ctx, number, status)
}
