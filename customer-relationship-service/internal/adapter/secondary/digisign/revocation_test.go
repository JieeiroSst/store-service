package digisign

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
)

func TestRevocation(t *testing.T) {
	ca := testsupport.NewCA(t)
	rev := testsupport.NewRevocation(t, ca)
	key := testsupport.ECKey(t)
	roots := ca.WriteRoots(t)

	issue := func(o testsupport.Options) string { return testsupport.Issue(t, ca, key, o) }
	withBoth := issue(testsupport.Options{OCSPURL: rev.OCSPURL(), CRLURL: rev.CRLURL()})
	ocspOnly := issue(testsupport.Options{OCSPURL: rev.OCSPURL()})
	crlOnly := issue(testsupport.Options{CRLURL: rev.CRLURL()})
	neither := issue(testsupport.Options{})

	verify := func(mode, cert string) (string, error) {
		v := newVerifier(t, config.SignatureConfig{TrustRootsFile: roots, Revocation: mode})
		got, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", cert), time.Now())
		if err != nil {
			return "", err
		}
		return got.Revocation, nil
	}
	setDown := func(ocspDown, crlDown bool) { rev.OCSPDown, rev.CRLDown = ocspDown, crlDown }

	t.Run("good over OCSP", func(t *testing.T) {
		setDown(false, true) // prove it was OCSP that answered
		defer setDown(false, false)
		if st, err := verify("hard", ocspOnly); err != nil || st != "good" {
			t.Fatalf("status %q, err %v", st, err)
		}
	})
	t.Run("good over CRL", func(t *testing.T) {
		setDown(true, false)
		defer setDown(false, false)
		if st, err := verify("hard", crlOnly); err != nil || st != "good" {
			t.Fatalf("status %q, err %v", st, err)
		}
	})
	t.Run("falls back from a dead OCSP responder to the CRL", func(t *testing.T) {
		setDown(true, false)
		defer setDown(false, false)
		if st, err := verify("hard", withBoth); err != nil || st != "good" {
			t.Fatalf("status %q, err %v", st, err)
		}
	})
	t.Run("hard mode refuses when nothing answers", func(t *testing.T) {
		setDown(true, true)
		defer setDown(false, false)
		_, err := verify("hard", withBoth)
		wantInvalid(t, err, "could not be confirmed")
	})
	t.Run("soft mode lets it through as unchecked", func(t *testing.T) {
		setDown(true, true)
		defer setDown(false, false)
		if st, err := verify("soft", withBoth); err != nil || st != "unchecked" {
			t.Fatalf("status %q, err %v", st, err)
		}
	})
	t.Run("certificate without revocation info", func(t *testing.T) {
		_, err := verify("hard", neither)
		wantInvalid(t, err, "no OCSP responder or CRL")
		if st, err := verify("soft", neither); err != nil || st != "unchecked" {
			t.Fatalf("soft: status %q, err %v", st, err)
		}
	})
	t.Run("off skips the check", func(t *testing.T) {
		setDown(true, true)
		defer setDown(false, false)
		if st, err := verify("off", withBoth); err != nil || st != "unchecked" {
			t.Fatalf("status %q, err %v", st, err)
		}
	})

	t.Run("revoked certificates are refused in every mode that checks", func(t *testing.T) {
		revokedBoth := issue(testsupport.Options{OCSPURL: rev.OCSPURL(), CRLURL: rev.CRLURL()})
		revokedOCSP := issue(testsupport.Options{OCSPURL: rev.OCSPURL()})
		revokedCRL := issue(testsupport.Options{CRLURL: rev.CRLURL()})
		for _, c := range []string{revokedBoth, revokedOCSP, revokedCRL} {
			rev.Revoke(testsupport.ParseCert(t, c).SerialNumber)
		}

		for _, mode := range []string{"hard", "soft"} {
			for name, c := range map[string]string{"both": revokedBoth, "ocsp": revokedOCSP, "crl": revokedCRL} {
				_, err := verify(mode, c)
				if err == nil {
					t.Fatalf("%s/%s: revoked certificate accepted", mode, name)
				}
				wantInvalid(t, err, "revoked")
			}
		}
		// A revoked certificate is still refused when OCSP is down and the CRL says so.
		setDown(true, false)
		defer setDown(false, false)
		_, err := verify("hard", revokedBoth)
		wantInvalid(t, err, "revoked")
	})
}
