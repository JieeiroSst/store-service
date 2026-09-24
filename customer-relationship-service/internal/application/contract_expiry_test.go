package application

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
)

type fakeExpirer struct {
	n   int64
	got time.Time
}

func (f *fakeExpirer) ExpireOne(context.Context, uint, time.Time) (bool, error) { return false, nil }

func (f *fakeExpirer) ExpireDue(_ context.Context, now time.Time) (int64, error) {
	f.got = now
	return f.n, nil
}

func TestContractExpiry_NotifiesOnlyWhenSomethingExpired(t *testing.T) {
	now := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	n := &fakeNotifier{}
	e := &fakeExpirer{n: 3}
	uc := &contractExpiry{expirer: e, notifier: n, now: func() time.Time { return now }}

	got, err := uc.ExpireOverdue(context.Background())
	if err != nil || got != 3 || !e.got.Equal(now) || len(n.got) != 1 {
		t.Fatalf("got %d, err %v, asked at %v, notified %d", got, err, e.got, len(n.got))
	}

	e.n = 0
	if _, _ = uc.ExpireOverdue(context.Background()); len(n.got) != 1 {
		t.Fatal("notified although nothing expired")
	}
}

type fakeVerifier struct {
	err     error
	payload []byte
	at      time.Time
}

func (f *fakeVerifier) Verify(_ context.Context, payload []byte, ev port.SignatureEvidence, at time.Time) (*port.VerifiedSignature, error) {
	f.payload, f.at = payload, at
	if f.err != nil {
		return nil, f.err
	}
	return &port.VerifiedSignature{
		Algorithm: ev.Algorithm, Signature: ev.Signature, Certificate: "PEM", Subject: "CN=Ada",
		Serial: "1a", Fingerprint: "fp", PayloadHash: "ph",
	}, nil
}

type fakeOrchestrator struct {
	synced []uint
	err    error
}

func (f *fakeOrchestrator) Sync(_ context.Context, id uint) error {
	f.synced = append(f.synced, id)
	return f.err
}

type fakeSealer struct {
	on    bool
	err   error
	calls int
	at    time.Time
}

func (f *fakeSealer) Enabled() bool { return f.on }

func (f *fakeSealer) CounterSign(_ context.Context, signed []byte, at time.Time) ([]byte, *port.VerifiedSignature, error) {
	f.calls++
	f.at = at
	if f.err != nil {
		return nil, nil, f.err
	}
	stamped := at.Add(time.Second)
	return append(append([]byte{}, signed...), []byte("\n% countersigned by A")...),
		&port.VerifiedSignature{Subject: "CN=Company", Serial: "c0", Fingerprint: "cfp", Timestamp: &port.TimestampEvidence{Time: stamped}}, nil
}

type passTx struct{}

func (passTx) InTx(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }

type fakeStamper struct {
	ev  *port.TimestampEvidence
	err error
	got []byte
}

func (f *fakeStamper) Stamp(_ context.Context, data []byte) (*port.TimestampEvidence, error) {
	f.got = data
	return f.ev, f.err
}

type fakePDF struct {
	err error
	got []byte
}

func (f *fakePDF) VerifyPDF(_ context.Context, doc []byte, _ time.Time) (*port.VerifiedSignature, error) {
	f.got = doc
	if f.err != nil {
		return nil, f.err
	}
	return &port.VerifiedSignature{Format: "pades", Revocation: "good", Algorithm: "PDF-CMS", Subject: "CN=Ada", PayloadHash: "dochash"}, nil
}

type fakeRenderer struct{}

func (fakeRenderer) Render(d port.RenewalDocument) ([]byte, error) {
	return []byte(fmt.Sprintf("%%PDF-1.7 issued for %d (%s) by %s", d.Contract.ID, d.Account.Name, d.SignedBy)), nil
}

type testWorkflow struct {
	*contractWorkflow
	repo     *memRepo[model.Contract]
	docs     docShim
	accounts *memRepo[model.Account]
	v        *fakeVerifier
	stamper  *fakeStamper
	pdf      *fakePDF
	store    *testsupport.MemStore
	sealer   *fakeSealer
}

