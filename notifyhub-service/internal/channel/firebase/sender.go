package firebase

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	fb "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/JIeeiroSst/notifyhub-service/internal/channel"
	"github.com/JIeeiroSst/notifyhub-service/internal/config"
	"github.com/JIeeiroSst/notifyhub-service/internal/model"
	"go.uber.org/zap"
	"google.golang.org/api/option"
)

const (
	fcmMaxMulticastTokens = 500

	defaultTTL = 24 * time.Hour
)

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityNormal Priority = "normal"
)

type PushMessage struct {
	channel.Message

	ImageURL string

	Topic string

	Condition string

	Data map[string]string

	Priority Priority

	CollapseKey string

	Sound string

	Badge *int

	ClickAction string

	TTL time.Duration

	DryRun bool
}

type BatchResult struct {
	SuccessCount int
	FailureCount int
	Responses    []*messaging.SendResponse
}

type Sender struct {
	mu     sync.Mutex
	client *messaging.Client
	logger *zap.Logger
}

func New(ctx context.Context, cfg config.FirebaseConfig, log *zap.Logger) (*Sender, error) {
	if log == nil {
		log = zap.NewNop()
	}
	if cfg.CredentialsFile == "" {
		return nil, fmt.Errorf("FIREBASE_CREDENTIALS_FILE is not set")
	}

	opt := option.WithCredentialsFile(cfg.CredentialsFile)
	app, err := fb.NewApp(ctx, &fb.Config{ProjectID: cfg.ProjectID}, opt)
	if err != nil {
		return nil, fmt.Errorf("firebase app init: %w", err)
	}

	mc, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("firebase messaging client: %w", err)
	}

	log.Info("firebase messaging client initialised",
		zap.String("project_id", cfg.ProjectID),
	)
	return &Sender{client: mc, logger: log}, nil
}

func (s *Sender) Type() model.ChannelType { return model.ChannelFirebase }

// Send dispatches a push notification to a single FCM registration token.
// msg.Recipient must be a valid FCM token.
// Extra options (topic, image, data, priority) can be set in msg.Data:
//
//	msg.Data["_topic"]        → send to topic instead of token
//	msg.Data["_condition"]    → send to condition
//	msg.Data["_image_url"]    → notification image
//	msg.Data["_priority"]     → "high" | "normal"
//	msg.Data["_collapse_key"] → collapse key
//	msg.Data["_sound"]        → sound file name
//	msg.Data["_ttl_sec"]      → TTL in seconds (integer string)
//	msg.Data["_dry_run"]      → "true" to validate only
//	All other keys are delivered as FCM data payload.
func (s *Sender) Send(ctx context.Context, msg channel.Message) error {
	pm := parsePushMessage(msg)
	_, err := s.sendOne(ctx, pm)
	return err
}

func (s *Sender) SendPush(ctx context.Context, pm PushMessage) (string, error) {
	return s.sendOne(ctx, pm)
}

func (s *Sender) sendOne(ctx context.Context, pm PushMessage) (string, error) {
	m, err := s.buildMessage(pm)
	if err != nil {
		return "", err
	}

	if pm.DryRun {
		msgID, err := s.client.SendDryRun(ctx, m)
		if err != nil {
			return "", fmt.Errorf("fcm dry-run: %w", err)
		}
		s.logger.Info("fcm dry-run success", zap.String("msg_id", msgID))
		return msgID, nil
	}

	msgID, err := s.client.Send(ctx, m)
	if err != nil {
		return "", fmt.Errorf("fcm send: %w", err)
	}

	s.logger.Info("fcm push sent",
		zap.String("msg_id", msgID),
		zap.String("to", targetDescription(pm)),
	)
	return msgID, nil
}


