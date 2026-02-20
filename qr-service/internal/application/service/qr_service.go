package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"

	"github.com/JIeeiroSst/qr-service/internal/domain/entity"
	"github.com/JIeeiroSst/qr-service/internal/domain/port"
	"github.com/JIeeiroSst/qr-service/internal/infrastructure/config"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"go.uber.org/zap"
)

const shortCodeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type qrCodeService struct {
	qrRepo   port.QRCodeRepository
	scanRepo port.ScanHistoryRepository
	cfg      *config.Config
	logger   *zap.Logger
}

func NewQRCodeService(
	qrRepo port.QRCodeRepository,
	scanRepo port.ScanHistoryRepository,
	cfg *config.Config,
	logger *zap.Logger,
) port.QRCodeService {
	return &qrCodeService{
		qrRepo:   qrRepo,
		scanRepo: scanRepo,
		cfg:      cfg,
		logger:   logger,
	}
}

func (s *qrCodeService) Generate(ctx context.Context, req *port.GenerateQRRequest) (*port.QRResponse, error) {
	shortCode := generateShortCode(8)
	scanURL := fmt.Sprintf("%s/qr/scan/%s", s.cfg.App.BaseURL, shortCode)

	encodedContent := req.Content
	if req.Type == entity.QRCodeTypeURL && req.RedirectURL == "" {
		req.RedirectURL = req.Content
		encodedContent = scanURL
	} else if req.RedirectURL != "" {
		encodedContent = scanURL
	}

	size := req.Size
	if size <= 0 {
		size = 256
	}
	foreColor := req.ForeColor
	if foreColor == "" {
		foreColor = "#000000"
	}
	backColor := req.BackColor
	if backColor == "" {
		backColor = "#FFFFFF"
	}

	imgBase64, err := generateQRImage(encodedContent, size)
	if err != nil {
		return nil, fmt.Errorf("generate qr image: %w", err)
	}

	now := time.Now()
	qr := &entity.QRCode{
		ID:          uuid.New().String(),
		ShortCode:   shortCode,
		Title:       req.Title,
		Type:        req.Type,
		Content:     req.Content,
		RedirectURL: req.RedirectURL,
		ImageBase64: imgBase64,
		Status:      entity.QRCodeStatusActive,
		ScanCount:   0,
		Size:        size,
		ForeColor:   foreColor,
		BackColor:   backColor,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   req.CreatedBy,
	}

	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			qr.ExpiresAt = &t
		}
	}

	created, err := s.qrRepo.Create(ctx, qr)
	if err != nil {
		return nil, fmt.Errorf("create qr code: %w", err)
	}

	return &port.QRResponse{
		QRCode:      created,
		ImageBase64: imgBase64,
		ScanURL:     scanURL,
	}, nil
}

func (s *qrCodeService) GetByID(ctx context.Context, id string) (*entity.QRCode, error) {
	return s.qrRepo.GetByID(ctx, id)
}

func (s *qrCodeService) GetByShortCode(ctx context.Context, shortCode string) (*entity.QRCode, error) {
	return s.qrRepo.GetByShortCode(ctx, shortCode)
}

func (s *qrCodeService) List(ctx context.Context, filter port.QRCodeFilter) (*port.QRListResponse, error) {
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	data, total, err := s.qrRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	totalPages := total / filter.Limit
	if total%filter.Limit != 0 {
		totalPages++
	}

	return &port.QRListResponse{
		Data:       data,
		Total:      total,
		Page:       filter.Page,
		Limit:      filter.Limit,
		TotalPages: totalPages,
	}, nil
}

func (s *qrCodeService) Update(ctx context.Context, id string, req *port.UpdateQRRequest) (*entity.QRCode, error) {
	existing, err := s.qrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != "" {
		existing.Title = req.Title
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.ForeColor != "" {
		existing.ForeColor = req.ForeColor
	}
	if req.BackColor != "" {
		existing.BackColor = req.BackColor
	}
	if req.ExpiresAt != nil && *req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, *req.ExpiresAt)
		if err == nil {
			existing.ExpiresAt = &t
		}
	}
	existing.UpdatedAt = time.Now()

	return s.qrRepo.Update(ctx, id, existing)
}

