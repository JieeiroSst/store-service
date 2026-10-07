package ekyc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.EkycGateway {
		return NewClient(cfg.EkycService.BaseURL, cfg.EkycService.Token, cfg.EkycService.Timeout)
	}),
)

const maxResponseBytes = 1 << 20

type Client struct {
	baseURL string
	token   string
	http    *http.Client
	now     func() time.Time
}

func NewClient(baseURL, token string, timeout time.Duration) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: &http.Client{Timeout: timeout}, now: time.Now}
}

type identity struct {
	Source         string  `json:"source"`
	DocumentNumber string  `json:"document_number"`
	Surname        string  `json:"surname"`
	GivenNames     string  `json:"given_names"`
	Nationality    string  `json:"nationality"`
	DateOfBirth    string  `json:"date_of_birth"`
	Sex            string  `json:"sex"`
	DateOfExpiry   string  `json:"date_of_expiry"`
	MRZLine1       string  `json:"mrz_line1"`
	ChecksumValid  bool    `json:"checksum_valid"`
	Confidence     float64 `json:"confidence"`
	NFCVerified    bool    `json:"nfc_verified"`
}

type verification struct {
	Status     string  `json:"status"`
	MatchScore float64 `json:"match_score"`
}

type status struct {
	Identity     *identity     `json:"identity"`
	Verification *verification `json:"verification"`
}

func (c *Client) SubmitCitizenCard(ctx context.Context, userID int64, front, back []byte) (domain.EkycResult, error) {
	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	for _, f := range []struct {
		field string
		data  []byte
	}{{"front", front}, {"back", back}} {
		w, err := mw.CreateFormFile(f.field, f.field+".jpg")
		if err != nil {
			return domain.EkycResult{}, err
		}
		if _, err := w.Write(f.data); err != nil {
			return domain.EkycResult{}, err
		}
	}
	if err := mw.Close(); err != nil {
		return domain.EkycResult{}, err
	}
	raw, code, err := c.do(ctx, http.MethodPost, userID, "/citizen-card", &body, mw.FormDataContentType())
	if err != nil {
		return domain.EkycResult{}, err
	}
	switch {
	case code == http.StatusUnprocessableEntity || code == http.StatusBadRequest:
		return domain.EkycResult{}, domain.Invalid("ekyc-service could not read the card: %s", errorMessage(raw))
	case code == http.StatusNotFound:
		return domain.EkycResult{}, domain.ErrUserNotFound
	case code != http.StatusCreated && code != http.StatusOK:
		return domain.EkycResult{}, fmt.Errorf("%w: ekyc-service returned %d: %s", domain.ErrUnavailable, code, errorMessage(raw))
	}
	return c.Status(ctx, userID)
}

func (c *Client) Status(ctx context.Context, userID int64) (domain.EkycResult, error) {
	raw, code, err := c.do(ctx, http.MethodGet, userID, "", nil, "")
	if err != nil {
		return domain.EkycResult{}, err
	}
	if code != http.StatusOK {
		return domain.EkycResult{}, fmt.Errorf("%w: ekyc-service returned %d: %s", domain.ErrUnavailable, code, errorMessage(raw))
	}
	var st status
	if err := json.Unmarshal(raw, &st); err != nil {
		return domain.EkycResult{}, fmt.Errorf("%w: ekyc-service: decode status: %v", domain.ErrUnavailable, err)
	}
	var res domain.EkycResult
	if st.Identity != nil {
		doc := c.toDocument(*st.Identity)
		res.Document = &doc
	}
	if st.Verification != nil {
		res.FaceVerified = st.Verification.Status == "verified"
		res.FaceScore = st.Verification.MatchScore
	}
	return res, nil
}

func (c *Client) do(ctx context.Context, method string, userID int64, suffix string, body io.Reader, contentType string) ([]byte, int, error) {
	url := c.baseURL + "/api/v1/ekyc/" + strconv.FormatInt(userID, 10) + suffix
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("%w: ekyc-service: %v", domain.ErrUnavailable, err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return nil, 0, fmt.Errorf("%w: ekyc-service: %v", domain.ErrUnavailable, err)
	}
	return raw, resp.StatusCode, nil
}

func (c *Client) toDocument(id identity) domain.IdentityDocument {
	now := c.now()
	return domain.IdentityDocument{
		Source:        id.Source,
		Number:        fullCCCDNumber(id.DocumentNumber, id.MRZLine1),
		FullName:      strings.TrimSpace(id.Surname + " " + id.GivenNames),
		DateOfBirth:   domain.ParseMRZDate(id.DateOfBirth, now, false),
		Gender:        map[string]string{"M": "MALE", "F": "FEMALE"}[id.Sex],
		Nationality:   id.Nationality,
		ExpiryDate:    domain.ParseMRZDate(id.DateOfExpiry, now, true),
		ChecksumValid: id.ChecksumValid,
		NFCVerified:   id.NFCVerified,
		Confidence:    id.Confidence,
	}
}

func fullCCCDNumber(docNumber, line1 string) string {
	if len(line1) >= 27 {
		full := line1[15:27]
		if isDigits(full) && strings.HasSuffix(full, docNumber) {
			return full
		}
	}
	return docNumber
}

func isDigits(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return s != ""
}

func errorMessage(raw []byte) string {
	var e struct {
		Error string `json:"error"`
	}
	if json.Unmarshal(raw, &e) == nil && e.Error != "" {
		return e.Error
	}
	return strings.TrimSpace(string(raw))
}
