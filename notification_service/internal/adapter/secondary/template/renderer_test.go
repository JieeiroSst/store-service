package template

import (
	"strings"
	"testing"
)

// Every template must render with the data its sender provides, and must escape it (the data comes from users).
func TestTicketTemplatesRender(t *testing.T) {
	r, err := NewRenderer()
	if err != nil {
		t.Fatal(err)
	}
	data := map[string]string{
		"Name": "Lan", "EventTitle": "<script>alert(1)</script>Concert", "EventDate": "Thứ Bảy 3 tháng 10, 20:00", "Venue": "Mỹ Đình",
		"OrderID": "42", "TicketCount": "2", "Total": "1,150,000", "Currency": "VND", "Message": "Enjoy!", "ExpiresAt": "05/10 20:00",
		"Note": "Cảm ơn bạn", "Reason": "mưa bão", "Amount": "900,000", "Fee": "50,000",
	}
	for _, name := range []string{"ticket_paid", "ticket_transfer_offer", "ticket_invited", "event_cancelled", "event_reminder", "resale_sold"} {
		subject, html, err := r.Render(name, data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if subject == "" || !strings.Contains(html, "Lan") {
			t.Errorf("%s: subject %q, body %q", name, subject, html)
		}
		if strings.Contains(html, "<script>") || strings.Contains(subject, "<script>") && strings.Contains(html, "<script>") {
			t.Errorf("%s: user data was not escaped: %s", name, html)
		}
	}
}
