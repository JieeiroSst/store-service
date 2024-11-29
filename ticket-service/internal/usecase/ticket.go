package usecase

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/common"
	"github.com/JIeeiroSst/ticket-service/dto"
	"github.com/JIeeiroSst/ticket-service/internal/repository"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/logger"
)

type Tickets interface {
	CreateTicket(ctx context.Context, req dto.CreateTicketsRequest) error
}

type TicketsUsecase struct {
	Repo *repository.Repository
}

func NewTicketsUsecase(repo *repository.Repository) *TicketsUsecase {
	return &TicketsUsecase{
		Repo: repo,
	}
}

func (u *TicketsUsecase) CreateTicket(ctx context.Context, req dto.CreateTicketsRequest) error {
	model := model.Tickets{
		TicketID:    geared_id.GearedIntID(),
		TicketName:  req.TicketName,
		StartDate:   req.StartDate,
		AddressRoom: req.AddressRoom,
		Amount:      req.Amount,
		Quantity:    req.Quantity,
		Status:      common.PENDING.Value(),
	}

	if err := u.Repo.Tickets.SaveTickets(ctx, model); err != nil {
		logger.Error(err)
		return err
	}
	return nil
}
