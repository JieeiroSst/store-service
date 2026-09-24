package ekyc

import (
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

func newClient(t *testing.T, h http.HandlerFunc) *Client {
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	return New(&config.Config{Ekyc: config.UpstreamConfig{BaseURL: srv.URL, Timeout: 2 * time.Second}})
}

func TestUploadsSendTheFieldsEkycExpects(t *testing.T) {
	got := map[string][]string{}
	var path string
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		path = r.URL.Path
		_, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
		mr := multipart.NewReader(r.Body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err != nil {
				break
			}
			b, _ := io.ReadAll(p)
			got[p.FormName()] = append(got[p.FormName()], string(b))
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id_number":"secret"}`))
	})

	if err := c.SubmitCitizenCard(context.Background(), "u/1", []byte("F"), []byte("B")); err != nil {
		t.Fatal(err)
	}
	if path != "/api/v1/ekyc/u%2F1/citizen-card" && path != "/api/v1/ekyc/u/1/citizen-card" {
		t.Errorf("path = %s", path)
	}
	if got["front"][0] != "F" || got["back"][0] != "B" {
		t.Errorf("fields = %v", got)
	}

	got = map[string][]string{}
	if err := c.SubmitFaceScan(context.Background(), "u1", [][]byte{[]byte("1"), []byte("2")}); err != nil {
		t.Fatal(err)
	}
	if len(got["frames"]) != 2 {
		t.Errorf("frames = %v", got)
	}
}

func TestStatusReducesEkycToAnOutcome(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"user_id":"u1","identity":{"id_number":"0123"},"face":{"embedding":[1,2]},
			"verification":{"status":"verified","match_score":0.91,"verified_at":"2026-01-02T03:04:05Z"}}`))
	})
	st, err := c.Status(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if st.State != model.IdentityVerified || st.MatchScore != 0.91 || !st.HasCard || !st.HasFace || st.VerifiedAt == nil {
		t.Errorf("status = %+v", st)
	}
}

func TestStatusForAnUnknownIdentityIsNotStarted(t *testing.T) {
	c := newClient(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"verification not found"}`, http.StatusNotFound)
	})
	if st, err := c.Status(context.Background(), "u1"); err != nil || st.State != model.IdentityNotStarted {
		t.Errorf("status=%+v err=%v", st, err)
	}
}

func TestErrorsAreClassified(t *testing.T) {
	cases := map[string]struct {
		code int
		body string
		want error
	}{
		"user missing":    {404, `{"error":"user not found in user-service"}`, model.ErrInvalid},
		"unreadable card": {422, `{"error":"MRZ could not be read with sufficient confidence"}`, model.ErrInvalid},
		"ekyc down":       {500, `{"error":"boom"}`, model.ErrUpstream},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			c := newClient(t, func(w http.ResponseWriter, _ *http.Request) { http.Error(w, tc.body, tc.code) })
			err := c.SubmitCitizenCard(context.Background(), "u", []byte("a"), []byte("b"))
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
		})
	}

	c := newClient(t, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"verification not found"}`, http.StatusNotFound)
	})
	if _, err := c.Verify(context.Background(), "u"); !errors.Is(err, model.ErrConflict) {
		t.Errorf("verify before submitting: got %v, want ErrConflict", err)
	}
}
