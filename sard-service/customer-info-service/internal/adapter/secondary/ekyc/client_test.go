package ekyc

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

const statusBody = `{"user_id":"7","identity":{"source":"card_ocr","document_number":"200012345","surname":"NGUYEN","given_names":"VAN AN",
"nationality":"VNM","date_of_birth":"000501","sex":"M","date_of_expiry":"400501","mrz_line1":"IDVNM2000123456079200012345<<8",
"checksum_valid":true,"confidence":0.93,"nfc_verified":false},"verification":{"status":"verified","match_score":0.81}}`

func TestClient(t *testing.T) {
	var gotFront, gotBack string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method + " " + r.URL.Path {
		case "POST /api/v1/ekyc/7/citizen-card":
			f, _, _ := r.FormFile("front")
			b, _, _ := r.FormFile("back")
			if f == nil || b == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			fb, _ := io.ReadAll(f)
			bb, _ := io.ReadAll(b)
			gotFront, gotBack = string(fb), string(bb)
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(`{}`))
		case "POST /api/v1/ekyc/8/citizen-card":
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"error":"mrz not readable"}`))
		case "POST /api/v1/ekyc/9/citizen-card":
			w.WriteHeader(http.StatusNotFound)
		case "GET /api/v1/ekyc/7":
			w.Write([]byte(statusBody))
		case "GET /api/v1/ekyc/10":
			w.Write([]byte(`{"user_id":"10"}`))
		default:
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()
	c := NewClient(srv.URL+"/", "", time.Second)
	c.now = func() time.Time { return time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC) }
	ctx := context.Background()

	res, err := c.SubmitCitizenCard(ctx, 7, []byte("F"), []byte("B"))
	if err != nil || gotFront != "F" || gotBack != "B" {
		t.Fatalf("submit: %v %q %q", err, gotFront, gotBack)
	}
	d := res.Document
	if d == nil || d.Number != "079200012345" || d.FullName != "NGUYEN VAN AN" || d.Gender != "MALE" ||
		domain.FormatDocumentDate(d.DateOfBirth) != "01/05/2000" || domain.FormatDocumentDate(d.ExpiryDate) != "01/05/2040" ||
		!d.ChecksumValid || d.Confidence != 0.93 || !res.FaceVerified {
		t.Fatalf("result %+v %+v", res, d)
	}
	if _, err := c.SubmitCitizenCard(ctx, 8, []byte("F"), []byte("B")); !domain.IsInvalid(err) {
		t.Fatalf("unreadable card: %v", err)
	}
	if _, err := c.SubmitCitizenCard(ctx, 9, []byte("F"), []byte("B")); !errors.Is(err, domain.ErrUserNotFound) {
		t.Fatalf("unknown user: %v", err)
	}
	if res, err := c.Status(ctx, 10); err != nil || res.Document != nil || res.FaceVerified {
		t.Fatalf("empty status: %v %+v", err, res)
	}
	if _, err := c.Status(ctx, 11); !errors.Is(err, domain.ErrUnavailable) {
		t.Fatalf("server error: %v", err)
	}
}

func TestFullCCCDNumber(t *testing.T) {
	if got := fullCCCDNumber("200012345", "IDVNM2000123456079200012345<<8"); got != "079200012345" {
		t.Fatalf("got %s", got)
	}
	if got := fullCCCDNumber("200012345", "IDVNM2000123456<<<<<<<<<<<<<<<"); got != "200012345" {
		t.Fatalf("no optional number: %s", got)
	}
}
