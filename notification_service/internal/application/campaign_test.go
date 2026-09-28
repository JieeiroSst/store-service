package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

var ctx = context.Background()

func pushNotification(f *fixture, userID uint) *model.Notification {
	n := &model.Notification{UserID: userID, Type: "push", Title: "t", Message: "m"}
	_ = notificationRepo{f.s}.Create(ctx, n)
	return n
}

func TestDispatchNoDeviceIsPermanent(t *testing.T) {
	f := newFixture()
	n := pushNotification(f, 7)
	err := f.notif.Dispatch(ctx, n, 1, false)
	if !errors.Is(err, common.ErrPermanent) || f.s.notifications[n.ID].Status != "failed" {
		t.Fatalf("err=%v status=%s", err, f.s.notifications[n.ID].Status)
	}
}

func TestDispatchTransientThenFinal(t *testing.T) {
	f := newFixture()
	f.s.addDevice(7, "tok-a", true)
	f.push.err = errors.New("connection reset")
	n := pushNotification(f, 7)

	err := f.notif.Dispatch(ctx, n, 2, false)
	if err == nil || errors.Is(err, common.ErrPermanent) || f.s.notifications[n.ID].Status != "retrying" || f.s.notifications[n.ID].RetryCount != 1 {
		t.Fatalf("transient: err=%v %+v", err, f.s.notifications[n.ID])
	}
	_ = f.notif.Dispatch(ctx, n, 5, true)
	if f.s.notifications[n.ID].Status != "failed" {
		t.Errorf("final attempt should be failed, got %s", f.s.notifications[n.ID].Status)
	}
	f.push.err = nil
	if err := f.notif.Dispatch(ctx, n, 3, false); err != nil || f.s.notifications[n.ID].Status != "sent" || f.s.notifications[n.ID].LastError != "" {
		t.Errorf("recovery: %v %+v", err, f.s.notifications[n.ID])
	}
}

func TestDispatchDeactivatesUnregisteredTokens(t *testing.T) {
	f := newFixture()
	dead := f.s.addDevice(7, "dead", true)
	f.s.addDevice(7, "live", true)
	f.push.outcome = func(tk string) port.TokenResult {
		if tk == "dead" {
			return port.TokenResult{Unregistered: true, Error: "not registered"}
		}
		return port.TokenResult{Success: true}
	}
	if err := f.notif.Dispatch(ctx, pushNotification(f, 7), 1, false); err != nil {
		t.Fatal(err)
	}
	if f.s.devices[dead.ID].IsActive || len(f.push.calls) != 1 || len(f.push.calls[0]) != 2 {
		t.Errorf("dead active=%v calls=%v", f.s.devices[dead.ID].IsActive, f.push.calls)
	}

	f.push.outcome = func(string) port.TokenResult { return port.TokenResult{Unregistered: true} }
	if err := f.notif.Dispatch(ctx, pushNotification(f, 7), 1, false); !errors.Is(err, common.ErrPermanent) {
		t.Errorf("all tokens dead should be permanent, got %v", err)
	}
}

func TestCreateCampaignValidation(t *testing.T) {
	f := newFixture()
	bad := []port.CreateCampaignInput{
		{Channel: "sms", Audience: model.AudienceAllDevices, Title: "a", Message: "b"},
		{Channel: model.CampaignChannelPush, Audience: model.AudienceEmails, Title: "a", Message: "b"},
		{Channel: model.CampaignChannelPush, Audience: model.AudienceUsers, Title: "a", Message: "b"},
		{Channel: model.CampaignChannelPush, Audience: model.AudienceUsers, Title: "a", Message: "b", UserIDs: []uint{0}},
		{Channel: model.CampaignChannelPush, Audience: model.AudienceTopic, Title: "a", Message: "b", Topic: "bad topic"},
		{Channel: model.CampaignChannelEmail, Audience: model.AudienceEmails, Title: "a", Message: "b", Emails: []string{"nope"}},
		{Channel: model.CampaignChannelEmail, Audience: model.AudienceEmails, TemplateType: "missing", Emails: []string{"a@b.vn"}},
		{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Message: "b"},
	}
	for i, in := range bad {
		if _, err := f.campaigns.Create(ctx, in); !errors.Is(err, common.ErrInvalidRequest) {
			t.Errorf("case %d: got %v", i, err)
		}
	}
}

