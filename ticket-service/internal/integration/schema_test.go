package integration

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/ticket-service/internal/adapter/outbound/postgres"
)

// A database created before ticket lifecycle existed (no holder columns, tickets that were scanned marked only by
// checked_in_at) is upgraded in place by applying the schema again, and applying it once more changes nothing.
func TestSchemaUpgradeKeepsExistingTickets(t *testing.T) {
	d := newDatabase(t)
	ctx := context.Background()
	pool, err := postgres.OpenPool(ctx, d.dsn, 2)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	// take the database back to the earlier shape
	for _, q := range []string{
		`alter table ticket drop column holder_id, drop column holder_name, drop column holder_email, drop column transfer_count`,
		`alter table ticket_order drop column invited_by`,
		`alter table event drop column transferable`,
		// seats used to be unique by row and number only
		`drop index seat_position_uq`,
		`alter table seat add constraint seat_ticket_type_id_row_label_number_key unique (ticket_type_id, row_label, number)`,
		`insert into event (organizer_id, title, category, city, venue, starts_at, ends_at, status)
			values (1, 'Old event', 'music', 'Hanoi', 'Hall', now() + interval '1 day', now() + interval '2 days', 3)`,
		`insert into ticket_type (event_id, name, price, total, available, sold) values (1, 'GA', 100, 10, 8, 2)`,
		`insert into ticket_order (user_id, event_id, status, currency, subtotal, total, buyer_name, buyer_email, request_id, expires_at)
			values (77, 1, 2, 'VND', 200, 200, 'Old Buyer', 'old@example.com', 'r1', now())`,
		`insert into order_item (order_id, ticket_type_id, name, quantity, unit_price) values (1, 1, 'GA', 2, 100)`,
		`insert into ticket (order_id, event_id, ticket_type_id, code, status, checked_in_at) values (1, 1, 1, 'OLDCODE1', 1, now())`,
		`insert into ticket (order_id, event_id, ticket_type_id, code, status) values (1, 1, 1, 'OLDCODE2', 1)`,
	} {
		if _, err := pool.Exec(ctx, q); err != nil {
			t.Fatalf("%s: %v", q, err)
		}
	}

	for i := 0; i < 2; i++ { // the second run must be a no-op
		if err := postgres.ApplySchema(ctx, pool); err != nil {
			t.Fatalf("apply %d: %v", i+1, err)
		}
	}
	var holder int64
	var name, email string
	var status int
	err = pool.QueryRow(ctx, `select holder_id, holder_name, holder_email, status from ticket where code = 'OLDCODE1'`).Scan(&holder, &name, &email, &status)
	if err != nil || holder != 77 || name != "Old Buyer" || email != "old@example.com" {
		t.Fatalf("old ticket after upgrade: holder=%d name=%q email=%q (%v)", holder, name, email, err)
	}
	if status != 3 {
		t.Fatalf("a ticket that had been scanned must be used (3), is %d", status)
	}
	if err := pool.QueryRow(ctx, `select status from ticket where code = 'OLDCODE2'`).Scan(&status); err != nil || status != 1 {
		t.Fatalf("an unscanned ticket stays valid: %d %v", status, err)
	}
	// two sections of one ticket type may now each have a row A
	for _, sec := range []string{"L", "R"} {
		if _, err := pool.Exec(ctx, `insert into seat (ticket_type_id, section, row_label, number) values (1, $1, 'A', 1)`, sec); err != nil {
			t.Fatalf("section %s row A after the upgrade: %v", sec, err)
		}
	}
	if _, err := pool.Exec(ctx, `insert into seat (ticket_type_id, section, row_label, number) values (1, 'L', 'A', 1)`); err == nil {
		t.Fatal("the same seat twice must still be refused")
	}
	var transferable bool
	if err := pool.QueryRow(ctx, `select transferable from event where id = 1`).Scan(&transferable); err != nil || !transferable {
		t.Fatalf("existing events stay transferable: %v %v", transferable, err)
	}
}
