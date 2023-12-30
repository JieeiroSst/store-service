package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func NetKafkaWriter(kafkaURL, topic string) *kafka.Writer {
	return &kafka.Writer{
		Addr:     kafka.TCP(kafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
}

func (p *QueueKakfa) Producer(kafkaWriter *kafka.Writer, remoteAddr string,body []byte,  ctx context.Context) {
	for {
		msg := kafka.Message{
			Key:   []byte(fmt.Sprintf("address-%s", remoteAddr)),
			Value: []byte(body),
		}
		err := kafkaWriter.WriteMessages(ctx, msg)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

func (p *QueueKakfa) Consume(ctx context.Context, group, topic, remoteAddr string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{remoteAddr},
		Topic:   topic,
		GroupID: group,
	})

	for {
		msg, err := r.ReadMessage(ctx)
		if err != nil {
			panic("could not read message " + err.Error())
		}
		fmt.Printf("%v received: %v", group, string(msg.Value))
		fmt.Println()
	}
}
