package usecase

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/ticket-service/common"
	"github.com/JIeeiroSst/ticket-service/dto"
	"github.com/JIeeiroSst/ticket-service/internal/repository"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/logger"
)

type Tickets interface {
	SaveTicket(ctx context.Context, req dto.CreateTicketsRequest) error
	UpdateTicket(ctx context.Context, ticketID, status int) error
}

type TicketsUsecase struct {
	Repo *repository.Repository
}

func NewTicketsUsecase(repo *repository.Repository) *TicketsUsecase {
	return &TicketsUsecase{
		Repo: repo,
	}
}

func (u *TicketsUsecase) SaveTicket(ctx context.Context, req dto.CreateTicketsRequest) error {
	model := model.Tickets{
		TicketID:    req.TicketID,
		TicketName:  req.TicketName,
		StartDate:   req.StartDate,
		AddressRoom: req.AddressRoom,
		Amount:      req.Amount,
		Quantity:    req.Quantity,
		Status:      common.PENDING.Value(),
	}

	if req.TicketID == 0 {
		model.TicketID = geared_id.GearedIntID()
	}

	if err := u.Repo.Tickets.SaveTickets(ctx, model); err != nil {
		logger.Error(ctx,"error %v",err)
		return err
	}
	return nil
}

func (u *TicketsUsecase) UpdateTicket(ctx context.Context, ticketID, status int) error {
	switch status {
	case common.APPROVE.Value():
		if err := u.Repo.Tickets.UpdateStatusTicket(ctx, common.APPROVE.Value(), ticketID); err != nil {
			logger.Error(ctx,"error %v",err)
			return err
		}
		return nil
	case common.REJECT.Value():
		if err := u.Repo.Tickets.UpdateStatusTicket(ctx, common.REJECT.Value(), ticketID); err != nil {
			logger.Error(ctx,"error %v",err)
			return err
		}
		return nil
	default:
		return errors.New("not found")
	}
}
