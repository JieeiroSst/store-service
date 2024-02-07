package kafka

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type QueueKakfa interface {
	Producer(ctx context.Context, topic string, data []byte)
	Consume(ctx context.Context, topic string) (*kafka.Message, error)
}

type queueKakfa struct {
	configMap map[string]kafka.ConfigValue
}

func NetKafkaWriter(kafkaURL string) QueueKakfa {
	config := kafka.ConfigMap{
		"bootstrap.servers": kafkaURL,
	}

	return &queueKakfa{
		configMap: config,
	}
}

func (p *queueKakfa) Producer(ctx context.Context, topic string, data []byte) {
	// Tạo producer
	producer, err := kafka.NewProducer((*kafka.ConfigMap)(&p.configMap))
	if err != nil {
		log.Fatal(err)
	}
	deliveryChannel := make(chan kafka.Event)

	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          data,
	}

	err = producer.Produce(message, deliveryChannel)

	// Use the deliveryChannel or handle potential errors
	if err != nil {
		log.Println("Error publishing message:", err)
	} else {
		select {
		case event := <-deliveryChannel:
			if strings.Contains(event.String(), "Message delivery failed") {
				log.Println("Error delivering message:", event.String())
			} else {
				log.Println("Message delivered successfully:", event.String())
			}
		case <-time.After(time.Second):
			log.Println("Timed out waiting for delivery event")
		}
	}
}

func (p *queueKakfa) Consume(ctx context.Context, topic string) (*kafka.Message, error) {
	consumer, err := kafka.NewConsumer((*kafka.ConfigMap)(&p.configMap))
	if err != nil {
		log.Fatal(err)
	}

	return consumer.ReadMessage(time.Minute)
}
