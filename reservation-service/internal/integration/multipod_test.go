package integration

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSSt/reservation-service/internal/adapter/outbound/postgres"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/inbound"
	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// The core promise: however many users, on however many pods, at most as many reservations succeed as there are rooms.

func TestOneRoomManyUsersManyPods(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	f := pods[0].seed(t, 1, 1)
	users, workers := envInt("MULTIPOD_USERS", 20_000), envInt("MULTIPOD_WORKERS", 2_000)

	start := time.Now()
	out := storm(t, pods, users, workers, func(int) inbound.ReserveCommand { return f.cmd(0, 1) })
	t.Logf("%d users on %d pods, %d in flight, %s: %s", users, len(pods), workers, time.Since(start).Round(time.Millisecond), out)

	out.mustHaveNoUnexpectedErrors(t)
	if out.ok != 1 {
		t.Fatalf("exactly one user must get the last room, got %d: %v", out.ok, out.winners)
	}
	if got := d.count(t, `select count(*)::int from reservation where status = 1`); got != 1 {
		t.Fatalf("the database holds %d reservations, want 1", got)
	}
	if got := d.count(t, `select total_reserved from room_type_inventory where room_type_id = $1`, f.typeID); got != 1 {
		t.Fatalf("total_reserved = %d, want 1", got)
	}
	if out.soldOut+out.busy+out.ok != int64(users) {
		t.Fatalf("every request must get an answer: %s", out)
	}
	if out.soldOut < int64(users)/2 {
		t.Errorf("most users should have been told 'sold out' quickly, only %d of %d were", out.soldOut, users)
	}
}

func TestSeveralRoomsGoToExactlyThatManyUsers(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 4, defaultPod())
	const rooms = 7
	f := pods[0].seed(t, rooms, 2)

	// Two-night stays, so each winner takes two rows.
	out := storm(t, pods, envInt("MULTIPOD_USERS", 20_000)/2, 1_000, func(int) inbound.ReserveCommand { return f.cmd(0, 2) })
	t.Logf("%s", out)
	out.mustHaveNoUnexpectedErrors(t)
	if out.ok != rooms {
		t.Fatalf("%d rooms must give exactly %d reservations, got %d", rooms, rooms, out.ok)
	}
	if got := d.count(t, `select count(*)::int from room_type_inventory where total_reserved <> $1`, rooms); got != 0 {
		t.Fatalf("%d nights are not fully reserved", got)
	}
}

// Stays that overlap each other lock overlapping rows. The rows are locked in date order, so pods asking for
// [1,3) and [2,4) and [0,3) cannot wait on each other in a cycle: no deadlock, and no night is ever oversold.
// (Postgres happens to scan the primary key in date order anyway, so this test does not fail if the explicit
// "order by date" is removed: that clause is insurance against a different query plan, and this is the check that
// nothing deadlocks or oversells under the plan we have.)
func TestOverlappingStaysAcrossPodsNeitherDeadlockNorOversell(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	const nights = 6
	f := pods[0].seed(t, 1, nights)
	before := d.deadlocks(t)

	out := storm(t, pods, envInt("MULTIPOD_USERS", 20_000)/4, 1_500, func(i int) inbound.ReserveCommand {
		length := 1 + i%3              // 1, 2 or 3 nights
		from := (i / 3) % (nights - 2) // spread the starting night
		return f.cmd(from, length)
	})
	t.Logf("%s", out)
	out.mustHaveNoUnexpectedErrors(t)

	if after := d.deadlocks(t); after != before {
		t.Fatalf("Postgres detected %d deadlock(s): rows are not locked in a fixed order", after-before)
	}
	if out.ok < 1 {
		t.Fatal("at least one stay must have been booked")
	}
	// No night holds more reservations than it has rooms, and the counter matches what is actually booked.
	if got := d.count(t, `select count(*)::int from room_type_inventory where total_reserved > total_inventory or total_reserved < 0`); got != 0 {
		t.Fatalf("%d nights are oversold", got)
	}
	if got := d.count(t, `
		select count(*)::int from room_type_inventory i
		where i.total_reserved <> (select count(*) from reservation r where r.room_type_id = i.room_type_id
			and r.status = 1 and i.date >= r.start_date and i.date < r.end_date)`); got != 0 {
		t.Fatalf("%d nights have a reserved counter that does not match the reservations covering them", got)
	}
}

