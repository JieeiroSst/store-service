package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	"github.com/JIeeiroSst/toggle-service/internal/domain/model"
	"github.com/JIeeiroSst/toggle-service/internal/domain/port"
)

type service struct {
	repo port.AuditRepository
}

func NewService(repo port.AuditRepository) port.AuditService {
	return &service{repo: repo}
}

func (s *service) Record(ctx context.Context, entityType string, entityID uuid.UUID, action model.AuditAction, projectID, environmentID *uuid.UUID, userID *string, before, after any) error {
	event := &model.AuditEvent{
		EntityType:    entityType,
		EntityID:      entityID,
		Action:        action,
		ProjectID:     projectID,
		EnvironmentID: environmentID,
		UserID:        userID,
		CreatedAt:     time.Now(),
	}
	if before != nil {
		if b, err := json.Marshal(before); err == nil {
			event.BeforeJSON = datatypes.JSON(b)
		}
	}
	if after != nil {
		if b, err := json.Marshal(after); err == nil {
			event.AfterJSON = datatypes.JSON(b)
		}
	}
	return s.repo.Create(ctx, event)
}

func (s *service) List(ctx context.Context, projectID uuid.UUID, entityType string, since, until *time.Time) ([]model.AuditEvent, error) {
	return s.repo.ListByProject(ctx, projectID, entityType, since, until)
}
