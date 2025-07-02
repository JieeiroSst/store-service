package services

import (
	"log"
	"time"

	"github.com/JIeeiroSst/kms/config"
	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/storage"
)

type Scheduler struct {
	keyService *KeyService
	stopCh     chan bool
}

func NewScheduler(db storage.Database, cache storage.Cache) *Scheduler {
	keyService = NewKeyService(db, cache)
	return &Scheduler{
		keyService: keyService,
		stopCh:     make(chan bool),
	}
}

func (s *Scheduler) Start() {
	rotationTicker := time.NewTicker(24 * time.Hour) // Check daily

	cleanupTicker := time.NewTicker(7 * 24 * time.Hour) // Weekly cleanup

	go func() {
		for {
			select {
			case <-rotationTicker.C:
				s.checkKeyRotation()
			case <-cleanupTicker.C:
				s.cleanupExpiredKeys()
			case <-s.stopCh:
				rotationTicker.Stop()
				cleanupTicker.Stop()
				return
			}
		}
	}()

	log.Println("Scheduler started")
}

func (s *Scheduler) Stop() {
	s.stopCh <- true
	log.Println("Scheduler stopped")
}

func (s *Scheduler) checkKeyRotation() {
	keys, err := s.keyService.db.ListKeys()
	if err != nil {
		log.Printf("Failed to list keys for rotation check: %v", err)
		return
	}

	rotationThreshold := time.Duration(config.AppConfig.KeyRotationDays) * 24 * time.Hour
	now := time.Now()

	for _, key := range keys {
		if key.Status != models.StatusActive {
			continue
		}

		var lastRotation time.Time
		if key.LastRotatedAt != nil {
			lastRotation = *key.LastRotatedAt
		} else {
			lastRotation = key.CreatedAt
		}

		if now.Sub(lastRotation) >= rotationThreshold {
			log.Printf("Auto-rotating key: %s", key.ID)
			if err := s.keyService.RotateKey(key.ID.String(), false); err != nil {
				log.Printf("Failed to auto-rotate key %s: %v", key.ID, err)
			}
		}
	}
}

func (s *Scheduler) cleanupExpiredKeys() {
	keys, err := s.keyService.db.ListKeys()
	if err != nil {
		log.Printf("Failed to list keys for cleanup: %v", err)
		return
	}

	now := time.Now()

	for _, key := range keys {
		if key.Status == models.StatusActive && now.After(key.ExpiresAt) {
			log.Printf("Marking expired key as deprecated: %s", key.ID)
			key.Status = models.StatusDeprecated
			if err := s.keyService.db.UpdateKey(&key); err != nil {
				log.Printf("Failed to mark key as deprecated %s: %v", key.ID, err)
			}
		}
	}
}
