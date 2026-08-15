package consumer

import (
	"context"
	"encoding/json"

	auditapp "github.com/JIeeiroSst/voucher-service/internal/application/audit"
	"go.uber.org/zap"
)

const auditGroup = "voucher-service-audit-log"
type AuditConsumer struct {
	reader   *Reader
	auditSvc auditapp.AuditService
	log      *zap.Logger
}

func NewAuditConsumer(topic string, factory ReaderFactory, auditSvc auditapp.AuditService, log *zap.Logger) *AuditConsumer {
	return &AuditConsumer{
		reader:   factory(topic, auditGroup),
		auditSvc: auditSvc,
		log:      log,
	}
}

func (c *AuditConsumer) Run(ctx context.Context) {
	for {
		msg, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			c.log.Error("audit consumer: read failed", zap.Error(err))
			continue
		}

		var payload map[string]any
		if err := json.Unmarshal(msg.Value, &payload); err != nil {
			c.log.Error("audit consumer: bad payload", zap.Error(err))
			continue
		}

		entry := auditapp.Entry{
			ActorType:  "system",
			Action:     stringField(payload, "Type"),
			EntityType: msg.Topic,
			EntityID:   stringField(payload, "AggID"),
			After:      payload,
		}
		if err := c.auditSvc.Record(ctx, entry); err != nil {
			c.log.Error("audit consumer: record failed", zap.Error(err))
		}
	}
}

func (c *AuditConsumer) Close() error {
	return c.reader.Close()
}

func stringField(m map[string]any, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
