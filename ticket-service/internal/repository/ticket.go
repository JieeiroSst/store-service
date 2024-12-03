package repository

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/common"
	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/JIeeiroSst/utils/pagination"
	"gorm.io/gorm"
)

type Tickets interface {
	SaveTickets(ctx context.Context, req model.Tickets) error
	FindByID(ctx context.Context, ticketID int) (*model.Tickets, error)
	FindPagination(ctx context.Context, p pagination.Pagination) (*pagination.Pagination, error)
	UpdateStatusTicket(ctx context.Context, status, ticketID int) error
	UpdateQuantityTickets(ctx context.Context, status, quantity, ticketID int) error
}

type TicketsRepository struct {
	db *gorm.DB
}

func NewTicketsRepository(db *gorm.DB) *TicketsRepository {
	return &TicketsRepository{
		db: db,
	}
}

func (r *TicketsRepository) SaveTickets(ctx context.Context, req model.Tickets) error {
	var ticket model.Tickets

	if err := r.db.Where("ticket_id = ?", req.TicketID).Error; err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}

	if ticket.TicketID == 0 {
		if err := r.db.Create(&req).Error; err != nil {
			logger.Error(ctx, "error %v", err)
			return err
		}
		return nil
	}
	if err := r.db.Save(&req).Error; err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}

	return nil
}

func (r *TicketsRepository) FindByID(ctx context.Context, ticketID int) (*model.Tickets, error) {
	var ticket model.Tickets
	if err := r.db.Where("ticket_id = ?", ticketID).Find(&ticket).Error; err != nil {
		logger.Error(ctx, "error %v", err)
		return nil, err
	}
	return &ticket, nil
}

func (r *TicketsRepository) FindPagination(ctx context.Context, param pagination.Pagination) (*pagination.Pagination, error) {
	var tickets []model.Tickets

	r.db.Scopes(pagination.Paginate(tickets, &param, r.db)).Find(&tickets)
	param.Rows = tickets

	return &param, nil
}

func (r *TicketsRepository) UpdateStatusTicket(ctx context.Context, status, ticketID int) error {
	if err := r.db.Model(model.Tickets{}).Where("ticket_id = ?", ticketID).
		Update("status = ?", status).Error; err != nil {
		logger.Error(ctx, "error %v", err)
		return err
	}
	return nil
}

func (r *TicketsRepository) UpdateQuantityTickets(ctx context.Context, status, quantity, ticketID int) error {
	ticket, err := r.FindByID(ctx, ticketID)
	if err != nil {
		return err
	}

	switch status {
	case common.PENDING.Value():
		ticket.Quantity -= quantity
		if err := r.SaveTickets(ctx, *ticket); err != nil {
			logger.Error(ctx, "error %v", err)
			return err
		}
		return nil
	case common.REJECT.Value():
		ticket.Quantity += quantity
		if err := r.SaveTickets(ctx, *ticket); err != nil {
			logger.Error(ctx, "error %v", err)
			return err
		}
		return nil
	default:
		return nil
	}
}
