package http

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/application"
	"github.com/JIeeiroSst/customer-info-service/internal/domain"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
)

const (
	maxBodyBytes      = 64 << 10
	maxMultipartBytes = domain.MaxDocumentImage + 1<<20
	healthTimeout     = 2 * time.Second
)

type Handler struct {
	customers *application.CustomerService
	health    port.HealthChecker
	clock     port.Clock
	tokens    [][]byte
}

func NewHandler(cfg *config.Config, customers *application.CustomerService, health port.HealthChecker, clock port.Clock) *Handler {
	h := &Handler{customers: customers, health: health, clock: clock}
	for _, t := range cfg.Server.AccessTokens {
		h.tokens = append(h.tokens, []byte(t))
	}
	return h
}

type documentDTO struct {
	Type          string  `json:"type,omitempty"`
	Source        string  `json:"source,omitempty"`
	NumberLast4   string  `json:"number_last4,omitempty"`
	FullName      string  `json:"full_name,omitempty"`
	DateOfBirth   string  `json:"date_of_birth,omitempty"`
	Gender        string  `json:"gender,omitempty"`
	Nationality   string  `json:"nationality,omitempty"`
	ExpiryDate    string  `json:"expiry_date,omitempty"`
	ChecksumValid bool    `json:"checksum_valid"`
	NFCVerified   bool    `json:"nfc_verified"`
	Confidence    float64 `json:"confidence"`
}

type kycDTO struct {
	Status       string       `json:"status"`
	FaceVerified bool         `json:"face_verified"`
	Document     *documentDTO `json:"document,omitempty"`
	Reasons      []string     `json:"reasons,omitempty"`
	SubmittedAt  *time.Time   `json:"submitted_at,omitempty"`
	VerifiedAt   *time.Time   `json:"verified_at,omitempty"`
}

type eligibilityDTO struct {
	Eligible bool     `json:"eligible"`
	Reasons  []string `json:"reasons,omitempty"`
}

