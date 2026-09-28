package application

import (
	"errors"
	"strings"
	"testing"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
)

func TestRawEmailWithRawData(t *testing.T) {
	f := newFixture()
	n, err := f.notif.CreateNotification(ctx, &model.Notification{
		Type: "email", Recipient: "ke.toan@shop.vn", Title: "Báo cáo doanh thu",
		Message: "<p>Tổng hợp ngày 28/09</p>",
		RawData: map[string]any{"revenue": 125000000.0, "orders": 342.0, "top": []any{"SKU-1", "SKU-2"}, "note": "<script>alert(1)</script>"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := f.notif.Dispatch(ctx, n, 1, false); err != nil {
		t.Fatal(err)
	}
	body := f.email.lastBody
	for _, want := range []string{"<p>Tổng hợp ngày 28/09</p>", ">revenue<", ">125000000<", `[&#34;SKU-1&#34;,&#34;SKU-2&#34;]`, "&lt;script&gt;"} {
		if !strings.Contains(body, want) {
			t.Errorf("body missing %q:\n%s", want, body)
		}
	}
	if strings.Contains(body, "<script>") {
		t.Error("raw_data was not HTML-escaped")
	}
	if f.email.lastSubj != "Báo cáo doanh thu" {
		t.Errorf("subject = %q", f.email.lastSubj)
	}
	if last := f.audit.contents[len(f.audit.contents)-1]; !strings.Contains(last.Body, ">orders<") {
		t.Error("audit should store the body that was actually sent, including raw_data")
	}
}

func TestTemplateEmailAlsoAppendsRawData(t *testing.T) {
	f := newFixture()
	n, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "email", Recipient: "a@b.vn", TemplateType: "welcome", TemplateData: map[string]string{"name": "An"}, RawData: map[string]any{"ref": "DH-1"}})
	_ = f.notif.Dispatch(ctx, n, 1, false)
	if !strings.HasPrefix(f.email.lastBody, "<p>xin chào</p>") || !strings.Contains(f.email.lastBody, ">DH-1<") {
		t.Errorf("body = %s", f.email.lastBody)
	}
}

func TestRawSlackWithRawData(t *testing.T) {
	f := newFixture()
	n, err := f.notif.CreateNotification(ctx, &model.Notification{
		Type: "slack", Title: "Deploy", Message: "payment-service v1.4.2 :rocket:",
		RawData: map[string]any{"env": "prod", "replicas": 3.0},
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = f.notif.Dispatch(ctx, n, 1, false)
	want := "payment-service v1.4.2 :rocket:\n```\n{\n  \"env\": \"prod\",\n  \"replicas\": 3\n}\n```"
	if f.slack.title != "Deploy" || f.slack.text != want {
		t.Errorf("slack = %q / %q", f.slack.title, f.slack.text)
	}

	onlyData, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "slack", RawData: map[string]any{"a": "```x```"}})
	_ = f.notif.Dispatch(ctx, onlyData, 1, false)
	if strings.Count(f.slack.text, "```") != 2 {
		t.Errorf("code fence inside raw_data must not break the block: %q", f.slack.text)
	}

	tpl, _ := f.notif.CreateNotification(ctx, &model.Notification{Type: "slack", TemplateType: "alert", TemplateData: map[string]string{"service": "api"}, RawData: map[string]any{"p95_ms": 900.0}})
	_ = f.notif.Dispatch(ctx, tpl, 1, false)
	if !strings.HasPrefix(f.slack.text, "service api is down\n```") {
		t.Errorf("template + raw_data = %q", f.slack.text)
	}
}

func TestEmailAndSlackValidation(t *testing.T) {
	f := newFixture()
	big := map[string]any{"x": strings.Repeat("a", model.MaxRawDataBytes)}
	bad := []*model.Notification{
		{Type: "email", Title: "s", Message: "b"},
		{Type: "email", Recipient: "a@b.vn", Message: "no subject"},
		{Type: "email", Recipient: "a@b.vn", Title: "subject only"},
		{Type: "slack"},
		{Type: "slack", Message: "x", RawData: big},
	}
	for i, n := range bad {
		if _, err := f.notif.CreateNotification(ctx, n); !errors.Is(err, common.ErrInvalidRequest) {
			t.Errorf("case %d: got %v", i, err)
		}
	}
	ok := []*model.Notification{
		{Type: "email", Recipient: "a@b.vn", Title: "s", RawData: map[string]any{"k": "v"}},
		{Type: "slack", RawData: map[string]any{"k": "v"}},
	}
	for i, n := range ok {
		if _, err := f.notif.CreateNotification(ctx, n); err != nil {
			t.Errorf("valid case %d: %v", i, err)
		}
	}
}

func TestTextToHTMLEscapes(t *testing.T) {
	if got := model.TextToHTML("a < b\nline 2"); !strings.Contains(got, "a &lt; b\nline 2") {
		t.Errorf("got %s", got)
	}
}

func TestRawDataNumbersAreNotScientific(t *testing.T) {
	got := model.RawDataHTML(map[string]any{"amount": 1250000000.0, "rate": 0.12, "ok": true})
	for _, want := range []string{">1250000000<", ">0.12<", ">true<"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s in %s", want, got)
		}
	}
}