// The winner's own retry may land on any pod, and must get the reservation back, not "sold out".
func TestWinnersRetryOnAnotherPodGetsTheirReservationBack(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 3, defaultPod())
	f := pods[0].seed(t, 1, 1)
	ctx := context.Background()

	winner := inbound.Principal{UserID: 7}
	cmd := f.cmd(0, 1)
	cmd.RequestID = "the-winner"
	first, err := pods[0].res.Reserve(ctx, winner, cmd)
	if err != nil {
		t.Fatal(err)
	}
	// Everyone else finds it sold out, on every pod, so every pod remembers that.
	out := storm(t, pods, 3_000, 300, func(int) inbound.ReserveCommand { return f.cmd(0, 1) })
	out.mustHaveNoUnexpectedErrors(t)
	for _, p := range pods {
		if !p.res.SoldOut(ctx, f.cmd(0, 1)) {
			t.Fatalf("%s should be answering 'sold out' from memory by now", p.name)
		}
	}
	// The retry, on each pod, with the same request id.
	for _, p := range pods {
		if p.res.SoldOut(ctx, cmd) {
			t.Fatalf("%s told the winner's retry 'sold out'", p.name)
		}
		again, err := p.res.Reserve(ctx, winner, cmd)
		if err != nil || again.ID != first.ID {
			t.Fatalf("%s: retry gave %+v, %v; want reservation %d", p.name, again, err, first.ID)
		}
	}
	// A different user cannot use the winner's request id to get anything.
	if _, err := pods[1].res.Reserve(ctx, inbound.Principal{UserID: 8}, cmd); err == nil {
		t.Fatal("someone else must not get the winner's reservation through their request id")
	}
	if got := d.count(t, `select count(*)::int from reservation`); got != 1 {
		t.Fatalf("%d reservations after all those retries, want 1", got)
	}
}

