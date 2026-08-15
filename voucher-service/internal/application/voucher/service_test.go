package voucher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/domain/shared"
	domainvoucher "github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	voucherapp "github.com/JIeeiroSst/voucher-service/internal/application/voucher"
	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/outbox"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// ---- fakes ----

type fakeRepo struct {
	mu       sync.Mutex
	vouchers map[shared.VoucherID]*domainvoucher.Voucher
}

func newFakeRepo() *fakeRepo { return &fakeRepo{vouchers: map[shared.VoucherID]*domainvoucher.Voucher{}} }

func (r *fakeRepo) Create(_ context.Context, v *domainvoucher.Voucher) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vouchers[v.ID] = v
	return nil
}
func (r *fakeRepo) FindByID(_ context.Context, id shared.VoucherID) (*domainvoucher.Voucher, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.vouchers[id]
	if !ok {
		return nil, domainvoucher.ErrVoucherNotFound
	}
	return v, nil
}
func (r *fakeRepo) FindByIDForUpdate(ctx context.Context, id shared.VoucherID) (*domainvoucher.Voucher, error) {
	return r.FindByID(ctx, id)
}
func (r *fakeRepo) FindByCode(_ context.Context, merchantID shared.MerchantID, code string) (*domainvoucher.Voucher, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, v := range r.vouchers {
		if v.MerchantID == merchantID && v.Code == code {
			return v, nil
		}
	}
	return nil, domainvoucher.ErrVoucherNotFound
}
func (r *fakeRepo) ListByOwner(_ context.Context, ownerType domainvoucher.OwnerType, ownerID string) ([]*domainvoucher.Voucher, error) {
	return nil, nil
}
func (r *fakeRepo) ListDueForExpiry(_ context.Context, now time.Time) ([]*domainvoucher.Voucher, error) {
	return nil, nil
}
func (r *fakeRepo) Save(_ context.Context, v *domainvoucher.Voucher) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.vouchers[v.ID] = v
	return nil
}

type fakeProvider struct {
	providerType shared.ProviderType
}

func (p *fakeProvider) Issue(_ context.Context, ref shared.ProductRef, qty int) ([]shared.VoucherCode, error) {
	codes := make([]shared.VoucherCode, qty)
	for i := range codes {
		codes[i] = shared.VoucherCode{Code: "CODE", PIN: ""}
	}
	return codes, nil
}
func (p *fakeProvider) Validate(_ context.Context, code, pin string) (shared.ValidationResult, error) {
	return shared.ValidationResult{Valid: true}, nil
}
func (p *fakeProvider) Redeem(_ context.Context, code, pin string, amount shared.Money) (shared.RedeemResult, error) {
	return shared.RedeemResult{Success: true, RedeemedAmount: amount, ProviderTxnRef: "provider-txn-1"}, nil
}
func (p *fakeProvider) Type() shared.ProviderType { return p.providerType }

type fakeRegistry struct {
	provider voucherapp.MerchantProvider
}

func (r *fakeRegistry) Resolve(shared.ProviderType) (voucherapp.MerchantProvider, error) {
	return r.provider, nil
}

type fakeMerchantLookup struct {
	info voucherapp.MerchantInfo
}

func (f *fakeMerchantLookup) GetMerchantInfo(_ context.Context, id shared.MerchantID) (voucherapp.MerchantInfo, error) {
	f.info.ID = id
	return f.info, nil
}

type fakeTxManager struct{}

func (fakeTxManager) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type fakeLocker struct {
	mu     sync.Mutex
	locked map[string]bool
}

func newFakeLocker() *fakeLocker { return &fakeLocker{locked: map[string]bool{}} }

func (l *fakeLocker) Acquire(_ context.Context, key string, ttl time.Duration) (func(context.Context) error, bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.locked[key] {
		return nil, false, nil
	}
	l.locked[key] = true
	return func(context.Context) error {
		l.mu.Lock()
		defer l.mu.Unlock()
		delete(l.locked, key)
		return nil
	}, true, nil
}

type fakeIdempStore struct {
	mu      sync.Mutex
	records map[string]*idempotency.Record
}

func newFakeIdempStore() *fakeIdempStore {
	return &fakeIdempStore{records: map[string]*idempotency.Record{}}
}
func (s *fakeIdempStore) Claim(_ context.Context, key, requestHash string, ttl time.Duration) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.records[key]; exists {
		return false, nil
	}
	s.records[key] = &idempotency.Record{Key: key, Status: idempotency.StatusInProgress}
	return true, nil
}
func (s *fakeIdempStore) Get(_ context.Context, key string) (*idempotency.Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.records[key], nil
}
func (s *fakeIdempStore) Complete(_ context.Context, key string, status int, body []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = &idempotency.Record{Key: key, Status: idempotency.StatusCompleted, ResponseStatus: status, ResponseBody: body}
	return nil
}
func (s *fakeIdempStore) Fail(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records[key] = &idempotency.Record{Key: key, Status: idempotency.StatusFailed}
	return nil
}
func (s *fakeIdempStore) Release(_ context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.records, key)
	return nil
}

