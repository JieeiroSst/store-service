package digisign

import (
	"bytes"
	"crypto/x509"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/adapter/secondary/pdfdoc"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
)

func TestVerifyPDF(t *testing.T) {
	ca := testsupport.NewCA(t)
	tsa := testsupport.NewTSA(t, ca, "")
	rev := testsupport.NewRevocation(t, ca)
	key := testsupport.ECKey(t)
	certPEM := testsupport.Issue(t, ca, key, testsupport.Options{OCSPURL: rev.OCSPURL()})
	cert := testsupport.ParseCert(t, certPEM)

	now := time.Now().UTC().Truncate(time.Second)
	doc, err := pdfdoc.NewRenderer(&config.Config{}).Render(port.RenewalDocument{
		Contract: &model.Contract{Base: model.Base{ID: 7}, AccountID: 3}, Account: &model.Account{Name: "Acme"},
		SignedBy: "Ada", EndDate: now.Add(48 * time.Hour), SigningTime: now,
	})
	if err != nil {
		t.Fatal(err)
	}
	newV := func(revocation string) *Verifier {
		return newVerifier(t, config.SignatureConfig{TrustRootsFile: ca.WriteRoots(t), Revocation: revocation})
	}
	signed := testsupport.SignPDF(t, doc, key, cert, []*x509.Certificate{ca.Cert}, "")

	t.Run("valid signature", func(t *testing.T) {
		got, err := newV("hard").VerifyPDF(ctx, signed, now)
		if err != nil {
			t.Fatal(err)
		}
		if got.Format != "pades" || got.Revocation != "good" || got.Timestamp != nil ||
			got.Subject != cert.Subject.String() || got.PayloadHash != hexSHA256(signed) {
			t.Fatalf("result = %+v", got)
		}
	})

	t.Run("signing preserves the issued bytes", func(t *testing.T) {
		if !bytes.HasPrefix(signed, doc) {
			t.Fatal("signed PDF does not start with the rendered document")
		}
	})

	t.Run("embedded time-stamp is verified and returned", func(t *testing.T) {
		stamped := testsupport.SignPDF(t, doc, key, cert, []*x509.Certificate{ca.Cert}, tsa.URL())
		got, err := newV("off").VerifyPDF(ctx, stamped, now)
		if err != nil {
			t.Fatal(err)
		}
		if got.Timestamp == nil || got.Timestamp.Authority == "" || time.Since(got.Timestamp.Time) > time.Minute {
			t.Fatalf("timestamp = %+v", got.Timestamp)
		}
	})

	t.Run("time-stamp from an untrusted authority", func(t *testing.T) {
		rogue := testsupport.NewTSA(t, testsupport.NewCA(t), "")
		stamped := testsupport.SignPDF(t, doc, key, cert, []*x509.Certificate{ca.Cert}, rogue.URL())
		_, err := newV("off").VerifyPDF(ctx, stamped, now)
		wantInvalid(t, err, "not trusted")
	})

	t.Run("modified after signing", func(t *testing.T) {
		tampered := bytes.Replace(signed, []byte("Ada"), []byte("Eve"), 1)
		_, err := newV("off").VerifyPDF(ctx, tampered, now)
		wantInvalid(t, err, "does not match the document")
	})

	t.Run("content appended after signing", func(t *testing.T) {
		appended := append(append([]byte{}, signed...), []byte("\n% appended\n")...)
		_, err := newV("off").VerifyPDF(ctx, appended, now)
		wantInvalid(t, err, "does not cover the whole document")
	})

	t.Run("unsigned document", func(t *testing.T) {
		_, err := newV("off").VerifyPDF(ctx, doc, now)
		wantInvalid(t, err, "")
	})

	t.Run("garbage", func(t *testing.T) {
		for _, junk := range [][]byte{nil, []byte("hello"), []byte("%PDF-1.4\n%%EOF"), bytes.Repeat([]byte{0xff}, 4096)} {
			_, err := newV("off").VerifyPDF(ctx, junk, now)
			wantInvalid(t, err, "")
		}
	})

	t.Run("untrusted signer", func(t *testing.T) {
		other := testsupport.NewCA(t)
		otherCert := testsupport.ParseCert(t, testsupport.Issue(t, other, key, testsupport.Options{}))
		rogueSigned := testsupport.SignPDF(t, doc, key, otherCert, []*x509.Certificate{other.Cert}, "")
		_, err := newV("off").VerifyPDF(ctx, rogueSigned, now)
		wantInvalid(t, err, "not trusted")
	})

	t.Run("revoked signer", func(t *testing.T) {
		rev.Revoke(cert.SerialNumber)
		_, err := newV("hard").VerifyPDF(ctx, signed, now)
		wantInvalid(t, err, "revoked")
	})
}