// "Sold out" is remembered per pod for a short time. When a room comes back, the pod that freed it forgets at once
// and the others catch up within the TTL: the stale window is bounded, and nobody is refused for longer than that.
func TestFreedRoomReachesEveryPodWithinTheSoldOutTTL(t *testing.T) {
	d := newDatabase(t)
	o := defaultPod()
	o.soldOutTTL = 400 * time.Millisecond
	pods := d.pods(t, 3, o)
	f := pods[0].seed(t, 1, 1)
	ctx := context.Background()

	held, err := pods[0].res.Reserve(ctx, inbound.Principal{UserID: 7}, f.cmd(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond) // "rooms were free" is trusted for 50 ms; let that pass so the probe sees the truth
	for _, p := range pods {           // every pod learns it is sold out
		if !p.res.SoldOut(ctx, f.cmd(0, 1)) {
			t.Fatalf("%s must see it sold out", p.name)
		}
	}
	// The holder cancels on pod 1.
	if _, err := pods[0].res.Cancel(ctx, inbound.Principal{UserID: 7}, held.ID); err != nil {
		t.Fatal(err)
	}
	if pods[0].res.SoldOut(ctx, f.cmd(0, 1)) {
		t.Fatal("the pod that freed the room must stop saying sold out immediately")
	}
	freed := time.Now()
	for _, p := range pods[1:] {
		deadline := freed.Add(o.soldOutTTL + 300*time.Millisecond)
		for p.res.SoldOut(ctx, f.cmd(0, 1)) {
			if time.Now().After(deadline) {
				t.Fatalf("%s was still saying sold out %s after the room came back (TTL %s)", p.name, time.Since(freed), o.soldOutTTL)
			}
			time.Sleep(20 * time.Millisecond)
		}
	}
	// The room can be had again on a pod that had been refusing.
	if _, err := pods[2].res.Reserve(ctx, inbound.Principal{UserID: 9}, f.cmd(0, 1)); err != nil {
		t.Fatalf("the freed room could not be reserved: %v", err)
	}
}

// One pod dying in the middle of a storm must not lose the invariant: never more reservations than rooms, and the
// counters match what is stored (an interrupted transaction is rolled back by Postgres, not half applied).
func TestAPodDyingMidStormLeavesTheDatabaseConsistent(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 4, defaultPod())
	f := pods[0].seed(t, 1, 1)
	users := envInt("MULTIPOD_USERS", 20_000)

	var mu sync.RWMutex
	live := append([]*pod(nil), pods...)
	var killed atomic.Bool
	alive := func() []*pod { mu.RLock(); defer mu.RUnlock(); return live }
	out := stormWith(t, alive, users, 1_500, func(i int) inbound.ReserveCommand {
		c := f.cmd(0, 1)
		c.RequestID = fmt.Sprintf("req-%d", i) // every client sends an idempotency key, so it can retry safely
		return c
	}, func(done int64) {
		if done == int64(users/10) && killed.CompareAndSwap(false, true) {
			mu.Lock()
			victim := live[0]
			live = live[1:] // the load balancer stops sending it traffic...
			mu.Unlock()
			victim.pool.Close() // ...and it dies, taking its in-flight requests and connections with it
		}
	})
	t.Logf("%s", out)

	reserved := d.count(t, `select count(*)::int from reservation where status = 1`)
	t.Logf("in the database: %d reservation(s); %d request(s) got an answer that was neither 'sold out' nor 'busy' (those were on the dead pod)", reserved, out.other)
	if reserved > 1 || out.ok > 1 {
		t.Fatalf("%d reservations stored, %d successes reported, for one room", reserved, out.ok)
	}
	// A success that was reported is really stored.
	if out.ok == 1 && reserved != 1 {
		t.Fatalf("a success was reported but nothing is stored")
	}
	// The winner's response may have been lost with the dead pod even though its transaction committed. It retries with
	// its request id, on a surviving pod, and gets its reservation back: not "sold out", and not a second room.
	if reserved == 1 && out.ok == 0 {
		pool, _ := postgres.OpenPool(context.Background(), d.dsn, 1)
		defer pool.Close()
		var reqID string
		var guest, id int64
		if err := pool.QueryRow(context.Background(), `select request_id, guest_id, id from reservation where status = 1`).Scan(&reqID, &guest, &id); err != nil {
			t.Fatal(err)
		}
		retry := f.cmd(0, 1)
		retry.RequestID = reqID
		p := live[len(live)-1]
		if p.res.SoldOut(context.Background(), retry) {
			t.Fatalf("the lost winner's retry was told 'sold out' by %s", p.name)
		}
		got, err := p.res.Reserve(context.Background(), inbound.Principal{UserID: guest}, retry)
		if err != nil || got.ID != id {
			t.Fatalf("the winner's retry got %+v, %v; want reservation %d", got, err, id)
		}
		t.Logf("the winner (user %d) whose answer was lost with the dead pod retried on %s and got reservation %d back", guest, p.name, got.ID)
	}
	if got := d.count(t, `select total_reserved from room_type_inventory where room_type_id = $1`, f.typeID); got != reserved {
		t.Fatalf("total_reserved %d does not match the %d stored reservation(s)", got, reserved)
	}
	// Only requests on the dead pod may fail with an unexpected error, and at most a handful were in flight there.
	if out.other > int64(1_500) {
		t.Fatalf("%d unexpected failures: more than the requests that could have been in flight on one pod", out.other)
	}
	// The response of a winner on the dead pod may have been lost even though its transaction committed. It retries
	// with its request id (any client should) and must not get a second room: there is only the one.
	if reserved == 0 {
		follow := storm(t, live, 2_000, 200, func(int) inbound.ReserveCommand { return f.cmd(0, 1) })
		if follow.ok != 1 {
			t.Fatalf("the room was free but the survivors booked it %d times", follow.ok)
		}
	}
}

