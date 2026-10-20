package dto

import (
	"time"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type RewardEventDTO struct {
	EventID    string    `json:"event_id"`
	Source     string    `json:"source"`
	UserID     string    `json:"user_id"`
	Title      string    `json:"title"`
	Points     int64     `json:"points"`
	OccurredAt time.Time `json:"occurred_at"`
}

func (e RewardEventDTO) ToEntity() entity.PointTransaction {
	return entity.PointTransaction{
		EventID:    e.EventID,
		UserID:     e.UserID,
		Source:     e.Source,
		Title:      e.Title,
		Points:     e.Points,
		OccurredAt: e.OccurredAt.UTC(),
	}
}

type PointAccountDTO struct {
	UserID      string    `json:"user_id"`
	Balance     int64     `json:"balance"`
	TotalEarned int64     `json:"total_earned"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PointTransactionDTO struct {
	ID           int64     `json:"id"`
	EventID      string    `json:"event_id"`
	Source       string    `json:"source"`
	Title        string    `json:"title"`
	Points       int64     `json:"points"`
	BalanceAfter int64     `json:"balance_after"`
	OccurredAt   time.Time `json:"occurred_at"`
}

type PointTransactionPageDTO struct {
	Items  []PointTransactionDTO `json:"items"`
	Total  int64                 `json:"total"`
	Limit  int                   `json:"limit"`
	Offset int                   `json:"offset"`
}

func AccountFromEntity(a entity.PointAccount) PointAccountDTO {
	return PointAccountDTO{UserID: a.UserID, Balance: a.Balance, TotalEarned: a.TotalEarned, UpdatedAt: a.UpdatedAt}
}

func TransactionFromEntity(t entity.PointTransaction) PointTransactionDTO {
	return PointTransactionDTO{
		ID:           t.ID,
		EventID:      t.EventID,
		Source:       t.Source,
		Title:        t.Title,
		Points:       t.Points,
		BalanceAfter: t.BalanceAfter,
		OccurredAt:   t.OccurredAt,
	}
}
