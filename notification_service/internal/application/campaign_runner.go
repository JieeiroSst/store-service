package application

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
	"golang.org/x/time/rate"
)

type campaignRunner struct {
	campaigns   port.CampaignRepository
	batches     port.CampaignBatchRepository
	targets     port.CampaignTargetSource
	publisher   port.BatchPublisher
	push        port.MulticastSender
	email       port.BatchEmailSender
	template    port.TemplateRenderer
	deactivator port.DeviceDeactivator
	clock       port.Clock
	limiter     *rate.Limiter
	settings    CampaignSettings
	audit       *auditor
}

func NewCampaignRunner(
	campaigns port.CampaignRepository,
	batches port.CampaignBatchRepository,
	targets port.CampaignTargetSource,
	publisher port.BatchPublisher,
	push port.MulticastSender,
	email port.BatchEmailSender,
	template port.TemplateRenderer,
	deactivator port.DeviceDeactivator,
	clock port.Clock,
	settings CampaignSettings,
	audit *auditor,
) port.CampaignRunner {
	r := &campaignRunner{
		campaigns: campaigns, batches: batches, targets: targets, publisher: publisher,
		push: push, email: email, template: template, deactivator: deactivator, clock: clock, settings: settings,
		audit: audit,
	}
	if settings.RatePerSecond > 0 {
		r.limiter = rate.NewLimiter(rate.Limit(settings.RatePerSecond), max(settings.PushBatchSize, settings.EmailBatchSize))
	}
	return r
}

var active = []model.CampaignStatus{model.CampaignPlanning}

func (r *campaignRunner) batchSize(c *model.Campaign) int {
	if c.Channel == model.CampaignChannelEmail {
		return min(r.settings.EmailBatchSize, r.email.MaxBatchSize())
	}
	return r.settings.PushBatchSize
}

func (r *campaignRunner) PlanNext(ctx context.Context) (bool, error) {
	now := r.clock.Now()
	c, err := r.campaigns.ClaimForPlanning(ctx, now, now.Add(r.settings.PlannerLease))
	if err != nil || c == nil {
		return false, err
	}
	if c.Audience == model.AudienceTopic {
		return true, r.sendTopic(ctx, c)
	}
	defer func() { _ = r.campaigns.ReleasePlanning(context.WithoutCancel(ctx), c.ID) }()

	size := r.batchSize(c)
	for i := 0; i < r.settings.BatchesPerPlan; i++ {
		if i > 0 && i%20 == 0 {
			fresh, err := r.campaigns.GetByID(ctx, c.ID)
			if err != nil {
				return true, err
			}
			if fresh.Status != model.CampaignPlanning {
				return true, nil
			}
		}

		ids, err := r.targets.NextIDs(ctx, c, c.Cursor, size)
		if err != nil {
			return true, err
		}
		if len(ids) == 0 {
			return true, r.finishPlanning(ctx, c)
		}

		b := &model.CampaignBatch{
			CampaignID:     c.ID,
			Seq:            c.BatchesPlanned + 1,
			AfterID:        c.Cursor,
			UntilID:        ids[len(ids)-1],
			State:          model.BatchQueued,
			Targets:        len(ids),
			RepublishAfter: r.clock.Now().Add(r.settings.RepublishGrace),
		}
		if err := r.batches.CreateAndAdvance(ctx, b, b.UntilID); err != nil {
			return true, err
		}
		c.Cursor, c.BatchesPlanned = b.UntilID, c.BatchesPlanned+1
		if err := r.publisher.PublishBatch(ctx, b.ID); err != nil {
			log.Printf("campaign %d: publish batch %d failed, sweeper will retry: %v", c.ID, b.ID, err)
		}
	}
	return true, nil
}

func (r *campaignRunner) finishPlanning(ctx context.Context, c *model.Campaign) error {
	now := r.clock.Now()
	if c.BatchesPlanned == 0 {
		_, err := r.campaigns.SetStatus(ctx, c.ID, active, model.CampaignCompleted, map[string]any{"completed_at": now})
		return err
	}
	if _, err := r.campaigns.SetStatus(ctx, c.ID, active, model.CampaignSending, nil); err != nil {
		return err
	}
	return r.campaigns.CompleteIfDrained(ctx, c.ID, now)
}

