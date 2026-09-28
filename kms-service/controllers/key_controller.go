package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/services"
	"github.com/gin-gonic/gin"
)

func CreateKeyV2(c *gin.Context) {
	var req models.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	key, err := services.CreateKey(req, callerID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create key: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, key)
}

func GetKeyForUse(c *gin.Context) {
	keyID := c.Param("id")
	if !authorizeKey(c, keyID) {
		return
	}

	keyUsage, err := services.GetKeyForUse(keyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found or inactive: " + err.Error()})
		return
	}

	response := map[string]interface{}{
		"key":     keyUsage.Key,
		"message": "Key retrieved successfully",
	}

	c.JSON(http.StatusOK, response)
}

func RotateKeyV2(c *gin.Context) {
	keyID := c.Param("id")
	if !authorizeKey(c, keyID) {
		return
	}

	var req models.RotateKeyRequest
	c.ShouldBindJSON(&req)

	err := services.RotateKeyV2(keyID, req.Force)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Rotation failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Key rotated successfully"})
}

func GetKeyUsageStats(c *gin.Context) {
	keyID := c.Param("id")
	if !authorizeKey(c, keyID) {
		return
	}

	key, err := services.GetKey(keyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}

	stats := map[string]interface{}{
		"key_id":          key.ID,
		"alias":           key.Alias,
		"use_count":       key.UseCount,
		"created_at":      key.CreatedAt,
		"last_rotated_at": key.LastRotatedAt,
		"expires_at":      key.ExpiresAt,
		"status":          key.Status,
		"version":         key.Version,
	}

	c.JSON(http.StatusOK, stats)
}

func GetAuditLogsV2(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := services.ListAuditLogsPaginated(strconv.FormatInt(callerID(c), 10), callerRole(c), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

func GetKeyAuditLogs(c *gin.Context) {
	keyID := c.Param("id")
	if !authorizeKey(c, keyID) {
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := services.GetAuditLogsByKeyID(keyID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get key audit logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key_id": keyID,
		"logs":   logs,
		"limit":  limit,
		"offset": offset,
	})
}

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"service":   "kms-service",
		"timestamp": time.Now().Unix(),
	})
}

func ListKeys(c *gin.Context) {
	keys, err := services.ListKeys(callerID(c), callerRole(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func GetKey(c *gin.Context) {
	id := c.Param("id")
	if !authorizeKey(c, id) {
		return
	}
	key, err := services.GetKey(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}
	c.JSON(http.StatusOK, key)
}

func DeleteKey(c *gin.Context) {
	id := c.Param("id")
	if !authorizeKey(c, id) {
		return
	}
	err := services.DeleteKey(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}
func callerID(c *gin.Context) int64 {
	v, _ := c.Get("user_id")
	id, _ := v.(int64)
	return id
}

func callerRole(c *gin.Context) models.UserRole {
	v, _ := c.Get("role")
	role, _ := v.(models.UserRole)
	return role
}

func authorizeKey(c *gin.Context, keyID string) bool {
	key, err := services.GetKey(keyID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return false
	}
	switch role := callerRole(c); {
	case role == models.RoleAdmin:
		return true
	case role == models.RoleAuditor && c.Request.Method == http.MethodGet && c.FullPath() != "/api/v1/keys/:id/use":
		return true
	case key.CreatedBy == callerID(c):
		return true
	}
	c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
	return false
}
