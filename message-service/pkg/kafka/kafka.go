package kafka

import (
	"context"
	"fmt"
	"log"

	"github.com/segmentio/kafka-go"
)

func NetKafkaWriter(kafkaURL string) *QueueKakfa {
	return &QueueKakfa{
		KafkaURL: kafkaURL,
	}
}

func (p *QueueKakfa) Producer(remoteAddr, topic string, body []byte, ctx context.Context) {
	kafkaWriter := kafka.Writer{
		Addr:     kafka.TCP(p.KafkaURL),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
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

func (p *QueueKakfa) Consume(ctx context.Context, group, topic, remoteAddr string) (kafka.Message, error) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{remoteAddr},
		Topic:    topic,
		GroupID:  group,
		MinBytes: 10e3,
		MaxBytes: 10e6,
	})

	// for {
	// 	msg, err := r.ReadMessage(ctx)
	// 	if err != nil {
	// 		panic("could not read message " + err.Error())
	// 	}
	// 	fmt.Printf("%v received: %v", group, string(msg.Value))
	// 	fmt.Println()
	// }
	return r.ReadMessage(ctx)
}
