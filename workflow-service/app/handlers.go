package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"
)

type OrderHandler struct {
	orderService *OrderService
}

func NewOrderHandler(orderService *OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

func (h *OrderHandler) ProcessOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	err := h.orderService.ProcessOrder(orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order processed successfully"})
}

func (h *OrderHandler) StartWorkflowHandler(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	tc, err := client.Dial(client.Options{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to Temporal"})
		return
	}
	defer tc.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "order-workflow-" + orderID,
		TaskQueue: "order-processing-task-queue",
	}

	we, err := tc.ExecuteWorkflow(c.Request.Context(), workflowOptions, OrderWorkflow, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Workflow started successfully",
		"workflowId": we.GetID(),
		"runId":      we.GetRunID(),
	})
}