func (s *Sender) SendMulticast(ctx context.Context, tokens []string, pm PushMessage) (*BatchResult, error) {
	if len(tokens) == 0 {
		return &BatchResult{}, nil
	}
	if len(tokens) > fcmMaxMulticastTokens {
		tokens = tokens[:fcmMaxMulticastTokens]
	}

	mm := s.buildMulticastMessage(tokens, pm)
	br, err := s.client.SendEachForMulticast(ctx, mm)
	if err != nil {
		return nil, fmt.Errorf("fcm multicast: %w", err)
	}

	result := &BatchResult{
		SuccessCount: br.SuccessCount,
		FailureCount: br.FailureCount,
		Responses:    br.Responses,
	}

	s.logger.Info("fcm multicast sent",
		zap.Int("tokens", len(tokens)),
		zap.Int("success", br.SuccessCount),
		zap.Int("failure", br.FailureCount),
	)
	return result, nil
}

func (s *Sender) SendMulticastBatch(ctx context.Context, tokens []string, pm PushMessage) (*BatchResult, error) {
	if len(tokens) == 0 {
		return &BatchResult{}, nil
	}

	type chunk struct {
		tokens []string
	}

	var chunks []chunk
	for i := 0; i < len(tokens); i += fcmMaxMulticastTokens {
		end := i + fcmMaxMulticastTokens
		if end > len(tokens) {
			end = len(tokens)
		}
		chunks = append(chunks, chunk{tokens: tokens[i:end]})
	}

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		total    BatchResult
		firstErr error
	)

	for _, ch := range chunks {
		wg.Add(1)
		go func(ch chunk) {
			defer wg.Done()
			br, err := s.SendMulticast(ctx, ch.tokens, pm)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
				return
			}
			if br != nil {
				total.SuccessCount += br.SuccessCount
				total.FailureCount += br.FailureCount
				total.Responses = append(total.Responses, br.Responses...)
			}
		}(ch)
	}
	wg.Wait()

	if firstErr != nil {
		return &total, firstErr
	}
	return &total, nil
}

func (s *Sender) SendToTopic(ctx context.Context, topic string, pm PushMessage) (string, error) {
	pm.Topic = topic
	return s.sendOne(ctx, pm)
}

func (s *Sender) SendToCondition(ctx context.Context, condition string, pm PushMessage) (string, error) {
	pm.Condition = condition
	return s.sendOne(ctx, pm)
}

func (s *Sender) SubscribeToTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := s.client.SubscribeToTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("subscribe to topic %q: %w", topic, err)
	}
	s.logger.Info("subscribed tokens to topic",
		zap.Int("count", len(tokens)),
		zap.String("topic", topic),
	)
	return nil
}

func (s *Sender) UnsubscribeFromTopic(ctx context.Context, tokens []string, topic string) error {
	_, err := s.client.UnsubscribeFromTopic(ctx, tokens, topic)
	if err != nil {
		return fmt.Errorf("unsubscribe from topic %q: %w", topic, err)
	}
	return nil
}

