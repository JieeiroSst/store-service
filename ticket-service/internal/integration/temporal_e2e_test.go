package integration

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// These tests run the whole service against a real Temporal server, to prove what the workflow test environment cannot:
// that the client starts and signals workflows with options the server accepts, and that a real worker runs them.
//
//	docker run -d -p 7233:7233 temporalio/temporal:latest server start-dev --ip 0.0.0.0
//	TEST_TEMPORAL_ADDRESS=localhost:7233 go test ./internal/integration -run TemporalServer -v
func temporalAddress(t *testing.T) string {
	addr := os.Getenv("TEST_TEMPORAL_ADDRESS")
	if addr == "" {
		t.Skip("set TEST_TEMPORAL_ADDRESS to run against a Temporal server")
	}
	return addr
}

func wfStatus(t *testing.T, c client.Client, id string) enumspb.WorkflowExecutionStatus {
	t.Helper()
	d, err := c.DescribeWorkflowExecution(context.Background(), id, "")
	if err != nil {
		return enumspb.WORKFLOW_EXECUTION_STATUS_UNSPECIFIED
	}
	return d.WorkflowExecutionInfo.Status
}

func TestTemporalServerRunsTheLifecycle(t *testing.T) {
	addr := temporalAddress(t)
	tc, err := client.Dial(client.Options{HostPort: addr})
	if err != nil {
		t.Fatal(err)
	}
	defer tc.Close()
	up := newFakeUploadService(t, "k")
	queue := fmt.Sprintf("ts-it-%d", time.Now().UnixNano())
	a := startApp(t, map[string]string{"TemporalAddress": addr, "TemporalTaskQueue": queue, "UploadServiceURL": up.URL, "UploadServiceKey": "k",
		"SweepIntervalSeconds": "3600"}) // the sweeper never runs in this test: whatever happens is the workflow's doing
	eid, tid := publishedEvent(a, 10)

	order := func(u int64, req string) map[string]any {
		return a.must(201, "POST", "/api/v1/orders", u, map[string]any{"event_id": eid, "buyer_name": "B", "request_id": req,
			"items": []map[string]any{{"ticket_type_id": tid, "quantity": 1}}})
	}
	wid := func(o map[string]any) string { return fmt.Sprintf("%s/order-%d", queue, int64(o["id"].(float64))) }

	// a held order has a running workflow
	o1 := order(5, "e2e-1")
	waitFor(t, "the workflow to start", 10*time.Second, func() bool { return wfStatus(t, tc, wid(o1)) == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING })

	// paying wakes it: the documents are made by its activity (there is no sweeper), and it goes on waiting for the reminder
	a.must(200, "POST", fmt.Sprintf("/api/v1/orders/%d/pay", int64(o1["id"].(float64))), 5, map[string]any{"method": "wallet"})
	waitFor(t, "the documents made by the workflow", 15*time.Second, func() bool { return len(up.of("ticket-user:5")) == 2 })
	if s := wfStatus(t, tc, wid(o1)); s != enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING {
		t.Fatalf("a paid order's workflow waits for its reminder: %s", s)
	}

	// giving up ends the workflow at once
	o2 := order(6, "e2e-2")
	waitFor(t, "the second workflow", 10*time.Second, func() bool { return wfStatus(t, tc, wid(o2)) == enumspb.WORKFLOW_EXECUTION_STATUS_RUNNING })
	a.must(200, "POST", fmt.Sprintf("/api/v1/orders/%d/cancel", int64(o2["id"].(float64))), 6, nil)
	waitFor(t, "the workflow to end", 15*time.Second, func() bool { return wfStatus(t, tc, wid(o2)) == enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED })

	// the same request sent again does not start another workflow, and does not fail
	if again := order(5, "e2e-1"); again["id"] != o1["id"] {
		t.Fatalf("a retry returned another order: %v", again["id"])
	}
}

// The hold of an order that is never paid ends by itself, in real time (the shortest hold is a minute).
func TestTemporalServerExpiresAnUnpaidOrder(t *testing.T) {
	addr := temporalAddress(t)
	if testing.Short() {
		t.Skip("takes a minute")
	}
	tc, err := client.Dial(client.Options{HostPort: addr})
	if err != nil {
		t.Fatal(err)
	}
	defer tc.Close()
	queue := fmt.Sprintf("ts-it-%d", time.Now().UnixNano())
	a := startApp(t, map[string]string{"TemporalAddress": addr, "TemporalTaskQueue": queue, "OrderHoldMinutes": "1", "SweepIntervalSeconds": "3600"})
	eid, tid := publishedEvent(a, 3)
	o := a.must(201, "POST", "/api/v1/orders", 5, map[string]any{"event_id": eid, "buyer_name": "B", "request_id": "e2e-exp",
		"items": []map[string]any{{"ticket_type_id": tid, "quantity": 2}}})
	w := fmt.Sprintf("%s/order-%d", queue, int64(o["id"].(float64)))
	waitFor(t, "the workflow to end", 100*time.Second, func() bool { return wfStatus(t, tc, w) == enumspb.WORKFLOW_EXECUTION_STATUS_COMPLETED })
	if got := a.must(200, "GET", fmt.Sprintf("/api/v1/orders/%d", int64(o["id"].(float64))), 5, nil); got["status"] != "expired" {
		t.Fatalf("the order is %v", got["status"])
	}
	if ev := a.must(200, "GET", fmt.Sprintf("/api/v1/events/%d", eid), 0, nil); ev["ticket_types"].([]any)[0].(map[string]any)["available"].(float64) != 3 {
		t.Fatalf("the tickets did not come back: %v", ev["ticket_types"])
	}
}

// With Temporal unreachable the service still starts and serves: orders are held, paid and expired by the sweeper as
// before, and the worker keeps trying in the background.
func TestServiceServesWhileTemporalIsDown(t *testing.T) {
	up := newFakeUploadService(t, "k")
	a := startApp(t, map[string]string{"TemporalAddress": "127.0.0.1:1", "UploadServiceURL": up.URL, "UploadServiceKey": "k", "SweepIntervalSeconds": "1"})
	eid, tid := publishedEvent(a, 5)
	paid := buy(a, 5, eid, tid, 1)
	if paid["status"] != "paid" {
		t.Fatalf("order: %v", paid["status"])
	}
	// no workflow, but the sweeper still makes the documents
	waitFor(t, "the sweeper to make the documents", 10*time.Second, func() bool { return len(up.of("ticket-user:5")) == 2 })
}
