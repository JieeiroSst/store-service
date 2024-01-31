package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/gin-gonic/gin"
)

func main() {
	// Khởi tạo cấu hình Kafka
	config := kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	}

	// Tạo producer
	producer, err := kafka.NewProducer(&config)
	if err != nil {
		log.Fatal(err)
	}

	// Tạo topic
	topic := "my-topic"

	// Tạo router
	router := gin.New()

	// Tạo endpoint để publish message
	router.POST("/publish", func(c *gin.Context) {
		// Lấy message từ request
		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusBadRequest, "Error reading request body")
			return
		}

		// Publish message
		deliveryChannel := make(chan kafka.Event)

		message := &kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}

		// Publish the message
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

		// Trả về phản hồi
		c.JSON(http.StatusOK, gin.H{
			"message": "Message published successfully",
		})
	})

	// Tạo endpoint để consume message
	router.GET("/consume", func(c *gin.Context) {
		// Tạo consumer
		consumer, err := kafka.NewConsumer(&config)
		if err != nil {
			log.Fatal(err)
		}

		// Subscribe topic
		consumer.Subscribe(topic, nil)

		// Consume message
		for {
			message, err := consumer.ReadMessage(time.Minute)
			if err != nil {
				log.Fatal(err)
			}

			// Trả về phản hồi
			c.JSON(http.StatusOK, gin.H{
				"message": string(message.Value),
			})
		}
	})

	// Khởi động server
	http.ListenAndServe(":8080", router)
}
