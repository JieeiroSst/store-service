package application

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

func statuses(ds []model.AuditDelivery) map[string]int {
	out := map[string]int{}
	for _, d := range ds {
		out[d.Status]++
	}
	return out
}

func TestAuditRecordsEveryPushDeliveryWithRequester(t *testing.T) {
	f := newFixture()
	devices := f.devices(DevicePolicy{})
	_, _ = devices.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 42, Token: fcmToken + "ios", DeviceType: "ios", Email: "an@shop.vn"})
	_, _ = devices.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 42, Token: fcmToken + "dead", DeviceType: "android"})
	f.push.outcome = func(tk string) port.TokenResult {
		if tk == fcmToken+"dead" {
			return port.TokenResult{Unregistered: true, Error: "not registered"}
		}
		return port.TokenResult{Success: true, MessageID: "projects/p/messages/1"}
	}

	n, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "push", Email: "an@shop.vn", Title: "Đơn DH-1", Message: "Đang giao", Data: map[string]string{"id": "1"}, RequestedBy: "order-service"})
	if err := f.notif.Dispatch(ctx, n, 1, false); err != nil {
		t.Fatal(err)
	}

	ds := f.audit.deliveries
	if len(ds) != 2 || statuses(ds)[model.AuditSent] != 1 || statuses(ds)[model.AuditInvalidToken] != 1 {
		t.Fatalf("deliveries = %+v", ds)
	}
	for _, d := range ds {
		if d.UserID != 42 || d.RequestedBy != "order-service" || d.SourceID != n.ID || d.TokenHash == "" || d.TokenHash == fcmToken+"ios" {
			t.Errorf("bad delivery row: %+v", d)
		}
	}
	if c := f.audit.contents[0]; c.Title != "Đơn DH-1" || c.Body != "Đang giao" || c.Data["id"] != "1" {
		t.Errorf("content = %+v", c)
	}
}

func TestAuditRecordsFailuresAndRenderedEmail(t *testing.T) {
	f := newFixture()
	n, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "push", UserID: 99, Title: "x"})
	_ = f.notif.Dispatch(ctx, n, 1, false)
	if d := f.audit.deliveries[0]; d.Status != model.AuditFailed || d.UserID != 99 || d.Error == "" || d.RequestedBy != model.DefaultRequester {
		t.Errorf("no-device attempt not audited: %+v", d)
	}

	e, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "email", Recipient: "Khach@Shop.vn", TemplateType: "welcome", TemplateData: map[string]string{"name": "Lan"}})
	_ = f.notif.Dispatch(ctx, e, 1, false)
	last := f.audit.deliveries[len(f.audit.deliveries)-1]
	content := f.audit.contents[last.ContentID-1]
	if last.Recipient != "khach@shop.vn" || last.Status != model.AuditSent || content.Title != "Chào Lan" || content.Body != "<p>xin chào</p>" {
		t.Errorf("email audit: %+v content=%+v", last, content)
	}
}

func TestAuditCampaignSharesOneContent(t *testing.T) {
	f := newFixture()
	for i := 0; i < 1100; i++ {
		f.s.addDevice(uint(i+1), fcmToken+string(rune('a'+i%26))+time.Duration(i).String(), true)
	}
	c, _ := f.campaigns.Create(ctx, port.CreateCampaignInput{Channel: model.CampaignChannelPush, Audience: model.AudienceAllDevices, Title: "Sale", Message: "50%", RequestedBy: "marketing"})
	f.drain(t)

	if len(f.audit.contents) != 1 || len(f.audit.deliveries) != 1100 {
		t.Fatalf("contents=%d deliveries=%d", len(f.audit.contents), len(f.audit.deliveries))
	}
	batches := map[uint]bool{}
	for _, d := range f.audit.deliveries {
		batches[d.BatchID] = true
		if d.SourceType != model.AuditSourceCampaign || d.SourceID != c.ID || d.RequestedBy != "marketing" || d.UserID == 0 {
			t.Fatalf("bad row %+v", d)
		}
	}
	if len(batches) != 3 {
		t.Errorf("rows span %d batches, want 3", len(batches))
	}
}

func TestAuditQueryByEmailPhoneAndRetention(t *testing.T) {
	f := newFixture()
	devices := f.devices(DevicePolicy{})
	_, _ = devices.RegisterDevice(ctx, port.RegisterDeviceInput{UserID: 42, Token: fcmToken, DeviceType: "ios", Email: "an@shop.vn", Phone: "0912345678"})
	for _, n := range []*model.Notification{
		{Type: "push", UserID: 42, Title: "a"},
		{Type: "email", Recipient: "an@shop.vn", Title: "b", Message: "c"},
		{Type: "push", UserID: 7, Title: "other user"},
	} {
		created, _ := f.notif.CreateNotification(ctx, n)
		_ = f.notif.Dispatch(ctx, created, 1, false)
	}

	cfg := &config.Config{Audit: config.AuditConfig{RetentionDays: 30}}
	svc := NewAuditService(f.audit, f.contacts, f.clock, cfg)

	byEmail, _ := svc.ListDeliveries(ctx, port.AuditQuery{Email: "AN@shop.vn"})
	if len(byEmail.Items) != 2 || len(byEmail.Contents) != 2 {
		t.Errorf("by email: %d items, %d contents", len(byEmail.Items), len(byEmail.Contents))
	}
	byPhone, _ := svc.ListDeliveries(ctx, port.AuditQuery{Phone: "+84 912 345 678", Channel: "push"})
	if len(byPhone.Items) != 1 || byPhone.Items[0].UserID != 42 {
		t.Errorf("by phone: %+v", byPhone.Items)
	}
	allForUser, _ := svc.ListDeliveries(ctx, port.AuditQuery{Phone: "0912345678"})
	if len(allForUser.Items) != 2 {
		t.Errorf("phone lookup should include emails sent to the user's address, got %d rows", len(allForUser.Items))
	}
	unknown, _ := svc.ListDeliveries(ctx, port.AuditQuery{Phone: "0999999999"})
	if len(unknown.Items) != 0 {
		t.Errorf("unknown phone returned %d rows", len(unknown.Items))
	}
	paged, _ := svc.ListDeliveries(ctx, port.AuditQuery{Limit: 2})
	if len(paged.Items) != 2 || paged.NextBeforeID == 0 {
		t.Errorf("paging: %+v", paged)
	}

	f.clock.now = f.clock.now.Add(31 * 24 * time.Hour)
	if n, _ := svc.PurgeExpired(ctx); n != 3 || len(f.audit.deliveries) != 0 {
		t.Errorf("purged %d, left %d", n, len(f.audit.deliveries))
	}
	keepForever := NewAuditService(f.audit, f.contacts, f.clock, &config.Config{})
	if n, _ := keepForever.PurgeExpired(ctx); n != 0 {
		t.Error("retention 0 must keep everything")
	}
}