func TestCreateCampaignDedupesRecipients(t *testing.T) {
	f := newFixture()
	c, err := f.campaigns.Create(ctx, port.CreateCampaignInput{
		Channel: model.CampaignChannelEmail, Audience: model.AudienceEmails, Title: "Khuyến mãi", Message: "<p>x</p>",
		Emails: []string{"A@Shop.vn", "a@shop.vn", "Tên <b@shop.vn>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if c.Status != model.CampaignPending || c.Recipients != 2 {
		t.Errorf("campaign = %+v", c)
	}
}

func TestPushCampaignToAllDevicesInBatches(t *testing.T) {
	f := newFixture()
	inactive := 0
	for i := 0; i < 1203; i++ {
		on := i%97 != 0
		if !on {
			inactive++
		}
		f.s.addDevice(uint(i/2+1), fmt.Sprintf("tok-%d", i), on)
	}
	f.push.outcome = func(tk string) port.TokenResult {
		if tk == "tok-5" || tk == "tok-15" {
			return port.TokenResult{Unregistered: true}
		}
		if tk == "tok-7" {
			return port.TokenResult{Error: "invalid argument"}
		}
		return port.TokenResult{Success: true}
	}
	c, err := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "Flash sale", Message: "Giảm 50%"})
	if err != nil {
		t.Fatal(err)
	}
	f.drain(t)

	view, _ := f.campaigns.Get(ctx, c.ID)
	active := 1203 - inactive
	p := view.Progress
	if view.Campaign.Status != model.CampaignCompleted || view.Campaign.BatchesPlanned != 3 || p.BatchesSent != 3 {
		t.Fatalf("campaign=%+v progress=%+v", view.Campaign, p)
	}
	if p.Targets != active || p.Sent != active-3 || p.Failed != 1 || p.InvalidTokens != 2 {
		t.Errorf("progress = %+v, active=%d", p, active)
	}
	for _, call := range f.push.calls {
		if len(call) > 500 {
			t.Errorf("batch of %d tokens exceeds FCM limit", len(call))
		}
	}
	for _, d := range f.s.devices {
		if (d.DeviceToken == "tok-5" || d.DeviceToken == "tok-15") && d.IsActive {
			t.Errorf("unregistered token %s still active", d.DeviceToken)
		}
	}
}

func TestPushCampaignToSelectedUsers(t *testing.T) {
	f := newFixture()
	f.s.addDevice(1, "u1-phone", true)
	f.s.addDevice(1, "u1-tablet", true)
	f.s.addDevice(2, "u2-phone", true)
	f.s.addDevice(3, "u3-phone", true)
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceUsers, Title: "a", Message: "b", UserIDs: []uint{1, 3, 3, 99}})
	f.drain(t)
	view, _ := f.campaigns.Get(ctx, c.ID)
	if view.Campaign.Recipients != 3 || view.Progress.Sent != 3 || len(f.push.calls) != 1 {
		t.Errorf("recipients=%d progress=%+v calls=%v", view.Campaign.Recipients, view.Progress, f.push.calls)
	}
}

func TestBatchRetriesThenFails(t *testing.T) {
	f := newFixture()
	f.s.addDevice(1, "tok", true)
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "a", Message: "b"})
	_, _ = f.runner.PlanNext(ctx)
	batchID := f.pub.published[0]

	f.push.outcome = func(string) port.TokenResult { return port.TokenResult{Retryable: true, Error: "unavailable"} }
	for attempt := 1; attempt <= 2; attempt++ {
		out, err := f.runner.SendBatch(ctx, batchID)
		if err != nil || !out.Retry || out.Attempt != attempt+1 || f.s.batches[batchID].State != model.BatchQueued {
			t.Fatalf("attempt %d: %+v %v %s", attempt, out, err, f.s.batches[batchID].State)
		}
	}
	out, _ := f.runner.SendBatch(ctx, batchID)
	view, _ := f.campaigns.Get(ctx, c.ID)
	if out.Retry || f.s.batches[batchID].State != model.BatchFailed || view.Campaign.Status != model.CampaignCompleted || view.Progress.Failed != 1 {
		t.Errorf("final: %+v batch=%+v campaign=%s", out, f.s.batches[batchID], view.Campaign.Status)
	}
}

func TestBatchIsClaimedOnce(t *testing.T) {
	f := newFixture()
	f.s.addDevice(1, "tok", true)
	_, _ = f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "a", Message: "b"})
	_, _ = f.runner.PlanNext(ctx)
	id := f.pub.published[0]
	_, _ = f.runner.SendBatch(ctx, id)
	_, _ = f.runner.SendBatch(ctx, id)
	if len(f.push.calls) != 1 {
		t.Errorf("duplicate delivery of the same batch: %d sends", len(f.push.calls))
	}
}

