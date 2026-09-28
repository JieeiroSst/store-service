package domain

import "time"

type SessionStatus string

const (
	SessionActive    SessionStatus = "active"
	SessionCompleted SessionStatus = "completed"
	SessionExpired   SessionStatus = "expired"
	SessionCancelled SessionStatus = "cancelled"
)

type Session struct {
	ID        string
	MachineID string
	StartedAt time.Time
	ExpiresAt time.Time
	Status    SessionStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (s *Session) IsActive(now time.Time) bool {
	return s.Status == SessionActive && now.Before(s.ExpiresAt)
}

func (s *Session) End(status SessionStatus) error {
	if s.Status != SessionActive {
		return ErrInvalidTransition
	}
	s.Status = status
	return nil
}
