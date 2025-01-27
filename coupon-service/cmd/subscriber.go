package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JIeeiroSst/coupon-service/internal/config"
	"github.com/JIeeiroSst/coupon-service/internal/delivery/subscriber"
	"github.com/nats-io/nats.go"
)

func runSubscriber() {
	log.Print(`Starting subscriber server`)

	config, _ := config.InitializeConfiguration(ecosystem)

	ms, err := subscriber.NewMultiSubscriber(config.Nats.DNS)
	if err != nil {
		log.Fatal(err)
	}

	// Add a JetStream subscription with queue group
	err = ms.SubscribeJetStream(subscriber.SubscriptionConfig{
		Subject:     "orders.persistent",
		DurableName: "order-processor",
		MaxInflight: 100,
		AckWait:     time.Second * 30,
		QueueGroup:  "order-processors",
	}, handlePersistentOrders)

	if err != nil { 
		log.Fatal(err)
	}

	// Wait for shutdown signal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	ms.Stop()
}

func handlePersistentOrders(msg *nats.Msg) {
	log.Printf("Received order: %s", string(msg.Data))
	msg.Ack()
}
