package ticketclient

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	base string
	key  string
	http *http.Client
}

type Option func(*Client)

func WithHTTPClient(h *http.Client) Option { return func(c *Client) { c.http = h } }

func New(baseURL, key string, opts ...Option) *Client {
	c := &Client{base: strings.TrimRight(baseURL, "/"), key: key, http: &http.Client{Timeout: 10 * time.Second}}
	for _, o := range opts {
		o(c)
	}
	return c
}

type actingUser struct{}

func WithActingUser(ctx context.Context, userID int64) context.Context {
	return context.WithValue(ctx, actingUser{}, userID)
}

type APIError struct {
	Status  int
	Message string
	Ticket  *Ticket
}

func (e *APIError) Error() string { return fmt.Sprintf("ticket-service: %d %s", e.Status, e.Message) }

func status(err error, code int) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status == code
}

func IsNotFound(err error) bool { return status(err, http.StatusNotFound) }

func IsConflict(err error) bool { return status(err, http.StatusConflict) }
func IsBusy(err error) bool     { return status(err, http.StatusTooManyRequests) }

func IsAlreadyCheckedIn(err error) bool {
	var e *APIError
	return errors.As(err, &e) && e.Status == http.StatusConflict && e.Ticket != nil
}

type Ticket struct {
	ID            int64      `json:"id"`
	OrderID       int64      `json:"order_id"`
	EventID       int64      `json:"event_id"`
	EventTitle    string     `json:"event_title"`
	EventStartsAt *time.Time `json:"event_starts_at"`
	SessionID     int64      `json:"session_id"`
	Venue         string     `json:"venue"`
	TicketTypeID  int64      `json:"ticket_type_id"`
	TypeName      string     `json:"type_name"`
	SeatID        int64      `json:"seat_id"`
	Seat          string     `json:"seat"`
	Code          string     `json:"code"`
	Status        string     `json:"status"`
	HolderID      int64      `json:"holder_id"`
	HolderName    string     `json:"holder_name"`
	HolderEmail   string     `json:"holder_email"`
	TransferCount int        `json:"transfer_count"`
	CheckedInAt   *time.Time `json:"checked_in_at"`
}

