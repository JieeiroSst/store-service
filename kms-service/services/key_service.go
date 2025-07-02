package services

import (
	"crypto/rand"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/storage"
	"github.com/JIeeiroSst/kms/utils"
	"github.com/google/uuid"
)

type KeyService struct {
	db    storage.Database
	cache storage.Cache
}

func NewKeyService(db storage.Database, cache storage.Cache) *KeyService {
	return &KeyService{
		db:    db,
		cache: cache,
	}
}

func (s *KeyService) CreateKey(req models.CreateKeyRequest, userID uuid.UUID) (*models.Key, error) {
	var plainKey []byte
	var keyLength int
	var err error

	switch req.Algorithm {
	case models.KeyTypeAES256:
		plainKey = make([]byte, 32) // 256 bits
		keyLength = 256
		_, err = rand.Read(plainKey)
	default:
		return nil, errors.New("unsupported algorithm")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	encryptedKey, err := utils.EncryptWithMasterKey(plainKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt key: %w", err)
	}

	expiresAt := time.Now().AddDate(0, 0, req.ExpiresIn)
	if req.ExpiresIn == 0 {
		expiresAt = time.Now().AddDate(1, 0, 0) // Default 1 year
	}

	key := &models.Key{
		ID:           uuid.New(),
		Alias:        req.Alias,
		EncryptedKey: []byte(encryptedKey),
		Algorithm:    req.Algorithm,
		KeyLength:    keyLength,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
		Status:       models.StatusActive,
		Version:      1,
		CreatedBy:    userID,
		Tags:         req.Tags,
		UseCount:     0,
	}

	if err := s.db.SaveKey(key); err != nil {
		return nil, fmt.Errorf("failed to save key: %w", err)
	}

	s.cache.SetKey(key.ID.String(), key.EncryptedKey, time.Hour)

	return key, nil
}

func (s *KeyService) GetKey(keyID string) (*models.Key, error) {
	key, err := s.db.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	if key.Status == models.StatusDeleted {
		return nil, errors.New("key not found")
	}

	return key, nil
}

func (s *KeyService) GetKeyForUse(keyID string) (*models.KeyUsageResponse, error) {
	if encryptedKey, err := s.cache.GetKey(keyID); err == nil && encryptedKey != nil {
		plainKey, err := utils.DecryptWithMasterKey(string(encryptedKey))
		if err == nil {
			go s.incrementUseCount(keyID)

			key, _ := s.GetKey(keyID)
			return &models.KeyUsageResponse{
				Key:      *key,
				PlainKey: plainKey,
			}, nil
		}
	}

	key, err := s.GetKey(keyID)
	if err != nil {
		return nil, err
	}

	if key.Status != models.StatusActive {
		return nil, errors.New("key is not active")
	}

	if time.Now().After(key.ExpiresAt) {
		return nil, errors.New("key has expired")
	}

	plainKey, err := utils.DecryptWithMasterKey(string(key.EncryptedKey))
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt key: %w", err)
	}

	s.cache.SetKey(keyID, key.EncryptedKey, time.Hour)

	go s.incrementUseCount(keyID)

	return &models.KeyUsageResponse{
		Key:      *key,
		PlainKey: plainKey,
	}, nil
}

func (s *KeyService) RotateKey(keyID string, force bool) error {
	key, err := s.GetKey(keyID)
	if err != nil {
		return err
	}

	if key.Status != models.StatusActive && !force {
		return errors.New("cannot rotate inactive key")
	}

	var newPlainKey []byte
	switch key.Algorithm {
	case models.KeyTypeAES256:
		newPlainKey = make([]byte, 32)
		_, err = rand.Read(newPlainKey)
	default:
		return errors.New("unsupported algorithm for rotation")
	}

	if err != nil {
		return fmt.Errorf("failed to generate new key: %w", err)
	}

	newEncryptedKey, err := utils.EncryptWithMasterKey(newPlainKey)
	if err != nil {
		return fmt.Errorf("failed to encrypt new key: %w", err)
	}

	now := time.Now()
	key.EncryptedKey = []byte(newEncryptedKey)
	key.UpdatedAt = now
	key.LastRotatedAt = &now
	key.Version++
	key.ExpiresAt = now.AddDate(1, 0, 0)

	if err := s.db.UpdateKey(key); err != nil {
		return fmt.Errorf("failed to update key: %w", err)
	}

	s.cache.SetKey(keyID, key.EncryptedKey, time.Hour)

	return nil
}

func (s *KeyService) DeleteKey(keyID string) error {
	if err := s.db.MarkKeyDeleted(keyID); err != nil {
		return err
	}

	s.cache.DeleteKey(keyID)

	return nil
}

func (s *KeyService) ListKeys(userID uuid.UUID, role models.UserRole) ([]models.Key, error) {
	if role == models.RoleAdmin {
		return s.db.ListKeys()
	}

	return s.db.ListKeysByUser(userID)
}

func (s *KeyService) incrementUseCount(keyID string) {
	s.db.IncrementKeyUseCount(keyID)
}
