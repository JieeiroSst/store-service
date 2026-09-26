package integration

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/ticket-service/pkg/ticketclient"
)

// raw is a request whose answer is not JSON.
func (a *api) raw(method, path string, as int64, headers map[string]string) (int, http.Header, string) {
	a.t.Helper()
	req, _ := http.NewRequest(method, a.base+path, nil)
	if as != 0 {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer user-%d", as))
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := a.c.Do(req)
	if err != nil {
		a.t.Fatal(err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, resp.Header, string(b)
}

// publishedEvent builds and publishes an event over the API and returns its id and the id of a GA ticket type.
func publishedEvent(a *api, total int) (eventID, typeID int64) {
	const org, adm = int64(1_000_000_002), int64(1_000_000_001)
	starts := time.Now().Add(96 * time.Hour).UTC().Truncate(time.Second)
	ev := a.must(201, "POST", "/api/v1/events", org, map[string]any{
		"title": "Sơn Tùng M-TP Concert", "category": "music", "city": "Hồ Chí Minh", "venue": "Thống Nhất", "wallet_id": "org-wallet",
		"starts_at": starts, "ends_at": starts.Add(3 * time.Hour)})
	if ev["transferable"] != true {
		a.t.Fatalf("events are transferable unless the organizer says otherwise: %v", ev["transferable"])
	}
	eventID = id(ev)
	ep := fmt.Sprintf("/api/v1/events/%d", eventID)
	typeID = id(a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "GA", "price": 500000, "total": total, "max_per_order": 5}))
	a.must(200, "POST", ep+"/submit", org, nil)
	a.must(200, "POST", ep+"/approve", adm, nil)
	return eventID, typeID
}

func buy(a *api, u, eventID, typeID int64, qty int) map[string]any {
	o := a.must(201, "POST", "/api/v1/orders", u, map[string]any{"event_id": eventID, "buyer_name": "Buyer", "request_id": fmt.Sprintf("buy-%d-%d", u, time.Now().UnixNano()),
		"items": []map[string]any{{"ticket_type_id": typeID, "quantity": qty}}})
	return a.must(200, "POST", fmt.Sprintf("/api/v1/orders/%d/pay", id(o)), u, map[string]any{"method": "wallet"})
}

func TestHTTPTicketLifecycle(t *testing.T) {
	a := startApp(t, nil)
	const org = int64(1_000_000_002)
	eid, tid := publishedEvent(a, 10)
	ep := fmt.Sprintf("/api/v1/events/%d", eid)

	// ---- my tickets
	paid := buy(a, 5, eid, tid, 2)
	tickets := paid["tickets"].([]any)
	t1, t2 := tickets[0].(map[string]any), tickets[1].(map[string]any)
	list := a.must(200, "GET", "/api/v1/tickets?status=valid", 5, nil)["items"].([]any)
	if len(list) != 2 || list[0].(map[string]any)["event_title"] != "Sơn Tùng M-TP Concert" || list[0].(map[string]any)["code"] == "" {
		t.Fatalf("my tickets: %v", list)
	}
	a.must(200, "GET", "/api/v1/my/tickets", 5, nil) // the older path still works
	a.must(400, "GET", "/api/v1/tickets?status=nonsense", 5, nil)
	a.must(401, "GET", "/api/v1/tickets", 0, nil)
	tp := fmt.Sprintf("/api/v1/tickets/%d", id(t1))
	a.must(404, "GET", tp, 6, nil)
	a.must(200, "GET", tp, org, nil)
	a.must(200, "PUT", tp+"/holder", 5, map[string]any{"name": "=HYPERLINK(\"http://evil\")", "email": "guest@example.com"})
	a.must(403, "PUT", tp+"/holder", org, map[string]any{"name": "x"})
	old := t1["code"].(string)
	re := a.must(200, "POST", tp+"/reissue", 5, nil)
	if re["code"] == old {
		t.Fatal("reissue kept the code")
	}

	// ---- transfer: the recipient never sees the code before accepting
	off := a.must(201, "POST", tp+"/transfer", 5, map[string]any{"to_user_id": 6, "message": "have fun"})
	oid := id(off)
	if off["status"] != "pending" || off["ticket"].(map[string]any)["code"] != nil {
		t.Fatalf("the offer leaks the code: %v", off)
	}
	a.must(409, "POST", tp+"/transfer", 5, map[string]any{"to_user_id": 7})
	in := a.must(200, "GET", "/api/v1/transfers?direction=incoming", 6, nil)["items"].([]any)
	if len(in) != 1 || in[0].(map[string]any)["ticket"].(map[string]any)["code"] != nil {
		t.Fatalf("incoming offers: %v", in)
	}
	out := a.must(200, "GET", "/api/v1/transfers", 5, nil)["items"].([]any)
	if len(out) != 1 {
		t.Fatalf("outgoing offers: %v", out)
	}
	a.must(400, "GET", "/api/v1/transfers?direction=sideways", 5, nil)
	op := fmt.Sprintf("/api/v1/transfers/%d", oid)
	a.must(404, "POST", op+"/accept", 7, nil)
	got := a.must(200, "POST", op+"/accept", 6, nil)
	if got["holder_id"].(float64) != 6 || got["code"] == "" || got["code"] == re["code"] || got["transfer_count"].(float64) != 1 {
		t.Fatalf("accepted ticket: %v", got)
	}
	a.must(409, "POST", op+"/accept", 6, nil)
	a.must(404, "GET", tp, 5, nil)
	hist := a.must(200, "GET", tp+"/history", 6, nil)["items"].([]any)
	var events []string
	for _, h := range hist {
		events = append(events, h.(map[string]any)["event"].(string))
	}
	if strings.Join(events, ",") != "issued,holder_changed,reissued,transfer_offered,transfer_accepted" {
		t.Fatalf("history: %v", events)
	}
	// decline and cancel
	o2 := a.must(201, "POST", fmt.Sprintf("/api/v1/tickets/%d/transfer", id(t2)), 5, map[string]any{"to_email": "user7@example.com"})
	a.must(200, "POST", fmt.Sprintf("/api/v1/transfers/%d/decline", id(o2)), 7, nil) // user 7's e-mail in the fake user-service
	o3 := a.must(201, "POST", fmt.Sprintf("/api/v1/tickets/%d/transfer", id(t2)), 5, map[string]any{"to_user_id": 8})
	if c := a.must(200, "POST", fmt.Sprintf("/api/v1/transfers/%d/cancel", id(o3)), 5, nil); c["status"] != "cancelled" {
		t.Fatalf("cancel: %v", c)
	}

	// ---- staff, lookup, scan, batch, revert
	a.must(403, "POST", ep+"/staff", 6, map[string]any{"user_id": 40})
	a.must(204, "POST", ep+"/staff", org, map[string]any{"user_id": 40})
	if st := a.must(200, "GET", ep+"/staff", org, nil)["items"].([]any); len(st) != 1 {
		t.Fatalf("staff: %v", st)
	}
	if mine := a.must(200, "GET", "/api/v1/my/staff-events", 40, nil)["items"].([]any); len(mine) != 1 {
		t.Fatalf("staff events: %v", mine)
	}
	a.must(403, "GET", ep+"/report", 40, nil)
	code2 := t2["code"].(string)
	if lk := a.must(200, "GET", ep+"/tickets/"+code2, 40, nil); lk["status"] != "valid" || lk["type_name"] != "GA" {
		t.Fatalf("lookup: %v", lk)
	}
	a.must(404, "GET", ep+"/tickets/NOSUCH", 40, nil)
	a.must(403, "GET", ep+"/tickets/"+code2, 6, nil)
	a.must(200, "POST", ep+"/check-in", 40, map[string]any{"code": code2})
	if st, out := a.do("POST", ep+"/check-in", 40, map[string]any{"code": code2}); st != 409 || out["ticket"] == nil {
		t.Fatalf("second scan: %d %v", st, out)
	}
	b := a.must(200, "POST", ep+"/check-in/batch", 40, map[string]any{"scans": []map[string]any{
		{"code": code2}, {"code": got["code"]}, {"code": "NOPE"}}})
	res := b["results"].([]any)
	if b["admitted"].(float64) != 1 || res[0].(map[string]any)["result"] != "already_used" || res[1].(map[string]any)["result"] != "admitted" || res[2].(map[string]any)["result"] != "not_found" {
		t.Fatalf("batch: %v", b)
	}
	if res[1].(map[string]any)["ticket"].(map[string]any)["code"] != nil {
		t.Fatal("a batch answer repeats the codes back")
	}
	a.must(400, "POST", ep+"/check-in/batch", 40, map[string]any{"scans": []any{}})
	a.must(403, "POST", fmt.Sprintf("/api/v1/tickets/%d/revert-check-in", id(t2)), 40, nil)
	if r := a.must(200, "POST", fmt.Sprintf("/api/v1/tickets/%d/revert-check-in", id(t2)), org, nil); r["status"] != "valid" {
		t.Fatalf("revert: %v", r)
	}
	a.must(400, "POST", fmt.Sprintf("/api/v1/tickets/%d/void", id(t2)), org, map[string]any{"reason": ""})
	if v := a.must(200, "POST", fmt.Sprintf("/api/v1/tickets/%d/void", id(t2)), org, map[string]any{"reason": "duplicate"}); v["status"] != "void" {
		t.Fatalf("void: %v", v)
	}

	// ---- attendee export
	st, hdr, body := a.raw("GET", ep+"/attendees.csv", org, nil)
	if st != 200 || !strings.HasPrefix(hdr.Get("Content-Type"), "text/csv") {
		t.Fatalf("csv: %d %v", st, hdr)
	}
	rows, err := csv.NewReader(strings.NewReader(body)).ReadAll()
	if err != nil || len(rows) != 3 || rows[0][0] != "ticket_id" {
		t.Fatalf("csv rows: %v %v", rows, err)
	}
	if strings.Contains(body, ",=HYPERLINK") || strings.Contains(body, "\"=HYPERLINK") {
		t.Fatal("a spreadsheet formula got into the export")
	}
	if st, _, _ := a.raw("GET", ep+"/attendees.csv", 5, nil); st != 403 {
		t.Fatalf("csv for a buyer: %d", st)
	}

	// ---- invitations and waiting lists
	inv := map[string]any{"ticket_type_id": tid, "quantity": 2, "user_id": 50, "name": "Sponsor", "email": "sponsor@example.com"}
	a.must(403, "POST", ep+"/invitations", 5, inv)
	if o := a.must(201, "POST", ep+"/invitations", org, inv); o["status"] != "paid" || o["total"].(float64) != 0 || len(o["tickets"].([]any)) != 2 {
		t.Fatalf("invitation: %v", o)
	}
	a.must(409, "PUT", fmt.Sprintf("%s/ticket-types/%d/waitlist", ep, tid), 5, nil) // tickets are still available
	a.must(200, "GET", "/api/v1/waitlist", 5, nil)
	rep := a.must(200, "GET", ep+"/report", org, nil)
	if rep["invited"].(float64) != 2 || len(rep["daily"].([]any)) != 1 {
		t.Fatalf("report: %v", rep)
	}
}

func TestInternalAPIAndClient(t *testing.T) {
	a := startApp(t, map[string]string{"InternalAPIKey": "cluster-secret"})
	const org = int64(1_000_000_002)
	eid, tid := publishedEvent(a, 6)
	paid := buy(a, 5, eid, tid, 2)
	orderID := id(paid)
	tk := paid["tickets"].([]any)[0].(map[string]any)
	code := tk["code"].(string)

	// no key, a wrong key and a user's token are all refused
	if st, _, _ := a.raw("GET", "/internal/v1/events/1", 0, nil); st != 401 {
		t.Fatalf("no key: %d", st)
	}
	if st, _, _ := a.raw("GET", "/internal/v1/events/1", 0, map[string]string{"X-Internal-Key": "guess"}); st != 401 {
		t.Fatalf("wrong key: %d", st)
	}
	if st, _, _ := a.raw("GET", "/internal/v1/events/1", 1_000_000_001, nil); st != 401 {
		t.Fatalf("an admin's user token is not the service key: %d", st)
	}
	// and the public API is not reachable with the key alone
	if st, _, _ := a.raw("GET", "/api/v1/tickets", 0, map[string]string{"X-Internal-Key": "cluster-secret"}); st != 401 {
		t.Fatalf("service key on the public API: %d", st)
	}

	c := ticketclient.New(a.base, "cluster-secret")
	ctx := ticketclient.WithActingUser(context.Background(), 777)

	ev, err := c.Event(ctx, eid)
	if err != nil || ev.Title == "" || len(ev.TicketTypes) != 1 || ev.TicketTypes[0].Sold != 2 || ev.TicketTypes[0].Available != 4 || !ev.Transferable {
		t.Fatalf("event: %+v %v", ev, err)
	}
	got, err := c.TicketByCode(ctx, code, 0)
	if err != nil || got.HolderID != 5 || got.Status != "valid" {
		t.Fatalf("by code: %+v %v", got, err)
	}
	if _, err := c.TicketByCode(ctx, "NOSUCH", 0); !ticketclient.IsNotFound(err) {
		t.Fatalf("unknown code: %v", err)
	}
	if _, err := c.TicketByCode(ctx, code, eid+100); !ticketclient.IsNotFound(err) {
		t.Fatalf("code of another event: %v", err)
	}
	if got, err := c.Ticket(ctx, got.ID); err != nil || got.Code != code {
		t.Fatalf("ticket: %+v %v", got, err)
	}
	page, err := c.UserTickets(ctx, 5, ticketclient.TicketFilter{EventID: eid, Status: "valid", Limit: 1})
	if err != nil || len(page.Items) != 1 || page.NextCursor == "" {
		t.Fatalf("user tickets: %+v %v", page, err)
	}
	if page, err = c.UserTickets(ctx, 5, ticketclient.TicketFilter{Cursor: page.NextCursor}); err != nil || len(page.Items) != 1 {
		t.Fatalf("second page: %+v %v", page, err)
	}
	if none, err := c.UserTickets(ctx, 99, ticketclient.TicketFilter{}); err != nil || len(none.Items) != 0 {
		t.Fatalf("a user without tickets: %+v %v", none, err)
	}

	// scanning through the cluster API records who acted
	if _, err := c.CheckIn(ctx, eid, code); err != nil {
		t.Fatal(err)
	}
	_, err = c.CheckIn(ctx, eid, code)
	var apiErr *ticketclient.APIError
	if !ticketclient.IsAlreadyCheckedIn(err) || !ticketclient.IsConflict(err) {
		t.Fatalf("second scan: %v", err)
	}
	if e, ok := err.(*ticketclient.APIError); ok {
		apiErr = e
	}
	if apiErr == nil || apiErr.Ticket.CheckedInAt == nil {
		t.Fatalf("the error must carry the ticket: %+v", apiErr)
	}
	hist, err := c.TicketHistory(ctx, got.ID)
	if err != nil || len(hist) != 2 || hist[1].Event != "checked_in" || hist[1].Actor != 777 {
		t.Fatalf("history: %+v %v", hist, err)
	}
	if r, err := c.RevertCheckIn(ctx, got.ID); err != nil || r.Status != "valid" {
		t.Fatalf("revert: %+v %v", r, err)
	}
	br, err := c.BatchCheckIn(ctx, eid, []ticketclient.Scan{{Code: code}, {Code: "NOPE"}})
	if err != nil || br.Admitted != 1 || br.Results[1].Result != "not_found" {
		t.Fatalf("batch: %+v %v", br, err)
	}
	if v, err := c.VoidTicket(ctx, got.ID, "fraud team"); err == nil {
		t.Fatalf("voiding a used ticket: %+v", v)
	} else if !ticketclient.IsConflict(err) {
		t.Fatalf("voiding a used ticket: %v", err)
	}

	// orders, invoices, invitations and reports
	if inv, err := c.Invoice(ctx, orderID); err != nil || inv.Status != "paid" || inv.Total != 1_000_000 || len(inv.Lines) != 1 || inv.EventTitle == "" {
		t.Fatalf("invoice: %+v %v", inv, err)
	}
	o, err := c.Order(ctx, orderID)
	if err != nil || o.Status != "paid" || len(o.Tickets) != 2 || o.UserID != 5 {
		t.Fatalf("order: %+v %v", o, err)
	}
	inv, err := c.Invite(ctx, eid, ticketclient.Invitation{TicketTypeID: tid, Quantity: 1, UserID: 60, Name: "Reward winner", Email: "winner@example.com", Note: "Loyalty reward"})
	if err != nil || inv.Total != 0 || len(inv.Tickets) != 1 || inv.UserID != 60 {
		t.Fatalf("invite: %+v %v", inv, err)
	}
	if _, err := c.Invite(ctx, eid, ticketclient.Invitation{TicketTypeID: tid, Quantity: 50, UserID: 61, Name: "x", Email: "x@example.com"}); !ticketclient.IsConflict(err) {
		t.Fatalf("inviting beyond the stock: %v", err)
	}
	rep, err := c.EventReport(ctx, eid)
	if err != nil || rep.TicketsSold != 3 || rep.Invited != 1 || rep.CheckedIn != 1 {
		t.Fatalf("report: %+v %v", rep, err)
	}
	// a ticket revoked through the cluster: the holder cannot get in
	free := inv.Tickets[0]
	if v, err := c.VoidTicket(ctx, free.ID, "reward withdrawn"); err != nil || v.Status != "void" {
		t.Fatalf("void: %+v %v", v, err)
	}
	// the order is refunded from another service
	if _, err := c.CancelOrder(ctx, orderID); !ticketclient.IsConflict(err) {
		t.Fatalf("refunding an order with a used ticket: %v", err)
	}
	_ = org
}

func TestInternalAPIIsOffWithoutAKey(t *testing.T) {
	a := startApp(t, nil) // InternalAPIKey unset
	if st, _, _ := a.raw("GET", "/internal/v1/events/1", 0, map[string]string{"X-Internal-Key": ""}); st != 404 {
		t.Fatalf("the internal API must not exist without a key: %d", st)
	}
}

func TestHTTPResaleInvoiceSeatMapSeries(t *testing.T) {
	a := startApp(t, map[string]string{"PlatformWalletID": "platform", "ResaleFeePercent": "10", "VATPercent": "10", "InvoiceSellerName": "Ticket Co"})
	const org, adm = int64(1_000_000_002), int64(1_000_000_001)
	starts := time.Now().Add(96 * time.Hour).UTC().Truncate(time.Second)
	ev := a.must(201, "POST", "/api/v1/events", org, map[string]any{
		"title": "Hà Anh Tuấn Live", "category": "music", "city": "Hà Nội", "venue": "Mỹ Đình", "wallet_id": "org-wallet",
		"starts_at": starts, "ends_at": starts.Add(3 * time.Hour), "resale_cap_percent": 100})
	if ev["resale_cap_percent"].(float64) != 100 {
		t.Fatalf("event: %v", ev)
	}
	eid := id(ev)
	ep := fmt.Sprintf("/api/v1/events/%d", eid)
	tid := id(a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "GA", "price": 500000, "total": 10}))
	a.must(400, "POST", "/api/v1/events", org, map[string]any{"title": "x y z", "category": "music", "city": "a", "venue": "b",
		"starts_at": starts, "ends_at": starts.Add(time.Hour), "resale_cap_percent": 999})

	// ---- a seat map
	seated := id(a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "Stalls", "price": 900000, "seated": true}))
	sp := fmt.Sprintf("%s/ticket-types/%d", ep, seated)
	a.must(201, "POST", sp+"/seats", org, map[string]any{"section": "Stalls", "rows": 2, "seats_per_row": 3, "offset_x": 100, "offset_y": 200})
	if pl := a.must(200, "PUT", sp+"/seat-map", org, map[string]any{"width": 800, "height": 600,
		"stage":    map[string]any{"label": "STAGE", "x": 250, "y": 20, "w": 300, "h": 60},
		"sections": []any{map[string]any{"name": "Stalls", "points": [][]float64{{80, 180}, {400, 180}, {400, 300}, {80, 300}}}},
		"seats":    []any{map[string]any{"row": "a", "number": 1, "x": 150, "y": 220}}}); pl["seats_placed"].(float64) != 1 {
		t.Fatalf("placed: %v", pl)
	}
	a.must(400, "PUT", sp+"/seat-map", org, map[string]any{"width": 10, "height": 10, "sections": []any{map[string]any{"name": "x", "points": [][]float64{{0, 0}, {99, 0}, {5, 5}}}}})
	a.must(404, "PUT", sp+"/seat-map", 5, map[string]any{"width": 10, "height": 10}) // a draft is invisible to strangers
	a.must(200, "POST", ep+"/submit", org, nil)
	a.must(200, "POST", ep+"/approve", adm, nil)
	m := a.must(200, "GET", sp+"/seats", 0, nil)
	seats := m["items"].([]any)
	if lay := m["layout"].(map[string]any); lay["width"].(float64) != 800 || lay["stage"].(map[string]any)["label"] != "STAGE" || len(seats) != 6 {
		t.Fatalf("seat map: %v", m)
	}
	if s0 := seats[0].(map[string]any); s0["x"].(float64) != 150 || s0["y"].(float64) != 220 || s0["label"] != "Stalls A-1" {
		t.Fatalf("seat A-1: %v", s0)
	}

	// ---- an invoice, and a hostile buyer name
	o := a.must(201, "POST", "/api/v1/orders", 5, map[string]any{"event_id": eid, "buyer_name": "<script>alert(1)</script>", "request_id": "inv1",
		"items": []map[string]any{{"ticket_type_id": tid, "quantity": 2}}})
	a.must(409, "GET", fmt.Sprintf("/api/v1/orders/%d/invoice", id(o)), 5, nil) // unpaid
	paid := a.must(200, "POST", fmt.Sprintf("/api/v1/orders/%d/pay", id(o)), 5, map[string]any{"method": "wallet"})
	ip := fmt.Sprintf("/api/v1/orders/%d/invoice", id(o))
	inv := a.must(200, "GET", ip, 5, nil)
	if inv["total"].(float64) != 1_000_000 || inv["vat"].(float64) != 90_909 || inv["seller_name"] != "Ticket Co" || len(inv["lines"].([]any)) != 1 ||
		!strings.HasPrefix(inv["number"].(string), "INV-") {
		t.Fatalf("invoice: %v", inv)
	}
	a.must(404, "GET", ip, 6, nil)
	st, hdr, body := a.raw("GET", ip+"?format=html", 5, nil)
	if st != 200 || !strings.HasPrefix(hdr.Get("Content-Type"), "text/html") || !strings.Contains(body, "1,000,000") || !strings.Contains(body, "Ticket Co") {
		t.Fatalf("html invoice: %d %q", st, body)
	}
	if strings.Contains(body, "<script>alert") || !strings.Contains(body, "&lt;script&gt;") {
		t.Fatalf("the buyer's name was not escaped in the invoice:\n%s", body)
	}

	// ---- resale
	tk := paid["tickets"].([]any)[0].(map[string]any)
	rp := fmt.Sprintf("/api/v1/tickets/%d/resale", id(tk))
	a.must(400, "POST", rp, 5, map[string]any{"price": 500001})
	a.must(404, "POST", rp, 6, map[string]any{"price": 100})
	l := a.must(201, "POST", rp, 5, map[string]any{"price": 450000})
	lp := fmt.Sprintf("/api/v1/resale/%d", id(l))
	if l["status"] != "open" || l["face_price"].(float64) != 500000 || l["seller_id"].(float64) != 5 {
		t.Fatalf("listing: %v", l)
	}
	pub := a.must(200, "GET", ep+"/resale?sort=price", 0, nil)["items"].([]any)
	if len(pub) != 1 || pub[0].(map[string]any)["code"] != nil || pub[0].(map[string]any)["seller_id"] != nil || pub[0].(map[string]any)["price"].(float64) != 450000 {
		t.Fatalf("public listings (no code, no seller): %v", pub)
	}
	a.must(200, "GET", "/api/v1/resale?status=open", 5, nil)
	a.must(400, "GET", "/api/v1/resale?status=weird", 5, nil)
	a.must(400, "POST", lp+"/buy", 5, nil) // your own
	a.must(401, "POST", lp+"/buy", 0, nil)
	bought := a.must(200, "POST", lp+"/buy", 6, nil)
	if bought["holder_id"].(float64) != 6 || bought["code"] == tk["code"] || bought["status"] != "valid" {
		t.Fatalf("bought: %v", bought)
	}
	a.must(409, "POST", lp+"/buy", 7, nil)
	a.must(409, "DELETE", lp, 5, nil)
	if pub = a.must(200, "GET", ep+"/resale", 0, nil)["items"].([]any); len(pub) != 0 {
		t.Fatalf("a sold ticket is still listed: %v", pub)
	}
	a.must(200, "POST", ep+"/check-in", org, map[string]any{"code": bought["code"]})

	// ---- another showtime
	next := starts.Add(24 * time.Hour)
	dup := a.must(201, "POST", ep+"/duplicate", org, map[string]any{"title": "Hà Anh Tuấn Live - đêm 2", "starts_at": next, "ends_at": next.Add(3 * time.Hour), "copy_promotions": true})
	if dup["status"] != "draft" || dup["series_id"].(float64) != float64(eid) || len(dup["ticket_types"].([]any)) != 2 || dup["resale_cap_percent"].(float64) != 100 {
		t.Fatalf("duplicate: %v", dup)
	}
	dp := fmt.Sprintf("/api/v1/events/%d", id(dup))
	a.must(403, "POST", ep+"/duplicate", 5, map[string]any{"starts_at": next, "ends_at": next.Add(time.Hour)})
	a.must(400, "POST", ep+"/duplicate", org, map[string]any{"starts_at": next, "ends_at": next.Add(-time.Hour)})
	a.must(200, "POST", dp+"/submit", org, nil)
	a.must(200, "POST", dp+"/approve", adm, nil)
	series := a.must(200, "GET", ep+"/series", 0, nil)["items"].([]any)
	if len(series) != 2 || series[0].(map[string]any)["id"].(float64) != float64(eid) || series[1].(map[string]any)["id"].(float64) != float64(id(dup)) {
		t.Fatalf("series: %v", series)
	}
	// fresh stock at the second showtime
	if d2 := a.must(200, "GET", dp, 0, nil); d2["ticket_types"].([]any)[0].(map[string]any)["sold"].(float64) != 0 {
		t.Fatalf("copied stock: %v", d2["ticket_types"])
	}
}

