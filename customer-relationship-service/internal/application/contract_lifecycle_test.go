package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

// statusRepo is a contract repository whose List honours the status filter.
type statusRepo struct{ *memRepo[model.Contract] }

func (r statusRepo) List(ctx context.Context, q port.ListQuery) ([]model.Contract, int64, error) {
	all, _, _ := r.memRepo.List(ctx, port.ListQuery{})
	var out []model.Contract
	for _, c := range all {
		if want, ok := q.Equals["contract_status"]; !ok || c.Status == want {
			out = append(out, c)
		}
	}
	total := int64(len(out))
	if q.Offset >= len(out) {
		return nil, total, nil
	}
	out = out[q.Offset:]
	if q.Limit > 0 && len(out) > q.Limit {
		out = out[:q.Limit]
	}
	return out, total, nil
}

type recordingExpirer struct {
	fakeExpirer
	one    bool
	gotID  uint
	gotAt  time.Time
	oneErr error
}

func (r *recordingExpirer) ExpireOne(_ context.Context, id uint, at time.Time) (bool, error) {
	r.gotID, r.gotAt = id, at
	return r.one, r.oneErr
}

func newLifecycle(repo port.Repository[model.Contract]) (*contractLifecycle, *recordingExpirer, *fakeOrchestrator, *fakeNotifier) {
	e, o, n := &recordingExpirer{}, &fakeOrchestrator{}, &fakeNotifier{}
	return &contractLifecycle{repo: repo, expirer: e, orchestrator: o, notifier: n, now: time.Now}, e, o, n
}

func TestLifecycle_State(t *testing.T) {
	repo := newMemRepo[model.Contract]()
	end := time.Now().Add(time.Hour)
	repo.items[1] = &model.Contract{Status: model.ContractApproved, EndDate: &end}
	repo.next = 1
	uc, _, _, _ := newLifecycle(repo)

	got, err := uc.State(context.Background(), 1)
	if err != nil || !got.Exists || got.Status != model.ContractApproved || !got.EndDate.Equal(end) {
		t.Fatalf("state = %+v, err %v", got, err)
	}
	got, err = uc.State(context.Background(), 2)
	if err != nil || got.Exists {
		t.Fatalf("missing contract: %+v, err %v; want Exists=false and no error", got, err)
	}
}

func TestLifecycle_ExpireNotifiesOnlyWhenItClosedSomething(t *testing.T) {
	uc, e, _, n := newLifecycle(newMemRepo[model.Contract]())
	at := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	if ok, err := uc.Expire(context.Background(), 7, at); err != nil || ok || len(n.got) != 0 {
		t.Fatalf("nothing to expire: ok %v, err %v, notified %d", ok, err, len(n.got))
	}
	e.one = true
	if ok, err := uc.Expire(context.Background(), 7, at); err != nil || !ok || len(n.got) != 1 {
		t.Fatalf("expired: ok %v, err %v, notified %d", ok, err, len(n.got))
	}
	if e.gotID != 7 || !e.gotAt.Equal(at) {
		t.Fatalf("asked to expire #%d at %v; the workflow's clock must be used", e.gotID, e.gotAt)
	}
}

func TestLifecycle_RemindOnlyForApprovedContracts(t *testing.T) {
	repo := newMemRepo[model.Contract]()
	repo.items[1] = &model.Contract{Status: model.ContractApproved}
	repo.items[2] = &model.Contract{Status: model.ContractExpired}
	repo.next = 2
	uc, _, _, n := newLifecycle(repo)

	_ = uc.Remind(context.Background(), 1, 7*24*time.Hour)
	_ = uc.Remind(context.Background(), 2, 7*24*time.Hour)
	if len(n.got) != 1 || n.got[0].Message != "Contract #1 ends in about 7 days" {
		t.Fatalf("notifications = %+v", n.got)
	}
}

func TestHumanize(t *testing.T) {
	for d, want := range map[time.Duration]string{
		30 * 24 * time.Hour: "30 days",
		24 * time.Hour:      "1 day",
		5 * time.Hour:       "5 hours",
		90 * time.Second:    "1m30s",
	} {
		if got := humanize(d); got != want {
			t.Errorf("humanize(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestLifecycle_ResyncSyncsOnlyContractsThatCanExpire(t *testing.T) {
	repo := newMemRepo[model.Contract]()
	end := time.Now().Add(time.Hour)
	for i, c := range []model.Contract{
		{Status: model.ContractApproved, EndDate: &end},
		{Status: model.ContractPending, EndDate: &end},
		{Status: model.ContractDraft, EndDate: &end},
		{Status: model.ContractApproved},                // no end date
		{Status: model.ContractExpired, EndDate: &end},  // waits for a signature, no timer needed
		{Status: model.ContractRejected, EndDate: &end}, // final
	} {
		c := c
		repo.items[uint(i+1)] = &c
		c.ID = uint(i + 1)
	}
	repo.next = 6
	uc, _, o, _ := newLifecycle(statusRepo{repo})

	n, err := uc.Resync(context.Background())
	if err != nil || n != 3 || len(o.synced) != 3 {
		t.Fatalf("synced %d (%v), err %v; want 3", n, o.synced, err)
	}
}

func TestContractService_SyncsTheLifecycleOnEveryChange(t *testing.T) {
	repo := newMemRepo[model.Contract]()
	orch := &fakeOrchestrator{}
	uc := NewContractService(repo, orch)
	ctx := context.Background()

	created, err := uc.Create(ctx, &model.Contract{AccountID: 1})
	if err != nil {
		t.Fatal(err)
	}
	created.ID = 1 // the in-memory repository does not assign it
	repo.items[1] = created
	if _, err := uc.Update(ctx, 1, &model.Contract{AccountID: 1}); err != nil {
		t.Fatal(err)
	}
	if err := uc.Delete(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if len(orch.synced) != 3 {
		t.Fatalf("synced %v, want one sync per create, update and delete", orch.synced)
	}
}

func TestContractService_OrchestratorFailureDoesNotFailTheRequest(t *testing.T) {
	repo := newMemRepo[model.Contract]()
	uc := NewContractService(repo, &fakeOrchestrator{err: errors.New("temporal is down")})
	if _, err := uc.Create(context.Background(), &model.Contract{AccountID: 1}); err != nil {
		t.Fatalf("create failed because Temporal is down: %v", err)
	}
}

func TestContractWorkflow_SyncsAfterEachTransition(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf := newTestWorkflow(t, now)
	orch := wf.orchestrator.(*fakeOrchestrator)
	wf.repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, Status: model.ContractDraft}
	wf.repo.next = 1
	ctx := context.Background()

	if _, err := wf.Submit(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if _, err := wf.Approve(ctx, 1); err != nil {
		t.Fatal(err)
	}
	if len(orch.synced) != 2 {
		t.Fatalf("synced %v, want one per transition", orch.synced)
	}
	if _, err := wf.Approve(ctx, 1); err == nil {
		t.Fatal("approving twice should fail")
	}
	if len(orch.synced) != 2 {
		t.Fatal("a refused transition synced the lifecycle")
	}
}
