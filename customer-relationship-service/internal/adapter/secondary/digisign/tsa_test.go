package digisign

import (
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"github.com/JIeeiroSst/customer-relationship-service/internal/testsupport"
	"github.com/digitorus/timestamp"
)

func TestStamp(t *testing.T) {
	ca := testsupport.NewCA(t)
	tsa := testsupport.NewTSA(t, ca, "")
	data := []byte("signature bytes")

	newV := func(cfg config.SignatureConfig) *Verifier {
		cfg.TSAURL = tsa.URL()
		return newVerifier(t, cfg)
	}
	trusting := config.SignatureConfig{TrustRootsFile: ca.WriteRoots(t)}

	t.Run("returns a verified token", func(t *testing.T) {
		got, err := newV(trusting).Stamp(ctx, data)
		if err != nil {
			t.Fatal(err)
		}
		if time.Since(got.Time) > time.Minute || got.Authority == "" {
			t.Fatalf("evidence = %+v", got)
		}
		raw, _ := base64.StdEncoding.DecodeString(got.Token)
		ts, err := timestamp.Parse(raw)
		if err != nil || ts.HashAlgorithm.New().Size() != 32 {
			t.Fatalf("token does not parse as a SHA-256 time-stamp: %v", err)
		}
	})

	t.Run("no authority configured", func(t *testing.T) {
		got, err := newVerifier(t, trusting).Stamp(ctx, data)
		if got != nil || err != nil {
			t.Fatalf("got %v, %v; want nil, nil", got, err)
		}
	})

	t.Run("authority down is an upstream failure", func(t *testing.T) {
		tsa.SetDown(true)
		defer tsa.SetDown(false)
		_, err := newV(trusting).Stamp(ctx, data)
		if !errors.Is(err, common.ErrUpstream) {
			t.Fatalf("err = %v, want ErrUpstream", err)
		}
	})

	t.Run("time too far from now", func(t *testing.T) {
		tsa.SetSkew(time.Hour)
		defer tsa.SetSkew(0)
		_, err := newV(trusting).Stamp(ctx, data)
		wantInvalid(t, err, "too far")
	})

	t.Run("authority from an untrusted CA", func(t *testing.T) {
		other := testsupport.NewCA(t)
		_, err := newV(config.SignatureConfig{TrustRootsFile: other.WriteRoots(t)}).Stamp(ctx, data)
		wantInvalid(t, err, "not trusted")
	})

	t.Run("no roots at all", func(t *testing.T) {
		_, err := newV(config.SignatureConfig{}).Stamp(ctx, data)
		wantInvalid(t, err, "no trusted time-stamp authority")
	})

	t.Run("separate TSA roots", func(t *testing.T) {
		other := testsupport.NewCA(t)
		cfg := config.SignatureConfig{TrustRootsFile: other.WriteRoots(t), TSARootsFile: ca.WriteRoots(t)}
		if _, err := newV(cfg).Stamp(ctx, data); err != nil {
			t.Fatal(err)
		}
	})
}
