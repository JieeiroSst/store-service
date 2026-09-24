package ekyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/adapter/secondary/httpx"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	base string
	http *http.Client
	json *httpx.Client
}

func New(cfg *config.Config) *Client {
	return &Client{
		base: cfg.Ekyc.BaseURL,
		// Card and frame uploads are larger and slower than the JSON calls.
		http: &http.Client{Timeout: cfg.Ekyc.Timeout * 3},
		json: httpx.New(cfg.Ekyc.BaseURL, cfg.Ekyc.Timeout),
	}
}

func (c *Client) Enabled() bool { return c.json.Enabled() }

func (c *Client) path(userID, step string) string {
	return "/api/v1/ekyc/" + url.PathEscape(userID) + step
}

func (c *Client) SubmitCitizenCard(ctx context.Context, userID string, front, back []byte) error {
	return c.upload(ctx, c.path(userID, "/citizen-card"), []part{{"front", front}, {"back", back}})
}

func (c *Client) SubmitFaceScan(ctx context.Context, userID string, frames [][]byte) error {
	parts := make([]part, len(frames))
	for i, f := range frames {
		parts[i] = part{"frames", f}
	}
	return c.upload(ctx, c.path(userID, "/face-scan"), parts)
}

func (c *Client) Verify(ctx context.Context, userID string) (*model.IdentityStatus, error) {
	var v verification
	if err := c.json.Do(ctx, http.MethodPost, c.path(userID, "/verify"), nil, &v); err != nil {
		return nil, mapErr(err, true)
	}
	return &model.IdentityStatus{State: state(&v), MatchScore: v.MatchScore, VerifiedAt: v.VerifiedAt}, nil
}

type verification struct {
	Status     string     `json:"status"`
	MatchScore float64    `json:"match_score"`
	VerifiedAt *time.Time `json:"verified_at"`
}

func (c *Client) Status(ctx context.Context, userID string) (*model.IdentityStatus, error) {
	var out struct {
		Identity     json.RawMessage `json:"identity"`
		Face         json.RawMessage `json:"face"`
		Verification *verification   `json:"verification"`
	}
	err := c.json.Do(ctx, http.MethodGet, c.path(userID, ""), nil, &out)
	if err != nil {
		if httpx.IsStatus(err, http.StatusNotFound) {
			if se, _ := httpx.AsStatus(err); se != nil && isUserMissing(se.Message) {
				return nil, model.Invalid("the linked user account does not exist in user-service")
			}
			return &model.IdentityStatus{State: model.IdentityNotStarted}, nil
		}
		return nil, mapErr(err, false)
	}
	st := &model.IdentityStatus{State: model.IdentityNotStarted, HasCard: present(out.Identity), HasFace: present(out.Face)}
	if out.Verification != nil {
		st.State, st.MatchScore, st.VerifiedAt = state(out.Verification), out.Verification.MatchScore, out.Verification.VerifiedAt
	}
	return st, nil
}

func present(raw json.RawMessage) bool { return len(raw) > 0 && string(raw) != "null" }

func state(v *verification) model.IdentityState {
	switch v.Status {
	case "verified":
		return model.IdentityVerified
	case "failed":
		return model.IdentityFailed
	default:
		return model.IdentityPending
	}
}

type part struct {
	field string
	data  []byte
}

// upload posts image parts as multipart; ekyc-service answers with the
// extracted identity or biometrics, which we never read or keep.
func (c *Client) upload(ctx context.Context, path string, parts []part) error {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for i, p := range parts {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name=%q; filename="%s-%d.jpg"`, p.field, p.field, i))
		h.Set("Content-Type", "image/jpeg")
		pw, err := w.CreatePart(h)
		if err != nil {
			return err
		}
		if _, err := pw.Write(p.data); err != nil {
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+path, &body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("%w: ekyc-service: %v", model.ErrUpstream, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<20))
		return nil
	}
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	return mapErr(&httpx.StatusError{Status: resp.StatusCode, Message: httpx.Message(raw)}, false)
}

func isUserMissing(msg string) bool {
	return bytes.Contains([]byte(msg), []byte("user not found"))
}

// mapErr sorts ekyc-service failures: 404 on a user is bad input, 404 on
// verify means a step is missing, 422 is an unreadable photo, and anything
// else (including transport errors) is the service being unavailable.
func mapErr(err error, verifying bool) error {
	se, ok := httpx.AsStatus(err)
	if !ok {
		return fmt.Errorf("%w: ekyc-service: %v", model.ErrUpstream, err)
	}
	switch {
	case se.Status == http.StatusNotFound && isUserMissing(se.Message):
		return model.Invalid("the linked user account does not exist in user-service")
	case se.Status == http.StatusNotFound && verifying:
		return model.Conflict("submit the citizen card and a face scan before verifying")
	case se.Status == http.StatusConflict:
		return model.Conflict("%s", se.Message)
	case se.Status >= 400 && se.Status < 500:
		return model.Invalid("%s", se.Message)
	default:
		return fmt.Errorf("%w: ekyc-service: %v", model.ErrUpstream, err)
	}
}

var Module = fx.Options(
	fx.Provide(fx.Annotate(New, fx.As(new(port.IdentityGateway)))),
)
