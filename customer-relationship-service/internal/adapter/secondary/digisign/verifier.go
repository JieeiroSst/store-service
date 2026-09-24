package digisign

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

const (
	maxCertPEM   = 32 << 10
	maxSignature = 8 << 10
	minRSABits   = 2048
	minECBits    = 256

	formatDetached = "detached-x509"
	formatPAdES    = "pades"

	statusGood      = "good"
	statusUnchecked = "unchecked"
)

type scheme struct {
	hash crypto.Hash
	kind string // ecdsa | pss | pkcs1
}

var schemes = map[string]scheme{
	"ECDSA-SHA256":     {crypto.SHA256, "ecdsa"},
	"ECDSA-SHA384":     {crypto.SHA384, "ecdsa"},
	"ECDSA-SHA512":     {crypto.SHA512, "ecdsa"},
	"RSA-PSS-SHA256":   {crypto.SHA256, "pss"},
	"RSA-PSS-SHA384":   {crypto.SHA384, "pss"},
	"RSA-PKCS1-SHA256": {crypto.SHA256, "pkcs1"},
	"RSA-PKCS1-SHA384": {crypto.SHA384, "pkcs1"},
}

type Verifier struct {
	roots          *x509.CertPool
	tsaRoots       *x509.CertPool
	allowUntrusted bool
	rev            *revocationChecker
	client         *http.Client
	tsaURL         string
	now            func() time.Time
}

func newPool(path string) (*x509.CertPool, error) {
	if path == "" {
		return nil, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read trust roots: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("no certificates found in %s", path)
	}
	return pool, nil
}

func NewVerifier(cfg *config.Config) (*Verifier, error) {
	sc := cfg.Signature
	roots, err := newPool(sc.TrustRootsFile)
	if err != nil {
		return nil, err
	}
	tsaRoots, err := newPool(sc.TSARootsFile)
	if err != nil {
		return nil, err
	}
	if tsaRoots == nil {
		tsaRoots = roots
	}
	timeout := sc.HTTPTimeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	return &Verifier{
		roots:          roots,
		tsaRoots:       tsaRoots,
		allowUntrusted: sc.AllowUntrusted,
		rev:            newRevocationChecker(sc.Revocation, client),
		client:         client,
		tsaURL:         sc.TSAURL,
		now:            time.Now,
	}, nil
}

func invalid(format string, args ...any) error {
	return fmt.Errorf("%w: %s", common.ErrInvalidSignature, fmt.Sprintf(format, args...))
}

func hexSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// Verify checks a detached signature over payload.
func (v *Verifier) Verify(ctx context.Context, payload []byte, ev port.SignatureEvidence, at time.Time) (*port.VerifiedSignature, error) {
	sc, ok := schemes[ev.Algorithm]
	if !ok {
		return nil, invalid("unsupported algorithm %q", ev.Algorithm)
	}
	if len(ev.CertificatePEM) > maxCertPEM || len(ev.Signature) > maxSignature {
		return nil, invalid("certificate or signature too large")
	}
	sig, err := base64.StdEncoding.DecodeString(ev.Signature)
	if err != nil || len(sig) == 0 {
		return nil, invalid("signature is not valid base64")
	}
	leaf, intermediates, err := parseChain(ev.CertificatePEM)
	if err != nil {
		return nil, err
	}
	if err := checkKeyUsage(leaf); err != nil {
		return nil, err
	}
	revocation, err := v.checkTrust(ctx, leaf, intermediates, at)
	if err != nil {
		return nil, err
	}

	h := sc.hash.New()
	h.Write(payload)
	if err := verifySignature(sc, leaf.PublicKey, h.Sum(nil), sig); err != nil {
		return nil, err
	}

	out := describe(leaf, formatDetached, ev.Algorithm, ev.Signature, revocation)
	out.PayloadHash = hexSHA256(payload)
	return out, nil
}

func describe(leaf *x509.Certificate, format, algorithm, signature, revocation string) *port.VerifiedSignature {
	return &port.VerifiedSignature{
		Format:      format,
		Revocation:  revocation,
		Algorithm:   algorithm,
		Signature:   signature,
		Certificate: string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: leaf.Raw})),
		Subject:     leaf.Subject.String(),
		Serial:      leaf.SerialNumber.Text(16),
		Fingerprint: hexSHA256(leaf.Raw),
	}
}

func checkKeyUsage(leaf *x509.Certificate) error {
	if leaf.KeyUsage&(x509.KeyUsageDigitalSignature|x509.KeyUsageContentCommitment) == 0 {
		return invalid("certificate is not allowed to sign (key usage)")
	}
	return nil
}

func parseChain(pemData string) (*x509.Certificate, *x509.CertPool, error) {
	var certs []*x509.Certificate
	rest := []byte(pemData)
	for {
		var block *pem.Block
		block, rest = pem.Decode(rest)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			return nil, nil, invalid("unexpected PEM block %q", block.Type)
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, nil, invalid("cannot parse certificate: %v", err)
		}
		certs = append(certs, c)
	}
	if len(certs) == 0 {
		return nil, nil, invalid("no certificate found")
	}
	pool := x509.NewCertPool()
	for _, c := range certs[1:] {
		pool.AddCert(c)
	}
	return certs[0], pool, nil
}

func checkKeyStrength(pub any) error {
	switch k := pub.(type) {
	case *ecdsa.PublicKey:
		if k.Curve.Params().BitSize < minECBits {
			return invalid("EC key too small")
		}
	case *rsa.PublicKey:
		if k.N.BitLen() < minRSABits {
			return invalid("RSA key too small")
		}
	default:
		return invalid("unsupported public key type")
	}
	return nil
}

func verifySignature(sc scheme, pub any, digest, sig []byte) error {
	switch sc.kind {
	case "ecdsa":
		k, ok := pub.(*ecdsa.PublicKey)
		if !ok {
			return invalid("algorithm does not match the certificate key")
		}
		if err := checkKeyStrength(k); err != nil {
			return err
		}
		if !ecdsa.VerifyASN1(k, digest, sig) {
			return invalid("signature does not match the payload")
		}
	default:
		k, ok := pub.(*rsa.PublicKey)
		if !ok {
			return invalid("algorithm does not match the certificate key")
		}
		if err := checkKeyStrength(k); err != nil {
			return err
		}
		var err error
		if sc.kind == "pss" {
			err = rsa.VerifyPSS(k, sc.hash, digest, sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
		} else {
			err = rsa.VerifyPKCS1v15(k, sc.hash, digest, sig)
		}
		if err != nil {
			return invalid("signature does not match the payload")
		}
	}
	return nil
}
