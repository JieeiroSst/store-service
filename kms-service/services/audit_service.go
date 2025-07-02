package services

import (
	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/storage"
	"github.com/google/uuid"
)

type AuditService struct {
	db storage.Database
}

func NewAuditService(db storage.Database) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) SaveAuditLog(log models.AuditLog) error {
	return s.db.SaveAuditLog(log)
}

func (s *AuditService) ListAuditLogs(userID string, role models.UserRole, limit, offset int) ([]models.AuditLog, error) {
	if role == models.RoleAdmin || role == models.RoleAuditor {
		return s.db.ListAuditLogs(limit, offset)
	}

	return s.db.ListAuditLogsByUser(userID, limit, offset)
}

func (s *AuditService) GetAuditLogsByKeyID(keyID string, limit, offset int) ([]models.AuditLog, error) {
	return s.db.GetAuditLogsByResource(keyID, limit, offset)
}

var (
	keyService   *KeyService
	auditService *AuditService
)

func InitServices(db storage.Database, cache storage.Cache) {
	keyService = NewKeyService(db, cache)
	auditService = NewAuditService(db)
}

func CreateKey(req models.CreateKeyRequest, userID uuid.UUID) (*models.Key, error) {
	return keyService.CreateKey(req, userID)
}

func GetKey(keyID string) (*models.Key, error) {
	return keyService.GetKey(keyID)
}

func RotateKey(keyID string) error {
	return keyService.RotateKey(keyID, false)
}

func DeleteKey(keyID string) error {
	return keyService.DeleteKey(keyID)
}

func ListKeys() ([]models.Key, error) {
	return keyService.db.ListKeys()
}

func ListAuditLogs() ([]models.AuditLog, error) {
	return auditService.db.ListAuditLogs(100, 0)
}

func SaveAuditLog(log models.AuditLog) error {
	return auditService.SaveAuditLog(log)
}
