package pdf

import (
	"bytes"
	_ "embed"
	"fmt"
	"strconv"
	"time"

	"github.com/go-pdf/fpdf"
	"github.com/skip2/go-qrcode"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

//go:embed fonts/DejaVuSansCondensed.ttf
var fontRegular []byte

//go:embed fonts/DejaVuSansCondensed-Bold.ttf
var fontBold []byte

const family = "dejavu"

type Renderer struct {
	compress bool
	now      func() time.Time
}

var _ outbound.DocumentRenderer = (*Renderer)(nil)

func NewRenderer() *Renderer { return &Renderer{compress: true, now: time.Now} }

func (r *Renderer) newDoc(title string) *fpdf.Fpdf {
	p := fpdf.New("P", "mm", "A4", "")
	p.SetCompression(r.compress)
	p.SetTitle(title, true)
	p.SetCreator("ticket-service", true)
	p.SetCreationDate(r.now())
	p.AddUTF8FontFromBytes(family, "", fontRegular)
	p.AddUTF8FontFromBytes(family, "B", fontBold)
	p.SetMargins(18, 18, 18)
	p.SetAutoPageBreak(true, 18)
	return p
}

func money(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.FormatInt(n, 10)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}

func stamp(t time.Time, loc *time.Location, layout string) string {
	if loc == nil {
		loc = time.UTC
	}
	return t.In(loc).Format(layout)
}

func (r *Renderer) out(p *fpdf.Fpdf) ([]byte, error) {
	var buf bytes.Buffer
	if err := p.Output(&buf); err != nil {
		return nil, fmt.Errorf("draw pdf: %w", err)
	}
	return buf.Bytes(), nil
}

// ---- invoice

func (r *Renderer) Invoice(inv domain.Invoice, loc *time.Location) ([]byte, error) {
	p := r.newDoc("Invoice " + inv.Number)
	p.AddPage()
	const w = 174.0 // usable width

	p.SetFont(family, "B", 20)
	p.CellFormat(w/2, 10, "Hoá đơn / Invoice", "", 0, "L", false, 0, "")
	p.SetFont(family, "B", 12)
	p.CellFormat(w/2, 10, inv.Number, "", 1, "R", false, 0, "")
	p.SetFont(family, "", 10)
	p.SetTextColor(100, 100, 100)
	p.CellFormat(w, 6, fmt.Sprintf("Ngày / Date: %s   ·   Đơn hàng / Order #%d", stamp(inv.IssuedAt, loc, "02/01/2006 15:04"), inv.OrderID), "", 1, "L", false, 0, "")
	p.SetTextColor(0, 0, 0)
	if inv.Status == "refunded" {
		p.SetFont(family, "B", 12)
		p.SetTextColor(190, 30, 30)
		p.CellFormat(w, 8, "ĐÃ HOÀN TIỀN / REFUNDED", "", 1, "L", false, 0, "")
		p.SetTextColor(0, 0, 0)
	}
	p.Ln(4)

	// seller and buyer side by side
	y := p.GetY()
	p.SetFont(family, "B", 10)
	p.CellFormat(w/2, 6, "Bên bán / Seller", "", 1, "L", false, 0, "")
	p.SetFont(family, "", 10)
	seller := inv.Seller.Name
	if inv.Seller.TaxID != "" {
		seller += "\nMST / Tax ID: " + inv.Seller.TaxID
	}
	if inv.Seller.Address != "" {
		seller += "\n" + inv.Seller.Address
	}
	p.MultiCell(w/2-4, 5, seller, "", "L", false)
	left := p.GetY()
	p.SetXY(18+w/2, y)
	p.SetFont(family, "B", 10)
	p.CellFormat(w/2, 6, "Bên mua / Billed to", "", 1, "L", false, 0, "")
	p.SetX(18 + w/2)
	p.SetFont(family, "", 10)
	p.MultiCell(w/2, 5, inv.BuyerName+"\n"+inv.BuyerEmail, "", "L", false)
	p.SetY(max(left, p.GetY()) + 4)

	// the event
	p.SetFillColor(243, 244, 246)
	p.SetFont(family, "B", 11)
	p.CellFormat(w, 8, " "+inv.EventTitle, "", 1, "L", true, 0, "")
	p.SetFont(family, "", 10)
	p.CellFormat(w, 6, " "+stamp(inv.EventDate, loc, "Monday 02/01/2006, 15:04")+"  ·  "+inv.Venue, "", 1, "L", true, 0, "")
	p.Ln(4)

	// lines
	cols := []float64{92, 20, 31, 31}
	p.SetFont(family, "B", 10)
	p.SetFillColor(230, 232, 236)
	for i, h := range []string{"Vé / Ticket", "SL / Qty", "Đơn giá / Unit", "Thành tiền / Amount"} {
		align := "R"
		if i == 0 {
			align = "L"
		}
		p.CellFormat(cols[i], 8, " "+h+" ", "B", 0, align, true, 0, "")
	}
	p.Ln(-1)
	p.SetFont(family, "", 10)
	for _, l := range inv.Lines {
		p.CellFormat(cols[0], 8, " "+l.Description, "B", 0, "L", false, 0, "")
		p.CellFormat(cols[1], 8, strconv.Itoa(l.Quantity)+" ", "B", 0, "R", false, 0, "")
		p.CellFormat(cols[2], 8, money(l.UnitPrice)+" ", "B", 0, "R", false, 0, "")
		p.CellFormat(cols[3], 8, money(l.Amount)+" ", "B", 1, "R", false, 0, "")
	}
	row := func(label, value string, bold bool) {
		style := ""
		if bold {
			style = "B"
		}
		p.SetFont(family, style, 10)
		p.CellFormat(cols[0]+cols[1]+cols[2], 7, label+" ", "", 0, "R", false, 0, "")
		p.CellFormat(cols[3], 7, value+" ", "", 1, "R", false, 0, "")
	}
	p.Ln(2)
	row("Tạm tính / Subtotal", money(inv.Subtotal), false)
	if inv.Discount > 0 {
		label := "Giảm giá / Discount"
		if inv.PromoCode != "" {
			label += " (" + inv.PromoCode + ")"
		}
		row(label, "-"+money(inv.Discount), false)
	}
	row(fmt.Sprintf("Tổng cộng / Total (%s)", inv.Currency), money(inv.Total), true)
	if inv.RefundAmount > 0 {
		row("Đã hoàn / Refunded", money(inv.RefundAmount), false)
	}
	p.Ln(4)
	p.SetFont(family, "", 9)
	p.SetTextColor(100, 100, 100)
	if inv.VATPercent > 0 {
		p.MultiCell(w, 5, fmt.Sprintf("Giá đã bao gồm VAT %d%%: %s %s. / Prices include VAT at %d%%.", inv.VATPercent, money(inv.VAT), inv.Currency, inv.VATPercent), "", "L", false)
	}
	p.MultiCell(w, 5, "Thanh toán bằng / Paid by: "+string(inv.Payment), "", "L", false)
	return r.out(p)
}

// ---- tickets

// QRPNG is the QR code of a ticket code as a PNG.
func QRPNG(code string) ([]byte, error) {
	return qrcode.Encode(code, qrcode.Medium, 512)
}

func (r *Renderer) Tickets(doc domain.TicketDocument) ([]byte, error) {
	if len(doc.Tickets) == 0 {
		return nil, fmt.Errorf("%w: no tickets to draw", domain.ErrInvalid)
	}
	p := r.newDoc(fmt.Sprintf("Tickets, order #%d", doc.OrderID))
	for i, t := range doc.Tickets {
		p.AddPage()
		const w = 174.0
		p.SetFillColor(24, 32, 56)
		p.Rect(18, 18, w, 34, "F")
		p.SetXY(24, 22)
		p.SetTextColor(255, 255, 255)
		p.SetFont(family, "B", 18)
		p.MultiCell(w-12, 8, doc.EventTitle, "", "L", false)
		p.SetX(24)
		p.SetFont(family, "", 11)
		p.CellFormat(w-12, 6, stamp(doc.StartsAt, doc.Location, "Monday 02/01/2006 · 15:04"), "", 1, "L", false, 0, "")
		p.SetTextColor(0, 0, 0)
		p.SetXY(18, 58)
		p.SetFont(family, "B", 12)
		p.CellFormat(w, 7, doc.Venue, "", 1, "L", false, 0, "")
		if doc.Address != "" {
			p.SetFont(family, "", 10)
			p.SetTextColor(100, 100, 100)
			p.CellFormat(w, 6, doc.Address, "", 1, "L", false, 0, "")
			p.SetTextColor(0, 0, 0)
		}
		p.Ln(6)

		field := func(label, value string) {
			p.SetFont(family, "", 9)
			p.SetTextColor(100, 100, 100)
			p.CellFormat(40, 6, label, "", 0, "L", false, 0, "")
			p.SetTextColor(0, 0, 0)
			p.SetFont(family, "B", 12)
			p.CellFormat(w-40, 6, value, "", 1, "L", false, 0, "")
			p.Ln(1)
		}
		field("Loại vé / Ticket", t.TypeName)
		if t.Seat != "" {
			field("Ghế / Seat", t.Seat)
		}
		if t.HolderName != "" {
			field("Người tham dự / Attendee", t.HolderName)
		}
		field("Đơn hàng / Order", fmt.Sprintf("#%d  ·  vé %d/%d", doc.OrderID, i+1, len(doc.Tickets)))

		png, err := QRPNG(t.Code)
		if err != nil {
			return nil, fmt.Errorf("qr code: %w", err)
		}
		name := fmt.Sprintf("qr%d", i)
		opts := fpdf.ImageOptions{ImageType: "PNG"}
		p.RegisterImageOptionsReader(name, opts, bytes.NewReader(png))
		const qr = 80.0
		x := 18 + (w-qr)/2
		y := p.GetY() + 8
		p.ImageOptions(name, x, y, qr, qr, false, opts, 0, "")
		p.SetXY(18, y+qr+3)
		p.SetFont(family, "B", 11)
		p.CellFormat(w, 7, t.Code, "", 1, "C", false, 0, "")
		p.SetFont(family, "", 9)
		p.SetTextColor(100, 100, 100)
		p.MultiCell(w, 5, "Xuất trình mã QR này tại cổng. Mỗi vé chỉ vào được một lần; mã thay đổi nếu vé được chuyển nhượng.\nShow this QR code at the gate. A ticket admits one person once; its code changes if the ticket is transferred.", "", "C", false)
		if t.Status != "" && t.Status != "valid" {
			p.SetFont(family, "B", 40)
			p.SetTextColor(190, 30, 30)
			p.SetXY(18, y+qr/2-8)
			p.CellFormat(w, 16, "VOID", "", 0, "C", false, 0, "")
		}
		p.SetTextColor(0, 0, 0)
	}
	return r.out(p)
}
