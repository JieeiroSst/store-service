package sealer

import (
	"bytes"
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/sign"
	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
)

type sealer struct {
	key      crypto.Signer
	chain    []*x509.Certificate
	tsaURL   string
	place    string
	verifier port.PDFSignatureVerifier
}

func New(cfg *config.Config, v port.PDFSignatureVerifier) (port.CounterSigner, error) {
	c := cfg.Company
	if c.SignKeyFile == "" && c.SignCertFile == "" {
		return disabled{}, nil
	}
	if c.SignKeyFile == "" || c.SignCertFile == "" {
		return nil, errors.New("COMPANY_SIGN_KEY_FILE and COMPANY_SIGN_CERT_FILE must be set together")
	}
	key, err := loadKey(c.SignKeyFile)
	if err != nil {
		return nil, err
	}
	chain, err := loadChain(c.SignCertFile)
	if err != nil {
		return nil, err
	}
	leaf := chain[0]
	pub, ok := leaf.PublicKey.(interface{ Equal(crypto.PublicKey) bool })
	if !ok || !pub.Equal(key.Public()) {
		return nil, errors.New("COMPANY_SIGN_KEY_FILE does not match the leaf of COMPANY_SIGN_CERT_FILE")
	}
	if now := time.Now(); now.After(leaf.NotAfter) {
		logrus.Warnf("the company signing certificate expired on %s: renewals will fail until it is replaced", leaf.NotAfter.Format(time.RFC3339))
	} else if leaf.NotAfter.Sub(now) < 30*24*time.Hour {
		logrus.Warnf("the company signing certificate expires on %s", leaf.NotAfter.Format(time.RFC3339))
	}
	return &sealer{key: key, chain: chain, tsaURL: cfg.Signature.TSAURL, place: c.Place, verifier: v}, nil
}

func (s *sealer) Enabled() bool { return true }

func (s *sealer) CounterSign(ctx context.Context, signed []byte, at time.Time) (final []byte, verified *port.VerifiedSignature, err error) {
	defer func() {
		if r := recover(); r != nil {
			final, verified, err = nil, nil, fmt.Errorf("%w: company countersignature: %v", common.ErrUpstream, r)
		}
	}()
	fail := func(format string, args ...any) error {
		return fmt.Errorf("%w: company countersignature: %s", common.ErrUpstream, fmt.Sprintf(format, args...))
	}

	rdr, err := pdf.NewReader(bytes.NewReader(signed), int64(len(signed)))
	if err != nil {
		return nil, nil, fail("%v", err)
	}
	leaf := s.chain[0]
	data := sign.SignData{
		Signature: sign.SignDataSignature{
			CertType: sign.ApprovalSignature,
			Info: sign.SignDataSignatureInfo{
				Name:     leaf.Subject.CommonName,
				Location: s.place,
				Reason:   "Bên A ký số đối ứng Phụ lục gia hạn hợp đồng",
				Date:     at,
			},
		},
		Signer:            s.key,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       leaf,
		CertificateChains: [][]*x509.Certificate{s.chain},
	}
	if s.tsaURL != "" {
		data.TSA = sign.TSA{URL: s.tsaURL}
	}

	var out bytes.Buffer
	if err := sign.Sign(bytes.NewReader(signed), &out, rdr, int64(len(signed)), data); err != nil {
		return nil, nil, fail("%v", err)
	}
	final = out.Bytes()
	if !bytes.HasPrefix(final, signed) {
		return nil, nil, fail("the countersigned file does not extend the signed one")
	}

	verified, err = s.verifier.VerifyPDF(ctx, final, at)
	if err != nil {
		return nil, nil, fail("%v", err)
	}
	if verified.Fingerprint != fingerprint(leaf) {
		return nil, nil, fail("the last signature in the file is not the company's")
	}
	return final, verified, nil
}

type disabled struct{}

func (disabled) Enabled() bool { return false }
func (disabled) CounterSign(context.Context, []byte, time.Time) ([]byte, *port.VerifiedSignature, error) {
	return nil, nil, errors.New("company signing is not configured")
}

func loadKey(path string) (crypto.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read company signing key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("%s: no PEM block found", path)
	}
	if _, encrypted := block.Headers["DEK-Info"]; encrypted || block.Type == "ENCRYPTED PRIVATE KEY" {
		return nil, fmt.Errorf("%s: passphrase-protected keys are not supported", path)
	}

	var key any
	switch block.Type {
	case "PRIVATE KEY":
		key, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "RSA PRIVATE KEY":
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		key, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("%s: unsupported PEM type %q", path, block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	switch k := key.(type) {
	case *rsa.PrivateKey:
		return k, nil
	case *ecdsa.PrivateKey:
		return k, nil
	case ed25519.PrivateKey:
		return nil, fmt.Errorf("%s: Ed25519 keys are not supported for PDF signatures", path)
	}
	return nil, fmt.Errorf("%s: unsupported key type", path)
}

func loadChain(path string) ([]*x509.Certificate, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read company signing certificate: %w", err)
	}
	var chain []*x509.Certificate
	for {
		var block *pem.Block
		block, data = pem.Decode(data)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		c, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		chain = append(chain, c)
	}
	if len(chain) == 0 {
		return nil, fmt.Errorf("%s: no certificate found", path)
	}
	return chain, nil
}

var Module = fx.Options(fx.Provide(New))

func fingerprint(c *x509.Certificate) string {
	sum := sha256.Sum256(c.Raw)
	return hex.EncodeToString(sum[:])
}
