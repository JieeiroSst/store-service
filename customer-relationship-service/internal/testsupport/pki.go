package testsupport

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type CA struct {
	Cert *x509.Certificate
	Key  *ecdsa.PrivateKey
	PEM  []byte
}

func NewCA(t testing.TB) *CA {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "Test Root CA"},
		NotBefore:             time.Now().Add(-24 * time.Hour),
		NotAfter:              time.Now().Add(10 * 365 * 24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, tpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	cert, _ := x509.ParseCertificate(der)
	return &CA{Cert: cert, Key: key, PEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})}
}

func (c *CA) WriteRoots(t testing.TB) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "roots.pem")
	if err := os.WriteFile(path, c.PEM, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

type Leaf struct {
	Key crypto.Signer
	PEM string
}

type Options struct {
	Usage     x509.KeyUsage
	NotBefore time.Time
	NotAfter  time.Time
	OCSPURL   string
	CRLURL    string
	EKU       []x509.ExtKeyUsage
	// CommonName defaults to "Ada Lovelace".
	CommonName string
}

func Issue(t testing.TB, ca *CA, key crypto.Signer, o Options) string {
	t.Helper()
	if o.Usage == 0 {
		o.Usage = x509.KeyUsageDigitalSignature
	}
	if o.NotBefore.IsZero() {
		o.NotBefore = time.Now().Add(-time.Hour)
	}
	if o.NotAfter.IsZero() {
		o.NotAfter = time.Now().Add(24 * time.Hour)
	}
	tpl := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      pkix.Name{CommonName: cn(o), Organization: []string{"Acme"}},
		NotBefore:    o.NotBefore,
		NotAfter:     o.NotAfter,
		KeyUsage:     o.Usage,
		ExtKeyUsage:  o.EKU,
	}
	if o.OCSPURL != "" {
		tpl.OCSPServer = []string{o.OCSPURL}
	}
	if o.CRLURL != "" {
		tpl.CRLDistributionPoints = []string{o.CRLURL}
	}
	parent, signer := tpl, key
	if ca != nil {
		parent, signer = ca.Cert, ca.Key
	}
	der, err := x509.CreateCertificate(rand.Reader, tpl, parent, key.Public(), signer)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func ECKey(t testing.TB) *ecdsa.PrivateKey {
	t.Helper()
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func RSAKey(t testing.TB, bits int) *rsa.PrivateKey {
	t.Helper()
	k, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		t.Fatal(err)
	}
	return k
}

func Sign(t testing.TB, key crypto.Signer, algorithm string, payload []byte) string {
	t.Helper()
	digest := sha256.Sum256(payload)
	var opts crypto.SignerOpts = crypto.SHA256
	if algorithm == "RSA-PSS-SHA256" {
		opts = &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash, Hash: crypto.SHA256}
	}
	sig, err := key.Sign(rand.Reader, digest[:], opts)
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(sig)
}

func ParseCert(t testing.TB, certPEM string) *x509.Certificate {
	t.Helper()
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		t.Fatal("no PEM block")
	}
	c, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func cn(o Options) string {
	if o.CommonName != "" {
		return o.CommonName
	}
	return "Ada Lovelace"
}

// WriteSeal creates a company signing key and a certificate for it issued by
// ca, as PEM files, and returns their paths (COMPANY_SIGN_KEY_FILE and
// COMPANY_SIGN_CERT_FILE).
func WriteSeal(t testing.TB, ca *CA, ocspURL string) (keyPath, certPath string) {
	t.Helper()
	key := ECKey(t)
	der, err := x509.MarshalPKCS8PrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	keyPath, certPath = filepath.Join(dir, "seal.key"), filepath.Join(dir, "seal.pem")
	if err := os.WriteFile(keyPath, pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
	cert := Issue(t, ca, key, Options{CommonName: "Công ty Cổ phần Giải pháp Số Việt", OCSPURL: ocspURL, EKU: []x509.ExtKeyUsage{x509.ExtKeyUsageEmailProtection}})
	if err := os.WriteFile(certPath, []byte(cert), 0o600); err != nil {
		t.Fatal(err)
	}
	return keyPath, certPath
}
