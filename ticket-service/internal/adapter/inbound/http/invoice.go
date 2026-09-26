package http

import (
	"fmt"
	"html/template"
	"io"
	"mime"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type invoiceLineResponse struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Amount      int64  `json:"amount"`
}

type invoiceResponse struct {
	Number       string                `json:"number"`
	IssuedAt     time.Time             `json:"issued_at"`
	OrderID      int64                 `json:"order_id"`
	SellerName   string                `json:"seller_name"`
	SellerTaxID  string                `json:"seller_tax_id,omitempty"`
	SellerAddr   string                `json:"seller_address,omitempty"`
	BuyerName    string                `json:"buyer_name"`
	BuyerEmail   string                `json:"buyer_email"`
	EventTitle   string                `json:"event_title"`
	EventDate    time.Time             `json:"event_date"`
	Venue        string                `json:"venue"`
	Currency     string                `json:"currency"`
	Lines        []invoiceLineResponse `json:"lines"`
	Subtotal     int64                 `json:"subtotal"`
	Discount     int64                 `json:"discount"`
	PromoCode    string                `json:"promo_code,omitempty"`
	Total        int64                 `json:"total"`
	VATPercent   int                   `json:"vat_percent"`
	VAT          int64                 `json:"vat"`
	Payment      string                `json:"payment_method"`
	Status       string                `json:"status"`
	RefundAmount int64                 `json:"refund_amount,omitempty"`
}

func toInvoiceResponse(i domain.Invoice) invoiceResponse {
	r := invoiceResponse{Number: i.Number, IssuedAt: i.IssuedAt, OrderID: i.OrderID, SellerName: i.Seller.Name, SellerTaxID: i.Seller.TaxID,
		SellerAddr: i.Seller.Address, BuyerName: i.BuyerName, BuyerEmail: i.BuyerEmail, EventTitle: i.EventTitle, EventDate: i.EventDate,
		Venue: i.Venue, Currency: i.Currency, Subtotal: i.Subtotal, Discount: i.Discount, PromoCode: i.PromoCode, Total: i.Total,
		VATPercent: i.VATPercent, VAT: i.VAT, Payment: string(i.Payment), Status: i.Status, RefundAmount: i.RefundAmount,
		Lines: make([]invoiceLineResponse, 0, len(i.Lines))}
	for _, l := range i.Lines {
		r.Lines = append(r.Lines, invoiceLineResponse{Description: l.Description, Quantity: l.Quantity, UnitPrice: l.UnitPrice, Amount: l.Amount})
	}
	return r
}

func money(n int64) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		s = "-" + s
	}
	return s
}