type customerDTO struct {
	ID              string         `json:"id"`
	UserID          int64          `json:"user_id"`
	Username        string         `json:"username"`
	FullName        string         `json:"full_name"`
	Email           string         `json:"email"`
	Phone           string         `json:"phone"`
	Address         string         `json:"address"`
	Gender          string         `json:"gender"`
	Status          string         `json:"status"`
	KYC             kycDTO         `json:"kyc"`
	CardEligibility eligibilityDTO `json:"card_eligibility"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	SyncedAt        time.Time      `json:"synced_at"`
}

func toDocument(d domain.IdentityDocument, docType string) *documentDTO {
	return &documentDTO{
		Type:          docType,
		Source:        d.Source,
		NumberLast4:   d.Last4(),
		FullName:      d.FullName,
		DateOfBirth:   domain.FormatDocumentDate(d.DateOfBirth),
		Gender:        d.Gender,
		Nationality:   d.Nationality,
		ExpiryDate:    domain.FormatDocumentDate(d.ExpiryDate),
		ChecksumValid: d.ChecksumValid,
		NFCVerified:   d.NFCVerified,
		Confidence:    d.Confidence,
	}
}

func (h *Handler) toCustomer(c *domain.Customer) customerDTO {
	k := kycDTO{Status: string(c.KYC.Status), FaceVerified: c.KYC.FaceVerified, Reasons: c.KYC.Reasons, SubmittedAt: c.KYC.SubmittedAt, VerifiedAt: c.KYC.VerifiedAt}
	if c.KYC.Document.Number != "" || c.KYC.Document.FullName != "" {
		k.Document = toDocument(c.KYC.Document, c.KYC.DocumentType)
	}
	e := c.CardEligibility(h.clock.Now())
	return customerDTO{
		ID: c.ID, UserID: c.UserID, Username: c.Username, FullName: c.FullName, Email: c.Email,
		Phone: c.Phone, Address: c.Address, Gender: c.Gender, Status: string(c.Status), KYC: k,
		CardEligibility: eligibilityDTO{Eligible: e.Eligible, Reasons: e.Reasons},
		CreatedAt:       c.CreatedAt, UpdatedAt: c.UpdatedAt, SyncedAt: c.SyncedAt,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()
	if err := h.health.Ping(ctx); err != nil {
		log.Printf("health check failed: %v", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable", "version": config.Version})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": config.Version})
}

func (h *Handler) Onboard(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := decodeBody(r, &req); err != nil {
		h.fail(w, err)
		return
	}
	c, created, err := h.customers.Onboard(r.Context(), req.UserID)
	if err != nil {
		h.fail(w, err)
		return
	}
	status := http.StatusOK
	if created {
		status = http.StatusCreated
	}
	writeJSON(w, status, h.toCustomer(c))
}

func (h *Handler) GetCustomer(w http.ResponseWriter, r *http.Request) {
	c, err := h.customers.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toCustomer(c))
}

func (h *Handler) GetByUser(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("userId"), 10, 64)
	if err != nil {
		h.fail(w, domain.Invalid("user id must be an integer"))
		return
	}
	c, err := h.customers.GetByUserID(r.Context(), id)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toCustomer(c))
}

func (h *Handler) Sync(w http.ResponseWriter, r *http.Request) {
	c, err := h.customers.Sync(r.Context(), r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toCustomer(c))
}

func (h *Handler) SubmitKYC(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 2*maxMultipartBytes)
	front, err := readImage(r, "front")
	if err != nil {
		h.fail(w, err)
		return
	}
	back, err := readImage(r, "back")
	if err != nil {
		h.fail(w, err)
		return
	}
	c, err := h.customers.SubmitKYC(r.Context(), r.PathValue("id"), front, back)
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toCustomer(c))
}

func (h *Handler) RefreshKYC(w http.ResponseWriter, r *http.Request) {
	c, err := h.customers.RefreshKYC(r.Context(), r.PathValue("id"))
	if err != nil {
		h.fail(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h.toCustomer(c))
}

func readImage(r *http.Request, field string) ([]byte, error) {
	f, _, err := r.FormFile(field)
	if err != nil {
		return nil, domain.Invalid("multipart field %q is required (max %d MB): %v", field, domain.MaxDocumentImage>>20, err)
	}
	defer f.Close()
	return io.ReadAll(io.LimitReader(f, domain.MaxDocumentImage+1))
}

func (h *Handler) authorize(next http.Handler) http.Handler {
	if len(h.tokens) == 0 {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !h.validToken(token(r)) {
			h.fail(w, domain.ErrUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) validToken(t string) bool {
	if t == "" {
		return false
	}
	ok := 0
	for _, want := range h.tokens {
		ok |= subtle.ConstantTimeCompare([]byte(t), want)
	}
	return ok == 1
}

func token(r *http.Request) string {
	if t, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer "); ok {
		return strings.TrimSpace(t)
	}
	return r.Header.Get("X-Api-Token")
}

func (h *Handler) fail(w http.ResponseWriter, err error) {
	status, msg := errorStatus(err)
	if status >= 500 {
		log.Printf("request failed: %v", err)
	}
	writeJSON(w, status, map[string]string{"error": msg})
}

func errorStatus(err error) (int, string) {
	switch {
	case domain.IsInvalid(err):
		return http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "missing or invalid access token"
	case errors.Is(err, domain.ErrUserNotFound):
		return http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "customer not found"
	case domain.IsConflict(err):
		return http.StatusConflict, err.Error()
	case errors.Is(err, domain.ErrUnavailable):
		return http.StatusServiceUnavailable, "a dependency is unavailable, try again later"
	case errors.Is(err, context.DeadlineExceeded):
		return http.StatusGatewayTimeout, "request timed out"
	}
	return http.StatusInternalServerError, "internal error"
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxBodyBytes))
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return domain.Invalid("invalid JSON body: %v", err)
	}
	return nil
}
