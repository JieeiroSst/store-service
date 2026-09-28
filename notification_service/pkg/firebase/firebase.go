package firebase

import (
	"context"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
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

func AndroidConfig() *messaging.AndroidConfig {
	return &messaging.AndroidConfig{
		Priority:     "high",
		Notification: &messaging.AndroidNotification{Sound: "default"},
	}
}

func APNSConfig() *messaging.APNSConfig {
	return &messaging.APNSConfig{
		Headers: map[string]string{"apns-priority": "10", "apns-push-type": "alert"},
		Payload: &messaging.APNSPayload{Aps: &messaging.Aps{Sound: "default"}},
	}
}

func (fm *FirebaseMessaging) SendToToken(ctx context.Context, token string, title, body string, data map[string]string) (string, error) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:    data,
		Token:   token,
		Android: AndroidConfig(),
		APNS:    APNSConfig(),
	}

	return fm.client.Send(ctx, message)
}

func (fm *FirebaseMessaging) SendToTopic(ctx context.Context, topic string, title, body string, data map[string]string) (string, error) {
	message := &messaging.Message{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:    data,
		Topic:   topic,
		Android: AndroidConfig(),
		APNS:    APNSConfig(),
	}

	return fm.client.Send(ctx, message)
}

func (fm *FirebaseMessaging) SendEachForMulticast(ctx context.Context, tokens []string, title, body string, data map[string]string) (*messaging.BatchResponse, error) {
	message := &messaging.MulticastMessage{
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data:    data,
		Tokens:  tokens,
		Android: AndroidConfig(),
		APNS:    APNSConfig(),
	}

	return fm.client.SendEachForMulticast(ctx, message)
}

func (fm *FirebaseMessaging) ValidateToken(ctx context.Context, token string) (string, error) {
	return fm.client.SendDryRun(ctx, &messaging.Message{
		Token:        token,
		Notification: &messaging.Notification{Title: "validate", Body: "validate"},
	})
}