type HistoryEntry struct {
	ID        int64     `json:"id"`
	Event     string    `json:"event"`
	From      string    `json:"from_status"`
	To        string    `json:"to_status"`
	Actor     int64     `json:"actor"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
}

type TicketPage struct {
	Items      []Ticket `json:"items"`
	NextCursor string   `json:"next_cursor"`
}

type TicketFilter struct {
	EventID int64
	Status  string
	Cursor  string
	Limit   int
}

type Scan struct {
	Code      string     `json:"code"`
	ScannedAt *time.Time `json:"scanned_at,omitempty"`
}

type ScanResult struct {
	Code   string  `json:"code"`
	Result string  `json:"result"`
	Ticket *Ticket `json:"ticket"`
}

type BatchResult struct {
	Results  []ScanResult `json:"results"`
	Admitted int          `json:"admitted"`
}

type OrderItem struct {
	TicketTypeID int64   `json:"ticket_type_id"`
	Name         string  `json:"name"`
	Quantity     int     `json:"quantity"`
	UnitPrice    int64   `json:"unit_price"`
	SeatIDs      []int64 `json:"seat_ids"`
}

type Order struct {
	ID            int64       `json:"id"`
	EventID       int64       `json:"event_id"`
	SessionID     int64       `json:"session_id"`
	UserID        int64       `json:"user_id"`
	Status        string      `json:"status"`
	Currency      string      `json:"currency"`
	Subtotal      int64       `json:"subtotal"`
	Discount      int64       `json:"discount"`
	Total         int64       `json:"total"`
	BuyerName     string      `json:"buyer_name"`
	BuyerEmail    string      `json:"buyer_email"`
	ExpiresAt     time.Time   `json:"expires_at"`
	PaymentMethod string      `json:"payment_method"`
	PaidAt        *time.Time  `json:"paid_at"`
	RefundAmount  int64       `json:"refund_amount"`
	Items         []OrderItem `json:"items"`
	Tickets       []Ticket    `json:"tickets"`
}

type InvoiceLine struct {
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
	UnitPrice   int64  `json:"unit_price"`
	Amount      int64  `json:"amount"`
}

type Invoice struct {
	Number       string        `json:"number"`
	IssuedAt     time.Time     `json:"issued_at"`
	OrderID      int64         `json:"order_id"`
	SellerName   string        `json:"seller_name"`
	BuyerName    string        `json:"buyer_name"`
	BuyerEmail   string        `json:"buyer_email"`
	EventTitle   string        `json:"event_title"`
	Currency     string        `json:"currency"`
	Lines        []InvoiceLine `json:"lines"`
	Subtotal     int64         `json:"subtotal"`
	Discount     int64         `json:"discount"`
	Total        int64         `json:"total"`
	VATPercent   int           `json:"vat_percent"`
	VAT          int64         `json:"vat"`
	Status       string        `json:"status"`
	RefundAmount int64         `json:"refund_amount"`
}

type TicketType struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Price     int64  `json:"price"`
	Total     int    `json:"total"`
	Available int    `json:"available"`
	Sold      int    `json:"sold"`
	SessionID int64  `json:"session_id"`
	Held      int    `json:"held"`
	Seated    bool   `json:"seated"`
	OnSale    bool   `json:"on_sale"`
	SoldOut   bool   `json:"sold_out"`
}

type Event struct {
	ID           int64        `json:"id"`
	OrganizerID  int64        `json:"organizer_id"`
	Title        string       `json:"title"`
	Category     string       `json:"category"`
	City         string       `json:"city"`
	Venue        string       `json:"venue"`
	StartsAt     time.Time    `json:"starts_at"`
	EndsAt       time.Time    `json:"ends_at"`
	Currency     string       `json:"currency"`
	Status       string       `json:"status"`
	Transferable bool         `json:"transferable"`
	SoldOut      bool         `json:"sold_out"`
	TicketTypes  []TicketType `json:"ticket_types"`
	Sessions     []Session    `json:"sessions"`
}

type Session struct {
	ID           int64     `json:"id"`
	EventID      int64     `json:"event_id"`
	StartsAt     time.Time `json:"starts_at"`
	EndsAt       time.Time `json:"ends_at"`
	Label        string    `json:"label"`
	Status       string    `json:"status"`
	CancelReason string    `json:"cancel_reason"`
}

type Report struct {
	EventID     int64 `json:"event_id"`
	Orders      int   `json:"orders"`
	TicketsSold int   `json:"tickets_sold"`
	Revenue     int64 `json:"revenue"`
	Refunded    int64 `json:"refunded"`
	CheckedIn   int   `json:"checked_in"`
	Invited     int   `json:"invited"`
}

type Invitation struct {
	TicketTypeID int64   `json:"ticket_type_id"`
	Quantity     int     `json:"quantity"`
	SeatIDs      []int64 `json:"seat_ids,omitempty"`
	UserID       int64   `json:"user_id"`
	Name         string  `json:"name"`
	Email        string  `json:"email"`
	Note         string  `json:"note,omitempty"`
}

func (c *Client) TicketByCode(ctx context.Context, code string, eventID int64) (*Ticket, error) {
	q := url.Values{}
	if eventID != 0 {
		q.Set("event_id", strconv.FormatInt(eventID, 10))
	}
	var t Ticket
	return &t, c.do(ctx, http.MethodGet, "/ticket-codes/"+url.PathEscape(code), q, nil, &t)
}

func (c *Client) Ticket(ctx context.Context, id int64) (*Ticket, error) {
	var t Ticket
	return &t, c.do(ctx, http.MethodGet, fmt.Sprintf("/tickets/%d", id), nil, nil, &t)
}

func (c *Client) TicketHistory(ctx context.Context, id int64) ([]HistoryEntry, error) {
	var out struct {
		Items []HistoryEntry `json:"items"`
	}
	err := c.do(ctx, http.MethodGet, fmt.Sprintf("/tickets/%d/history", id), nil, nil, &out)
	return out.Items, err
}

func (c *Client) UserTickets(ctx context.Context, userID int64, f TicketFilter) (*TicketPage, error) {
	q := url.Values{}
	if f.EventID != 0 {
		q.Set("event_id", strconv.FormatInt(f.EventID, 10))
	}
	if f.Status != "" {
		q.Set("status", f.Status)
	}
	if f.Cursor != "" {
		q.Set("cursor", f.Cursor)
	}
	if f.Limit > 0 {
		q.Set("limit", strconv.Itoa(f.Limit))
	}
	var p TicketPage
	return &p, c.do(ctx, http.MethodGet, fmt.Sprintf("/users/%d/tickets", userID), q, nil, &p)
}

func (c *Client) VoidTicket(ctx context.Context, id int64, reason string) (*Ticket, error) {
	var t Ticket
	return &t, c.do(ctx, http.MethodPost, fmt.Sprintf("/tickets/%d/void", id), nil, map[string]string{"reason": reason}, &t)
}

func (c *Client) RevertCheckIn(ctx context.Context, id int64) (*Ticket, error) {
	var t Ticket
	return &t, c.do(ctx, http.MethodPost, fmt.Sprintf("/tickets/%d/revert-check-in", id), nil, nil, &t)
}

func (c *Client) CheckIn(ctx context.Context, eventID int64, code string) (*Ticket, error) {
	var t Ticket
	return &t, c.do(ctx, http.MethodPost, fmt.Sprintf("/events/%d/check-in", eventID), nil, map[string]string{"code": code}, &t)
}

func (c *Client) BatchCheckIn(ctx context.Context, eventID int64, scans []Scan) (*BatchResult, error) {
	var r BatchResult
	return &r, c.do(ctx, http.MethodPost, fmt.Sprintf("/events/%d/check-in/batch", eventID), nil, map[string]any{"scans": scans}, &r)
}

func (c *Client) Order(ctx context.Context, id int64) (*Order, error) {
	var o Order
	return &o, c.do(ctx, http.MethodGet, fmt.Sprintf("/orders/%d", id), nil, nil, &o)
}

func (c *Client) Invoice(ctx context.Context, orderID int64) (*Invoice, error) {
	var i Invoice
	return &i, c.do(ctx, http.MethodGet, fmt.Sprintf("/orders/%d/invoice", orderID), nil, nil, &i)
}

func (c *Client) CancelOrder(ctx context.Context, id int64) (*Order, error) {
	var o Order
	return &o, c.do(ctx, http.MethodPost, fmt.Sprintf("/orders/%d/cancel", id), nil, nil, &o)
}

func (c *Client) Event(ctx context.Context, id int64) (*Event, error) {
	var e Event
	return &e, c.do(ctx, http.MethodGet, fmt.Sprintf("/events/%d", id), nil, nil, &e)
}

func (c *Client) EventReport(ctx context.Context, id int64) (*Report, error) {
	var r Report
	return &r, c.do(ctx, http.MethodGet, fmt.Sprintf("/events/%d/report", id), nil, nil, &r)
}

func (c *Client) Invite(ctx context.Context, eventID int64, in Invitation) (*Order, error) {
	var o Order
	return &o, c.do(ctx, http.MethodPost, fmt.Sprintf("/events/%d/invitations", eventID), nil, in, &o)
}

func (c *Client) do(ctx context.Context, method, path string, q url.Values, in, out any) error {
	var body io.Reader
	if in != nil {
		b, err := json.Marshal(in)
		if err != nil {
			return err
		}
		body = bytes.NewReader(b)
	}
	u := c.base + "/internal/v1" + path
	if len(q) > 0 {
		u += "?" + q.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, u, body)
	if err != nil {
		return err
	}
	req.Header.Set("X-Internal-Key", c.key)
	if id, ok := ctx.Value(actingUser{}).(int64); ok && id > 0 {
		req.Header.Set("X-Acting-User", strconv.FormatInt(id, 10))
	}
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode/100 != 2 {
		var e struct {
			Error  string  `json:"error"`
			Ticket *Ticket `json:"ticket"`
		}
		_ = json.Unmarshal(raw, &e)
		if e.Error == "" {
			e.Error = http.StatusText(resp.StatusCode)
		}
		return &APIError{Status: resp.StatusCode, Message: e.Error, Ticket: e.Ticket}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(raw, out)
}
