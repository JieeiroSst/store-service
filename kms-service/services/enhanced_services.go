package services

import (
	"github.com/JIeeiroSst/kms/models"
)

func GetKeyForUse(keyID string) (*models.KeyUsageResponse, error) {
	return keyService.GetKeyForUse(keyID)
}

func RotateKeyV2(keyID string, force bool) error {
	return keyService.RotateKey(keyID, force)
}

func ListAuditLogsPaginated(userID string, role models.UserRole, limit, offset int) ([]models.AuditLog, error) {
	return auditService.ListAuditLogs(userID, role, limit, offset)
}

func GetAuditLogsByKeyID(keyID string, limit, offset int) ([]models.AuditLog, error) {
	return auditService.GetAuditLogsByKeyID(keyID, limit, offset)
}
