package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/kms/models"
	"github.com/JIeeiroSst/kms/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateKeyV2(c *gin.Context) {
	var req models.CreateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("user_id")
	key, err := services.CreateKey(req, userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create key: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, key)
}

func GetKeyForUse(c *gin.Context) {
	keyID := c.Param("id")

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

	userID, _ := c.Get("user_id")
	role, _ := c.Get("role")

	logs, err := services.ListAuditLogsPaginated(userID.(uuid.UUID).String(), role.(models.UserRole), limit, offset)
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
	keys, err := services.ListKeys()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keys"})
		return
	}
	c.JSON(http.StatusOK, keys)
}

func GetKey(c *gin.Context) {
	id := c.Param("id")
	key, err := services.GetKey(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Key not found"})
		return
	}
	c.JSON(http.StatusOK, key)
}

func DeleteKey(c *gin.Context) {
	id := c.Param("id")
	err := services.DeleteKey(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
		return
	}
	c.Status(http.StatusNoContent)
}