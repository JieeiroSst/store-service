package firebase

import (
	"context"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"google.golang.org/api/option"
)

type FirebaseMessaging struct {
	client *messaging.Client
}

func NewFirebaseMessaging(credentialsFile string) (*FirebaseMessaging, error) {
	opt := option.WithCredentialsFile(credentialsFile)
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		return nil, err
	}

	return &FirebaseMessaging{
		client: client,
	}, nil
}

func (fm *FirebaseMessaging) SendToToken(ctx context.Context, token string, title, body string, data map[string]string) (string, error) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Token: token,
	}

	return fm.client.Send(ctx, message)
}

func (fm *FirebaseMessaging) SendToTopic(ctx context.Context, topic string, title, body string, data map[string]string) (string, error) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:  data,
		Topic: topic,
	}

	return fm.client.Send(ctx, message)
}

func (fm *FirebaseMessaging) SendMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:   data,
		Tokens: tokens,
	}

	return fm.client.SendMulticast(ctx, message)
}
