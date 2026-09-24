package digisign

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/digitorus/timestamp"
)

const maxTimestampSkew = 5 * time.Minute

func (v *Verifier) Stamp(ctx context.Context, data []byte) (*port.TimestampEvidence, error) {
	if v.tsaURL == "" {
		return nil, nil
	}
	nonce, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 63))
	if err != nil {
		return nil, err
	}
	reqDER, err := timestamp.CreateRequest(bytes.NewReader(data), &timestamp.RequestOptions{
		Hash: crypto.SHA256, Certificates: true, Nonce: nonce,
	})
	if err != nil {
		return nil, err
	}

	upstream := func(format string, args ...any) error {
		return fmt.Errorf("%w: time-stamp authority: %s", common.ErrUpstream, fmt.Sprintf(format, args...))
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, v.tsaURL, bytes.NewReader(reqDER))
	if err != nil {
		return nil, upstream("%v", err)
	}
	req.Header.Set("Content-Type", "application/timestamp-query")
	resp, err := v.client.Do(req)
	if err != nil {
		return nil, upstream("%v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, upstream("returned %s", resp.Status)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxRevocationBody))
	if err != nil {
		return nil, upstream("%v", err)
	}
	ts, err := timestamp.ParseResponse(body)
	if err != nil {
		return nil, upstream("bad response: %v", err)
	}
	if ts.Nonce == nil || ts.Nonce.Cmp(nonce) != 0 {
		return nil, upstream("response does not echo the nonce")
	}

	now := v.now()
	authority, err := v.validateTimestamp(ctx, ts, data, now)
	if err != nil {
		return nil, err
	}
	return &port.TimestampEvidence{
		Token:     base64.StdEncoding.EncodeToString(ts.RawToken),
		Time:      ts.Time.UTC(),
		Authority: authority,
	}, nil
}

func (v *Verifier) validateTimestamp(ctx context.Context, ts *timestamp.Timestamp, data []byte, expected time.Time) (string, error) {
	switch ts.HashAlgorithm {
	case crypto.SHA256, crypto.SHA384, crypto.SHA512:
	default:
		return "", invalid("time-stamp uses a weak hash")
	}
	h := ts.HashAlgorithm.New()
	h.Write(data)
	if !bytes.Equal(h.Sum(nil), ts.HashedMessage) {
		return "", invalid("time-stamp does not cover this signature")
	}
	if d := ts.Time.Sub(expected); d > maxTimestampSkew || d < -maxTimestampSkew {
		return "", invalid("time-stamp time %s is too far from %s", ts.Time.UTC().Format(time.RFC3339), expected.UTC().Format(time.RFC3339))
	}

	var tsa *x509.Certificate
	pool := x509.NewCertPool()
	for _, c := range ts.Certificates {
		if tsa == nil && hasEKU(c, x509.ExtKeyUsageTimeStamping) {
			tsa = c
			continue
		}
		pool.AddCert(c)
	}
	if tsa == nil {
		return "", invalid("time-stamp carries no time-stamping certificate")
	}

	switch {
	case v.tsaRoots != nil:
		chains, err := tsa.Verify(x509.VerifyOptions{
			Roots:         v.tsaRoots,
			Intermediates: pool,
			CurrentTime:   ts.Time,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageTimeStamping},
		})
		if err != nil {
			return "", invalid("time-stamp authority is not trusted: %v", err)
		}
		if _, err := v.rev.check(ctx, chains[0], v.now()); err != nil {
			return "", err
		}
	case !v.allowUntrusted:
		return "", invalid("no trusted time-stamp authority configured")
	}
	return tsa.Subject.String(), nil
}

func hasEKU(c *x509.Certificate, want x509.ExtKeyUsage) bool {
	for _, e := range c.ExtKeyUsage {
		if e == want {
			return true
		}
	}
	return false
}
