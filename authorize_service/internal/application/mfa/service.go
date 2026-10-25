package mfa

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JieeiroSst/authorize-service/common"
	"github.com/JieeiroSst/authorize-service/internal/domain/model"
	"github.com/JieeiroSst/authorize-service/internal/domain/port"
)

const qrCodeSize = 256

type Config struct {
	Issuer string
}

type Service struct {
	repo     port.TOTPRepository
	box      port.SecretBox
	qr       port.QRCodeEncoder
	throttle port.FailureThrottle
	cfg      Config
	now      func() time.Time
}

func New(repo port.TOTPRepository, box port.SecretBox, qr port.QRCodeEncoder, throttle port.FailureThrottle, cfg Config) *Service {
	return &Service{repo: repo, box: box, qr: qr, throttle: throttle, cfg: cfg, now: time.Now}
}

func (s *Service) Status(ctx context.Context, userID string) (model.TOTPStatus, error) {
	if userID == "" {
		return model.TOTPStatus{}, common.ErrInvalidUserID
	}
	rec, err := s.repo.Get(ctx, userID)
	if errors.Is(err, common.ErrMFANotSetUp) || (err == nil && !rec.Enabled) {
		return model.TOTPStatus{}, nil
	}
	if err != nil {
		return model.TOTPStatus{}, err
	}
	return model.TOTPStatus{
		Enabled:                true,
		EnabledAt:              rec.EnabledAt,
		RecoveryCodesRemaining: len(parseRecoveryCodes(rec.RecoveryCodes)),
	}, nil
}

func (s *Service) SetupTOTP(ctx context.Context, userID, account string) (model.TOTPSetup, error) {
	if userID == "" {
		return model.TOTPSetup{}, common.ErrInvalidUserID
	}
	secret, err := newSecret()
	if err != nil {
		return model.TOTPSetup{}, err
	}
	sealed, err := s.box.Seal([]byte(secret))
	if err != nil {
		return model.TOTPSetup{}, err
	}
	if err := s.repo.SavePending(ctx, userID, sealed); err != nil {
		return model.TOTPSetup{}, err
	}
	if account = strings.TrimSpace(account); account == "" {
		account = userID
	}
	uri := otpauthURI(s.cfg.Issuer, account, secret)
	png, err := s.qr.PNG(uri, qrCodeSize)
	if err != nil {
		return model.TOTPSetup{}, err
	}
	return model.TOTPSetup{
		Secret:     secret,
		OtpauthURI: uri,
		QRCodePNG:  png,
		Issuer:     s.cfg.Issuer,
		Account:    account,
		Digits:     totpDigits,
		Period:     totpPeriod,
		Algorithm:  totpAlgorithm,
	}, nil
}

func (s *Service) EnableTOTP(ctx context.Context, userID, code string) ([]string, error) {
	if userID == "" {
		return nil, common.ErrInvalidUserID
	}
	rec, err := s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	if rec.Enabled {
		return nil, common.ErrMFAAlreadyEnabled
	}
	if err := s.guard(ctx, userID); err != nil {
		return nil, err
	}
	key, err := s.key(rec)
	if err != nil {
		return nil, err
	}
	step, ok := matchStep(key, strings.TrimSpace(code), s.now(), rec.LastUsedStep)
	if !ok {
		return nil, s.reject(ctx, userID)
	}
	plain, hashed, err := newRecoveryCodes()
	if err != nil {
		return nil, err
	}
	if err := s.repo.Enable(ctx, userID, step, hashed, s.now()); err != nil {
		return nil, err
	}
	_ = s.throttle.Reset(ctx, userID)
	return plain, nil
}

func (s *Service) DisableTOTP(ctx context.Context, userID, code string) error {
	rec, err := s.enabledRecord(ctx, userID)
	if err != nil {
		return err
	}
	if _, err := s.verify(ctx, rec, code, true); err != nil {
		return err
	}
	return s.repo.Delete(ctx, userID)
}

func (s *Service) RegenerateRecoveryCodes(ctx context.Context, userID, code string) ([]string, error) {
	rec, err := s.enabledRecord(ctx, userID)
	if err != nil {
		return nil, err
	}
	if _, err := s.verify(ctx, rec, code, false); err != nil {
		return nil, err
	}
	rec, err = s.repo.Get(ctx, userID)
	if err != nil {
		return nil, err
	}
	plain, hashed, err := newRecoveryCodes()
	if err != nil {
		return nil, err
	}
	ok, err := s.repo.SwapRecoveryCodes(ctx, userID, rec.RecoveryCodes, hashed)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, common.ErrMFAInvalidCode
	}
	return plain, nil
}

func (s *Service) VerifyTOTP(ctx context.Context, userID, code string, allowRecovery bool) (model.TOTPVerification, error) {
	rec, err := s.enabledRecord(ctx, userID)
	if err != nil {
		return model.TOTPVerification{}, err
	}
	return s.verify(ctx, rec, code, allowRecovery)
}

func (s *Service) verify(ctx context.Context, rec model.UserTOTP, code string, allowRecovery bool) (model.TOTPVerification, error) {
	if err := s.guard(ctx, rec.UserID); err != nil {
		return model.TOTPVerification{}, err
	}
	code = strings.TrimSpace(code)
	ok, recovery, err := s.check(ctx, rec, code, allowRecovery)
	if err != nil {
		return model.TOTPVerification{}, err
	}
	if !ok {
		return model.TOTPVerification{}, s.reject(ctx, rec.UserID)
	}
	_ = s.throttle.Reset(ctx, rec.UserID)
	return model.TOTPVerification{UsedRecoveryCode: recovery}, nil
}

func (s *Service) check(ctx context.Context, rec model.UserTOTP, code string, allowRecovery bool) (bool, bool, error) {
	if isTOTPCode(code) {
		key, err := s.key(rec)
		if err != nil {
			return false, false, err
		}
		step, ok := matchStep(key, code, s.now(), rec.LastUsedStep)
		if !ok {
			return false, false, nil
		}
		used, err := s.repo.MarkUsed(ctx, rec.UserID, step)
		return used, false, err
	}
	if !allowRecovery || code == "" {
		return false, false, nil
	}
	rest, ok := consumeRecoveryCode(rec.RecoveryCodes, code)
	if !ok {
		return false, false, nil
	}
	swapped, err := s.repo.SwapRecoveryCodes(ctx, rec.UserID, rec.RecoveryCodes, rest)
	return swapped, swapped, err
}

func (s *Service) guard(ctx context.Context, userID string) error {
	blocked, err := s.throttle.Blocked(ctx, userID)
	if err != nil {
		return err
	}
	if blocked {
		return common.ErrMFATooManyAttempts
	}
	return nil
}

func (s *Service) reject(ctx context.Context, userID string) error {
	if err := s.throttle.RecordFailure(ctx, userID); err != nil {
		return err
	}
	return common.ErrMFAInvalidCode
}

func (s *Service) enabledRecord(ctx context.Context, userID string) (model.UserTOTP, error) {
	if userID == "" {
		return model.UserTOTP{}, common.ErrInvalidUserID
	}
	rec, err := s.repo.Get(ctx, userID)
	if errors.Is(err, common.ErrMFANotSetUp) || (err == nil && !rec.Enabled) {
		return model.UserTOTP{}, common.ErrMFANotEnabled
	}
	return rec, err
}

func (s *Service) key(rec model.UserTOTP) ([]byte, error) {
	secret, err := s.box.Open(rec.Secret)
	if err != nil {
		return nil, err
	}
	return decodeSecret(string(secret))
}
