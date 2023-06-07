package dto

import "github.com/JIeeiroSst/workflow-service/internal/activities/card"

type CreateCard struct {
	Cart       card.CartState
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
