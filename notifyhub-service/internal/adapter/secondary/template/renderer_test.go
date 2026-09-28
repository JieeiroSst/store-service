package template

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/model"
)

func TestRenderEscapingPerChannel(t *testing.T) {
	r := New()
	data := map[string]interface{}{"name": "Tom & Jerry's <b>"}

	sms := &model.Template{ID: "1", Name: "sms", Channel: model.ChannelSMS, Subject: "Hi {{.name}}", Body: "Hi {{.name}}"}
	subj, body, err := r.Render(sms, data)
	if err != nil {
		t.Fatal(err)
	}
	if body != "Hi Tom & Jerry's <b>" || subj != "Hi Tom & Jerry's <b>" {
		t.Fatalf("sms must not be HTML-escaped: subject=%q body=%q", subj, body)
	}

	mail := &model.Template{ID: "2", Name: "mail", Channel: model.ChannelEmail, Subject: "Hi {{.name}}", Body: "<p>{{.name}}</p>"}
	subj, body, err = r.Render(mail, data)
	if err != nil {
		t.Fatal(err)
	}
	if subj != "Hi Tom & Jerry's <b>" {
		t.Fatalf("email subject must be plain text: %q", subj)
	}
	if body != "<p>Tom &amp; Jerry&#39;s &lt;b&gt;</p>" {
		t.Fatalf("email body must be HTML-escaped: %q", body)
	}
}

func TestRenderRecompilesChangedTemplate(t *testing.T) {
	r := New()
	tpl := &model.Template{ID: "1", Name: "t", Channel: model.ChannelSMS, Body: "v1", UpdatedAt: time.Unix(1, 0)}
	if err := r.Compile(tpl); err != nil {
		t.Fatal(err)
	}
	tpl2 := *tpl
	tpl2.Body, tpl2.UpdatedAt = "v2", time.Unix(2, 0)
	_, body, err := r.Render(&tpl2, nil)
	if err != nil {
		t.Fatal(err)
	}
	if body != "v2" {
		t.Fatalf("stale cache used: %q", body)
	}
}

func TestCompileRejectsInvalid(t *testing.T) {
	if err := New().Compile(&model.Template{ID: "x", Name: "x", Channel: model.ChannelSMS, Body: "{{.name"}); err == nil {
		t.Fatal("want parse error")
	}
}
