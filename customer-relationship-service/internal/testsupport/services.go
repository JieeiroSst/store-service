package testsupport

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/digitorus/pdf"
	"github.com/digitorus/pdfsign/sign"
	"github.com/digitorus/timestamp"
	"golang.org/x/crypto/ocsp"
)

type Revocation struct {
	Server *httptest.Server
	ca     *CA

	mu                sync.Mutex
	revoked           map[string]bool
	OCSPDown, CRLDown bool
}

func (r *Revocation) OCSPURL() string { return r.Server.URL + "/ocsp" }
func (r *Revocation) CRLURL() string  { return r.Server.URL + "/crl" }

func NewRevocation(t testing.TB, ca *CA) *Revocation {
	r := &Revocation{ca: ca, revoked: map[string]bool{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/ocsp", r.serveOCSP)
	mux.HandleFunc("/crl", r.serveCRL)
	r.Server = httptest.NewServer(mux)
	t.Cleanup(r.Server.Close)
	return r
}

func (r *Revocation) Revoke(serial *big.Int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.revoked[serial.String()] = true
}

func (r *Revocation) serveOCSP(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.OCSPDown {
		http.Error(w, "down", http.StatusServiceUnavailable)
		return
	}
	body, _ := io.ReadAll(req.Body)
	parsed, err := ocsp.ParseRequest(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	tpl := ocsp.Response{
		Status:       ocsp.Good,
		SerialNumber: parsed.SerialNumber,
		ThisUpdate:   time.Now().Add(-time.Minute),
		NextUpdate:   time.Now().Add(time.Hour),
	}
	if r.revoked[parsed.SerialNumber.String()] {
		tpl.Status = ocsp.Revoked
		tpl.RevokedAt = time.Now().Add(-time.Hour)
	}
	der, err := ocsp.CreateResponse(r.ca.Cert, r.ca.Cert, tpl, r.ca.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/ocsp-response")
	_, _ = w.Write(der)
}

func (r *Revocation) serveCRL(w http.ResponseWriter, req *http.Request) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.CRLDown {
		http.Error(w, "down", http.StatusServiceUnavailable)
		return
	}
	var entries []x509.RevocationListEntry
	for s := range r.revoked {
		n, _ := new(big.Int).SetString(s, 10)
		entries = append(entries, x509.RevocationListEntry{SerialNumber: n, RevocationTime: time.Now().Add(-time.Hour)})
	}
	der, err := x509.CreateRevocationList(rand.Reader, &x509.RevocationList{
		Number:                    big.NewInt(1),
		ThisUpdate:                time.Now().Add(-time.Minute),
		NextUpdate:                time.Now().Add(time.Hour),
		RevokedCertificateEntries: entries,
	}, r.ca.Cert, r.ca.Key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/pkix-crl")
	_, _ = w.Write(der)
}

type TSA struct {
	Server *httptest.Server
	Cert   *x509.Certificate
	mu     sync.Mutex
	Down   bool
	Skew   time.Duration
}

func (a *TSA) URL() string { return a.Server.URL }

func NewTSA(t testing.TB, ca *CA, ocspURL string) *TSA {
	t.Helper()
	key := ECKey(t)
	certPEM := Issue(t, ca, key, Options{EKU: []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping}, OCSPURL: ocspURL})
	cert := ParseCert(t, certPEM)

	a := &TSA{Cert: cert}
	a.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		a.mu.Lock()
		down, skew := a.Down, a.Skew
		a.mu.Unlock()
		if down {
			http.Error(w, "down", http.StatusServiceUnavailable)
			return
		}
		body, _ := io.ReadAll(req.Body)
		tsReq, err := timestamp.ParseRequest(body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		ts := &timestamp.Timestamp{
			HashAlgorithm:     tsReq.HashAlgorithm,
			HashedMessage:     tsReq.HashedMessage,
			Time:              time.Now().Add(skew),
			Nonce:             tsReq.Nonce,
			Policy:            asn1.ObjectIdentifier{1, 2, 3, 4},
			Accuracy:          time.Second,
			AddTSACertificate: true,
		}
		resp, err := ts.CreateResponse(cert, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/timestamp-reply")
		_, _ = w.Write(resp)
	}))
	t.Cleanup(a.Server.Close)
	return a
}

func (a *TSA) SetDown(v bool) { a.mu.Lock(); a.Down = v; a.mu.Unlock() }
func (a *TSA) SetSkew(d time.Duration) {
	a.mu.Lock()
	a.Skew = d
	a.mu.Unlock()
}

func SignPDF(t testing.TB, doc []byte, key crypto.Signer, cert *x509.Certificate, chain []*x509.Certificate, tsaURL string) []byte {
	t.Helper()
	rdr, err := pdf.NewReader(bytes.NewReader(doc), int64(len(doc)))
	if err != nil {
		t.Fatal(err)
	}
	data := sign.SignData{
		Signature: sign.SignDataSignature{
			CertType: sign.ApprovalSignature,
			Info:     sign.SignDataSignatureInfo{Name: "Ada Lovelace", Date: time.Now()},
		},
		Signer:            key,
		DigestAlgorithm:   crypto.SHA256,
		Certificate:       cert,
		CertificateChains: [][]*x509.Certificate{append([]*x509.Certificate{cert}, chain...)},
	}
	if tsaURL != "" {
		data.TSA = sign.TSA{URL: tsaURL}
	}
	var out bytes.Buffer
	if err := sign.Sign(bytes.NewReader(doc), &out, rdr, int64(len(doc)), data); err != nil {
		t.Fatal(err)
	}
	return out.Bytes()
}
