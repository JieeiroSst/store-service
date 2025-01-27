package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/config"
	"github.com/JIeeiroSst/coupon-service/internal/delivery/consumer"
	"github.com/nats-io/nats.go"
)

func runConsumer() {
	log.Print(`Starting consumer server`)

	config, _ := config.InitializeConfiguration(ecosystem)

	mc, err := consumer.NewMultiConsumer(config.Nats.DNS)
	if err != nil {
		log.Fatalf(`Error creating consumer: %v`, err)
	}

	err = mc.AddConsumer(consumer.ConsumerConfig{
		Subject:    "orders.new",
		QueueGroup: "order-processors",
		NumWorkers: 3,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}, handleNewOrder)
	if err != nil {
		log.Fatalf(`Error adding consumer: %v`, err)
	}

	err = mc.AddConsumer(consumer.ConsumerConfig{
		Subject:    "orders.update",
		QueueGroup: "order-updaters",
		NumWorkers: 2,
		MaxRetries: 3,
		RetryDelay: time.Second,
	}, handleOrderUpdate)
	if err != nil {
		log.Fatalf(`Error adding consumer: %v`, err)
	}

	go mc.Start()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	mc.Stop()
}

func handleNewOrder(msg *nats.Msg) error {
	return nil
}

func handleOrderUpdate(msg *nats.Msg) error {
	return nil
}
