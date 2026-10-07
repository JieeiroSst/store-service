package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/customer-info-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

type fixedClock struct{ t time.Time }

func (c fixedClock) Now() time.Time { return c.t }

type nopMetrics struct{}

func (nopMetrics) ObserveOnboarding(string)    {}
func (nopMetrics) ObserveKYC(domain.KYCStatus) {}

type fakeUsers struct {
	mu    sync.Mutex
	users map[int64]domain.UserProfile
	calls int
}

func (f *fakeUsers) GetUser(_ context.Context, id int64) (domain.UserProfile, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	u, ok := f.users[id]
	if !ok {
		return domain.UserProfile{}, domain.ErrUserNotFound
	}
	return u, nil
}

func (f *fakeUsers) set(u domain.UserProfile) {
	f.mu.Lock()
	f.users[u.ID] = u
	f.mu.Unlock()
}

type fakeEkyc struct {
	mu   sync.Mutex
	doc  domain.IdentityDocument
	face map[int64]bool
	docs map[int64]bool
}

func (f *fakeEkyc) SubmitCitizenCard(ctx context.Context, userID int64, _, _ []byte) (domain.EkycResult, error) {
	f.mu.Lock()
	f.docs[userID] = true
	f.face[userID] = false
	f.mu.Unlock()
	return f.Status(ctx, userID)
}

func (f *fakeEkyc) Status(_ context.Context, userID int64) (domain.EkycResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if !f.docs[userID] {
		return domain.EkycResult{}, nil
	}
	d := f.doc
	return domain.EkycResult{Document: &d, FaceVerified: f.face[userID]}, nil
}

func (f *fakeEkyc) verifyFace(userID int64) {
	f.mu.Lock()
	f.face[userID] = true
	f.mu.Unlock()
}

var now = time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)

func dob(y int) *time.Time {
	t := time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC)
	return &t
}

func newService(users *fakeUsers, doc domain.IdentityDocument) (*CustomerService, *fakeEkyc) {
	ekyc := &fakeEkyc{doc: doc, face: map[int64]bool{}, docs: map[int64]bool{}}
	return NewCustomerService(domain.KYCPolicy{MinAge: 18, MinConfidence: 0.5, RequireFaceMatch: true}, memory.NewCustomers(), memory.NewLocker(),
		users, ekyc, crypto.NewHasher([]byte("application-test-document-hash-key-0123")), fixedClock{now}, nopMetrics{}), ekyc
}

func TestOnboardIsIdempotent(t *testing.T) {
	users := &fakeUsers{users: map[int64]domain.UserProfile{7: {ID: 7, Username: "an", Name: "Nguyen Van An", Active: true}}}
	svc, _ := newService(users, domain.IdentityDocument{})
	ctx := context.Background()

	var wg sync.WaitGroup
	ids := make(chan string, 10)
	created := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, isNew, err := svc.Onboard(ctx, 7)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- c.ID
			created <- isNew
		}()
	}
	wg.Wait()
	close(ids)
	close(created)
	first := <-ids
	for id := range ids {
		if id != first {
			t.Fatalf("two customers for one user: %s %s", first, id)
		}
	}
	n := 0
	for c := range created {
		if c {
			n++
		}
	}
	if n != 1 || users.calls != 1 {
		t.Fatalf("created %d times, user-service called %d times", n, users.calls)
	}

	if _, _, err := svc.Onboard(ctx, 99); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("unknown user: %v", err)
	}
	users.set(domain.UserProfile{ID: 8, Name: "Locked", Active: false})
	if _, _, err := svc.Onboard(ctx, 8); !domain.IsConflict(err) {
		t.Fatalf("locked user: %v", err)
	}
	if _, _, err := svc.Onboard(ctx, 0); !domain.IsInvalid(err) {
		t.Fatalf("zero id: %v", err)
	}
}

func TestKYCAndSync(t *testing.T) {
	users := &fakeUsers{users: map[int64]domain.UserProfile{
		7: {ID: 7, Name: "Nguyễn Văn An", Active: true},
		8: {ID: 8, Name: "Nguyen Van An", Active: true},
	}}
	doc := domain.IdentityDocument{Number: "079200012345", FullName: "NGUYEN VAN AN", DateOfBirth: dob(2000), ChecksumValid: true, Confidence: 0.9}
	svc, ekyc := newService(users, doc)
	ctx := context.Background()
	c, _, _ := svc.Onboard(ctx, 7)

	if _, err := svc.SubmitKYC(ctx, c.ID, []byte("front"), nil); !domain.IsInvalid(err) {
		t.Fatalf("missing back image: %v", err)
	}
	if got, err := svc.RefreshKYC(ctx, c.ID); err != nil || got.KYC.Status != domain.KYCNone {
		t.Fatalf("nothing submitted yet: %v %+v", err, got.KYC)
	}
	got, err := svc.SubmitKYC(ctx, c.ID, []byte("front"), []byte("back"))
	if err != nil || got.KYC.Status != domain.KYCPending {
		t.Fatalf("card read, face missing: %v %+v", err, got.KYC)
	}
	ekyc.verifyFace(7)
	got, err = svc.RefreshKYC(ctx, c.ID)
	if err != nil || got.KYC.Status != domain.KYCVerified || !got.KYC.FaceVerified || got.KYC.DocumentHash == "" || got.KYC.DocumentHash == doc.Number {
		t.Fatalf("kyc: %v %+v", err, got.KYC)
	}

	other, _, _ := svc.Onboard(ctx, 8)
	ekyc.verifyFace(8)
	if _, err := svc.SubmitKYC(ctx, other.ID, []byte("front"), []byte("back")); err != nil {
		t.Fatalf("pending kyc does not claim the document: %v", err)
	}
	ekyc.verifyFace(8)
	if _, err := svc.RefreshKYC(ctx, other.ID); !domain.IsConflict(err) {
		t.Fatalf("same document for a second customer: %v", err)
	}

	users.set(domain.UserProfile{ID: 7, Name: "Nguyen Van An", Active: false})
	synced, err := svc.Sync(ctx, c.ID)
	if err != nil || synced.Status != domain.CustomerSuspended || synced.KYC.Status != domain.KYCVerified {
		t.Fatalf("sync: %v %+v", err, synced)
	}
	if _, err := svc.SubmitKYC(ctx, c.ID, []byte("front"), []byte("back")); !domain.IsConflict(err) {
		t.Fatalf("kyc on suspended customer: %v", err)
	}
	if byUser, err := svc.GetByUserID(ctx, 7); err != nil || byUser.ID != c.ID {
		t.Fatalf("by user: %v", err)
	}
}
