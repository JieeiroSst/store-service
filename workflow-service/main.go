package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/JIeeiroSst/workflow-service/app"
	"github.com/JIeeiroSst/workflow-service/repository"
	"github.com/JIeeiroSst/workflow-service/services"
	"github.com/gin-gonic/gin"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	var wg sync.WaitGroup

	ctx, cancel := setupSignalHandler()
	defer cancel()

	temporalClient := setupTemporalClient()
	defer temporalClient.Close()

	wg.Add(1)
	go runTemporalWorker(ctx, &wg, temporalClient)

	wg.Add(1)
	go runGinServer(ctx, &wg)

	wg.Wait()
	log.Println("Application shutdown complete")
}

func setupSignalHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-signalCh
		log.Printf("Received signal: %v. Initiating graceful shutdown...", sig)
		cancel()
	}()

	return ctx, cancel
}

func setupTemporalClient() client.Client {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create Temporal client:", err)
	}
	return temporalClient
}

func runTemporalWorker(ctx context.Context, wg *sync.WaitGroup, c client.Client) {
	defer wg.Done()

	w := worker.New(c, "order-processing-task-queue", worker.Options{})

	registerWorkflowAndActivities(w)

	interruptCh := worker.InterruptCh()
	errorCh := make(chan error, 1)

	log.Println("Starting Temporal worker")
	go func() {
		errorCh <- w.Run(interruptCh)
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down Temporal worker...")
		interrupt := make(chan os.Signal, 1)
		signal.Notify(interrupt, os.Interrupt)
		interrupt <- os.Interrupt
	case err := <-errorCh:
		if err != nil {
			log.Printf("Temporal worker error: %v\n", err)
		}
	}
}

func registerWorkflowAndActivities(w worker.Worker) {
	w.RegisterWorkflow(app.OrderWorkflow)
	w.RegisterActivity(app.ValidateOrderActivity)
	w.RegisterActivity(app.ProcessPaymentActivity)
	w.RegisterActivity(app.FulfillOrderActivity)
	w.RegisterActivity(app.RefundPaymentActivity)
	w.RegisterActivity(app.SendConfirmationActivity)
}

func runGinServer(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	router := setupRouter()

	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		log.Println("Starting Gin HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("HTTP server error: %v\n", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("HTTP server shutdown error: %v\n", err)
	}
}

func setupRouter() *gin.Engine {
	var ordersMap = map[string]*app.Order{
		"order1": {ID: "order1", Status: "pending", Amount: 100.0, PaymentID: "", TrackingNumber: ""},
		"order2": {ID: "order2", Status: "completed", Amount: 100.0, PaymentID: "", TrackingNumber: ""},
		"order3": {ID: "order3", Status: "shipped", Amount: 100.0, PaymentID: "", TrackingNumber: ""},
	}

	orderRepo := repository.NewOrderRepository(ordersMap)
	paymentClient := services.NewPaymentClient()
	fulfillmentService := services.NewFulfillmentService()
	notificationService := services.NewNotificationService()

	orderService := app.NewOrderService(
		orderRepo,
		paymentClient,
		fulfillmentService,
		notificationService,
	)

	orderHandler := app.NewOrderHandler(orderService)

	router := gin.Default()

	router.POST("/orders/:id/process", orderHandler.ProcessOrder)
	router.POST("/orders/:id/workflow", orderHandler.StartWorkflowHandler)

	return router
}
