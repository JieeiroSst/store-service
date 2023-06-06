package dto

import "github.com/JIeeiroSst/workflow-service/internal/activities"

type CreateCard struct {
	Cart       activities.CartState
	WorkflowID string
}

type UpdateEmailRequest struct {
	Email string
}

type CheckoutRequest struct {
	Email string
}

type Product struct {
	Id          int
	Name        string
	Description string
	Image       string
	Price       float32
}
