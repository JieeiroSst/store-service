package kafka

import (
	"errors"
	"log"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"github.com/JieeiroSst/banking-service/pkg/idempotency"
	"github.com/Shopify/sarama"
)

type TransactionConsumer struct {
	transaction port.TransactionUsecase
	idempotency *idempotency.IdempotencyGuard
}

func NewTransactionConsumer(transaction port.TransactionUsecase, guard *idempotency.IdempotencyGuard) *TransactionConsumer {
	return &TransactionConsumer{transaction: transaction, idempotency: guard}
}

func (c *TransactionConsumer) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (c *TransactionConsumer) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (c *TransactionConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			c.handle(session, msg)
		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *TransactionConsumer) handle(session sarama.ConsumerGroupSession, msg *sarama.ConsumerMessage) {
	evt, err := parseTransactionEvent(msg.Value)
	if err != nil {
		log.Printf("kafka: dropping unprocessable transaction event (partition=%d offset=%d): %v", msg.Partition, msg.Offset, err)
		session.MarkMessage(msg, "")
		return
	}

	ctx := session.Context()

	if err := c.idempotency.Acquire(ctx, evt.ExternalRef); err != nil {
		switch {
		case errors.Is(err, idempotency.ErrDuplicate):
			log.Printf("kafka: message external_ref=%s already processed, skipping", evt.ExternalRef)
			session.MarkMessage(msg, "")
			return
		case errors.Is(err, idempotency.ErrProcessing):
			log.Printf("kafka: message external_ref=%s is currently being processed, will retry", evt.ExternalRef)
			return
		default:
			log.Printf("kafka: idempotency check unavailable for external_ref=%s: %v — falling back to DB uniqueness guard", evt.ExternalRef, err)
		}
	}

	externalRef := evt.ExternalRef
	tx := &model.Transaction{
		AccountID:       evt.AccountID,
		ExternalRef:     &externalRef,
		TransactionType: evt.TransactionType,
		Amount:          evt.Amount,
		TransactionDate: evt.TransactionDate,
	}

	if _, err := c.transaction.CreateTransaction(ctx, tx); err != nil {
		if errors.Is(err, common.ErrDuplicate) {
			log.Printf("kafka: transaction external_ref=%s already recorded, skipping", evt.ExternalRef)
			if err := c.idempotency.MarkDone(ctx, evt.ExternalRef); err != nil {
				log.Printf("kafka: failed to reconcile idempotency state for external_ref=%s: %v", evt.ExternalRef, err)
			}
			session.MarkMessage(msg, "")
			return
		}

		log.Printf("kafka: failed to record transaction external_ref=%s (partition=%d offset=%d): %v — will retry", evt.ExternalRef, msg.Partition, msg.Offset, err)
		if err := c.idempotency.Release(ctx, evt.ExternalRef); err != nil {
			log.Printf("kafka: failed to release idempotency lock for external_ref=%s: %v", evt.ExternalRef, err)
		}
		return
	}

	if err := c.idempotency.MarkDone(ctx, evt.ExternalRef); err != nil {
		log.Printf("kafka: failed to mark external_ref=%s done in redis: %v", evt.ExternalRef, err)
	}
	session.MarkMessage(msg, "")
}
