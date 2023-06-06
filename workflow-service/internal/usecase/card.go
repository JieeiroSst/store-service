package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/workflow-service/dto"
	"github.com/JIeeiroSst/workflow-service/internal/activities"
	"go.temporal.io/sdk/client"
)

type Cards interface {
	CreateCard() (*dto.CreateCard, error)
	AddToCard(workflowID string, item activities.CartItem) error
	RemoveFromCard(workflowID string, item activities.CartItem) error
	UpdateEmailToCard(workflowID string, body dto.UpdateEmailRequest) error
	CheckoutCard(workflowID string, body dto.CheckoutRequest) error
	GetCard(workflowID string) (*dto.Product, error)
}

type CardUsecase struct {
	Temporal client.Client
}

func NewCardUsecase(Temporal client.Client) *CardUsecase {
	return &CardUsecase{
		Temporal: Temporal,
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

	cart := activities.CartState{Items: make([]activities.CartItem, 0)}
	we, err := u.Temporal.ExecuteWorkflow(context.Background(), options, activities.CartWorkflow, cart)
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

func (u *CardUsecase) AddToCard(workflowID string, item activities.CartItem) error {
	update := activities.AddToCartSignal{Route: activities.RouteTypes.ADD_TO_CART, Item: item}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", activities.SignalChannels.ADD_TO_CART_CHANNEL, update)
	if err != nil {
		return err
	}
	return nil
}

func (u *CardUsecase) RemoveFromCard(workflowID string, item activities.CartItem) error {
	update := activities.RemoveFromCartSignal{
		Route: activities.RouteTypes.REMOVE_FROM_CART,
		Item:  item,
	}

	if err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", activities.SignalChannels.REMOVE_FROM_CART_CHANNEL, update); err != nil {
		return err
	}

	return nil
}

func (u *CardUsecase) UpdateEmailToCard(workflowID string, body dto.UpdateEmailRequest) error {
	updateEmail := activities.UpdateEmailSignal{Route: activities.RouteTypes.UPDATE_EMAIL, Email: body.Email}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", activities.SignalChannels.UPDATE_EMAIL_CHANNEL, updateEmail)
	if err != nil {
		return err
	}
	return nil
}

func (u *CardUsecase) CheckoutCard(workflowID string, body dto.CheckoutRequest) error {
	checkout := activities.CheckoutSignal{Route: activities.RouteTypes.CHECKOUT, Email: body.Email}

	err := u.Temporal.SignalWorkflow(context.Background(), workflowID, "", activities.SignalChannels.CHECKOUT_CHANNEL, checkout)
	if err != nil {
		return err
	}
	return nil
}
