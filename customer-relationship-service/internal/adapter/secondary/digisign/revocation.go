package digisign

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ocsp"
)

const (
	maxRevocationBody = 4 << 20
	revocationSkew    = 5 * time.Minute

	revHard = "hard"
	revSoft = "soft"
	revOff  = "off"
)

type revocationChecker struct {
	mode   string
	client *http.Client

	mu   sync.Mutex
	crls map[string]*x509.RevocationList
}

func newRevocationChecker(mode string, client *http.Client) *revocationChecker {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode != revSoft && mode != revOff {
		mode = revHard
	}
	return &revocationChecker{mode: mode, client: client, crls: map[string]*x509.RevocationList{}}
}

type certStatus int

const (
	statusUnknown certStatus = iota
	statusOK
	statusRevoked
)

func (r *revocationChecker) check(ctx context.Context, chain []*x509.Certificate, now time.Time) (string, error) {
	if r.mode == revOff {
		return statusUnchecked, nil
	}
	result := statusGood
	for i := 0; i < len(chain)-1; i++ {
		cert, issuer := chain[i], chain[i+1]
		st, reason := r.status(ctx, cert, issuer, now)
		switch {
		case st == statusRevoked:
			return "", invalid("certificate %q (serial %s) is revoked", cert.Subject.CommonName, cert.SerialNumber.Text(16))
		case st == statusUnknown && r.mode == revHard:
			return "", invalid("revocation status of %q could not be confirmed: %s", cert.Subject.CommonName, reason)
		case st == statusUnknown:
			result = statusUnchecked
		}
	}
	return result, nil
}

func (r *revocationChecker) status(ctx context.Context, cert, issuer *x509.Certificate, now time.Time) (certStatus, string) {
	var reasons []string
	for _, url := range cert.OCSPServer {
		st, err := r.ocspStatus(ctx, url, cert, issuer, now)
		if err == nil {
			return st, ""
		}
		reasons = append(reasons, "OCSP: "+err.Error())
	}
	for _, url := range cert.CRLDistributionPoints {
		st, err := r.crlStatus(ctx, url, cert, issuer, now)
		if err == nil {
			return st, ""
		}
		reasons = append(reasons, "CRL: "+err.Error())
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "certificate names no OCSP responder or CRL")
	}
	return statusUnknown, strings.Join(reasons, "; ")
}

func (r *revocationChecker) fetch(ctx context.Context, method, url, contentType string, body []byte) ([]byte, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return nil, fmt.Errorf("unsupported scheme in %q", url)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s returned %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxRevocationBody+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxRevocationBody {
		return nil, fmt.Errorf("response from %s is too large", url)
	}
	return data, nil
}

func (r *revocationChecker) ocspStatus(ctx context.Context, url string, cert, issuer *x509.Certificate, now time.Time) (certStatus, error) {
	reqDER, err := ocsp.CreateRequest(cert, issuer, nil)
	if err != nil {
		return statusUnknown, err
	}
	body, err := r.fetch(ctx, http.MethodPost, url, "application/ocsp-request", reqDER)
	if err != nil {
		return statusUnknown, err
	}
	resp, err := ocsp.ParseResponseForCert(body, cert, issuer)
	if err != nil {
		return statusUnknown, err
	}
	if resp.ThisUpdate.After(now.Add(revocationSkew)) || (!resp.NextUpdate.IsZero() && now.After(resp.NextUpdate)) {
		return statusUnknown, fmt.Errorf("response is not current")
	}
	switch resp.Status {
	case ocsp.Good:
		return statusOK, nil
	case ocsp.Revoked:
		return statusRevoked, nil
	}
	return statusUnknown, fmt.Errorf("responder does not know the certificate")
}

func (r *revocationChecker) crlStatus(ctx context.Context, url string, cert, issuer *x509.Certificate, now time.Time) (certStatus, error) {
	crl, err := r.loadCRL(ctx, url, issuer, now)
	if err != nil {
		return statusUnknown, err
	}
	for _, e := range crl.RevokedCertificateEntries {
		if e.SerialNumber.Cmp(cert.SerialNumber) == 0 {
			return statusRevoked, nil
		}
	}
	return statusOK, nil
}

func (r *revocationChecker) loadCRL(ctx context.Context, url string, issuer *x509.Certificate, now time.Time) (*x509.RevocationList, error) {
	r.mu.Lock()
	crl := r.crls[url]
	r.mu.Unlock()
	if crl != nil && now.Before(crl.NextUpdate) {
		return crl, nil
	}

	data, err := r.fetch(ctx, http.MethodGet, url, "", nil)
	if err != nil {
		return nil, err
	}
	if block, _ := pem.Decode(data); block != nil {
		data = block.Bytes
	}
	crl, err = x509.ParseRevocationList(data)
	if err != nil {
		return nil, err
	}
	if err := crl.CheckSignatureFrom(issuer); err != nil {
		return nil, fmt.Errorf("CRL is not signed by the issuer: %w", err)
	}
	if crl.ThisUpdate.After(now.Add(revocationSkew)) || now.After(crl.NextUpdate) {
		return nil, fmt.Errorf("CRL is not current")
	}

	r.mu.Lock()
	r.crls[url] = crl
	r.mu.Unlock()
	return crl, nil
}
