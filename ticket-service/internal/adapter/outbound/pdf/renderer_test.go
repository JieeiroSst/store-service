package pdf

import (
	"bytes"
	"image/png"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	"golang.org/x/image/font/sfnt"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func testRenderer() *Renderer {
	r := NewRenderer()
	r.compress = false // so the structure can be read
	r.now = func() time.Time { return time.Date(2026, 10, 1, 9, 0, 0, 0, time.UTC) }
	return r
}

func isPDF(b []byte) bool {
	return bytes.HasPrefix(b, []byte("%PDF-")) && bytes.HasSuffix(bytes.TrimSpace(b), []byte("%%EOF"))
}

func pages(pdf []byte) int { return len(regexp.MustCompile(`/Type /Page[^s]`).FindAll(pdf, -1)) }

func TestFontCoversVietnamese(t *testing.T) {
	for name, data := range map[string][]byte{"regular": fontRegular, "bold": fontBold} {
		f, err := sfnt.Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		var b sfnt.Buffer
		for _, r := range "ạảãấầẩẫậắằẳẵặẹẻẽếềểễệỉịọỏốồổỗộớờởỡợụủứừửữựỳỵỷỹđĐƠơƯưÀÁÂÃÈÉÊÌÍÒÓÔÕÙÚÝăĂ₫" {
			if i, err := f.GlyphIndex(&b, r); err != nil || i == 0 {
				t.Errorf("%s font has no glyph for %q: a Vietnamese name would print with holes", name, r)
			}
		}
	}
}

func invoice(status string) domain.Invoice {
	paid := time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC)
	return domain.Invoice{Number: "INV-2026-00000042", IssuedAt: paid, OrderID: 42,
		Seller:    domain.Seller{Name: "Công ty Vé Việt", TaxID: "0312345678", Address: "1 Lê Lợi, Quận 1, TP. Hồ Chí Minh"},
		BuyerName: "Nguyễn Thị Ánh", BuyerEmail: "anh@example.com", EventTitle: "Hà Anh Tuấn — Live Concert",
		EventDate: time.Date(2026, 11, 20, 20, 0, 0, 0, time.UTC), Venue: "Sân vận động Mỹ Đình", Currency: "VND",
		Lines:    []domain.InvoiceLine{{Description: "VIP", Quantity: 2, UnitPrice: 1_500_000, Amount: 3_000_000}, {Description: "GA", Quantity: 1, UnitPrice: 500_000, Amount: 500_000}},
		Subtotal: 3_500_000, Discount: 350_000, PromoCode: "EARLY10", Total: 3_150_000, VATPercent: 10, VAT: 286_364, Payment: domain.MethodWallet, Status: status}
}

func TestInvoicePDF(t *testing.T) {
	r := testRenderer()
	loc := time.FixedZone("UTC+7", 7*3600)
	out, err := r.Invoice(invoice("paid"), loc)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(out, []byte("%PDF-")) || !bytes.HasSuffix(bytes.TrimSpace(out), []byte("%%EOF")) || pages(out) != 1 {
		t.Fatalf("not a one-page PDF: %d bytes, %d pages", len(out), pages(out))
	}
	if !bytes.Contains(out, []byte("/Title (")) {
		t.Error("the title is missing from the document info")
	}
	refunded := invoice("refunded")
	refunded.RefundAmount = 3_150_000
	rf, err := r.Invoice(refunded, nil)
	if err != nil || !isPDF(rf) || len(rf) == 0 {
		t.Fatalf("a refunded invoice: %v", err)
	}
	// many lines flow onto a second page
	many := invoice("paid")
	for i := 0; i < 40; i++ {
		many.Lines = append(many.Lines, domain.InvoiceLine{Description: "Extra", Quantity: 1, UnitPrice: 1, Amount: 1})
	}
	if out, err = r.Invoice(many, loc); err != nil || pages(out) < 2 {
		t.Fatalf("a long invoice: %d pages, %v", pages(out), err)
	}
}

func TestTicketsPDFHasOneScannablePagePerTicket(t *testing.T) {
	r := testRenderer()
	doc := domain.TicketDocument{OrderID: 42, EventTitle: "Hà Anh Tuấn — Live Concert", Venue: "Sân vận động Mỹ Đình", Address: "Hà Nội",
		StartsAt: time.Date(2026, 11, 20, 13, 0, 0, 0, time.UTC), Location: time.FixedZone("UTC+7", 7*3600), BuyerName: "Nguyễn Thị Ánh"}
	codes := []string{"K7N4Q2ZP5T3V6XW9B8C1D0FGHJ4LM2NR", "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", "ZZ7Y6X5W4V3U2T1SRQPONMLKJIHGFEDC"}
	for i, c := range codes {
		doc.Tickets = append(doc.Tickets, domain.TicketPage{TicketID: int64(i + 1), TypeName: "VIP", Seat: "Stalls A-" + string(rune('1'+i)), HolderName: "Lê Văn Đức", Code: c, Status: "valid"})
	}
	out, err := r.Tickets(doc)
	if err != nil {
		t.Fatal(err)
	}
	if pages(out) != 3 || strings.Count(string(out), "/Subtype /Image") != 3 {
		t.Fatalf("%d pages, %d images for 3 tickets", pages(out), strings.Count(string(out), "/Subtype /Image"))
	}
	// what a gate scans is the code, exactly
	for _, c := range codes {
		img, err := QRPNG(c)
		if err != nil {
			t.Fatal(err)
		}
		pic, err := png.Decode(bytes.NewReader(img))
		if err != nil {
			t.Fatal(err)
		}
		bmp, err := gozxing.NewBinaryBitmapFromImage(pic)
		if err != nil {
			t.Fatal(err)
		}
		res, err := qrcode.NewQRCodeReader().Decode(bmp, nil)
		if err != nil || res.GetText() != c {
			t.Fatalf("QR of %q decodes to %v (%v)", c, res, err)
		}
	}
	if _, err := r.Tickets(domain.TicketDocument{OrderID: 1}); err == nil {
		t.Fatal("a document without tickets")
	}
	doc.Tickets[0].Status = "void"
	if v, err := r.Tickets(doc); err != nil || !isPDF(v) || pages(v) != 3 {
		t.Fatalf("a void ticket is still drawn, marked: %v", err)
	}
}

func TestMoney(t *testing.T) {
	for n, want := range map[int64]string{0: "0", 999: "999", 1000: "1,000", 1150000: "1,150,000", -50000: "-50,000"} {
		if got := money(n); got != want {
			t.Errorf("money(%d) = %q", n, got)
		}
	}
}
