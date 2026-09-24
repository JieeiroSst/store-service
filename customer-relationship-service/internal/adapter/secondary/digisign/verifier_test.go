package digisign

import (
	"context"
	"crypto"
	"crypto/x509"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
)

var ctx = context.Background()

var payload = []byte(`{"type":"crm-contract-signature/v1","contract_id":1}`)

func newVerifier(t *testing.T, cfg config.SignatureConfig) *Verifier {
	t.Helper()
	if cfg.Revocation == "" {
		cfg.Revocation = "off"
	}
	v, err := NewVerifier(&config.Config{Signature: cfg})
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func evidence(t *testing.T, key crypto.Signer, algorithm, cert string) port.SignatureEvidence {
	return port.SignatureEvidence{Algorithm: algorithm, Signature: testsupport.Sign(t, key, algorithm, payload), CertificatePEM: cert}
}

func wantInvalid(t *testing.T, err error, contains string) {
	t.Helper()
	if !errors.Is(err, common.ErrInvalidSignature) || !strings.Contains(err.Error(), contains) {
		t.Fatalf("err = %v, want ErrInvalidSignature containing %q", err, contains)
	}
}

func TestVerify_AcceptsEachAlgorithmWithTrustedChain(t *testing.T) {
	ca := testsupport.NewCA(t)
	v := newVerifier(t, config.SignatureConfig{TrustRootsFile: ca.WriteRoots(t)})

	ec, rsaKey := testsupport.ECKey(t), testsupport.RSAKey(t, 2048)
	ecCert, rsaCert := testsupport.Issue(t, ca, ec, testsupport.Options{}), testsupport.Issue(t, ca, rsaKey, testsupport.Options{})

	cases := []struct {
		alg  string
		key  crypto.Signer
		cert string
	}{
		{"ECDSA-SHA256", ec, ecCert},
		{"RSA-PSS-SHA256", rsaKey, rsaCert},
		{"RSA-PKCS1-SHA256", rsaKey, rsaCert},
	}
	for _, c := range cases {
		got, err := v.Verify(ctx, payload, evidence(t, c.key, c.alg, c.cert), time.Now())
		if err != nil {
			t.Fatalf("%s: %v", c.alg, err)
		}
		if got.Subject == "" || len(got.Fingerprint) != 64 || len(got.PayloadHash) != 64 || got.Serial == "" {
			t.Fatalf("%s: incomplete result %+v", c.alg, got)
		}
	}
}

func TestVerify_Rejects(t *testing.T) {
	ca, other := testsupport.NewCA(t), testsupport.NewCA(t)
	v := newVerifier(t, config.SignatureConfig{TrustRootsFile: ca.WriteRoots(t)})
	key := testsupport.ECKey(t)
	good := testsupport.Issue(t, ca, key, testsupport.Options{})

	t.Run("tampered payload", func(t *testing.T) {
		ev := evidence(t, key, "ECDSA-SHA256", good)
		_, err := v.Verify(ctx, append([]byte("x"), payload...), ev, time.Now())
		wantInvalid(t, err, "does not match")
	})
	t.Run("untrusted CA", func(t *testing.T) {
		cert := testsupport.Issue(t, other, key, testsupport.Options{})
		_, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", cert), time.Now())
		wantInvalid(t, err, "not trusted")
	})
	t.Run("self-signed", func(t *testing.T) {
		cert := testsupport.Issue(t, nil, key, testsupport.Options{})
		_, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", cert), time.Now())
		wantInvalid(t, err, "not trusted")
	})
	t.Run("expired certificate", func(t *testing.T) {
		cert := testsupport.Issue(t, ca, key, testsupport.Options{NotBefore: time.Now().Add(-48 * time.Hour), NotAfter: time.Now().Add(-24 * time.Hour)})
		_, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", cert), time.Now())
		wantInvalid(t, err, "not trusted")
	})
	t.Run("certificate not valid yet at signing time", func(t *testing.T) {
		_, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", good), time.Now().Add(-72*time.Hour))
		wantInvalid(t, err, "not trusted")
	})
	t.Run("key usage forbids signing", func(t *testing.T) {
		cert := testsupport.Issue(t, ca, key, testsupport.Options{Usage: x509.KeyUsageKeyEncipherment})
		_, err := v.Verify(ctx, payload, evidence(t, key, "ECDSA-SHA256", cert), time.Now())
		wantInvalid(t, err, "key usage")
	})
	t.Run("algorithm does not match key", func(t *testing.T) {
		ev := evidence(t, key, "ECDSA-SHA256", good)
		ev.Algorithm = "RSA-PSS-SHA256"
		_, err := v.Verify(ctx, payload, ev, time.Now())
		wantInvalid(t, err, "does not match the certificate key")
	})
	t.Run("small RSA key", func(t *testing.T) {
		small := testsupport.RSAKey(t, 1024)
		cert := testsupport.Issue(t, ca, small, testsupport.Options{})
		_, err := v.Verify(ctx, payload, evidence(t, small, "RSA-PKCS1-SHA256", cert), time.Now())
		wantInvalid(t, err, "too small")
	})
	t.Run("unknown algorithm", func(t *testing.T) {
		ev := evidence(t, key, "ECDSA-SHA256", good)
		ev.Algorithm = "MD5-RSA"
		_, err := v.Verify(ctx, payload, ev, time.Now())
		wantInvalid(t, err, "unsupported algorithm")
	})
	t.Run("garbage", func(t *testing.T) {
		_, err := v.Verify(ctx, payload, port.SignatureEvidence{Algorithm: "ECDSA-SHA256", Signature: "%%%", CertificatePEM: good}, time.Now())
		wantInvalid(t, err, "base64")
		_, err = v.Verify(ctx, payload, port.SignatureEvidence{Algorithm: "ECDSA-SHA256", Signature: "AAAA", CertificatePEM: "nope"}, time.Now())
		wantInvalid(t, err, "no certificate")
	})
}

func TestVerify_WithoutRootsIsDeniedUnlessUntrustedAllowed(t *testing.T) {
	key := testsupport.ECKey(t)
	selfSigned := testsupport.Issue(t, nil, key, testsupport.Options{})
	ev := evidence(t, key, "ECDSA-SHA256", selfSigned)

	_, err := newVerifier(t, config.SignatureConfig{}).Verify(ctx, payload, ev, time.Now())
	wantInvalid(t, err, "no trusted CA")

	dev := newVerifier(t, config.SignatureConfig{AllowUntrusted: true})
	if _, err := dev.Verify(ctx, payload, ev, time.Now()); err != nil {
		t.Fatalf("dev mode: %v", err)
	}
	_, err = dev.Verify(ctx, payload, ev, time.Now().Add(-48*time.Hour))
	wantInvalid(t, err, "not valid at the signing time")
}

func TestNew_FailsOnBadRootsFile(t *testing.T) {
	if _, err := NewVerifier(&config.Config{Signature: config.SignatureConfig{TrustRootsFile: "/nonexistent.pem"}}); err == nil {
		t.Fatal("expected an error")
	}
}
