package publisher

import (
	"context"
	"encoding/json"

	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
)

type BatchMessage struct {
	BatchID uint `json:"batch_id"`
}

type batchPublisher struct {
	mq rabbitmq.RabbitMQ
}

func NewBatchPublisher(mq rabbitmq.RabbitMQ) port.BatchPublisher {
	return &batchPublisher{mq: mq}
}

func (p *batchPublisher) PublishBatch(_ context.Context, batchID uint) error {
	body, err := json.Marshal(BatchMessage{BatchID: batchID})
	if err != nil {
		return err
	}
	return p.mq.Publish(rabbitmq.QueueCampaignBatch, body, 1)
}
