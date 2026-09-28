package worker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/application"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"github.com/JIeeiroSst/nofitifaction-service/pkg/rabbitmq"
)

type fakeMQ struct {
	retries     []int
	deadLetters []string
	publishErr  error
}

func (m *fakeMQ) PublishToQueue(*model.Notification) error   { return nil }
func (m *fakeMQ) DeclareQueue(string, []time.Duration) error { return nil }
func (m *fakeMQ) Publish(string, []byte, int) error          { return nil }
func (m *fakeMQ) Consume(context.Context, string, int, int, rabbitmq.Handler) error {
	return nil
}
func (m *fakeMQ) Retry(queue string, _ []byte, next int) (bool, error) {
	if m.publishErr != nil {
		return false, m.publishErr
	}
	if next-1 > len(application.NotificationRetryDelays) {
		return false, nil
	}
	m.retries = append(m.retries, next)
	return true, nil
}
func (m *fakeMQ) DeadLetter(_ string, _ []byte, reason string) error {
	m.deadLetters = append(m.deadLetters, reason)
	return nil
}

type fakeNotifications struct {
	port.NotificationUsecase
	err      error
	attempts []int
	finals   []bool
}

func (n *fakeNotifications) Dispatch(_ context.Context, _ *model.Notification, attempt int, final bool) error {
	n.attempts = append(n.attempts, attempt)
	n.finals = append(n.finals, final)
	return n.err
}

func TestHandleNotificationRouting(t *testing.T) {
	body := []byte(`{"id":1,"type":"push","user_id":7}`)
	maxAttempts := len(application.NotificationRetryDelays) + 1

	cases := []struct {
		name        string
		err         error
		attempt     int
		wantRetries []int
		wantDLQ     int
	}{
		{"success acks", nil, 1, nil, 0},
		{"permanent acks without retry", fmt.Errorf("%w: no device", common.ErrPermanent), 1, nil, 0},
		{"transient schedules next attempt", errors.New("timeout"), 2, []int{3}, 0},
		{"transient on last attempt goes to dlq", errors.New("timeout"), maxAttempts, nil, 1},
	}
	for _, tc := range cases {
		mq, n := &fakeMQ{}, &fakeNotifications{err: tc.err}
		c := &Consumer{mq: mq, notification: n}
		if err := c.HandleNotification(context.Background(), rabbitmq.Delivery{Body: body, Attempt: tc.attempt}); err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if fmt.Sprint(mq.retries) != fmt.Sprint(tc.wantRetries) || len(mq.deadLetters) != tc.wantDLQ {
			t.Errorf("%s: retries=%v dlq=%v", tc.name, mq.retries, mq.deadLetters)
		}
		if n.finals[0] != (tc.attempt >= maxAttempts) {
			t.Errorf("%s: final flag %v", tc.name, n.finals[0])
		}
	}
}

func TestHandleNotificationBadJSONGoesToDLQ(t *testing.T) {
	mq, n := &fakeMQ{}, &fakeNotifications{}
	c := &Consumer{mq: mq, notification: n}
	if err := c.HandleNotification(context.Background(), rabbitmq.Delivery{Body: []byte("{"), Attempt: 1}); err != nil {
		t.Fatal(err)
	}
	if len(mq.deadLetters) != 1 || len(n.attempts) != 0 {
		t.Errorf("dlq=%v dispatched=%v", mq.deadLetters, n.attempts)
	}
}

func TestHandleNotificationRequeuesWhenRetryPublishFails(t *testing.T) {
	mq, n := &fakeMQ{publishErr: errors.New("broker down")}, &fakeNotifications{err: errors.New("timeout")}
	c := &Consumer{mq: mq, notification: n}
	if err := c.HandleNotification(context.Background(), rabbitmq.Delivery{Body: []byte(`{"id":1}`), Attempt: 1}); err == nil {
		t.Error("message would be acked and lost when the retry could not be published")
	}
}
