package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type memPatientRepo struct {
	fakePatients
	p       model.Patient
	updated *model.Patient
	failErr error
}

func (m *memPatientRepo) Get(context.Context, int32) (*model.Patient, error) {
	c := m.p
	return &c, nil
}
func (m *memPatientRepo) Update(_ context.Context, p *model.Patient) error {
	if m.failErr != nil {
		return m.failErr
	}
	m.updated = p
	m.p = *p
	return nil
}

type fakeIdentity struct {
	off       bool
	cards     int
	frames    int
	verifiedU string
}

func (f *fakeIdentity) Enabled() bool { return !f.off }
func (f *fakeIdentity) SubmitCitizenCard(_ context.Context, _ string, _, _ []byte) error {
	f.cards++
	return nil
}
func (f *fakeIdentity) SubmitFaceScan(_ context.Context, _ string, fr [][]byte) error {
	f.frames += len(fr)
	return nil
}
func (f *fakeIdentity) Verify(_ context.Context, u string) (*model.IdentityStatus, error) {
	f.verifiedU = u
	return &model.IdentityStatus{State: model.IdentityVerified, MatchScore: 0.93}, nil
}
func (f *fakeIdentity) Status(context.Context, string) (*model.IdentityStatus, error) {
	return &model.IdentityStatus{State: model.IdentityPending, HasCard: true}, nil
}

func TestIdentityNeedsALinkedUser(t *testing.T) {
	ctx := context.Background()
	repo := &memPatientRepo{}
	gw := &fakeIdentity{}
	svc := NewIdentityService(repo, gw)

	if _, err := svc.Verify(ctx, 1); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("unlinked patient: got %v, want ErrConflict", err)
	}
	if _, err := svc.LinkUser(ctx, 1, " u-42 "); err != nil || repo.updated == nil || *repo.updated.UserID != "u-42" {
		t.Fatalf("link: %+v err=%v", repo.updated, err)
	}
	st, err := svc.Verify(ctx, 1)
	if err != nil || st.State != model.IdentityVerified || gw.verifiedU != "u-42" {
		t.Errorf("verify: %+v err=%v user=%q", st, err, gw.verifiedU)
	}

	// Unlinking clears the account.
	if p, err := svc.LinkUser(ctx, 1, ""); err != nil || p.UserID != nil {
		t.Errorf("unlink: %+v err=%v", p, err)
	}
}

func TestLinkingAUserTwiceIsAConflict(t *testing.T) {
	repo := &memPatientRepo{failErr: model.Conflict("already exists")}
	if _, err := NewIdentityService(repo, &fakeIdentity{}).LinkUser(context.Background(), 1, "u"); !errors.Is(err, model.ErrConflict) {
		t.Errorf("got %v, want ErrConflict", err)
	}
}

func TestSubmissionsAreValidatedAndForwarded(t *testing.T) {
	ctx := context.Background()
	uid := "u1"
	repo := &memPatientRepo{p: model.Patient{ID: 1, UserID: &uid}}
	gw := &fakeIdentity{}
	svc := NewIdentityService(repo, gw)

	if _, err := svc.SubmitCitizenCard(ctx, 1, []byte("f"), nil); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("missing back: got %v, want ErrInvalid", err)
	}
	if _, err := svc.SubmitFaceScan(ctx, 1, nil); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("no frames: got %v, want ErrInvalid", err)
	}
	st, err := svc.SubmitCitizenCard(ctx, 1, []byte("f"), []byte("b"))
	if err != nil || gw.cards != 1 || !st.HasCard {
		t.Errorf("card: %+v err=%v cards=%d", st, err, gw.cards)
	}
	if _, err := svc.SubmitFaceScan(ctx, 1, [][]byte{[]byte("a"), []byte("b")}); err != nil || gw.frames != 2 {
		t.Errorf("face: err=%v frames=%d", err, gw.frames)
	}
}

func TestIdentityDisabledIsUnavailable(t *testing.T) {
	uid := "u1"
	svc := NewIdentityService(&memPatientRepo{p: model.Patient{UserID: &uid}}, &fakeIdentity{off: true})
	if _, err := svc.Status(context.Background(), 1); !errors.Is(err, model.ErrUpstream) {
		t.Errorf("got %v, want ErrUpstream", err)
	}
}

func TestPatientUpdateKeepsTheUserLink(t *testing.T) {
	uid := "u9"
	repo := &memPatientRepo{p: model.Patient{ID: 1, UserID: &uid}}
	svc := NewPatientService(struct {
		*memPatientRepo
	}{repo}, fixedClock(now), nil)
	in := &model.Patient{ID: 1, FirstName: "A", LastName: "B", DateOfBirth: now.AddDate(-20, 0, 0), Gender: model.GenderMale}
	if _, err := svc.Update(context.Background(), in); err != nil {
		t.Fatal(err)
	}
	if repo.updated == nil || repo.updated.UserID == nil || *repo.updated.UserID != "u9" {
		t.Errorf("user link lost on update: %+v", repo.updated)
	}
}