func TestLostPublishIsRepublishedBySweeper(t *testing.T) {
	f := newFixture()
	f.s.addDevice(1, "tok", true)
	_, _ = f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "a", Message: "b"})
	f.pub.err = errors.New("broker down")
	if _, err := f.runner.PlanNext(ctx); err != nil {
		t.Fatal(err)
	}
	f.pub.err = nil
	if n, _ := f.runner.RepublishStale(ctx); n != 0 {
		t.Errorf("republished before grace period: %d", n)
	}
	f.clock.now = f.clock.now.Add(f.settings.RepublishGrace + time.Second)
	if n, _ := f.runner.RepublishStale(ctx); n != 1 || len(f.pub.published) != 1 {
		t.Errorf("sweeper republished %d", n)
	}
}

func TestCancelStopsRemainingBatches(t *testing.T) {
	f := newFixture()
	for i := 0; i < 1200; i++ {
		f.s.addDevice(1, fmt.Sprint(i), true)
	}
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "a", Message: "b"})
	_, _ = f.runner.PlanNext(ctx)
	_, _ = f.runner.SendBatch(ctx, f.pub.published[0])
	if _, err := f.campaigns.Cancel(ctx, c.ID); err != nil {
		t.Fatal(err)
	}
	for _, id := range f.pub.published[1:] {
		_, _ = f.runner.SendBatch(ctx, id)
	}
	view, _ := f.campaigns.Get(ctx, c.ID)
	if view.Campaign.Status != model.CampaignCancelled || view.Progress.BatchesSent != 1 || view.Progress.BatchesCancelled != 2 || len(f.push.calls) != 1 {
		t.Errorf("status=%s progress=%+v sends=%d", view.Campaign.Status, view.Progress, len(f.push.calls))
	}
	if _, err := f.campaigns.Cancel(ctx, c.ID); !errors.Is(err, common.ErrInvalidRequest) {
		t.Errorf("second cancel: %v", err)
	}
}

func TestEmailCampaignUsesBatchesWithIdempotencyKey(t *testing.T) {
	f := newFixture()
	emails := make([]string, 250)
	for i := range emails {
		emails[i] = fmt.Sprintf("user%d@shop.vn", i)
	}
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelEmail, Audience: model.AudienceEmails, TemplateType: "welcome", TemplateData: map[string]string{"name": "bạn"}, Emails: emails})
	f.drain(t)
	view, _ := f.campaigns.Get(ctx, c.ID)
	if len(f.email.batches) != 3 || len(f.email.batches[0]) != 100 || view.Progress.Sent != 250 || view.Campaign.Status != model.CampaignCompleted {
		t.Fatalf("batches=%d progress=%+v", len(f.email.batches), view.Progress)
	}
	if f.email.batches[0][0].Subject != "Chào bạn" || !strings.HasPrefix(f.email.keys[0], fmt.Sprintf("campaign-%d-batch-", c.ID)) {
		t.Errorf("subject=%q key=%q", f.email.batches[0][0].Subject, f.email.keys[0])
	}
}

func TestEmailPermanentFailureMarksBatchFailed(t *testing.T) {
	f := newFixture()
	f.email.err = fmt.Errorf("%w: 422 invalid from", common.ErrPermanent)
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelEmail, Audience: model.AudienceEmails, Title: "a", Message: "b", Emails: []string{"a@b.vn"}})
	f.drain(t)
	view, _ := f.campaigns.Get(ctx, c.ID)
	if view.Progress.BatchesFailed != 1 || view.Progress.Failed != 1 || len(f.email.batches) != 1 {
		t.Errorf("progress=%+v calls=%d", view.Progress, len(f.email.batches))
	}
}

func TestTopicCampaign(t *testing.T) {
	f := newFixture()
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceTopic, Topic: "all-users", Title: "a", Message: "b"})
	f.push.topicErr = errors.New("unavailable")
	_, _ = f.runner.PlanNext(ctx)
	if got, _ := f.campaigns.Get(ctx, c.ID); got.Campaign.Status != model.CampaignPlanning || got.Campaign.PlannerLeaseUntil == nil {
		t.Fatalf("after transient failure: %+v", got.Campaign)
	}
	if worked, _ := f.runner.PlanNext(ctx); worked {
		t.Error("retried before backoff elapsed")
	}
	f.push.topicErr = nil
	f.clock.now = f.clock.now.Add(time.Hour)
	_, _ = f.runner.PlanNext(ctx)
	got, _ := f.campaigns.Get(ctx, c.ID)
	if got.Campaign.Status != model.CampaignCompleted || len(f.push.topics) != 2 {
		t.Errorf("status=%s topic sends=%d", got.Campaign.Status, len(f.push.topics))
	}
}
