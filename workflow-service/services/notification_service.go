package services

type NotificationServiceImpl struct {
}

func NewNotificationService() *NotificationServiceImpl {
	return &NotificationServiceImpl{}
}

func (s *NotificationServiceImpl) SendOrderConfirmation(orderID string, trackingNumber string) error {
	return nil
}
