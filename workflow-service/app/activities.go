package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

type OrderDetails struct {
	ID          string
	CustomerID  string
	Items       []OrderItem
	TotalAmount float64
	Status      string
	CreatedAt   time.Time
}

type OrderItem struct {
	ProductID   string
	Quantity    int
	Price       float64
	Description string
}

type PaymentDetails struct {
	ID            string
	OrderID       string
	Amount        float64
	TransactionID string
	Status        string
	CreatedAt     time.Time
}

func ValidateOrderActivity(ctx context.Context, orderID string) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting validation of order", "orderID", orderID)

	activity.RecordHeartbeat(ctx, "retrieving order")

	order, err := getOrderFromDB(ctx, orderID)
	if err != nil {
		logger.Error("Error retrieving order", "orderID", orderID, "error", err)
		return false, temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve order: %v", err),
			"ORDER_RETRIEVAL_ERROR",
			err,
		)
	}

	if order == nil {
		logger.Error("Order not found", "orderID", orderID)
		return false, temporal.NewNonRetryableApplicationError(
			"order not found",
			"ORDER_NOT_FOUND",
			nil,
		)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Activity cancelled", "orderID", orderID)
		return false, ctx.Err() 
	}

	activity.RecordHeartbeat(ctx, "validating customer")
	customerValid, err := validateCustomer(ctx, order.CustomerID)
	if err != nil {
		logger.Error("Error validating customer", "orderID", orderID, "customerID", order.CustomerID, "error", err)
		return false, temporal.NewApplicationError(
			fmt.Sprintf("customer validation failed: %v", err),
			"CUSTOMER_VALIDATION_ERROR",
			err,
		)
	}

	if !customerValid {
		logger.Info("Customer is not valid", "orderID", orderID, "customerID", order.CustomerID)
		return false, temporal.NewNonRetryableApplicationError(
			"customer validation failed: invalid customer",
			"INVALID_CUSTOMER",
			nil,
		)
	}

	activity.RecordHeartbeat(ctx, "checking inventory")
	for _, item := range order.Items {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("Activity timed out while checking inventory", "orderID", orderID)
			return false, ctx.Err()
		}

		available, err := checkInventory(ctx, item.ProductID, item.Quantity)
		if err != nil {
			logger.Error("Error checking inventory", "orderID", orderID, "productID", item.ProductID, "error", err)
			return false, temporal.NewApplicationError(
				fmt.Sprintf("inventory check failed: %v", err),
				"INVENTORY_CHECK_ERROR",
				err,
			)
		}

		if !available {
			logger.Info("Insufficient inventory", "orderID", orderID, "productID", item.ProductID)
			return false, temporal.NewNonRetryableApplicationError(
				"insufficient inventory",
				"INSUFFICIENT_INVENTORY",
				nil,
			)
		}
	}

	logger.Info("Order successfully validated", "orderID", orderID)
	return true, nil
}

func ProcessPaymentActivity(ctx context.Context, orderID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing payment", "orderID", orderID)

	activity.RecordHeartbeat(ctx, "retrieving order")

	order, err := getOrderFromDB(ctx, orderID)
	if err != nil {
		logger.Error("Error retrieving order", "orderID", orderID, "error", err)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve order: %v", err),
			"ORDER_RETRIEVAL_ERROR",
			err,
		)
	}

	if shouldContinueAsNew(order) {
		chunkID := getNextChunkID(order)
		logger.Info("Continuing as new for large order processing", "orderID", orderID, "chunkID", chunkID)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve order: %v", err),
			"ORDER_RETRIEVAL_ERROR",
			true, 
		)
	}

	activity.RecordHeartbeat(ctx, "retrieving payment method")
	paymentMethod, err := getCustomerPaymentMethod(ctx, order.CustomerID)
	if err != nil {
		logger.Error("Error retrieving payment method", "orderID", orderID, "customerID", order.CustomerID, "error", err)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve payment method: %v", err),
			"PAYMENT_METHOD_ERROR",
			err,
		)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Activity cancelled", "orderID", orderID)
		return "", ctx.Err()
	}

	paymentID := uuid.New().String()
	activity.RecordHeartbeat(ctx, "processing payment with gateway")

	transactionID, err := processPaymentWithGateway(ctx, paymentMethod, order.TotalAmount)
	if err != nil {
		logger.Error("Payment gateway error", "orderID", orderID, "error", err)
		return "", temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("payment gateway error: %v", err),
			"PAYMENT_GATEWAY_ERROR",
			err,
		)
	}

	payment := PaymentDetails{
		ID:            paymentID,
		OrderID:       orderID,
		Amount:        order.TotalAmount,
		TransactionID: transactionID,
		Status:        "completed",
		CreatedAt:     time.Now(),
	}

	activity.RecordHeartbeat(ctx, "saving payment")
	if err := savePaymentToDB(ctx, payment); err != nil {
		logger.Error("Error saving payment record", "orderID", orderID, "paymentID", paymentID, "error", err)
		_ = voidTransaction(ctx, transactionID)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("failed to save payment record: %v", err),
			"PAYMENT_RECORD_ERROR",
			err,
		)
	}

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("Activity timed out after payment processing", "orderID", orderID)
		return "", ctx.Err() 
	}

	if err := updateOrderStatus(ctx, orderID, "paid"); err != nil {
		logger.Error("Error updating order status", "orderID", orderID, "error", err)
		logger.Error("CRITICAL: Payment processed but order status not updated. Manual intervention required", "orderID", orderID)
	}

	logger.Info("Payment processed successfully", "orderID", orderID, "paymentID", paymentID)
	return paymentID, nil
}