func newTestWorkflow(tb testing.TB, now time.Time) *testWorkflow {
	fe := newFileEnv(tb)
	t := &testWorkflow{
		repo: newMemRepo[model.Contract](), docs: fe.docs, accounts: newMemRepo[model.Account](),
		v: &fakeVerifier{}, stamper: &fakeStamper{}, pdf: &fakePDF{}, store: testsupport.NewMemStore(), sealer: &fakeSealer{},
	}
	t.store.Off = true // the database holds documents unless a test switches MinIO on
	t.contractWorkflow = &contractWorkflow{
		repo: t.repo, files: fe.build(t.repo, t.store, passTx{}), accounts: t.accounts, sealer: t.sealer, tx: passTx{}, verifier: t.v, pdfVerifier: t.pdf, stamper: t.stamper,
		renderer: fakeRenderer{}, notifier: &fakeNotifier{}, orchestrator: &fakeOrchestrator{}, now: func() time.Time { return now },
	}
	return t
}

func newContractWorkflow(tb testing.TB, now time.Time) (*contractWorkflow, *memRepo[model.Contract], *fakeVerifier) {
	t := newTestWorkflow(tb, now)
	return t.contractWorkflow, t.repo, t.v
}

func TestContractSign_OnlyReopensExpiredContracts(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, repo, _ := newContractWorkflow(t, now)
	ctx := context.Background()
	past := now.Add(-24 * time.Hour)
	repo.items[1] = &model.Contract{Status: model.ContractExpired, EndDate: &past}
	repo.items[2] = &model.Contract{Status: model.ContractApproved}
	repo.next = 2

	valid := port.SignContractInput{
		SignedBy: "Ada", EndDate: now.Add(24 * time.Hour), SigningTime: now,
		Algorithm: "ECDSA-SHA256", Signature: "c2ln", Certificate: "PEM",
	}

	if _, err := wf.Sign(ctx, 2, valid); !errors.Is(err, common.ErrInvalidTransition) {
		t.Fatalf("sign active contract: %v, want ErrInvalidTransition", err)
	}

	mutate := map[string]func(*port.SignContractInput){
		"no signer":      func(in *port.SignContractInput) { in.SignedBy = "" },
		"only spaces":    func(in *port.SignContractInput) { in.SignedBy = "  " },
		"past end date":  func(in *port.SignContractInput) { in.EndDate = past },
		"zero end date":  func(in *port.SignContractInput) { in.EndDate = time.Time{} },
		"stale signing":  func(in *port.SignContractInput) { in.SigningTime = now.Add(-time.Hour) },
		"future signing": func(in *port.SignContractInput) { in.SigningTime = now.Add(time.Hour) },
	}
	for name, m := range mutate {
		in := valid
		m(&in)
		if _, err := wf.Sign(ctx, 1, in); !errors.Is(err, common.ErrInvalidRequest) {
			t.Fatalf("%s: %v, want ErrInvalidRequest", name, err)
		}
	}

	got, err := wf.Sign(ctx, 1, valid)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.ContractApproved || got.SignedBy != "Ada" || !got.SignedAt.Equal(now) || !got.EndDate.Equal(valid.EndDate) {
		t.Fatalf("signed contract = %+v", got)
	}
	if got.SignatureAlgorithm != "ECDSA-SHA256" || got.SignerSubject != "CN=Ada" || got.SignerFingerprint != "fp" || got.SignaturePayloadHash != "ph" {
		t.Fatalf("signature evidence not stored: %+v", got)
	}
}

func TestContractSign_VerifiesTheCanonicalPayloadAndKeepsContractExpiredOnFailure(t *testing.T) {
	now := time.Date(2026, 6, 1, 12, 0, 0, 123456789, time.UTC)
	wf, repo, v := newContractWorkflow(t, now)
	past := now.Add(-time.Hour)
	repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, AccountID: 7, Status: model.ContractExpired, EndDate: &past}
	repo.next = 1

	end := now.Add(24 * time.Hour)
	in := port.SignContractInput{SignedBy: "Ada", EndDate: end, SigningTime: now, Algorithm: "ECDSA-SHA256", Signature: "c2ln", Certificate: "PEM"}

	v.err = common.ErrInvalidSignature
	if _, err := wf.Sign(context.Background(), 1, in); !errors.Is(err, common.ErrInvalidSignature) {
		t.Fatalf("err = %v, want ErrInvalidSignature", err)
	}
	if repo.items[1].Status != model.ContractExpired || repo.items[1].Signature != "" {
		t.Fatalf("failed signature changed the contract: %+v", repo.items[1])
	}

	want := model.SigningPayload(&model.Contract{Base: model.Base{ID: 1}, AccountID: 7}, "Ada", end.Truncate(time.Second), now.Truncate(time.Second))
	if string(v.payload) != string(want) || !v.at.Equal(now.Truncate(time.Second)) {
		t.Fatalf("verified payload %s at %v, want %s", v.payload, v.at, want)
	}
	if !strings.Contains(string(want), `"end_date":"`+end.Truncate(time.Second).Format(time.RFC3339)+`"`) {
		t.Fatalf("payload = %s", want)
	}
}

