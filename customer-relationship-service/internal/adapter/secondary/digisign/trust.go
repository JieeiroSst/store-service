package digisign

import (
	"context"
	"crypto/x509"
	"time"
)

func (v *Verifier) checkTrust(ctx context.Context, leaf *x509.Certificate, intermediates *x509.CertPool, at time.Time) (string, error) {
	if v.roots != nil {
		chains, err := leaf.Verify(x509.VerifyOptions{
			Roots:         v.roots,
			Intermediates: intermediates,
			CurrentTime:   at,
			KeyUsages:     []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		})
		if err != nil {
			return "", invalid("certificate is not trusted: %v", err)
		}
		return v.rev.check(ctx, chains[0], v.now())
	}
	if !v.allowUntrusted {
		return "", invalid("no trusted CA configured")
	}
	if at.Before(leaf.NotBefore) || at.After(leaf.NotAfter) {
		return "", invalid("certificate is not valid at the signing time")
	}
	return statusUnchecked, nil
}
