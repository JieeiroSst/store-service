package redis

import (
	"context"
	"encoding/json"

	"github.com/go-redis/redis/v8"
)

type PubSub struct {
	client *redis.Client
}

func NewPubSub(dns string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     dns,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (r *PubSub) PublishMessage(ctx context.Context, channel string, message interface{}) error {
	payload, err := json.Marshal(message)
	if err != nil {
		panic(err)
	}
	err = r.client.Publish(ctx, channel, payload).Err()
	return err
}

//	for {
//		msg, err := subscriber.ReceiveMessage(ctx)
//		if err != nil {
//			panic(err)
//		}
//		if err := json.Unmarshal([]byte(msg.Payload), &user); err != nil {
//			panic(err)
//		}
//	}
func (r *PubSub) Subscribe(ctx context.Context, channel string) (*redis.Message, error) {
	subscriber := r.client.Subscribe(ctx, channel)
	defer subscriber.Close()
	return subscriber.ReceiveMessage(ctx)
}
