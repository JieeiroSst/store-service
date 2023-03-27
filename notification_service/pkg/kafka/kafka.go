package kafka

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/Shopify/sarama"
)

type WorkerInterface interface {
	PushToQueue(topic string, message []byte) error
	Consumer(topic string, fn func(data []byte) error) error
	connectConsumer() (sarama.Consumer, error)
}
type Worker struct {
	URL string
}

type URL struct {
	Url string
}

func NewWorker(url string) WorkerInterface {
	return &Worker{
		URL: url,
	}
}

func ConnectProducer(brokersUrl []string) (sarama.SyncProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5

	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (w *Worker) PushToQueue(topic string, message []byte) error {

	brokersUrl := []string{w.URL}
	producer, err := ConnectProducer(brokersUrl)
	if err != nil {
		return err
	}

	defer producer.Close()

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}

	fmt.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)

	return nil
}

func (w *Worker) connectConsumer() (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	conn, err := sarama.NewConsumer([]string{w.URL}, config)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (w *Worker) Consumer(topic string, fn func(data []byte) error) error {
	worker, err := w.connectConsumer()
	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}
	fmt.Println("Consumer started ")
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	msgCount := 0

	// Get signal for finish
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCount++
				fmt.Printf("Received message Count %d: | Topic(%s) | Message(%s) \n", msgCount, string(msg.Topic), string(msg.Value))
				if err := fn(msg.Value); err != nil {
					log.Println(err)
				}
			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCount, "messages")

	if err := worker.Close(); err != nil {
		panic(err)
	}

	return nil
}