// A guest's cap on unpaid reservations is enforced across pods: their parallel requests are serialised in the database.
func TestPerGuestHoldCapHoldsAcrossPods(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	f := pods[0].seed(t, 3, 40)

	var out outcome
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 40; i++ { // one bot, forty different nights, sent to whichever pod
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_, err := pods[i%len(pods)].res.Reserve(context.Background(), inbound.Principal{UserID: 424242}, f.cmd(i, 1))
			out.record(424242, err)
		}()
	}
	close(start)
	wg.Wait()
	t.Logf("%s", &out)
	if out.ok != 3 || out.capped != 37 {
		t.Fatalf("a guest may hold 3 unpaid reservations: got %s", &out)
	}
	if got := d.count(t, `select count(*)::int from reservation where guest_id = 424242`); got != 3 {
		t.Fatalf("%d reservations stored for the bot", got)
	}
}

// The last use of a promo code goes to exactly one guest; everybody else's attempt is rolled back completely, rooms
// included.
func TestLastUseOfAPromoCodeAcrossPods(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	f := pods[0].seed(t, 300, 2)
	ctx := context.Background()
	if _, err := pods[0].hotels.SavePromotion(ctx, domain.Promotion{HotelID: f.hotelID, Code: "LASTONE", PercentOff: 20, MinNights: 1, MaxUses: 1, Active: true}); err != nil {
		t.Fatal(err)
	}

	var out outcome
	var wg sync.WaitGroup
	start := make(chan struct{})
	for i := 0; i < 250; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			cmd := f.cmd(0, 2)
			cmd.PromoCode = "LASTONE"
			_, err := pods[i%len(pods)].res.Reserve(ctx, inbound.Principal{UserID: int64(firstUser + i)}, cmd)
			if errors.Is(err, domain.ErrInvalid) {
				out.record(0, domain.ErrConflict) // refused because the code was used up
				return
			}
			out.record(int64(firstUser+i), err)
		}()
	}
	close(start)
	wg.Wait()
	t.Logf("%s", &out)
	if out.ok != 1 {
		t.Fatalf("the code has one use left, %d guests got it", out.ok)
	}
	if got := d.count(t, `select used_count from promotion where code = 'LASTONE'`); got != 1 {
		t.Fatalf("used_count = %d", got)
	}
	// The 249 refused attempts took no rooms (their transactions rolled back), and only one has a discount.
	if got := d.count(t, `select max(total_reserved) from room_type_inventory where room_type_id = $1`, f.typeID); got != 1 {
		t.Fatalf("rooms reserved: %d, want 1", got)
	}
	if got := d.count(t, `select count(*)::int from reservation where promo_id is not null`); got != 1 {
		t.Fatalf("%d reservations carry the code", got)
	}
}

// ---- the background loops, run by every pod at once

