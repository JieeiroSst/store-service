package sealer

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/digisign"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/pdfdoc"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
	"github.com/digitorus/pdfsign/verify"
)

type env struct {
	ca      *testsupport.CA
	tsa     *testsupport.TSA
	cfg     *config.Config
	verify  *digisign.Verifier
	signedB []byte // the renewal PDF after Bên B signed it
}

func setup(t *testing.T) *env {
	t.Helper()
	ca := testsupport.NewCA(t)
	rev := testsupport.NewRevocation(t, ca)
	tsa := testsupport.NewTSA(t, ca, rev.OCSPURL())
	keyPath, certPath := testsupport.WriteSeal(t, ca, rev.OCSPURL())

	cfg := &config.Config{
		Company:   config.CompanyConfig{Name: "Công ty Cổ phần Giải pháp Số Việt", Place: "Hà Nội", SignKeyFile: keyPath, SignCertFile: certPath},
		Signature: config.SignatureConfig{TrustRootsFile: ca.WriteRoots(t), Revocation: "hard", TSAURL: tsa.URL()},
	}
	v, err := digisign.NewVerifier(cfg)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC().Truncate(time.Second)
	doc, err := pdfdoc.NewRenderer(cfg).Render(port.RenewalDocument{
		Contract: &model.Contract{Base: model.Base{ID: 3}, AccountID: 1}, Account: &model.Account{Name: "Acme"},
		SignedBy: "Ada", EndDate: now.Add(24 * time.Hour), SigningTime: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	bKey := testsupport.ECKey(t)
	bCert := testsupport.ParseCert(t, testsupport.Issue(t, ca, bKey, testsupport.Options{OCSPURL: rev.OCSPURL()}))
	signedB := testsupport.SignPDF(t, doc, bKey, bCert, []*x509.Certificate{ca.Cert}, tsa.URL())
	return &env{ca: ca, tsa: tsa, cfg: cfg, verify: v, signedB: signedB}
}

func TestCounterSign(t *testing.T) {
	e := setup(t)
	s, err := New(e.cfg, e.verify)
	if err != nil {
		t.Fatal(err)
	}
	if !s.Enabled() {
		t.Fatal("sealer not enabled")
	}

	final, seal, err := s.CounterSign(context.Background(), e.signedB, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.HasPrefix(final, e.signedB) {
		t.Fatal("countersigning rewrote the file instead of appending to it")
	}
	if n := bytes.Count(final, []byte("/ByteRange")); n != 2 {
		t.Fatalf("signatures in the file: %d, want 2", n)
	}
	if seal.Format != "pades" || seal.Revocation != "good" || seal.Subject == "" || seal.Timestamp == nil {
		t.Fatalf("evidence = %+v", seal)
	}
	if !bytes.Contains([]byte(seal.Subject), []byte("Công ty Cổ phần Giải pháp Số Việt")) {
		t.Fatalf("subject = %q", seal.Subject)
	}

	// Both signatures verify on the finished file, and Bên B's is still the
	// one made before Bên A's.
	resp, err := verify.Verify(bytes.NewReader(final), int64(len(final)))
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Signers) != 2 {
		t.Fatalf("signers = %d, want 2", len(resp.Signers))
	}
	for i, sg := range resp.Signers {
		if !sg.ValidSignature {
			t.Errorf("signer %d (%s) does not verify", i, sg.Name)
		}
	}
	if resp.Signers[0].Name != "Ada Lovelace" || resp.Signers[1].Name != "Công ty Cổ phần Giải pháp Số Việt" {
		t.Fatalf("signer order = %q, %q", resp.Signers[0].Name, resp.Signers[1].Name)
	}
}

func TestCounterSign_RefusesWhenTheCompanyCertificateIsNotTrusted(t *testing.T) {
	e := setup(t)
	// The verifier trusts a different CA than the one that issued the company's certificate.
	other := testsupport.NewCA(t)
	cfg := *e.cfg
	cfg.Signature.TrustRootsFile = other.WriteRoots(t)
	v, err := digisign.NewVerifier(&cfg)
	if err != nil {
		t.Fatal(err)
	}
	s, err := New(e.cfg, v)
	if err != nil {
		t.Fatal(err)
	}
	_, _, err = s.CounterSign(context.Background(), e.signedB, time.Now())
	if !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("err = %v, want ErrUpstream", err)
	}
}

func TestCounterSign_RefusesGarbage(t *testing.T) {
	e := setup(t)
	s, _ := New(e.cfg, e.verify)
	for _, junk := range [][]byte{nil, []byte("nope"), []byte("%PDF-1.4\n%%EOF")} {
		if _, _, err := s.CounterSign(context.Background(), junk, time.Now()); !errors.Is(err, common.ErrUpstream) {
			t.Fatalf("%q: err = %v, want ErrUpstream", junk, err)
		}
	}
}

func TestNew(t *testing.T) {
	e := setup(t)
	keyPEM, _ := os.ReadFile(e.cfg.Company.SignKeyFile)
	dir := t.TempDir()
	write := func(name string, data []byte) string {
		p := filepath.Join(dir, name)
		if err := os.WriteFile(p, data, 0o600); err != nil {
			t.Fatal(err)
		}
		return p
	}
	with := func(key, cert string) *config.Config {
		c := *e.cfg
		c.Company.SignKeyFile, c.Company.SignCertFile = key, cert
		return &c
	}

	if s, err := New(&config.Config{}, e.verify); err != nil || s.Enabled() {
		t.Fatalf("no key configured: enabled=%v err=%v", s.Enabled(), err)
	}
	if _, err := New(with(e.cfg.Company.SignKeyFile, ""), e.verify); err == nil {
		t.Fatal("a key without a certificate was accepted")
	}
	if _, err := New(with("", e.cfg.Company.SignCertFile), e.verify); err == nil {
		t.Fatal("a certificate without a key was accepted")
	}

	// A key that does not belong to the certificate.
	otherKey, _ := testsupport.WriteSeal(t, e.ca, "")
	if _, err := New(with(otherKey, e.cfg.Company.SignCertFile), e.verify); err == nil {
		t.Fatal("a mismatched key and certificate were accepted")
	}

	// Encrypted keys are refused with a clear error rather than misparsed.
	encrypted := pem.EncodeToMemory(&pem.Block{Type: "ENCRYPTED PRIVATE KEY", Bytes: []byte("x")})
	if _, err := New(with(write("enc.key", encrypted), e.cfg.Company.SignCertFile), e.verify); err == nil {
		t.Fatal("an encrypted key was accepted")
	}
	if _, err := New(with(write("bad.key", []byte("not pem")), e.cfg.Company.SignCertFile), e.verify); err == nil {
		t.Fatal("garbage key accepted")
	}
	if _, err := New(with(e.cfg.Company.SignKeyFile, write("bad.pem", keyPEM)), e.verify); err == nil {
		t.Fatal("a private key in place of the certificate was accepted")
	}
}
