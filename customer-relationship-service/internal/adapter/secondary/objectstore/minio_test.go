package objectstore

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/config"
	"go.uber.org/fx/fxtest"
)

func TestDisabledWithoutAnEndpoint(t *testing.T) {
	s, err := New(fxtest.NewLifecycle(t), &config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if s.Enabled() {
		t.Fatal("a store without an endpoint reports itself enabled")
	}
	if err := s.Put(context.Background(), "k", []byte("x"), "text/plain"); !errors.Is(err, common.ErrUpstream) {
		t.Fatalf("put: %v, want ErrUpstream", err)
	}
	if err := s.Delete(context.Background(), "k"); err != nil {
		t.Fatalf("delete on a disabled store: %v", err)
	}
}

// TestAgainstARealMinIO needs a server:
// MINIO_TEST_ENDPOINT=localhost:9000 MINIO_TEST_USER=minioadmin MINIO_TEST_PASSWORD=... go test ./...
func TestAgainstARealMinIO(t *testing.T) {
	endpoint := os.Getenv("MINIO_TEST_ENDPOINT")
	if endpoint == "" {
		t.Skip("MINIO_TEST_ENDPOINT not set")
	}
	ctx := context.Background()
	lc := fxtest.NewLifecycle(t)
	s, err := New(lc, &config.Config{Storage: config.StorageConfig{
		Endpoint:  endpoint,
		AccessKey: os.Getenv("MINIO_TEST_USER"),
		SecretKey: os.Getenv("MINIO_TEST_PASSWORD"),
		// A bucket that does not exist yet: the store creates it.
		Bucket: "crm-test-" + time.Now().Format("150405"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	lc.RequireStart()
	defer lc.RequireStop()

	if !s.Enabled() {
		t.Fatal("store not enabled")
	}
	pdf := append([]byte("%PDF-1.7\n"), bytes.Repeat([]byte{0, 1, 2, 250, 255}, 50_000)...)
	const key = "contracts/7/signed-abc.pdf"

	if _, err := s.Get(ctx, key); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("missing key: %v, want ErrNotFound", err)
	}
	if err := s.Put(ctx, key, pdf, "application/pdf"); err != nil {
		t.Fatal(err)
	}
	got, err := s.Get(ctx, key)
	if err != nil || !bytes.Equal(got, pdf) {
		t.Fatalf("round trip: %d bytes back (want %d), err %v", len(got), len(pdf), err)
	}

	// Same key again overwrites: a retried renewal is harmless.
	if err := s.Put(ctx, key, pdf, "application/pdf"); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, key); !errors.Is(err, common.ErrNotFound) {
		t.Fatalf("after delete: %v, want ErrNotFound", err)
	}

	// Without the start hook the bucket does not exist yet: the first Put
	// creates it and retries.
	late, err := New(fxtest.NewLifecycle(t), &config.Config{Storage: config.StorageConfig{
		Endpoint:  endpoint,
		AccessKey: os.Getenv("MINIO_TEST_USER"),
		SecretKey: os.Getenv("MINIO_TEST_PASSWORD"),
		Bucket:    "crm-late-" + time.Now().Format("150405"),
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := late.Put(ctx, "a.pdf", []byte("%PDF"), "application/pdf"); err != nil {
		t.Fatalf("put into a missing bucket: %v", err)
	}
}
