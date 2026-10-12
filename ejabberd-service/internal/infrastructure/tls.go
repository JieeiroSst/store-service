package infrastructure

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
)

func NewTLSConfig(cfg *config.Config) (*tls.Config, error) {
	cert, err := loadCertificate(cfg.CertFiles)
	if err != nil {
		return nil, err
	}
	if cert == nil {
		slog.Warn("no certfiles configured, generating a self-signed certificate", "hosts", cfg.Hosts)
		if cert, err = selfSigned(cfg.Hosts); err != nil {
			return nil, err
		}
	}
	return &tls.Config{Certificates: []tls.Certificate{*cert}, MinVersion: tls.VersionTLS12}, nil
}

func loadCertificate(files []string) (*tls.Certificate, error) {
	if len(files) == 0 {
		return nil, nil
	}
	var certPEM, keyPEM bytes.Buffer
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			return nil, fmt.Errorf("read certfile %s: %w", f, err)
		}
		for {
			var block *pem.Block
			block, raw = pem.Decode(raw)
			if block == nil {
				break
			}
			switch block.Type {
			case "CERTIFICATE":
				_ = pem.Encode(&certPEM, block)
			case "PRIVATE KEY", "RSA PRIVATE KEY", "EC PRIVATE KEY":
				if keyPEM.Len() == 0 {
					_ = pem.Encode(&keyPEM, block)
				}
			}
		}
	}
	if certPEM.Len() == 0 || keyPEM.Len() == 0 {
		return nil, fmt.Errorf("certfiles %v must contain a certificate and a private key", files)
	}
	cert, err := tls.X509KeyPair(certPEM.Bytes(), keyPEM.Bytes())
	if err != nil {
		return nil, fmt.Errorf("load certfiles: %w", err)
	}
	return &cert, nil
}

func selfSigned(hosts []string) (*tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 127))
	if err != nil {
		return nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: hosts[0], Organization: []string{"ejabberd"}},
		DNSNames:              hosts,
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}, nil
}
