package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type fakeUsecase struct {
	front, back []byte
	frames      int
	err         error
}

func (f *fakeUsecase) LinkUser(_ context.Context, id int32, uid string) (*model.Patient, error) {
	if f.err != nil {
		return nil, f.err
	}
	p := &model.Patient{ID: id}
	if uid != "" {
		p.UserID = &uid
	}
	return p, nil
}

func (f *fakeUsecase) status() (*model.IdentityStatus, error) {
	if f.err != nil {
		return nil, f.err
	}
	at := time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	return &model.IdentityStatus{State: model.IdentityVerified, MatchScore: 0.9, VerifiedAt: &at, HasCard: true, HasFace: true}, nil
}
func (f *fakeUsecase) Status(context.Context, int32) (*model.IdentityStatus, error) {
	return f.status()
}
func (f *fakeUsecase) Verify(context.Context, int32) (*model.IdentityStatus, error) {
	return f.status()
}
func (f *fakeUsecase) SubmitCitizenCard(_ context.Context, _ int32, fr, bk []byte) (*model.IdentityStatus, error) {
	f.front, f.back = fr, bk
	return f.status()
}
func (f *fakeUsecase) SubmitFaceScan(_ context.Context, _ int32, fr [][]byte) (*model.IdentityStatus, error) {
	f.frames = len(fr)
	return f.status()
}

func serve(uc *fakeUsecase) http.Handler {
	mux := http.NewServeMux()
	NewIdentityHandler(uc).Register(mux)
	return mux
}

func multipartBody(t *testing.T, files map[string][]string) (*bytes.Buffer, string) {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	for field, contents := range files {
		for _, c := range contents {
			fw, _ := w.CreateFormFile(field, "x.jpg")
			_, _ = fw.Write([]byte(c))
		}
	}
	_ = w.Close()
	return &buf, w.FormDataContentType()
}

func do(h http.Handler, method, path string, body *bytes.Buffer, ctype string) *httptest.ResponseRecorder {
	if body == nil {
		body = &bytes.Buffer{}
	}
	req := httptest.NewRequest(method, path, body)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLinkUser(t *testing.T) {
	h := serve(&fakeUsecase{})
	rec := do(h, http.MethodPut, "/v1/patients/3/user", bytes.NewBufferString(`{"user_id":"u9"}`), "application/json")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"user_id":"u9"`) {
		t.Errorf("%d %s", rec.Code, rec.Body)
	}
	if rec := do(h, http.MethodPut, "/v1/patients/abc/user", bytes.NewBufferString(`{}`), ""); rec.Code != 400 {
		t.Errorf("bad id: %d", rec.Code)
	}
	if rec := do(h, http.MethodPut, "/v1/patients/3/user", bytes.NewBufferString(`nope`), ""); rec.Code != 400 {
		t.Errorf("bad body: %d", rec.Code)
	}
}

func TestStatusExposesOnlyTheOutcome(t *testing.T) {
	rec := do(serve(&fakeUsecase{}), http.MethodGet, "/v1/patients/1/identity", nil, "")
	var got map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if rec.Code != 200 || got["status"] != "verified" || got["has_face_scan"] != true || len(got) != 6 {
		t.Errorf("%d %v", rec.Code, got)
	}
}

func TestCitizenCardNeedsBothImages(t *testing.T) {
	uc := &fakeUsecase{}
	h := serve(uc)

	body, ct := multipartBody(t, map[string][]string{"front": {"F"}})
	if rec := do(h, http.MethodPost, "/v1/patients/1/identity/citizen-card", body, ct); rec.Code != 400 {
		t.Errorf("missing back: %d", rec.Code)
	}
	body, ct = multipartBody(t, map[string][]string{"front": {"F"}, "back": {"B"}})
	if rec := do(h, http.MethodPost, "/v1/patients/1/identity/citizen-card", body, ct); rec.Code != 201 || string(uc.front) != "F" || string(uc.back) != "B" {
		t.Errorf("%d front=%q back=%q", rec.Code, uc.front, uc.back)
	}
	if rec := do(h, http.MethodPost, "/v1/patients/1/identity/citizen-card", bytes.NewBufferString("x"), "application/json"); rec.Code != 400 {
		t.Errorf("not multipart: %d", rec.Code)
	}
}

func TestFaceScanForwardsEveryFrame(t *testing.T) {
	uc := &fakeUsecase{}
	body, ct := multipartBody(t, map[string][]string{"frames": {"1", "2", "3"}})
	if rec := do(serve(uc), http.MethodPost, "/v1/patients/1/identity/face-scan", body, ct); rec.Code != 201 || uc.frames != 3 {
		t.Errorf("%d frames=%d", rec.Code, uc.frames)
	}
}

func TestErrorsUseTheGatewayShapeAndStatuses(t *testing.T) {
	cases := map[string]struct {
		err  error
		code int
	}{
		"not found": {model.ErrNotFound, 404},
		"conflict":  {model.Conflict("not linked"), 409},
		"invalid":   {model.Invalid("bad"), 400},
		"upstream":  {model.ErrUpstream, 503},
		"unknown":   {context.DeadlineExceeded, 500},
	}
	for name, tc := range cases {
		rec := do(serve(&fakeUsecase{err: tc.err}), http.MethodPost, "/v1/patients/1/identity/verify", nil, "")
		var got map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &got)
		if rec.Code != tc.code || got["message"] == nil || got["code"] == nil {
			t.Errorf("%s: %d %v", name, rec.Code, got)
		}
	}
}