func TestContractApprove_RefusesPastEndDate(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, repo, _ := newContractWorkflow(t, now)
	past := now.Add(-time.Hour)
	repo.items[1] = &model.Contract{Status: model.ContractPending, EndDate: &past}
	repo.next = 1

	if _, err := wf.Approve(context.Background(), 1); !errors.Is(err, common.ErrInvalidTransition) {
		t.Fatalf("approve: %v, want ErrInvalidTransition", err)
	}
	if _, err := wf.Reject(context.Background(), 1); err != nil {
		t.Fatalf("reject should still work: %v", err)
	}
}

func TestContractSign_TimestampsDetachedSignatures(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf := newTestWorkflow(t, now)
	past := now.Add(-time.Hour)
	wf.repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, Status: model.ContractExpired, EndDate: &past}
	wf.repo.next = 1
	in := port.SignContractInput{SignedBy: "Ada", EndDate: now.Add(time.Hour), SigningTime: now, Algorithm: "ECDSA-SHA256", Signature: "c2lnbmF0dXJl", Certificate: "PEM"}

	stamp := &port.TimestampEvidence{Token: "dG9rZW4=", Time: now, Authority: "CN=TSA"}
	wf.stamper.ev = stamp
	got, err := wf.Sign(context.Background(), 1, in)
	if err != nil {
		t.Fatal(err)
	}
	if string(wf.stamper.got) != "signature" {
		t.Fatalf("time-stamped %q, want the decoded signature bytes", wf.stamper.got)
	}
	if got.TimestampToken != stamp.Token || got.TimestampAuthority != "CN=TSA" || !got.TimestampAt.Equal(now) || got.SignatureFormat != "detached-x509" && got.SignatureFormat != "" {
		t.Fatalf("timestamp evidence = %+v", got)
	}

	// A failing authority refuses the renewal and leaves the contract expired.
	wf.repo.items[1].Status = model.ContractExpired
	wf.stamper.err = fmt.Errorf("%w: down", common.ErrUpstream)
	if _, err := wf.Sign(context.Background(), 1, in); !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
	if wf.repo.items[1].Status != model.ContractExpired {
		t.Fatal("contract reopened without a time-stamp")
	}
}

func TestContractSign_PAdES(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf := newTestWorkflow(t, now)
	past := now.Add(-time.Hour)
	wf.repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, AccountID: 1, Status: model.ContractExpired, EndDate: &past}
	wf.repo.next = 1
	ctx := context.Background()
	wf.accounts.items[1] = &model.Account{Name: "Acme"}
	wf.accounts.next = 1
	issued := "%PDF-1.7 issued for 1 (Acme) by Ada"
	in := port.SignContractInput{SignedBy: "Ada", EndDate: now.Add(time.Hour), SigningTime: now}

	// Both modes at once, or a PDF that is not the issued one, are refused.
	both := in
	both.SignedPDF, both.Signature = base64.StdEncoding.EncodeToString([]byte(issued)), "c2ln"
	if _, err := wf.Sign(ctx, 1, both); !errors.Is(err, common.ErrInvalidRequest) {
		t.Fatalf("both modes: %v", err)
	}
	in.SignedPDF = base64.StdEncoding.EncodeToString([]byte("%PDF-1.7 something else entirely"))
	if _, err := wf.Sign(ctx, 1, in); !errors.Is(err, common.ErrInvalidSignature) {
		t.Fatalf("foreign PDF: %v, want ErrInvalidSignature", err)
	}
	if wf.pdf.got != nil {
		t.Fatal("a foreign PDF reached the verifier")
	}

	// An incremental update appended to the issued bytes is accepted, stored and downloadable.
	signedBytes := issued + "\n% incremental update with the signature"
	in.SignedPDF = base64.StdEncoding.EncodeToString([]byte(signedBytes))
	got, err := wf.Sign(ctx, 1, in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Status != model.ContractApproved || got.SignatureFormat != "pades" || got.SignaturePayloadHash != "dochash" {
		t.Fatalf("contract = %+v", got)
	}
	stored, err := wf.SignedDocument(ctx, 1)
	if err != nil || string(stored) != signedBytes {
		t.Fatalf("stored document %q, err %v", stored, err)
	}

	// A verifier failure keeps the contract expired and stores nothing.
	wf.repo.items[1].Status = model.ContractExpired
	wf.docs.reset()
	wf.pdf.err = fmt.Errorf("%w: revoked", common.ErrInvalidSignature)
	if _, err := wf.Sign(ctx, 1, in); !errors.Is(err, common.ErrInvalidSignature) {
		t.Fatalf("err = %v", err)
	}
	if wf.repo.items[1].Status != model.ContractExpired || wf.docs.count() != 0 {
		t.Fatal("failed verification changed state")
	}
}