// Every pod runs the sweeper. Each expired hold must be released exactly once: the rooms are counted back once, and
// the history says so once.
func TestConcurrentExpirySweepsReleaseEachHoldOnce(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	const holds = 300
	f := pods[0].seed(t, holds, 1)
	ctx := context.Background()

	for i := 0; i < holds; i++ {
		_, err := pods[0].reservations.Reserve(ctx, outbound.ReserveParams{GuestID: int64(firstUser + i), HotelID: f.hotelID, RoomTypeID: f.typeID,
			Start: f.start, End: f.start.AddDate(0, 0, 1), Rooms: 1, Adults: 2, Currency: "VND", AddonsTotal: "0.0000",
			RequestID: fmt.Sprintf("old-%d", i), ExpiresAt: time.Now().Add(-time.Minute)})
		if err != nil {
			t.Fatal(err)
		}
	}
	if got := d.count(t, `select total_reserved from room_type_inventory where room_type_id = $1`, f.typeID); got != holds {
		t.Fatalf("setup: %d reserved", got)
	}

	var released atomic.Int64
	var wg sync.WaitGroup
	for _, p := range pods {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for round := 0; round < 6; round++ {
				n, err := p.res.ReleaseExpired(ctx)
				if err != nil {
					t.Errorf("%s: %v", p.name, err)
					return
				}
				released.Add(int64(n))
			}
		}()
	}
	wg.Wait()

	if released.Load() != holds {
		t.Errorf("the pods report %d holds released, want %d: some were released twice or not at all", released.Load(), holds)
	}
	if got := d.count(t, `select count(*)::int from reservation where status = 3`); got != holds {
		t.Errorf("%d reservations cancelled, want %d", got, holds)
	}
	if got := d.count(t, `select total_reserved from room_type_inventory where room_type_id = $1`, f.typeID); got != 0 {
		t.Errorf("total_reserved = %d after everything was released (a double release would push it negative and be clamped; a miss leaves it above 0)", got)
	}
	if got := d.count(t, `select count(*)::int from reservation_event where event = 'canceled'`); got != holds {
		t.Errorf("%d 'canceled' events, want %d: each hold must be released once", got, holds)
	}
}

// Every pod runs the waiting-list round. Two pods that meet on the same entry must make ONE offer, and the second must
// leave the first's reservation alone.
func TestConcurrentWaitingListRoundsMakeOneOffer(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 5, defaultPod())
	f := pods[0].seed(t, 1, 1)
	ctx := context.Background()

	holder := inbound.Principal{UserID: 7}
	held, err := pods[0].res.Reserve(ctx, holder, f.cmd(0, 1))
	if err != nil {
		t.Fatal(err)
	}
	const waiting = 40
	for i := 0; i < waiting; i++ { // 40 guests wait for the one room, oldest first
		_, err := pods[i%len(pods)].waitlist.Join(ctx, inbound.Principal{UserID: int64(firstUser + i)}, inbound.WaitCommand{
			HotelID: f.hotelID, RoomTypeID: f.typeID, Start: f.start, End: f.start.AddDate(0, 0, 1), Rooms: 1, Adults: 2})
		if err != nil {
			t.Fatal(err)
		}
	}
	// The room comes back through the repository, so no pod has run its waiting-list round yet.
	if _, err := pods[0].reservations.Transition(ctx, outbound.TransitionParams{ID: held.ID,
		From: []domain.ReservationStatus{domain.ReservationPending}, To: domain.ReservationCanceled, Actor: 7}); err != nil {
		t.Fatal(err)
	}

	// All pods run the round together, several times, as the sweepers and cancellations would.
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, p := range pods {
		for k := 0; k < 3; k++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				if _, err := p.waitlist.PromoteAll(ctx); err != nil {
					t.Errorf("%s: %v", p.name, err)
				}
			}()
		}
	}
	close(start)
	wg.Wait()

	if got := d.count(t, `select count(*)::int from waitlist where status = 2`); got != 1 {
		t.Fatalf("%d entries were offered the one room, want 1", got)
	}
	// The oldest entry got it, and its reservation is alive (a second pod used to cancel it as "the guest left").
	if got := d.count(t, `select count(*)::int from waitlist w join reservation r on r.id = w.reservation_id
		where w.status = 2 and r.status = 1 and w.guest_id = $1`, firstUser); got != 1 {
		t.Fatalf("the first guest in line does not hold a live offer")
	}
	if got := d.count(t, `select count(*)::int from reservation where status = 1`); got != 1 {
		t.Fatalf("%d live reservations for one room", got)
	}
	if got := d.count(t, `select count(*)::int from reservation where status = 3 and status_note like '%waiting list%'`); got != 0 {
		t.Fatalf("%d offers were cancelled as if the guest had left", got)
	}
	if got := d.count(t, `select total_reserved from room_type_inventory where room_type_id = $1`, f.typeID); got != 1 {
		t.Fatalf("total_reserved = %d, want 1", got)
	}
	if got := d.count(t, `select count(*)::int from notification where kind = 'waitlist.offer'`); got != 1 {
		t.Fatalf("the winner was told %d times", got)
	}
}