func (s *Sender) buildMessage(pm PushMessage) (*messaging.Message, error) {
	if pm.Topic == "" && pm.Condition == "" && pm.Recipient == "" {
		return nil, fmt.Errorf("fcm: one of Recipient (token), Topic, or Condition must be set")
	}

	m := &messaging.Message{
		Notification: &messaging.Notification{
			Title:    pm.Subject,
			Body:     pm.Body,
			ImageURL: pm.ImageURL,
		},
		Data: cleanData(pm.Data),
	}

	switch {
	case pm.Condition != "":
		m.Condition = pm.Condition
	case pm.Topic != "":
		m.Topic = pm.Topic
	default:
		m.Token = pm.Recipient
	}

	android := &messaging.AndroidConfig{
		Priority:    string(pm.Priority),
		CollapseKey: pm.CollapseKey,
	}
	if pm.TTL > 0 {
		ttl := pm.TTL
		android.TTL = &ttl
	} else {
		ttl := defaultTTL
		android.TTL = &ttl
	}
	if pm.Subject != "" || pm.Body != "" {
		android.Notification = &messaging.AndroidNotification{
			Title:       pm.Subject,
			Body:        pm.Body,
			Sound:       pm.Sound,
			ClickAction: pm.ClickAction,
			ImageURL:    pm.ImageURL,
		}
	}
	m.Android = android

	apns := &messaging.APNSConfig{
		Headers: map[string]string{},
		Payload: &messaging.APNSPayload{
			Aps: &messaging.Aps{
				Alert: &messaging.ApsAlert{
					Title: pm.Subject,
					Body:  pm.Body,
				},
				Sound: pm.Sound,
			},
		},
	}
	if pm.Badge != nil {
		badge := *pm.Badge
		apns.Payload.Aps.Badge = &badge
	}
	if pm.Priority == PriorityHigh {
		apns.Headers["apns-priority"] = "10"
	} else {
		apns.Headers["apns-priority"] = "5"
	}
	m.APNS = apns

	m.Webpush = &messaging.WebpushConfig{
		Notification: &messaging.WebpushNotification{
			Title: pm.Subject,
			Body:  pm.Body,
			Icon:  pm.ImageURL,
		},
	}

	return m, nil
}

func (s *Sender) buildMulticastMessage(tokens []string, pm PushMessage) *messaging.MulticastMessage {
	ttl := pm.TTL
	if ttl == 0 {
		ttl = defaultTTL
	}

	return &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    pm.Subject,
			Body:     pm.Body,
			ImageURL: pm.ImageURL,
		},
		Data: cleanData(pm.Data),
		Android: &messaging.AndroidConfig{
			Priority:    string(pm.Priority),
			CollapseKey: pm.CollapseKey,
			TTL:         &ttl,
			Notification: &messaging.AndroidNotification{
				Title:       pm.Subject,
				Body:        pm.Body,
				Sound:       pm.Sound,
				ClickAction: pm.ClickAction,
				ImageURL:    pm.ImageURL,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: pm.Subject,
						Body:  pm.Body,
					},
					Sound: pm.Sound,
				},
			},
		},
	}
}

func parsePushMessage(msg channel.Message) PushMessage {
	pm := PushMessage{
		Message:  msg,
		Priority: PriorityHigh,
		TTL:      defaultTTL,
	}
	if msg.Data == nil {
		return pm
	}

	// Extract control keys
	if v, ok := msg.Data["_topic"]; ok && v != "" {
		pm.Topic = v
	}
	if v, ok := msg.Data["_condition"]; ok && v != "" {
		pm.Condition = v
	}
	if v, ok := msg.Data["_image_url"]; ok {
		pm.ImageURL = v
	}
	if v, ok := msg.Data["_priority"]; ok {
		pm.Priority = Priority(v)
	}
	if v, ok := msg.Data["_collapse_key"]; ok {
		pm.CollapseKey = v
	}
	if v, ok := msg.Data["_sound"]; ok {
		pm.Sound = v
	}
	if v, ok := msg.Data["_click_action"]; ok {
		pm.ClickAction = v
	}
	if v, ok := msg.Data["_dry_run"]; ok && v == "true" {
		pm.DryRun = true
	}

	pm.Data = make(map[string]string)
	for k, v := range msg.Data {
		if !strings.HasPrefix(k, "_") {
			pm.Data[k] = v
		}
	}
	return pm
}

func cleanData(data map[string]string) map[string]string {
	if len(data) == 0 {
		return nil
	}
	out := make(map[string]string, len(data))
	for k, v := range data {
		if !strings.HasPrefix(k, "_") {
			out[k] = v
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func targetDescription(pm PushMessage) string {
	switch {
	case pm.Condition != "":
		return "condition: " + pm.Condition
	case pm.Topic != "":
		return "topic: " + pm.Topic
	default:
		if len(pm.Recipient) > 20 {
			return pm.Recipient[:20] + "…"
		}
		return pm.Recipient
	}
}