func (r *campaignRunner) sendTopic(ctx context.Context, c *model.Campaign) error {
	messageID, err := r.push.SendToTopic(ctx, c.Topic, c.Title, c.Message, c.Data)
	t := newTrail(model.AuditSourceCampaign, c.ID, 0, c.PlannerAttempts+1, string(c.Channel), c.RequestedBy)
	t.setContent(c.Title, c.Message, c.Data, "")
	entry := model.AuditDelivery{Recipient: "topic:" + c.Topic, Status: model.AuditSent, MessageID: messageID}
	if err != nil {
		entry.Status, entry.Error = model.AuditFailed, err.Error()
	}
	t.add(entry)
	r.audit.record(ctx, t)
	now := r.clock.Now()
	if err == nil {
		_, err = r.campaigns.SetStatus(ctx, c.ID, active, model.CampaignCompleted, map[string]any{"completed_at": now, "last_error": ""})
		return err
	}
	attempts := c.PlannerAttempts + 1
	fields := map[string]any{"planner_attempts": attempts, "last_error": err.Error(), "planner_lease_until": now.Add(retryDelay(BatchRetryDelays, attempts))}
	if IsPermanent(err) || attempts >= r.settings.MaxAttempts {
		fields["completed_at"] = now
		_, setErr := r.campaigns.SetStatus(ctx, c.ID, active, model.CampaignFailed, fields)
		return setErr
	}
	_, setErr := r.campaigns.SetStatus(ctx, c.ID, active, model.CampaignPlanning, fields)
	return setErr
}

func (r *campaignRunner) RepublishStale(ctx context.Context) (int, error) {
	now := r.clock.Now()
	due, err := r.batches.DueForRepublish(ctx, now, r.settings.RepublishPerTick)
	if err != nil {
		return 0, err
	}
	for _, b := range due {
		if err := r.batches.TouchRepublish(ctx, b.ID, now.Add(r.settings.RepublishGrace)); err != nil {
			return 0, err
		}
		if err := r.publisher.PublishBatch(ctx, b.ID); err != nil {
			return 0, err
		}
	}
	return len(due), nil
}

func (r *campaignRunner) SendBatch(ctx context.Context, batchID uint) (port.BatchOutcome, error) {
	now := r.clock.Now()
	b, claimed, err := r.batches.Claim(ctx, batchID, now, now.Add(r.settings.BatchLease))
	if err != nil || !claimed {
		return port.BatchOutcome{}, err
	}
	c, err := r.campaigns.GetByID(ctx, b.CampaignID)
	if err != nil {
		return port.BatchOutcome{}, err
	}
	if c.Status == model.CampaignCancelled {
		b.State = model.BatchCancelled
		return port.BatchOutcome{}, r.batches.Finish(ctx, b)
	}

	targets, err := r.targets.Targets(ctx, c, b.AfterID, b.UntilID)
	if err != nil {
		return port.BatchOutcome{}, err
	}
	if r.limiter != nil && len(targets) > 0 {
		if err := r.limiter.WaitN(ctx, min(len(targets), r.limiter.Burst())); err != nil {
			return port.BatchOutcome{}, err
		}
	}

	t := newTrail(model.AuditSourceCampaign, c.ID, b.ID, b.Attempts, string(c.Channel), c.RequestedBy)
	var sendErr error
	if c.Channel == model.CampaignChannelEmail {
		sendErr = r.sendEmails(ctx, c, b, targets, t)
	} else {
		sendErr = r.sendPush(ctx, c, b, targets, t)
	}
	if sendErr != nil {
		t.failAll(sendErr)
	}
	r.audit.record(ctx, t)
	if sendErr == nil || IsPermanent(sendErr) {
		b.State = model.BatchSent
		if sendErr != nil {
			b.State, b.Failed, b.LastError = model.BatchFailed, len(targets)-b.Sent-b.Invalid, sendErr.Error()
		}
		return port.BatchOutcome{}, r.finish(ctx, b)
	}

	if b.Attempts >= r.settings.MaxAttempts {
		b.State, b.Sent, b.Invalid, b.Failed, b.LastError = model.BatchFailed, 0, 0, len(targets), sendErr.Error()
		return port.BatchOutcome{}, r.finish(ctx, b)
	}
	republishAfter := r.clock.Now().Add(retryDelay(BatchRetryDelays, b.Attempts) + r.settings.RepublishGrace)
	if err := r.batches.Requeue(ctx, b.ID, republishAfter, sendErr.Error()); err != nil {
		return port.BatchOutcome{}, err
	}
	return port.BatchOutcome{Retry: true, Attempt: b.Attempts + 1}, nil
}

