package application

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/JIeeiroSst/nofitifaction-service/common"
	"github.com/JIeeiroSst/nofitifaction-service/config"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/model"
	"github.com/JIeeiroSst/nofitifaction-service/internal/domain/port"
)

const maxTokenLength = 4096

type DevicePolicy struct {
	SingleDevicePerUser bool
	ValidateOnRegister  bool
	StaleAfter          time.Duration
}

func NewDevicePolicy(cfg *config.Config) DevicePolicy {
	p := DevicePolicy{
		SingleDevicePerUser: cfg.Push.SingleDevicePerUser,
		ValidateOnRegister:  true,
		StaleAfter:          270 * 24 * time.Hour,
	}
	if cfg.Push.ValidateOnRegister != nil {
		p.ValidateOnRegister = *cfg.Push.ValidateOnRegister
	}
	if cfg.Push.StaleTokenDays > 0 {
		p.StaleAfter = time.Duration(cfg.Push.StaleTokenDays) * 24 * time.Hour
	}
	return p
}

type userDeviceService struct {
	repo      port.UserDeviceRepository
	contacts  port.ContactRepository
	validator port.TokenValidator
	policy    DevicePolicy
}

func NewUserDeviceService(repo port.UserDeviceRepository, contacts port.ContactRepository, validator port.TokenValidator, policy DevicePolicy) port.UserDeviceUsecase {
	return &userDeviceService{repo: repo, contacts: contacts, validator: validator, policy: policy}
}

func normalizeContact(in port.ContactInput) (*model.UserContact, error) {
	if in.UserID == 0 {
		return nil, invalid("user_id is required")
	}
	c := &model.UserContact{UserID: in.UserID, UpdatedAt: time.Now()}
	if email := model.NormalizeEmail(in.Email); email != "" {
		if _, err := mail.ParseAddress(email); err != nil || strings.ContainsAny(email, "<> ") {
			return nil, invalid("email %q is not valid", in.Email)
		}
		c.Email = &email
	}
	if strings.TrimSpace(in.Phone) != "" {
		phone, ok := model.NormalizePhone(in.Phone)
		if !ok {
			return nil, invalid("phone %q is not valid", in.Phone)
		}
		c.Phone = &phone
	}
	return c, nil
}

func (s *userDeviceService) UpsertContact(ctx context.Context, in port.ContactInput) (*model.UserContact, error) {
	c, err := normalizeContact(in)
	if err != nil {
		return nil, err
	}
	if c.Email == nil && c.Phone == nil {
		return nil, invalid("email or phone is required")
	}
	if err := s.contacts.Upsert(ctx, c); err != nil {
		return nil, err
	}
	return s.contacts.Get(ctx, in.UserID)
}

func (s *userDeviceService) GetContact(ctx context.Context, userID uint) (*model.UserContact, error) {
	return s.contacts.Get(ctx, userID)
}

func (s *userDeviceService) DeactivateStaleTokens(ctx context.Context) (int64, error) {
	return s.repo.DeactivateStale(ctx, time.Now().Add(-s.policy.StaleAfter))
}

func (s *userDeviceService) RegisterDevice(ctx context.Context, in port.RegisterDeviceInput) (*model.UserDevice, error) {
	token := strings.TrimSpace(in.Token)
	deviceType := strings.ToLower(strings.TrimSpace(in.DeviceType))
	deviceID := strings.TrimSpace(in.DeviceID)
	switch {
	case in.UserID == 0:
		return nil, invalid("user_id is required: register the token after the user has logged in")
	case token == "":
		return nil, invalid("device_token is required")
	case len(token) > maxTokenLength:
		return nil, invalid("device_token is too long")
	case len(deviceID) > 128:
		return nil, invalid("device_id must be at most 128 characters")
	case !model.ValidDeviceType(deviceType):
		return nil, invalid("device_type must be ios, android or web")
	case model.LooksLikeAPNsToken(token):
		return nil, invalid("device_token looks like an APNs device token; send the FCM registration token from Messaging.messaging().token instead")
	}
	contact, err := normalizeContact(port.ContactInput{UserID: in.UserID, Email: in.Email, Phone: in.Phone})
	if err != nil {
		return nil, err
	}
	if s.policy.ValidateOnRegister {
		if err := s.validator.ValidateToken(ctx, token); errors.Is(err, common.ErrInvalidToken) {
			return nil, invalid("%v", err)
		}
	}

	hash := model.HashToken(token)
	now := time.Now()
	device := &model.UserDevice{
		UserID:      in.UserID,
		DeviceToken: token,
		TokenHash:   &hash,
		DeviceID:    deviceID,
		DeviceType:  deviceType,
		IsActive:    true,
		LastUsedAt:  now,
		CreatedAt:   now,
	}
	saved, err := s.repo.Register(ctx, device, port.RegisterOptions{
		ReplaceSameInstall:  deviceID != "",
		SingleDevicePerUser: s.policy.SingleDevicePerUser,
	})
	if err != nil {
		return nil, err
	}
	if contact.Email != nil || contact.Phone != nil {
		if err := s.contacts.Upsert(ctx, contact); err != nil {
			return nil, err
		}
	}
	return saved, nil
}

func (s *userDeviceService) UnregisterDevice(ctx context.Context, in port.UnregisterDeviceInput) (int64, error) {
	token, deviceID := strings.TrimSpace(in.Token), strings.TrimSpace(in.DeviceID)
	switch {
	case token != "":
		return s.repo.DeactivateByTokenHash(ctx, model.HashToken(token))
	case deviceID != "" && in.UserID != 0:
		return s.repo.DeactivateByDeviceID(ctx, in.UserID, deviceID)
	case deviceID != "":
		return 0, invalid("user_id is required together with device_id")
	case in.UserID != 0:
		return s.repo.DeactivateByUserID(ctx, in.UserID)
	default:
		return 0, invalid("user_id (all devices) or user_id + device_id (one device) is required")
	}
}

func (s *userDeviceService) GetDevice(ctx context.Context, id uint) (*model.UserDevice, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *userDeviceService) ListDevices(ctx context.Context, userID uint) ([]model.UserDevice, error) {
	if userID != 0 {
		return s.repo.ListByUserID(ctx, userID)
	}
	return s.repo.List(ctx)
}

func (s *userDeviceService) UpdateDevice(ctx context.Context, device *model.UserDevice) (*model.UserDevice, error) {
	if _, err := s.repo.GetByID(ctx, device.ID); err != nil {
		return nil, err
	}
	if device.DeviceType != "" && !model.ValidDeviceType(strings.ToLower(device.DeviceType)) {
		return nil, invalid("device_type must be ios, android or web")
	}
	device.DeviceType = strings.ToLower(device.DeviceType)
	device.DeviceToken, device.TokenHash = "", nil
	if err := s.repo.Update(ctx, device); err != nil {
		return nil, err
	}
	return s.repo.GetByID(ctx, device.ID)
}

func (s *userDeviceService) DeleteDevice(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}