type failTx struct{}

func (failTx) InTx(context.Context, func(context.Context) error) error { return common.ErrDBFailed }

func padesFixture(t *testing.T, now time.Time) (*testWorkflow, port.SignContractInput, string) {
	wf := newTestWorkflow(t, now)
	past := now.Add(-time.Hour)
	wf.repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, AccountID: 1, Status: model.ContractExpired, EndDate: &past}
	wf.repo.next = 1
	wf.accounts.items[1] = &model.Account{Name: "Acme"}
	wf.accounts.next = 1
	signed := "%PDF-1.7 issued for 1 (Acme) by Ada\n% signature"
	return wf, port.SignContractInput{
		SignedBy: "Ada", EndDate: now.Add(time.Hour), SigningTime: now,
		SignedPDF: base64.StdEncoding.EncodeToString([]byte(signed)),
	}, signed
}

func TestContractSign_KeepsTheSignedPDFInObjectStorage(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, in, signed := padesFixture(t, now)
	wf.store.Off = false
	ctx := context.Background()

	if _, err := wf.Sign(ctx, 1, in); err != nil {
		t.Fatal(err)
	}

	// The bytes are in the bucket under a content-addressed key; the row keeps
	// the key and the hash, not the file.
	if wf.docs.count() != 1 {
		t.Fatalf("documents = %d", wf.docs.count())
	}
	row := wf.docs.row(1)
	sum := sha256.Sum256([]byte(signed))
	hash := hex.EncodeToString(sum[:])
	if row.Data != nil || !strings.HasPrefix(row.ObjectKey, "contracts/1/renewal/") || !strings.HasSuffix(row.ObjectKey, ".pdf") ||
		row.SHA256 != hash || row.Kind != model.FileKindRenewal || row.Source != model.FileSourceSystem || row.Size != int64(len(signed)) ||
		row.Name != "Phu-luc-gia-han-1.pdf" || row.ContentType != "application/pdf" || row.UploadedBy != "Ada" {
		t.Fatalf("row = %+v", row)
	}
	if string(wf.store.Objects[row.ObjectKey]) != signed || wf.store.Types[row.ObjectKey] != "application/pdf" {
		t.Fatalf("stored %q as %q", wf.store.Objects[row.ObjectKey], wf.store.Types[row.ObjectKey])
	}
}

func TestSignedDocument_ReadsFromObjectStorageAndChecksTheHash(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf := newTestWorkflow(t, now)
	wf.store.Off = false
	ctx := context.Background()

	pdf := []byte("%PDF-1.7 signed")
	sum := sha256.Sum256(pdf)
	key := "contracts/1/signed-x.pdf"
	wf.docs.insert(model.ContractFile{ContractID: 1, Kind: model.FileKindRenewal, Latest: true, Version: 1, SHA256: hex.EncodeToString(sum[:]), ObjectKey: key})
	_ = wf.store.Put(ctx, key, pdf, "application/pdf")

	got, err := wf.SignedDocument(ctx, 1)
	if err != nil || !bytes.Equal(got, pdf) {
		t.Fatalf("got %q, err %v", got, err)
	}

	// Someone edits the object in the bucket.
	wf.store.Tamper(key, []byte("%PDF-1.7 forged"))
	if _, err := wf.SignedDocument(ctx, 1); !errors.Is(err, common.ErrIntegrity) {
		t.Fatalf("tampered object: %v, want ErrIntegrity", err)
	}

	// The object is gone.
	delete(wf.store.Objects, key)
	if _, err := wf.SignedDocument(ctx, 1); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("missing object: %v, want ErrNotFound", err)
	}

	// A document saved in MinIO cannot be read once MinIO is switched off.
	wf.store.Off = true
	if _, err := wf.SignedDocument(ctx, 1); err == nil {
		t.Fatal("read succeeded without a document store")
	}
}

