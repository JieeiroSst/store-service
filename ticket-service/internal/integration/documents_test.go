package integration

import (
	"bytes"
	"context"
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func isPDF(b []byte) bool {
	return bytes.HasPrefix(b, []byte("%PDF-")) && bytes.HasSuffix(bytes.TrimSpace(b), []byte("%%EOF"))
}

// The documents of a paid order are made once, kept under the buyer's id, and readable by the buyer alone.
func TestOrderDocuments(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 20, nil)

	pending, err := p.orders.Reserve(ctx, user(5), cmd(e.ID, tt.ID, 2, "doc-1"))
	if err != nil {
		t.Fatal(err)
	}
	if made, err := p.docs.Generate(ctx, pending.ID); err != nil || made {
		t.Fatalf("documents for an unpaid order: %v %v", made, err)
	}
	o, err := p.orders.Pay(ctx, user(5), pending.ID, inbound.PayCommand{Method: domain.MethodWallet})
	if err != nil {
		t.Fatal(err)
	}
	if ids, _ := p.docs.GenerateMissing(ctx); ids != 1 {
		t.Fatalf("generate missing: %d", ids)
	}
	files := p.store.of(5)
	if len(files) != 2 || len(p.store.of(6)) != 0 {
		t.Fatalf("files under the buyer's id: %d", len(files))
	}
	names := map[string]bool{}
	for _, f := range files {
		names[f.name] = true
		if !isPDF(f.data) {
			t.Fatalf("%s is not a PDF", f.name)
		}
	}
	if !names["tickets-order-"+itoa(o.ID)+".pdf"] || !names["invoice-INV-"+time4(o)+"-"+pad8(o.ID)+".pdf"] {
		t.Fatalf("file names: %v", names)
	}
	// once made, not made again
	if made, err := p.docs.Generate(ctx, o.ID); err != nil || made {
		t.Fatalf("second generation: %v %v", made, err)
	}
	if n, _ := p.docs.GenerateMissing(ctx); n != 0 || len(p.store.of(5)) != 2 {
		t.Fatalf("a finished order was picked up again: %d, %d files", n, len(p.store.of(5)))
	}

	// who reads what
	list, err := p.docs.List(ctx, user(5), o.ID)
	if err != nil || len(list) != 2 {
		t.Fatalf("list: %+v %v", list, err)
	}
	if _, err := p.docs.List(ctx, user(6), o.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger listing: %v", err)
	}
	f, err := p.docs.Open(ctx, user(5), o.ID, domain.DocInvoice)
	if err != nil {
		t.Fatal(err)
	}
	data, _ := io.ReadAll(f.Body)
	f.Body.Close()
	if !isPDF(data) || f.Name == "" {
		t.Fatalf("download: %q", f.Name)
	}
	if _, err := p.docs.Open(ctx, user(6), o.ID, domain.DocInvoice); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger downloading: %v", err)
	}
	if _, err := p.docs.Open(ctx, admin, o.ID, domain.DocTickets); err != nil {
		t.Fatalf("an admin downloading: %v", err)
	}
	if _, err := p.docs.Open(ctx, user(5), o.ID, "passport"); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("an unknown kind: %v", err)
	}

	// PDFs on request: the invoice as it is now, a ticket for its holder alone
	inv, err := p.docs.InvoicePDF(ctx, user(5), o.ID)
	if err != nil || !isPDF(inv) {
		t.Fatalf("invoice pdf: %v", err)
	}
	if _, err := p.docs.InvoicePDF(ctx, user(6), o.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger's invoice pdf: %v", err)
	}
	tk := o.Tickets[0]
	pdf1, err := p.docs.TicketPDF(ctx, user(5), tk.ID)
	if err != nil || !isPDF(pdf1) {
		t.Fatalf("ticket pdf: %v", err)
	}
	if _, err := p.docs.TicketPDF(ctx, user(6), tk.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("a stranger's ticket pdf: %v", err)
	}
	// after a transfer the ticket's PDF is the new holder's, with the new code
	x, _ := p.tickets.Offer(ctx, user(5), tk.ID, inbound.TransferCommand{ToUserID: 6})
	if _, err := p.docs.TicketPDF(ctx, user(5), tk.ID); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("a ticket on offer has no printable version: %v", err)
	}
	if _, err := p.tickets.AcceptTransfer(ctx, user(6), x.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := p.docs.TicketPDF(ctx, user(5), tk.ID); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("the old holder's ticket pdf: %v", err)
	}
	pdf2, err := p.docs.TicketPDF(ctx, user(6), tk.ID)
	if err != nil || bytes.Equal(pdf1, pdf2) {
		t.Fatalf("the new holder's ticket pdf must carry the new code: %v", err)
	}
	// a refunded order's invoice says so
	if _, err := p.orders.Cancel(ctx, organizer, o.ID); err != nil {
		t.Fatal(err)
	}
	if rf, err := p.docs.InvoicePDF(ctx, user(5), o.ID); err != nil || bytes.Equal(rf, inv) {
		t.Fatalf("refunded invoice: %v", err)
	}
}

func itoa(n int64) string { return fmtInt(n) }

// A failing store leaves the order to be retried; nothing is half-recorded.
func TestDocumentsAreRetried(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o := paidOrder(t, p, 5, e, tt, 1)

	p.store.putFail = 1
	if n, _ := p.docs.GenerateMissing(ctx); n != 0 {
		t.Fatalf("a failed upload counted: %d", n)
	}
	if list, _ := p.docs.List(ctx, user(5), o.ID); len(list) > 1 {
		t.Fatalf("recorded: %+v", list)
	}
	if n, err := p.docs.GenerateMissing(ctx); err != nil || n != 1 {
		t.Fatalf("retry: %d %v", n, err)
	}
	if list, _ := p.docs.List(ctx, user(5), o.ID); len(list) != 2 || len(p.store.of(5)) != 2 {
		t.Fatalf("after the retry: %d documents, %d files", len(list), len(p.store.of(5)))
	}
}

// Replicas generating the same order's documents at once end with exactly one copy of each, and no orphan files.
func TestDocumentsRace(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o := paidOrder(t, p, 5, e, tt, 3)

	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := p.docs.Generate(ctx, o.ID); err != nil {
				t.Errorf("generate: %v", err)
			}
		}()
	}
	wg.Wait()
	list, _ := p.docs.List(ctx, user(5), o.ID)
	if len(list) != 2 || len(p.store.of(5)) != 2 {
		t.Fatalf("%d documents recorded, %d files stored: the losers must remove their copies", len(list), len(p.store.of(5)))
	}
}

// Without an upload-service nothing is stored, but PDFs are still drawn on request.
func TestDocumentsWithoutAStore(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	p := newPod(t, d, newFakeWallets(), podOptions{ratePerSec: 1000})
	e, tt := seedEvent(t, p, 5, nil)
	o := paidOrder(t, p, 5, e, tt, 1)
	bare := newBareDocs(p)
	if n, err := bare.GenerateMissing(ctx); err != nil || n != 0 {
		t.Fatalf("no store: %d %v", n, err)
	}
	if pdf, err := bare.InvoicePDF(ctx, user(5), o.ID); err != nil || !isPDF(pdf) {
		t.Fatalf("on-request invoice: %v", err)
	}
	if _, err := bare.Open(ctx, user(5), o.ID, domain.DocInvoice); !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("nothing stored: %v", err)
	}
}