func (r *campaignRunner) finish(ctx context.Context, b *model.CampaignBatch) error {
	if err := r.batches.Finish(ctx, b); err != nil {
		return err
	}
	return r.campaigns.CompleteIfDrained(ctx, b.CampaignID, r.clock.Now())
}

var errAllRetryable = errors.New("every token failed with a retryable error")

func (r *campaignRunner) sendPush(ctx context.Context, c *model.Campaign, b *model.CampaignBatch, targets []model.BatchTarget, t *trail) error {
	b.Sent, b.Failed, b.Invalid = 0, 0, 0
	if len(targets) == 0 {
		return nil
	}
	tokens := make([]string, len(targets))
	for i, t := range targets {
		tokens[i] = t.Token
	}
	t.setContent(c.Title, c.Message, c.Data, "")
	results, err := r.push.SendToTokens(ctx, tokens, c.Title, c.Message, c.Data)
	for i, tg := range targets {
		entry := model.AuditDelivery{UserID: tg.UserID, DeviceID: tg.ID, DeviceType: tg.DeviceType, TokenHash: model.HashToken(tg.Token)}
		if err == nil && i < len(results) {
			entry.Status, entry.MessageID, entry.Error = pushStatus(results[i]), results[i].MessageID, results[i].Error
		}
		t.add(entry)
	}
	if err != nil {
		return err
	}

	var unregistered []uint
	retryable := 0
	for i, res := range results {
		switch {
		case res.Success:
			b.Sent++
		case res.Unregistered:
			b.Invalid++
			unregistered = append(unregistered, targets[i].ID)
		default:
			b.Failed++
			if res.Retryable {
				retryable++
			}
			b.LastError = res.Error
		}
	}
	if err := r.deactivator.DeactivateByIDs(ctx, unregistered); err != nil {
		return err
	}
	if b.Sent == 0 && b.Failed > 0 && retryable == b.Failed {
		return fmt.Errorf("%w: %s", errAllRetryable, b.LastError)
	}
	return nil
}

func (r *campaignRunner) sendEmails(ctx context.Context, c *model.Campaign, b *model.CampaignBatch, targets []model.BatchTarget, t *trail) error {
	b.Sent, b.Failed, b.Invalid = 0, 0, 0
	if len(targets) == 0 {
		return nil
	}
	subject, html := c.Title, c.Message
	if c.TemplateType != "" {
		var err error
		if subject, html, err = r.template.Render(c.TemplateType, c.TemplateData); err != nil {
			return fmt.Errorf("%w: %v", common.ErrPermanent, err)
		}
	}
	t.setContent(subject, html, nil, c.TemplateType)
	messages := make([]port.EmailMessage, len(targets))
	for i, tg := range targets {
		messages[i] = port.EmailMessage{To: tg.Email, Subject: subject, HTML: html}
		t.add(model.AuditDelivery{Recipient: tg.Email})
	}
	if err := r.email.SendBatch(ctx, messages, fmt.Sprintf("campaign-%d-batch-%d", c.ID, b.ID)); err != nil {
		return err
	}
	for i := range t.deliveries {
		t.deliveries[i].Status = model.AuditSent
	}
	b.Sent = len(targets)
	return nil
}
