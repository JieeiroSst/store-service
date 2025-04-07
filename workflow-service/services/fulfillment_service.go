package services

import (
	"fmt"
)

type FulfillmentServiceImpl struct {
}

func NewFulfillmentService() *FulfillmentServiceImpl {
	return &FulfillmentServiceImpl{}
}

func (f *FulfillmentServiceImpl) FulfillOrder(orderID string) (string, error) {
	trackingNumber := fmt.Sprintf("trk-%s", orderID)
	return trackingNumber, nil
}