func (s *qrCodeService) UpdateContent(ctx context.Context, id string, req *port.UpdateContentRequest) (*entity.QRCode, error) {
	existing, err := s.qrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	redirectURL := req.RedirectURL
	if redirectURL == "" {
		redirectURL = req.Content
	}

	updated, err := s.qrRepo.UpdateContent(ctx, id, req.Content, redirectURL)
	if err != nil {
		return nil, err
	}

	s.logger.Info("QR code content updated dynamically",
		zap.String("id", id),
		zap.String("short_code", existing.ShortCode),
	)

	return updated, nil
}

func (s *qrCodeService) Delete(ctx context.Context, id string) error {
	if err := s.scanRepo.DeleteByQRCodeID(ctx, id); err != nil {
		s.logger.Warn("failed to delete scan history", zap.String("qr_id", id), zap.Error(err))
	}
	return s.qrRepo.Delete(ctx, id)
}

func (s *qrCodeService) Redirect(ctx context.Context, shortCode string, meta *port.ScanMeta) (string, error) {
	qr, err := s.qrRepo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return "", fmt.Errorf("qr code not found: %w", err)
	}

	if !qr.IsActive() {
		return "", fmt.Errorf("qr code is not active or has expired")
	}

	go func() {
		history := &entity.ScanHistory{
			ID:        uuid.New().String(),
			QRCodeID:  qr.ID,
			ShortCode: shortCode,
			IPAddress: meta.IPAddress,
			UserAgent: meta.UserAgent,
			Referer:   meta.Referer,
			ScannedAt: time.Now(),
		}
		history.DeviceType, history.OS, history.Browser = parseUserAgent(meta.UserAgent)

		bgCtx := context.Background()
		if err := s.scanRepo.Create(bgCtx, history); err != nil {
			s.logger.Error("failed to save scan history", zap.Error(err))
		}
		if err := s.qrRepo.IncrementScanCount(bgCtx, qr.ID); err != nil {
			s.logger.Error("failed to increment scan count", zap.Error(err))
		}
	}()

	if qr.RedirectURL != "" {
		return qr.RedirectURL, nil
	}
	return qr.Content, nil
}

func (s *qrCodeService) Regenerate(ctx context.Context, id string) (*port.QRResponse, error) {
	qr, err := s.qrRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	scanURL := fmt.Sprintf("%s/qr/scan/%s", s.cfg.App.BaseURL, qr.ShortCode)
	encodedContent := qr.Content
	if qr.RedirectURL != "" {
		encodedContent = scanURL
	}

	imgBase64, err := generateQRImage(encodedContent, qr.Size)
	if err != nil {
		return nil, fmt.Errorf("generate qr image: %w", err)
	}

	qr.ImageBase64 = imgBase64
	qr.UpdatedAt = time.Now()
	updated, err := s.qrRepo.Update(ctx, id, qr)
	if err != nil {
		return nil, err
	}

	return &port.QRResponse{
		QRCode:      updated,
		ImageBase64: imgBase64,
		ScanURL:     scanURL,
	}, nil
}

func generateQRImage(content string, size int) (string, error) {
	var buf bytes.Buffer
	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", err
	}
	buf.Write(png)
	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func generateShortCode(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = shortCodeChars[rand.Intn(len(shortCodeChars))]
	}
	return string(b)
}

func parseUserAgent(ua string) (deviceType, os, browser string) {
	deviceType = "desktop"
	os = "unknown"
	browser = "unknown"

	if ua == "" {
		return
	}

	switch {
	case contains(ua, "Mobile") || contains(ua, "Android") || contains(ua, "iPhone"):
		deviceType = "mobile"
	case contains(ua, "Tablet") || contains(ua, "iPad"):
		deviceType = "tablet"
	}

	switch {
	case contains(ua, "Windows"):
		os = "Windows"
	case contains(ua, "Mac OS"):
		os = "macOS"
	case contains(ua, "Linux"):
		os = "Linux"
	case contains(ua, "Android"):
		os = "Android"
	case contains(ua, "iOS") || contains(ua, "iPhone") || contains(ua, "iPad"):
		os = "iOS"
	}

	switch {
	case contains(ua, "Chrome") && !contains(ua, "Edg"):
		browser = "Chrome"
	case contains(ua, "Firefox"):
		browser = "Firefox"
	case contains(ua, "Safari") && !contains(ua, "Chrome"):
		browser = "Safari"
	case contains(ua, "Edg"):
		browser = "Edge"
	}

	return
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		len(s) > 0 && len(substr) > 0 &&
			func() bool {
				for i := 0; i <= len(s)-len(substr); i++ {
					if s[i:i+len(substr)] == substr {
						return true
					}
				}
				return false
			}())
}