type countingPush struct {
	mu     sync.Mutex
	pushed map[int64]int
}

func (c *countingPush) Push(_ context.Context, n domain.Notification) error {
	c.mu.Lock()
	c.pushed[n.ID]++
	c.mu.Unlock()
	time.Sleep(time.Millisecond) // a real HTTP call: long enough for the other pods to look at the same rows
	return nil
}

// Every pod forwards notifications. Claiming is atomic, so nothing is pushed to notification-service twice.
func TestReplicasNeverPushTheSameNotificationTwice(t *testing.T) {
	d := newDatabase(t)
	push := &countingPush{pushed: map[int64]int{}}
	o := defaultPod()
	o.push = push
	pods := d.pods(t, 5, o)
	ctx := context.Background()

	const total = 500
	for i := 0; i < total; i++ {
		if err := pods[0].inbox.Add(ctx, domain.Notification{UserID: int64(i % 50), Kind: "test", Title: "t", Body: "b", ReservationID: int64(i + 1)}); err != nil {
			t.Fatal(err)
		}
	}
	var wg sync.WaitGroup
	start := make(chan struct{})
	for _, p := range pods {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for round := 0; round < 12; round++ {
				if _, err := p.notifications.PushPending(ctx); err != nil {
					t.Errorf("%s: %v", p.name, err)
				}
			}
		}()
	}
	close(start)
	wg.Wait()

	push.mu.Lock()
	defer push.mu.Unlock()
	if len(push.pushed) != total {
		t.Fatalf("%d of %d notifications were pushed", len(push.pushed), total)
	}
	for id, n := range push.pushed {
		if n != 1 {
			t.Fatalf("notification %d was pushed %d times", id, n)
		}
	}
	if got := d.count(t, `select count(*)::int from notification where pushed_at is null`); got != 0 {
		t.Fatalf("%d notifications are not marked as pushed", got)
	}
}

// The classic failure: the reservation is committed, and the pod dies before it can answer. The user sees an error,
// not knowing whether they hold the room. Retrying with the same request id, on another pod, must give them their
// reservation, not a second one and not "sold out".
func TestALostAnswerIsRecoveredByRetryingOnAnotherPod(t *testing.T) {
	d := newDatabase(t)
	pods := d.pods(t, 3, defaultPod())
	f := pods[0].seed(t, 1, 1)
	ctx := context.Background()

	user := inbound.Principal{UserID: 555}
	cmd := f.cmd(0, 1)
	cmd.RequestID = "checkout-77"
	won, err := pods[0].res.Reserve(ctx, user, cmd) // committed...
	if err != nil {
		t.Fatal(err)
	}
	pods[0].pool.Close() // ...and the pod dies before the answer reaches the user. They never see `won`.

	// Everyone else piles onto the survivors and is told sold out.
	out := storm(t, pods[1:], 2_000, 200, func(int) inbound.ReserveCommand { return f.cmd(0, 1) })
	out.mustHaveNoUnexpectedErrors(t)
	if out.ok != 0 {
		t.Fatalf("the room was already taken, %d more users got it", out.ok)
	}
	for _, p := range pods[1:] {
		if p.res.SoldOut(ctx, cmd) {
			t.Fatalf("%s told the user with the lost answer 'sold out'", p.name)
		}
		got, err := p.res.Reserve(ctx, user, cmd)
		if err != nil || got.ID != won.ID {
			t.Fatalf("%s: retry gave %+v, %v; want reservation %d", p.name, got, err, won.ID)
		}
	}
	if got := d.count(t, `select count(*)::int from reservation`); got != 1 {
		t.Fatalf("%d reservations, want 1", got)
	}
}
