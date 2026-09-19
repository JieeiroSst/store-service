package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/parking-lot-service/internal/domain/model"
	"github.com/JIeeiroSst/parking-lot-service/internal/domain/port"
	"gorm.io/gorm"
)

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) port.TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) Create(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Create(ticket).Error
}

func (r *ticketRepository) GetByID(ctx context.Context, id string) (*model.Ticket, error) {
	var ticket model.Ticket
	if err := r.db.WithContext(ctx).Where("history_id = ?", id).First(&ticket).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, port.ErrNotFound
		}
		return nil, err
	}
	return &ticket, nil
}

func (r *ticketRepository) Update(ctx context.Context, ticket *model.Ticket) error {
	return r.db.WithContext(ctx).Model(&model.Ticket{}).
		Where("history_id = ?", ticket.TicketID).
		Updates(map[string]interface{}{
			"leave_time":   ticket.LeaveTime,
			"exit_gate_id": ticket.ExitGateID,
			"status":       ticket.Status,
		}).Error
}

func (r *ticketRepository) ListByPlate(ctx context.Context, plate string, limit, offset int) ([]model.Ticket, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var tickets []model.Ticket
	err := r.db.WithContext(ctx).
		Where("vehicle_plate = ?", plate).
		Order("parked_time DESC").
		Limit(limit).
		Offset(offset).
		Find(&tickets).Error
	return tickets, err
}
