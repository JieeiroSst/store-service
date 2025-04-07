package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/workflow-service/app"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
)

type MockDependencies struct {
	OrderRepo           *MockOrderRepository
	PaymentClient       *MockPaymentClient
	FulfillmentService  *MockFulfillmentService
	NotificationService *MockNotificationService
}

type MockOrderRepository struct{ mock.Mock }
type MockPaymentClient struct{ mock.Mock }
type MockFulfillmentService struct{ mock.Mock }
type MockNotificationService struct{ mock.Mock }

func (m *MockPaymentClient) ProcessPayment(orderID string, amount float64) (string, error) {
	args := m.Called(orderID, amount)
	return args.String(0), args.Error(1)
}

func (m *MockPaymentClient) RefundPayment(paymentID string) error {
	args := m.Called(paymentID)
	return args.Error(0)
}

type ActivityUnitTestSuite struct {
	suite.Suite
	mocks MockDependencies
}

func (s *ActivityUnitTestSuite) SetupTest() {
	s.mocks = MockDependencies{
		OrderRepo:           new(MockOrderRepository),
		PaymentClient:       new(MockPaymentClient),
		FulfillmentService:  new(MockFulfillmentService),
		NotificationService: new(MockNotificationService),
	}
}

func (s *ActivityUnitTestSuite) TestProcessPaymentActivity() {
	ctx := context.Background()
	orderID := "test-order-123"
	expectedPaymentID := "pmt-test-order-123"

	s.mocks.PaymentClient.On("ProcessPayment", orderID, mock.Anything).Return(expectedPaymentID, nil)

	result, err := ProcessPaymentActivityWithDeps(ctx, orderID, s.mocks.PaymentClient)

	s.NoError(err)
	s.Equal(expectedPaymentID, result)
	s.mocks.PaymentClient.AssertExpectations(s.T())
}

type WorkflowTestSuite struct {
	testsuite.WorkflowTestSuite
	suite.Suite
}

func (s *WorkflowTestSuite) Test_OrderWorkflow() {
	env := s.NewTestWorkflowEnvironment()

	env.OnActivity(app.ValidateOrderActivity, mock.Anything, "test-order-123").Return(true, nil)
	env.OnActivity(app.ProcessPaymentActivity, mock.Anything, "test-order-123").Return("payment-123", nil)
	env.OnActivity(app.FulfillOrderActivity, mock.Anything, "test-order-123", "payment-123").Return("tracking-123", nil)
	env.OnActivity(app.SendConfirmationActivity, mock.Anything, "test-order-123", "tracking-123").Return(nil)

	env.ExecuteWorkflow(app.OrderWorkflow, "test-order-123")

	s.True(env.IsWorkflowCompleted())
	s.NoError(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}

func (s *WorkflowTestSuite) Test_OrderWorkflowFailure() {
	env := s.NewTestWorkflowEnvironment()

	env.OnActivity(app.ValidateOrderActivity, mock.Anything, "test-order-123").Return(true, nil)
	env.OnActivity(app.ProcessPaymentActivity, mock.Anything, "test-order-123").Return("payment-123", nil)
	env.OnActivity(app.FulfillOrderActivity, mock.Anything, "test-order-123", "payment-123").Return("",
		errors.New("fulfillment failed"))
	env.OnActivity(app.RefundPaymentActivity, mock.Anything, "payment-123").Return(true, nil)

	env.ExecuteWorkflow(app.OrderWorkflow, "test-order-123")

	s.True(env.IsWorkflowCompleted())
	s.Error(env.GetWorkflowError())

	env.AssertExpectations(s.T())
}

func setupLocalTemporalTestServer(t *testing.T) *testsuite.TestWorkflowEnvironment {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	return env
}

func TestIntegration_OrderWorkflow(t *testing.T) {
	setupLocalTemporalTestServer(t)

	c, err := client.Dial(client.Options{
		HostPort: "localhost:7233", 
	})
	if err != nil {
		t.Fatal("Unable to create Temporal client", err)
	}
	defer c.Close()

	taskQueue := "integration-test-queue"
	w := worker.New(c, taskQueue, worker.Options{})

	w.RegisterWorkflow(app.OrderWorkflow)
	w.RegisterActivity(app.ValidateOrderActivity)
	w.RegisterActivity(app.ProcessPaymentActivity)
	w.RegisterActivity(app.FulfillOrderActivity)
	w.RegisterActivity(app.RefundPaymentActivity)
	w.RegisterActivity(app.SendConfirmationActivity)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		err := w.Run(worker.InterruptCh())
		if err != nil {
			t.Logf("Worker error: %v", err)
		}
	}()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "integration-test-workflow",
		TaskQueue: taskQueue,
	}

	we, err := c.ExecuteWorkflow(ctx, workflowOptions, app.OrderWorkflow, "integration-test-order")
	if err != nil {
		t.Fatal("Unable to execute workflow", err)
	}

	var result interface{}
	err = we.Get(ctx, &result)
	if err != nil {
		t.Fatal("Workflow failed", err)
	}
}

func ProcessPaymentActivityWithDeps(ctx context.Context, orderID string, client app.PaymentClient) (string, error) {
	return client.ProcessPayment(orderID, 0) 
}

func TestActivityUnitTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityUnitTestSuite))
}

func TestWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(WorkflowTestSuite))
}
