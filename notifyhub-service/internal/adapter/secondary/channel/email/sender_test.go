package email

import (
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
)

func TestBuildMIMEMessageBlocksHeaderInjection(t *testing.T) {
	raw := string(buildMIMEMessage("noreply@example.com", port.Message{
		Recipient: "a@example.com",
		Subject:   "Hello\r\nBcc: victim@example.com",
		Body:      "<p>hi</p>",
	}, time.Unix(0, 0)))

	headers, _, _ := strings.Cut(raw, "\r\n\r\n")
	for _, line := range strings.Split(headers, "\r\n") {
		if strings.HasPrefix(strings.ToLower(line), "bcc:") {
			t.Fatalf("injected header present:\n%s", headers)
		}
	}
	if !strings.Contains(headers, "Content-Type: text/html; charset=utf-8") {
		t.Fatalf("html body not detected:\n%s", headers)
	}
}

func TestBuildMIMEMessageEncodesUnicodeSubject(t *testing.T) {
	raw := string(buildMIMEMessage("noreply@example.com", port.Message{
		Recipient: "a@example.com",
		Subject:   "Xin chào",
		Body:      "plain",
	}, time.Unix(0, 0)))
	if !strings.Contains(raw, "Subject: =?utf-8?q?") {
		t.Fatalf("subject not MIME-encoded:\n%s", raw)
	}
}