func FulfillOrderActivity(ctx context.Context, orderID, paymentID string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Fulfilling order", "orderID", orderID, "paymentID", paymentID)

	activity.RecordHeartbeat(ctx, "retrieving order")
	order, err := getOrderFromDB(ctx, orderID)
	if err != nil {
		logger.Error("Error retrieving order", "orderID", orderID, "error", err)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve order: %v", err),
			"ORDER_RETRIEVAL_ERROR",
			err,
		)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Activity cancelled", "orderID", orderID)
		return "", ctx.Err() 
	}

	activity.RecordHeartbeat(ctx, "verifying payment")
	paymentVerified, err := verifyPayment(ctx, paymentID, order.TotalAmount)
	if err != nil {
		logger.Error("Error verifying payment", "orderID", orderID, "paymentID", paymentID, "error", err)
		return "", temporal.NewApplicationError(
			fmt.Sprintf("payment verification failed: %v", err),
			"PAYMENT_VERIFICATION_ERROR",
			err,
		)
	}

	if !paymentVerified {
		logger.Error("Payment failed verification", "orderID", orderID, "paymentID", paymentID)
		return "", temporal.NewNonRetryableApplicationError(
			"payment verification failed",
			"INVALID_PAYMENT",
			nil,
		)
	}

	activity.RecordHeartbeat(ctx, "reserving inventory")
	for _, item := range order.Items {
		if ctx.Err() == context.DeadlineExceeded {
			logger.Error("Activity timed out while reserving inventory", "orderID", orderID)
			return "", ctx.Err() 
		}

		if err := reserveInventory(ctx, item.ProductID, item.Quantity); err != nil {
			logger.Error("Failed to reserve inventory", "orderID", orderID, "productID", item.ProductID, "error", err)
			return "", temporal.NewApplicationError(
				fmt.Sprintf("inventory reservation failed: %v", err),
				"INVENTORY_RESERVATION_ERROR",
				err,
			)
		}
	}

	activity.RecordHeartbeat(ctx, "creating shipment")
	trackingNumber, err := createShipment(ctx, orderID, order)
	if err != nil {
		logger.Error("Failed to create shipment", "orderID", orderID, "error", err)
		for _, item := range order.Items {
			_ = releaseInventory(ctx, item.ProductID, item.Quantity)
		}
		return "", temporal.NewApplicationError(
			fmt.Sprintf("shipment creation failed: %v", err),
			"SHIPMENT_CREATION_ERROR",
			err,
		)
	}

	if err := updateOrderStatus(ctx, orderID, "fulfilled"); err != nil {
		logger.Error("Failed to update order status", "orderID", orderID, "error", err)
		logger.Warn("Order fulfilled but status not updated. Manual intervention required", "orderID", orderID)
	}

	logger.Info("Order fulfilled successfully", "orderID", orderID, "trackingNumber", trackingNumber)
	return trackingNumber, nil
}