func TestHTTPVenues(t *testing.T) {
	a := startApp(t, nil)
	const org, adm = int64(1_000_000_002), int64(1_000_000_001)

	// the templates say what they take
	ts := a.must(200, "GET", "/api/v1/venue-templates", 0, nil)["items"].([]any)
	names := map[string]bool{}
	for _, x := range ts {
		tm := x.(map[string]any)
		names[tm["name"].(string)] = len(tm["params"].([]any)) > 0
	}
	if !names["theatre"] || !names["hall"] || !names["arena"] {
		t.Fatalf("templates: %v", ts)
	}

	// preview: counts and, on request, the seats with their places; nothing is saved
	pv := a.must(200, "POST", "/api/v1/venues/preview?seats=true", org, map[string]any{"template": "arena", "params": map[string]any{"rows": 2, "seats_per_row": 3}})
	if pv["total"].(float64) != 24 || len(pv["sections"].([]any)) != 4 || len(pv["seats"].([]any)) != 24 {
		t.Fatalf("preview: %v", pv)
	}
	if lay := pv["layout"].(map[string]any); lay["stages"].([]any)[0].(map[string]any)["shape"] != "ellipse" {
		t.Fatalf("layout: %v", lay)
	}
	if pv = a.must(200, "POST", "/api/v1/venues/preview", org, map[string]any{"template": "hall"}); pv["seats"] != nil {
		t.Fatal("seats were not asked for")
	}
	a.must(400, "POST", "/api/v1/venues/preview", org, map[string]any{"template": "hall", "params": map[string]any{"rows": -1}})
	a.must(400, "POST", "/api/v1/venues/preview", org, map[string]any{"map": map[string]any{"width": 10, "height": 10, "sections": []any{}}})
	a.must(401, "POST", "/api/v1/venues/preview", 0, map[string]any{"template": "hall"})

	// a hall drawn by hand, with a fan-shaped balcony
	v := a.must(201, "POST", "/api/v1/venues", org, map[string]any{"name": "Grand Théâtre", "city": "Hà Nội", "map": map[string]any{
		"width": 900, "height": 700,
		"stages": []any{map[string]any{"label": "STAGE", "x": 350, "y": 10, "w": 200, "h": 40}},
		"sections": []any{
			map[string]any{"key": "STALLS", "name": "Stalls", "color": "#c33", "layout": "grid", "origin": map[string]any{"x": 300, "y": 120}, "row_seats": []int{8, 10, 10, 12}, "aisle_after": []int{5}},
			map[string]any{"key": "BALC", "name": "Balcony", "color": "#36c", "layout": "arc", "center": map[string]any{"x": 450, "y": 60}, "radius": 380, "rows": 3, "angle_start": 40, "angle_end": 140}}}})
	vid := id(v)
	if v["sections"].(float64) != 2 || v["seats"].(float64) <= 40 {
		t.Fatalf("venue: %v", v)
	}
	vp := fmt.Sprintf("/api/v1/venues/%d", vid)
	if got := a.must(200, "GET", vp+"?seats=true", org, nil); len(got["preview"].(map[string]any)["seats"].([]any)) != int(v["seats"].(float64)) {
		t.Fatalf("venue with seats: %v", got["preview"].(map[string]any)["total"])
	}
	a.must(404, "GET", vp, 5, nil)
	a.must(200, "PUT", vp+"/shared", adm, map[string]any{"shared": true})
	a.must(200, "GET", vp, 5, nil)
	a.must(403, "PUT", vp, 5, map[string]any{"name": "Hijack", "template": "hall"})
	a.must(400, "GET", "/api/v1/venues?scope=nowhere", org, nil)
	if lib := a.must(200, "GET", "/api/v1/venues?scope=shared", 5, nil)["items"].([]any); len(lib) != 1 || lib[0].(map[string]any)["map"] != nil {
		t.Fatalf("library (a list carries no blueprint): %v", lib)
	}
	cp := a.must(201, "POST", vp+"/copy", 5, map[string]any{"name": "Our copy"})
	if cp["name"] != "Our copy" || cp["owner_id"].(float64) != 5 {
		t.Fatalf("copy: %v", cp)
	}

	// an event with two ticket types: stalls and balcony
	starts := time.Now().Add(96 * time.Hour).UTC().Truncate(time.Second)
	ev := a.must(201, "POST", "/api/v1/events", org, map[string]any{"title": "Swan Lake", "category": "arts", "city": "Hà Nội", "venue": "Grand Théâtre",
		"wallet_id": "w", "starts_at": starts, "ends_at": starts.Add(3 * time.Hour)})
	eid := id(ev)
	ep := fmt.Sprintf("/api/v1/events/%d", eid)
	stalls := id(a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "Stalls", "price": 1200000, "seated": true}))
	balc := id(a.must(201, "POST", ep+"/ticket-types", org, map[string]any{"name": "Balcony", "price": 600000, "seated": true}))
	a.must(404, "POST", ep+"/seat-map/apply", 5, map[string]any{"venue_id": vid}) // a draft is not theirs
	a.must(400, "POST", ep+"/seat-map/apply", org, map[string]any{"venue_id": vid, "assignments": []any{}})
	a.must(400, "POST", ep+"/seat-map/apply", org, map[string]any{"venue_id": vid, "assignments": []any{map[string]any{"section": "PIT", "ticket_type_id": stalls}}})
	res := a.must(200, "POST", ep+"/seat-map/apply", org, map[string]any{"venue_id": vid, "assignments": []any{
		map[string]any{"section": "STALLS", "ticket_type_id": stalls}, map[string]any{"section": "BALC", "ticket_type_id": balc}}})
	if res["seats"].(float64) != v["seats"].(float64) || len(res["sections"].([]any)) != 2 {
		t.Fatalf("apply: %v", res)
	}
	a.must(409, "POST", ep+"/seat-map/apply", org, map[string]any{"venue_id": vid, "assignments": []any{map[string]any{"section": "STALLS", "ticket_type_id": stalls}}})
	a.must(404, "GET", ep+"/seat-map", 0, nil) // still a draft
	a.must(200, "POST", ep+"/submit", org, nil)
	a.must(200, "POST", ep+"/approve", adm, nil)
	if got := a.must(200, "GET", ep, 0, nil); got["venue_id"] != nil {
		t.Fatalf("the public event page shows the venue id? %v", got["venue_id"])
	}

	// one call gives the whole map
	m := a.must(200, "GET", ep+"/seat-map", 0, nil)
	lay := m["layout"].(map[string]any)
	if len(lay["sections"].([]any)) != 2 || len(lay["stages"].([]any)) != 1 || len(m["ticket_types"].([]any)) != 2 {
		t.Fatalf("seat map: %v", m)
	}
	sold := map[string]float64{}
	for _, x := range lay["sections"].([]any) {
		sec := x.(map[string]any)
		sold[sec["key"].(string)] = sec["ticket_type_id"].(float64)
		if len(sec["points"].([]any)) < 3 {
			t.Fatalf("section without an outline: %v", sec)
		}
	}
	if sold["STALLS"] != float64(stalls) || sold["BALC"] != float64(balc) {
		t.Fatalf("who sells what: %v", sold)
	}
	seats := m["seats"].([]any)
	var pick []int64
	for _, x := range seats {
		s := x.(map[string]any)
		if s["section"] == "BALC" && len(pick) < 2 {
			pick = append(pick, int64(s["id"].(float64)))
		}
		if s["x"] == nil || s["y"] == nil || s["ticket_type_id"] == nil {
			t.Fatalf("seat without a place or a type: %v", s)
		}
	}
	// buy two balcony seats straight from the map
	o := a.must(201, "POST", "/api/v1/orders", 5, map[string]any{"event_id": eid, "buyer_name": "B", "request_id": "map-1",
		"items": []any{map[string]any{"ticket_type_id": balc, "quantity": 2, "seat_ids": pick}}})
	if o["total"].(float64) != 1_200_000 {
		t.Fatalf("order: %v", o)
	}
	// the map shows them held (the public copy is cached for a moment)
	held := 0
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline) && held < 2; time.Sleep(200 * time.Millisecond) {
		held = 0
		for _, x := range a.must(200, "GET", ep+"/seat-map", 0, nil)["seats"].([]any) {
			if x.(map[string]any)["status"] == "held" {
				held++
			}
		}
	}
	if held != 2 {
		t.Fatalf("%d seats held on the map, want 2", held)
	}
	a.must(204, "DELETE", vp, org, nil)
	a.must(404, "GET", vp, org, nil)
	if m2 := a.must(200, "GET", ep+"/seat-map", 0, nil); len(m2["seats"].([]any)) != len(seats) {
		t.Fatal("deleting the venue took seats from an event")
	}
}