func TestContractSign_ObjectStorageFailures(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	// MinIO refuses the upload: the renewal fails and the contract stays expired.
	wf, in, _ := padesFixture(t, now)
	wf.store.Off = false
	wf.store.PutErr = fmt.Errorf("%w: down", common.ErrUpstream)
	if _, err := wf.Sign(ctx, 1, in); !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
	if wf.repo.items[1].Status != model.ContractExpired || wf.docs.count() != 0 {
		t.Fatal("a failed upload changed state")
	}

	// The database write fails after the upload: no orphan object is left behind.
	wf, in, _ = padesFixture(t, now)
	wf.store.Off = false
	wf.contractWorkflow.tx = failTx{}
	if _, err := wf.Sign(ctx, 1, in); !errors.Is(err, common.ErrDBFailed) {
		t.Fatalf("err = %v, want ErrDBFailed", err)
	}
	if len(wf.store.Objects) != 0 || len(wf.store.Deleted) != 1 {
		t.Fatalf("orphaned objects: %v (deleted %v)", wf.store.Objects, wf.store.Deleted)
	}
}

func TestContractSign_CompanyCountersignsAfterBenB(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, in, signedByB := padesFixture(t, now)
	wf.sealer.on = true
	wf.store.Off = false
	ctx := context.Background()

	got, err := wf.Sign(ctx, 1, in)
	if err != nil {
		t.Fatal(err)
	}

	final := signedByB + "\n% countersigned by A"
	sum := sha256.Sum256([]byte(final))
	hash := hex.EncodeToString(sum[:])
	row := wf.docs.row(1)
	if row.SHA256 != hash || string(wf.store.Objects[row.ObjectKey]) != final {
		t.Fatalf("the stored file is not the countersigned one: %+v", row)
	}
	if wf.sealer.calls != 1 || !wf.sealer.at.Equal(now) {
		t.Fatalf("sealer called %d times at %v", wf.sealer.calls, wf.sealer.at)
	}
	// Bên B's evidence is unchanged; Bên A's is recorded, at the time-stamp's time.
	if got.SignerSubject != "CN=Ada" || got.CountersignerSubject != "CN=Company" || got.CountersignerSerial != "c0" ||
		got.CountersignerFingerprint != "cfp" || got.CountersignedAt == nil || !got.CountersignedAt.Equal(now.Add(time.Second)) {
		t.Fatalf("contract = %+v", got)
	}

	// The evidence survives a PUT.
	kept := &model.Contract{}
	kept.CopySignatureFrom(got)
	if kept.CountersignerSubject != "CN=Company" || kept.CountersignedAt == nil {
		t.Fatal("CopySignatureFrom dropped the countersignature")
	}
}

func TestContractSign_WithoutACompanyKeyOnlyBenBSigns(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, in, signedByB := padesFixture(t, now)
	got, err := wf.Sign(context.Background(), 1, in)
	if err != nil {
		t.Fatal(err)
	}
	if wf.sealer.calls != 0 || got.CountersignerSubject != "" || got.CountersignedAt != nil {
		t.Fatalf("sealer used although not configured: calls %d, %+v", wf.sealer.calls, got)
	}
	if string(wf.docs.row(1).Data) != signedByB {
		t.Fatal("the stored file is not the one Bên B signed")
	}
}

func TestContractSign_FailedCountersignatureRefusesTheRenewal(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf, in, _ := padesFixture(t, now)
	wf.sealer.on = true
	wf.sealer.err = fmt.Errorf("%w: company countersignature: certificate expired", common.ErrUpstream)
	wf.store.Off = false

	if _, err := wf.Sign(context.Background(), 1, in); !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
	if wf.repo.items[1].Status != model.ContractExpired || wf.docs.count() != 0 || len(wf.store.Objects) != 0 {
		t.Fatal("a failed countersignature changed state or stored a file")
	}
}

func TestContractSign_DetachedSignaturesAreNotCountersigned(t *testing.T) {
	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	wf := newTestWorkflow(t, now)
	wf.sealer.on = true
	past := now.Add(-time.Hour)
	wf.repo.items[1] = &model.Contract{Base: model.Base{ID: 1}, Status: model.ContractExpired, EndDate: &past}
	wf.repo.next = 1

	_, err := wf.Sign(context.Background(), 1, port.SignContractInput{
		SignedBy: "Ada", EndDate: now.Add(time.Hour), SigningTime: now, Algorithm: "ECDSA-SHA256", Signature: "c2ln", Certificate: "PEM",
	})
	if err != nil || wf.sealer.calls != 0 {
		t.Fatalf("err %v, sealer calls %d; a detached signature has no PDF to countersign", err, wf.sealer.calls)
	}
}