func RefundPaymentActivity(ctx context.Context, paymentID string) (bool, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing refund", "paymentID", paymentID)

	activity.RecordHeartbeat(ctx, "retrieving payment")
	payment, err := getPaymentFromDB(ctx, paymentID)
	if err != nil {
		logger.Error("Error retrieving payment", "paymentID", paymentID, "error", err)
		return false, temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve payment: %v", err),
			"PAYMENT_RETRIEVAL_ERROR",
			err,
		)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Activity cancelled", "paymentID", paymentID)
		return false, ctx.Err() 
	}

	if payment.Status != "completed" {
		logger.Error("Payment cannot be refunded", "paymentID", paymentID, "status", payment.Status)
		return false, temporal.NewNonRetryableApplicationError(
			fmt.Sprintf("payment cannot be refunded (status: %s)", payment.Status),
			"INVALID_PAYMENT_STATUS",
			nil,
		)
	}

	activity.RecordHeartbeat(ctx, "processing refund with gateway")
	refundID, err := processRefundWithGateway(ctx, payment.TransactionID, payment.Amount)
	if err != nil {
		logger.Error("Payment gateway refund error", "paymentID", paymentID, "error", err)
		return false, temporal.NewApplicationError(
			fmt.Sprintf("refund gateway error: %v", err),
			"REFUND_GATEWAY_ERROR",
			err,
		)
	}

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("Activity timed out after refund processing", "paymentID", paymentID)
		logger.Error("CRITICAL: Refund was processed but status not updated. Manual intervention required", "paymentID", paymentID)
		return false, ctx.Err()
	}

	if err := updatePaymentStatus(ctx, paymentID, "refunded", refundID); err != nil {
		logger.Error("Error updating payment status", "paymentID", paymentID, "error", err)
		logger.Error("CRITICAL: Refund processed but payment record not updated. Manual intervention required", "paymentID", paymentID)
	}

	if err := updateOrderStatus(ctx, payment.OrderID, "refunded"); err != nil {
		logger.Error("Error updating order status", "paymentID", paymentID, "orderID", payment.OrderID, "error", err)
		logger.Warn("Payment refunded but order status not updated. Manual intervention required", "paymentID", paymentID)
	}

	logger.Info("Refund processed successfully", "paymentID", paymentID, "refundID", refundID)
	return true, nil
}

func SendConfirmationActivity(ctx context.Context, orderID, trackingNumber string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending confirmation", "orderID", orderID, "trackingNumber", trackingNumber)

	activity.RecordHeartbeat(ctx, "retrieving order")
	order, err := getOrderFromDB(ctx, orderID)
	if err != nil {
		logger.Error("Error retrieving order", "orderID", orderID, "error", err)
		return temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve order: %v", err),
			"ORDER_RETRIEVAL_ERROR",
			err,
		)
	}

	if ctx.Err() == context.Canceled {
		logger.Info("Activity cancelled", "orderID", orderID)
		return ctx.Err() 
	}

	activity.RecordHeartbeat(ctx, "retrieving customer details")
	customer, err := getCustomerDetails(ctx, order.CustomerID)
	if err != nil {
		logger.Error("Error retrieving customer details", "orderID", orderID, "customerID", order.CustomerID, "error", err)
		return temporal.NewApplicationError(
			fmt.Sprintf("failed to retrieve customer details: %v", err),
			"CUSTOMER_DETAILS_ERROR",
			err,
		)
	}

	confirmationDetails := map[string]interface{}{
		"orderID":        orderID,
		"trackingNumber": trackingNumber,
		"items":          order.Items,
		"totalAmount":    order.TotalAmount,
		"shippingInfo":   customer.ShippingAddress,
	}

	activity.RecordHeartbeat(ctx, "sending email")
	if err := sendEmail(ctx, customer.Email, "Order Confirmation", confirmationDetails); err != nil {
		logger.Error("Failed to send email confirmation", "orderID", orderID, "email", customer.Email, "error", err)
		return temporal.NewApplicationError(
			fmt.Sprintf("email notification failed: %v", err),
			"EMAIL_NOTIFICATION_ERROR",
			err,
		)
	}

	if ctx.Err() == context.DeadlineExceeded {
		logger.Error("Activity timed out after sending email", "orderID", orderID)
		return ctx.Err() 
	}

	if customer.Phone != "" {
		activity.RecordHeartbeat(ctx, "sending SMS")
		if err := sendSMS(ctx, customer.Phone, fmt.Sprintf("Your order %s has shipped! Track: %s", orderID, trackingNumber)); err != nil {
			logger.Warn("Failed to send SMS confirmation", "orderID", orderID, "phone", customer.Phone, "error", err)
		}
	}

	if err := updateOrderNotificationStatus(ctx, orderID, "sent"); err != nil {
		logger.Warn("Failed to update notification status", "orderID", orderID, "error", err)
	}

	logger.Info("Confirmation sent successfully", "orderID", orderID)
	return nil
}