var invoiceHTML = template.Must(template.New("invoice").Funcs(template.FuncMap{"money": money}).Parse(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"><title>Invoice {{.Number}}</title>
<style>
 body{font-family:system-ui,sans-serif;color:#1a1a1a;max-width:720px;margin:2rem auto;padding:0 1rem}
 h1{margin:0}.muted{color:#666}.right{text-align:right}
 table{width:100%;border-collapse:collapse;margin:1.5rem 0}th,td{padding:.5rem;border-bottom:1px solid #ddd;text-align:left}
 th.right,td.right{text-align:right}.total td{font-weight:bold;border-bottom:none}.stamp{color:#b00;font-weight:bold;text-transform:uppercase}
 @media print{body{margin:0}}
</style></head><body>
<h1>Invoice {{.Number}}</h1>
<p class="muted">Issued {{.IssuedAt.Format "2 January 2006"}} · order #{{.OrderID}}{{if eq .Status "refunded"}} · <span class="stamp">refunded</span>{{end}}</p>
<p><strong>{{.SellerName}}</strong>{{if .SellerTaxID}}<br>Tax ID: {{.SellerTaxID}}{{end}}{{if .SellerAddr}}<br>{{.SellerAddr}}{{end}}</p>
<p><strong>Billed to</strong><br>{{.BuyerName}}<br>{{.BuyerEmail}}</p>
<p><strong>{{.EventTitle}}</strong><br>{{.EventDate.Format "Monday 2 January 2006, 15:04"}} · {{.Venue}}</p>
<table>
<tr><th>Ticket</th><th class="right">Qty</th><th class="right">Unit price</th><th class="right">Amount</th></tr>
{{range .Lines}}<tr><td>{{.Description}}</td><td class="right">{{.Quantity}}</td><td class="right">{{money .UnitPrice}}</td><td class="right">{{money .Amount}}</td></tr>
{{end}}
<tr><td colspan="3" class="right">Subtotal</td><td class="right">{{money .Subtotal}}</td></tr>
{{if .Discount}}<tr><td colspan="3" class="right">Discount{{if .PromoCode}} ({{.PromoCode}}){{end}}</td><td class="right">-{{money .Discount}}</td></tr>
{{end}}
<tr class="total"><td colspan="3" class="right">Total ({{.Currency}})</td><td class="right">{{money .Total}}</td></tr>
</table>
{{if .VATPercent}}<p class="muted">Prices include VAT at {{.VATPercent}}%: {{money .VAT}} {{.Currency}}.</p>{{end}}
<p class="muted">Paid by {{.Payment}}.{{if .RefundAmount}} Refunded: {{money .RefundAmount}} {{.Currency}}.{{end}}</p>
</body></html>`))

// getInvoice serves the receipt as JSON, as a printable page with ?format=html, or as a PDF with ?format=pdf.
func (h *Handler) getInvoice(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	if r.URL.Query().Get("format") == "pdf" {
		pdf, err := h.documents.InvoicePDF(r.Context(), mustPrincipal(r), id)
		if err != nil {
			h.fail(w, err)
			return
		}
		h.writePDF(w, fmt.Sprintf("invoice-order-%d.pdf", id), "inline", pdf)
		return
	}
	inv, err := h.orders.Invoice(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	if r.URL.Query().Get("format") == "html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "private, no-store")
		if err := invoiceHTML.Execute(w, toInvoiceResponse(inv)); err != nil {
			h.log.Error("render invoice", "err", err)
		}
		return
	}
	h.write(w, http.StatusOK, toInvoiceResponse(inv))
}

func (h *Handler) writePDF(w http.ResponseWriter, name, disposition string, pdf []byte) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", strconv.Itoa(len(pdf)))
	w.Header().Set("Content-Disposition", mime.FormatMediaType(disposition, map[string]string{"filename": name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	if _, err := w.Write(pdf); err != nil {
		h.log.Warn("write pdf", "err", err)
	}
}

// ticketPDF is one ticket with its QR code, drawn fresh: the code on it is the ticket's code now.
func (h *Handler) ticketPDF(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	pdf, err := h.documents.TicketPDF(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	h.writePDF(w, fmt.Sprintf("ticket-%d.pdf", id), "attachment", pdf)
}

type documentResponse struct {
	Kind      string    `json:"kind"`
	FileName  string    `json:"file_name"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	URL       string    `json:"url"`
}

func (h *Handler) listDocuments(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	docs, err := h.documents.List(r.Context(), mustPrincipal(r), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	out := make([]documentResponse, len(docs))
	for i, d := range docs {
		out[i] = documentResponse{Kind: d.Kind, FileName: d.FileName, Size: d.Size, CreatedAt: d.CreatedAt, URL: fmt.Sprintf("/api/v1/orders/%d/documents/%s", id, d.Kind)}
	}
	h.write(w, http.StatusOK, map[string]any{"items": out})
}

func (h *Handler) downloadDocument(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r, "id")
	if !ok {
		return
	}
	f, err := h.documents.Open(r.Context(), mustPrincipal(r), id, r.PathValue("kind"))
	if err != nil {
		h.fail(w, err)
		return
	}
	defer f.Body.Close()
	w.Header().Set("Content-Type", "application/pdf")
	if f.Size > 0 {
		w.Header().Set("Content-Length", strconv.FormatInt(f.Size, 10))
	}
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": f.Name}))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Cache-Control", "private, no-store")
	if _, err := io.Copy(w, f.Body); err != nil {
		h.log.Warn("download interrupted", "order", id, "err", err)
	}
}
