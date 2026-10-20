package messaging

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/point-service/config"
	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/domain/service"
)

type stubLedger struct {
	service.PointLedgerService
	applied bool
	err     error
	got     dto.RewardEventDTO
}

func (s *stubLedger) Apply(_ context.Context, ev dto.RewardEventDTO) (bool, error) {
	s.got = ev
	return s.applied, s.err
}

func TestHandle(t *testing.T) {
	body := []byte(`{"event_id":"lucky-play-10","source":"lucky-wheel","game_type":"wheel","user_id":"7","campaign_id":1,"reference_id":10,"title":"500 diem","points":500,"occurred_at":"2026-10-20T09:19:03Z"}`)
	cases := []struct {
		name   string
		body   []byte
		ledger *stubLedger
		want   Outcome
	}{
		{"credited", body, &stubLedger{applied: true}, Ack},
		{"duplicate", body, &stubLedger{applied: false}, Ack},
		{"garbage", []byte("{"), &stubLedger{}, Drop},
		{"invalid", body, &stubLedger{err: service.ErrInvalidRewardEvent}, Drop},
		{"database down", body, &stubLedger{err: errors.New("dial tcp: refused")}, Retry},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := NewRewardConsumer(config.RabbitMQConfig{}, tc.ledger)
			if got := c.Handle(context.Background(), tc.body); got != tc.want {
				t.Fatalf("outcome = %v, want %v", got, tc.want)
			}
		})
	}
	s := &stubLedger{applied: true}
	NewRewardConsumer(config.RabbitMQConfig{}, s).Handle(context.Background(), body)
	if s.got.EventID != "lucky-play-10" || s.got.UserID != "7" || s.got.Points != 500 || s.got.Title != "500 diem" || s.got.OccurredAt.IsZero() {
		t.Fatalf("lucky-service event decoded: %+v", s.got)
	}
}

func TestConfigDefaults(t *testing.T) {
	t.Setenv("RABBITMQ_ENABLED", "true")
	t.Setenv("RABBITMQ_PASSWORD", "secret")
	c := config.RabbitMQConfig{Host: "rabbitmq-service", User: "admin"}
	c.ApplyEnv()
	if !c.Enabled || c.Password != "secret" || c.Port != "5672" || c.Queue != "point-service.rewards" || c.Bindings[0] != "lucky:lucky.reward.granted" || c.Prefetch != 20 {
		t.Fatalf("defaults: %+v", c)
	}
	if u := NewRewardConsumer(c, nil).url(); u != "amqp://admin:secret@rabbitmq-service:5672/" {
		t.Fatalf("url: %s", u)
	}
}