func shouldContinueAsNew(order *OrderDetails) bool {
	return len(order.Items) > 10
}

func getNextChunkID(order *OrderDetails) string {
	return uuid.New().String()[:8]
}

func getOrderFromDB(ctx context.Context, orderID string) (*OrderDetails, error) {
	return &OrderDetails{
		ID:          orderID,
		CustomerID:  "cust-123",
		TotalAmount: 99.99,
		Status:      "pending",
		Items: []OrderItem{
			{ProductID: "prod-1", Quantity: 2, Price: 29.99, Description: "Widget"},
			{ProductID: "prod-2", Quantity: 1, Price: 40.01, Description: "Gadget"},
		},
		CreatedAt: time.Now().Add(-1 * time.Hour),
	}, nil
}

func validateCustomer(ctx context.Context, customerID string) (bool, error) {
	return true, nil
}

func checkInventory(ctx context.Context, productID string, quantity int) (bool, error) {
	return true, nil
}

func getCustomerPaymentMethod(ctx context.Context, customerID string) (string, error) {
	return "pm_token_12345", nil
}

func processPaymentWithGateway(ctx context.Context, paymentMethod string, amount float64) (string, error) {
	return fmt.Sprintf("txn-%s", uuid.New().String()), nil
}

func savePaymentToDB(ctx context.Context, payment PaymentDetails) error {
	return nil
}

func updateOrderStatus(ctx context.Context, orderID, status string) error {
	return nil
}

func voidTransaction(ctx context.Context, transactionID string) error {
	return nil
}

func verifyPayment(ctx context.Context, paymentID string, expectedAmount float64) (bool, error) {
	return true, nil
}

func reserveInventory(ctx context.Context, productID string, quantity int) error {
	return nil
}

func releaseInventory(ctx context.Context, productID string, quantity int) error {
	return nil
}

func createShipment(ctx context.Context, orderID string, order *OrderDetails) (string, error) {
	return fmt.Sprintf("trk-%s-%s", orderID, uuid.New().String()[:8]), nil
}

func getPaymentFromDB(ctx context.Context, paymentID string) (*PaymentDetails, error) {
	return &PaymentDetails{
		ID:            paymentID,
		OrderID:       fmt.Sprintf("order-%s", paymentID[4:]),
		Amount:        99.99,
		TransactionID: fmt.Sprintf("txn-%s", paymentID[4:]),
		Status:        "completed",
		CreatedAt:     time.Now().Add(-30 * time.Minute),
	}, nil
}

func processRefundWithGateway(ctx context.Context, transactionID string, amount float64) (string, error) {
	return fmt.Sprintf("ref-%s", uuid.New().String()), nil
}

func updatePaymentStatus(ctx context.Context, paymentID, status, refundID string) error {
	return nil
}

type CustomerDetails struct {
	ID              string
	Email           string
	Phone           string
	ShippingAddress Address
}

type Address struct {
	Street  string
	City    string
	State   string
	ZipCode string
	Country string
}

func getCustomerDetails(ctx context.Context, customerID string) (*CustomerDetails, error) {
	return &CustomerDetails{
		ID:    customerID,
		Email: "customer@example.com",
		Phone: "+1234567890",
		ShippingAddress: Address{
			Street:  "123 Main St",
			City:    "Anytown",
			State:   "CA",
			ZipCode: "12345",
			Country: "USA",
		},
	}, nil
}

func sendEmail(ctx context.Context, email, subject string, data map[string]interface{}) error {
	return nil
}

func sendSMS(ctx context.Context, phone, message string) error {
	return nil
}

func updateOrderNotificationStatus(ctx context.Context, orderID, status string) error {
	return nil
}