type fakeOutbox struct{}

func (fakeOutbox) Enqueue(context.Context, outbox.Event) error { return nil }

type fixedClock struct{ now time.Time }

func (c fixedClock) Now() time.Time { return c.now }

// ---- test setup ----

func newTestService(t *testing.T, provider voucherapp.MerchantProvider, providerType shared.ProviderType) (voucherapp.VoucherService, *fakeRepo, *fakeIdempStore) {
	t.Helper()
	repo := newFakeRepo()
	registry := &fakeRegistry{provider: provider}
	lookup := &fakeMerchantLookup{info: voucherapp.MerchantInfo{ProviderType: providerType, Active: true}}
	locker := newFakeLocker()
	idemp := newFakeIdempStore()
	log := zap.NewNop()

	svc := voucherapp.NewService(repo, registry, lookup, fakeTxManager{}, locker, idemp, fakeOutbox{}, fixedClock{now: time.Now()}, log)
	return svc, repo, idemp
}

func TestRedeemVoucher_HappyPath_SelfProvider(t *testing.T) {
	provider := &fakeProvider{providerType: shared.ProviderTypeSelf}
	svc, repo, _ := newTestService(t, provider, shared.ProviderTypeSelf)

	ref := shared.ProductRef{MerchantID: shared.NewMerchantID(), SKU: "sku", Denomination: shared.NewMoney(50000, "VND")}
	v, err := domainvoucher.NewVoucher(ref.MerchantID, ref, time.Now())
	require.NoError(t, err)
	require.NoError(t, v.Issue(shared.VoucherCode{Code: "C1"}, nil, time.Now()))
	require.NoError(t, v.Activate(domainvoucher.OwnerTypeUser, "user-1", time.Now()))
	require.NoError(t, repo.Create(context.Background(), v))

	out, err := svc.RedeemVoucher(context.Background(), voucherapp.RedeemVoucherInput{
		VoucherID:      v.ID,
		Amount:         shared.NewMoney(50000, "VND"),
		IdempotencyKey: "idem-1",
	})
	require.NoError(t, err)
	require.Equal(t, string(domainvoucher.StatusRedeemed), out.Status)
}

func TestRedeemVoucher_DuplicateRequest_ReplaysResultWithoutReRedeeming(t *testing.T) {
	provider := &fakeProvider{providerType: shared.ProviderTypeSelf}
	svc, repo, _ := newTestService(t, provider, shared.ProviderTypeSelf)

	ref := shared.ProductRef{MerchantID: shared.NewMerchantID(), SKU: "sku", Denomination: shared.NewMoney(50000, "VND")}
	v, _ := domainvoucher.NewVoucher(ref.MerchantID, ref, time.Now())
	_ = v.Issue(shared.VoucherCode{Code: "C1"}, nil, time.Now())
	_ = v.Activate(domainvoucher.OwnerTypeUser, "user-1", time.Now())
	_ = repo.Create(context.Background(), v)

	in := voucherapp.RedeemVoucherInput{VoucherID: v.ID, Amount: shared.NewMoney(50000, "VND"), IdempotencyKey: "idem-2"}

	out1, err := svc.RedeemVoucher(context.Background(), in)
	require.NoError(t, err)

	out2, err := svc.RedeemVoucher(context.Background(), in)
	require.NoError(t, err)
	require.Equal(t, out1.VoucherID, out2.VoucherID)
	require.Equal(t, out1.Status, out2.Status)
}

func TestRedeemVoucher_AlreadyRedeemed_ReturnsInvalidTransition(t *testing.T) {
	provider := &fakeProvider{providerType: shared.ProviderTypeSelf}
	svc, repo, idemp := newTestService(t, provider, shared.ProviderTypeSelf)

	ref := shared.ProductRef{MerchantID: shared.NewMerchantID(), SKU: "sku", Denomination: shared.NewMoney(50000, "VND")}
	v, _ := domainvoucher.NewVoucher(ref.MerchantID, ref, time.Now())
	_ = v.Issue(shared.VoucherCode{Code: "C1"}, nil, time.Now())
	_ = v.Activate(domainvoucher.OwnerTypeUser, "user-1", time.Now())
	_ = v.Redeem(shared.NewMoney(50000, "VND"), "", time.Now())
	_ = repo.Create(context.Background(), v)

	_, err := svc.RedeemVoucher(context.Background(), voucherapp.RedeemVoucherInput{
		VoucherID: v.ID, Amount: shared.NewMoney(50000, "VND"), IdempotencyKey: "idem-3",
	})
	require.Error(t, err)

	rec, _ := idemp.Get(context.Background(), "idem-3")
	require.Equal(t, idempotency.StatusFailed, rec.Status)
}
