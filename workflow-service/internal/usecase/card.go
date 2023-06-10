package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/activities/card"
	cardWorkflow "github.com/JIeeiroSst/workflow-service/internal/activities/card"
	"go.temporal.io/sdk/client"
)

type Cards interface {
	CreateCard() (*dto.CreateCard, error)
	AddToCard(workflowID string, item cardWorkflow.CartItem) error
	RemoveFromCard(workflowID string, item cardWorkflow.CartItem) error
	UpdateEmailToCard(workflowID string, body dto.UpdateEmailRequest) error
	CheckoutCard(workflowID string, body dto.CheckoutRequest) error
	GetCard(workflowID string) (*dto.Product, error)
}

type CardUsecase struct {
	Temporal client.Client
	Card     card.CardWorkflow
}

func NewCardUsecase(Temporal client.Client, Card card.CardWorkflow) *CardUsecase {
	return &CardUsecase{
		Temporal: Temporal,
		Card:     Card,
	}
}

func (u *CardUsecase) GetProducts() ([]dto.Product, error) {

	return nil, nil
}

func (u *CardUsecase) CreateCard() (*dto.CreateCard, error) {
	workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "CART_TASK_QUEUE",
	}

	cart := cardWorkflow.CartState{Items: make([]cardWorkflow.CartItem, 0)}
	we, err := u.Temporal.ExecuteWorkflow(context.Background(), options, u.Card.CartWorkflow, cart)
	if err != nil {
		return nil, err
	}
	return &dto.CreateCard{
		Cart:       cart,
		WorkflowID: we.GetID(),
	}, nil
}

func (u *CardUsecase) GetCard(workflowID string) (*dto.Product, error) {
	response, err := u.Temporal.QueryWorkflow(context.Background(), workflowID, "", "getCart")
	if err != nil {
		return nil, err
	}
	var res *dto.Product
	if err := response.Get(&res); err != nil {
		return nil, err
	}
	return res, nil
}

func (u *CardUsecase) AddToCard(workflowID string, item cardWorkflow.CartItem) error {
	update := cardWorkflow.AddToCartSignal{Route: cardWorkflow.RouteTypes.ADD_TO_CART, Item: item}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", cardWorkflow.SignalChannels.ADD_TO_CART_CHANNEL, update)
	if err != nil {
		return err
	}
	return nil
}

func (u *CardUsecase) RemoveFromCard(workflowID string, item cardWorkflow.CartItem) error {
	update := cardWorkflow.RemoveFromCartSignal{
		Route: cardWorkflow.RouteTypes.REMOVE_FROM_CART,
		Item:  item,
	}

	if err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", cardWorkflow.SignalChannels.REMOVE_FROM_CART_CHANNEL, update); err != nil {
		return err
	}

	return nil
}

func (u *CardUsecase) UpdateEmailToCard(workflowID string, body dto.UpdateEmailRequest) error {
	updateEmail := cardWorkflow.UpdateEmailSignal{Route: cardWorkflow.RouteTypes.UPDATE_EMAIL, Email: body.Email}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", cardWorkflow.SignalChannels.UPDATE_EMAIL_CHANNEL, updateEmail)
	if err != nil {
		return err
	}
	return nil
}

func (u *CardUsecase) CheckoutCard(workflowID string, body dto.CheckoutRequest) error {
	checkout := cardWorkflow.CheckoutSignal{Route: cardWorkflow.RouteTypes.CHECKOUT, Email: body.Email}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", cardWorkflow.SignalChannels.CHECKOUT_CHANNEL, checkout)
	if err != nil {
		return err
	}
	return nil
}
