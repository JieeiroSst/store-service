package constants

const (
	OrderProcessingTaskQueue = "ORDER_PROCESSING_TASK_QUEUE"
	CronTaskQueue            = "ORDER_CRON_TASK_QUEUE"

	OrderWorkflowIDPrefix = "order-workflow"
	CronWorkflowID        = "order-cleanup-cron"

	OrderCleanupCronSchedule = "0 */6 * * *" 

	OrderStatusPending    = "PENDING"
	OrderStatusValidated  = "VALIDATED"
	OrderStatusPaymentOK  = "PAYMENT_CONFIRMED"
	OrderStatusShipping   = "SHIPPING"
	OrderStatusCompleted  = "COMPLETED"
	OrderStatusCancelled  = "CANCELLED"
	OrderStatusFailed     = "FAILED"

	ActivityStartToCloseTimeout = 30  
	ChildWorkflowTimeout        = 120 
)
