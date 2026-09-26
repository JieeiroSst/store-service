package integration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/service"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type fakeMail struct {
	mu   sync.Mutex
	fail int // fail this many calls first
	sent []domain.Notification
}

func (f *fakeMail) Email(_ context.Context, n domain.Notification) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail > 0 {
		f.fail--
		return errors.New("notification-service is down")
	}
	f.sent = append(f.sent, n)
	return nil
}

// E-mails are queued with the notification, sent from the queue, retried when the sender is down, and never sent twice.
func TestEmailQueue(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000, email: true})
	mail := &fakeMail{fail: 1}
	notes := service.NewNotificationService(postgres.NewNotificationRepository(p.pool), nil, mail, nil)
	e, tt := seedEvent(t, p, 10, nil)

	o := paidOrder(t, p, 5, e, tt, 2)
	if n, err := notes.SendEmails(ctx); err != nil || n != 0 { // the sender is down
		t.Fatalf("first attempt: %d %v", n, err)
	}
	if _, err := p.pool.Exec(ctx, `update notification set email_claimed_until = null`); err != nil { // skip the lease
		t.Fatal(err)
	}
	if n, err := notes.SendEmails(ctx); err != nil || n != 1 {
		t.Fatalf("retry: %d %v", n, err)
	}
	if n, _ := notes.SendEmails(ctx); n != 0 {
		t.Fatal("an e-mail that was sent must not be sent again")
	}
	m := mail.sent[0]
	if m.Email.To != "buyer@example.com" || m.Email.Template != "ticket_paid" || m.Email.Data["TicketCount"] != "2" ||
		m.Email.Data["Total"] != "1,000,000" || m.Email.Data["EventTitle"] != e.Title || m.Email.Data["OrderID"] == "" || m.UserID != 5 {
		t.Fatalf("the receipt e-mail: %+v", m.Email)
	}
	if o.ID == 0 {
		t.Fatal()
	}

	// an offer to an address nobody registered still reaches them
	p.tickets.WithDirectory(nil)
	tk := o.Tickets[0]
	if _, err := p.tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToEmail: "friend@example.com", Message: "see you"}); err != nil {
		t.Fatal(err)
	}
	if n, err := notes.SendEmails(ctx); err != nil || n != 1 {
		t.Fatalf("offer e-mail: %d %v", n, err)
	}
	if m := mail.sent[1]; m.UserID != 0 || m.Email.To != "friend@example.com" || m.Email.Template != "ticket_transfer_offer" || m.Email.Data["Message"] != "see you" {
		t.Fatalf("the offer e-mail: %+v", m)
	}
	// an e-mail-only row is nobody's inbox item
	var inbox int
	if err := p.pool.QueryRow(ctx, `select count(*) from notification where user_id = 0`).Scan(&inbox); err != nil || inbox != 1 {
		t.Fatalf("e-mail-only rows: %d %v", inbox, err)
	}

	// an invitation e-mails the invited person, and the reminder and cancellation e-mails exist too
	if _, err := p.orders.Invite(ctx, organizer, e.ID, inbound.InviteCommand{TicketTypeID: tt.ID, Quantity: 1, UserID: 60, Name: "Guest", Email: "guest@example.com"}); err != nil {
		t.Fatal(err)
	}
	if _, err := p.events.Cancel(ctx, organizer, e.ID, "storm"); err != nil {
		t.Fatal(err)
	}
	if n, err := p.orders.SettleCancelledEvents(ctx); err != nil || n < 1 {
		t.Fatalf("settle: %d %v", n, err)
	}
	if _, err := notes.SendEmails(ctx); err != nil {
		t.Fatal(err)
	}
	templates := map[string]string{}
	for _, s := range mail.sent {
		templates[s.Email.Template] = s.Email.To
	}
	if templates["ticket_invited"] != "guest@example.com" || templates["event_cancelled"] == "" {
		t.Fatalf("e-mails sent: %v", templates)
	}
}

// Without an e-mail gateway nothing is queued, so the table does not fill with mail nobody sends.
func TestNoEmailWithoutGateway(t *testing.T) {
	d := newDatabase(t)
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000}) // email off
	e, tt := seedEvent(t, p, 5, nil)
	paidOrder(t, p, 5, e, tt, 1)
	var n int
	if err := p.pool.QueryRow(context.Background(), `select count(*) from notification where email_template is not null`).Scan(&n); err != nil || n != 0 {
		t.Fatalf("%d e-mails queued without a gateway (%v)", n, err)
	}
}

// Through the real fx module: the sweeper hands the e-mail to a (fake) notification-service in its own JSON shape.
func TestEmailReachesNotificationService(t *testing.T) {
	var mu sync.Mutex
	var got []map[string]any
	ns := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		mu.Lock()
		if r.URL.Path == "/api/v1/notifications/email" {
			got = append(got, body)
		}
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte("{}"))
	}))
	defer ns.Close()
	a := startApp(t, map[string]string{"NotificationServiceURL": ns.URL, "SweepIntervalSeconds": "1"})
	eid, tid := publishedEvent(a, 5)
	buy(a, 5, eid, tid, 1)

	deadline := time.Now().Add(8 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(got)
		mu.Unlock()
		if n > 0 {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) == 0 {
		t.Fatal("no e-mail reached notification-service")
	}
	b := got[0]
	data, _ := b["template_data"].(map[string]any)
	if b["recipient"] != "user5@example.com" && b["recipient"] != "buyer@example.com" || b["template_type"] != "ticket_paid" || data["EventTitle"] == "" || b["user_id"].(float64) != 5 {
		t.Fatalf("e-mail request: %v", b)
	}
}
