package eventbus

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/recruitment-platform-service/internal/domain/shared"
	"github.com/JIeeiroSst/recruitment-platform-service/internal/port"
	"go.uber.org/zap"
)

type Handler func(context.Context, shared.DomainEvent) error

type inProcessEventBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *zap.Logger
}

func New(logger *zap.Logger) port.EventBus {
	return &inProcessEventBus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

func (b *inProcessEventBus) Subscribe(eventType string, handler func(context.Context, shared.DomainEvent) error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
	b.logger.Info("event handler registered", zap.String("event_type", eventType))
}

func (b *inProcessEventBus) Publish(ctx context.Context, event shared.DomainEvent) error {
	b.mu.RLock()
	handlers := append([]Handler{}, b.handlers[event.EventType]...)
	b.mu.RUnlock()

	if len(handlers) == 0 {
		return nil
	}

	for _, h := range handlers {
		h := h
		go func() {
			if err := h(ctx, event); err != nil {
				b.logger.Error("event handler error",
					zap.String("event_type", event.EventType),
					zap.String("event_id", event.EventID.String()),
					zap.Error(err),
				)
			}
		}()
	}
	return nil
}

func WireAuditLog(bus port.EventBus, logger *zap.Logger) {
	allEvents := []string{
		"CandidateCreated", "CandidateStatusChanged", "CandidateScoredByAI",
		"JobPublished", "JobPaused", "JobClosed",
		"ApplicationCreated", "ApplicationStageMoved", "ApplicationRejected",
		"ApplicationWithdrawn", "OfferExtended", "OfferAccepted", "CandidateHired",
		"InterviewScheduled", "InterviewFeedbackSubmitted",
		"PartnerHireRecorded",
	}
	for _, evt := range allEvents {
		evt := evt
		bus.Subscribe(evt, func(ctx context.Context, e shared.DomainEvent) error {
			logger.Debug("domain event",
				zap.String("type", e.EventType),
				zap.String("event_id", e.EventID.String()),
			)
			return nil
		})
	}
}
