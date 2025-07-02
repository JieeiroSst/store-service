package models

import (
	"time"

	"github.com/google/uuid"
)

type KeyType string
type KeyStatus string
type UserRole string
type LogLevel string

const (
	KeyTypeAES256  KeyType = "AES-256"
	KeyTypeRSA2048 KeyType = "RSA-2048"
	KeyTypeRSA4096 KeyType = "RSA-4096"
	KeyTypeECCP256 KeyType = "ECC-P256"

	StatusActive     KeyStatus = "active"
	StatusDeprecated KeyStatus = "deprecated"
	StatusDeleted    KeyStatus = "deleted"
	StatusPending    KeyStatus = "pending"

	RoleAdmin   UserRole = "admin"
	RoleUser    UserRole = "user"
	RoleAuditor UserRole = "auditor"

	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
)

type Key struct {
	ID            uuid.UUID         `json:"id" db:"id"`
	Alias         string            `json:"alias" db:"alias"`
	EncryptedKey  []byte            `json:"-" db:"encrypted_key"`
	Algorithm     KeyType           `json:"algorithm" db:"algorithm"`
	KeyLength     int               `json:"key_length" db:"key_length"`
	CreatedAt     time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time         `json:"updated_at" db:"updated_at"`
	ExpiresAt     time.Time         `json:"expires_at" db:"expires_at"`
	LastRotatedAt *time.Time        `json:"last_rotated_at" db:"last_rotated_at"`
	Status        KeyStatus         `json:"status" db:"status"`
	Version       int               `json:"version" db:"version"`
	CreatedBy     uuid.UUID         `json:"created_by" db:"created_by"`
	Tags          map[string]string `json:"tags" db:"tags"`
	UseCount      int64             `json:"use_count" db:"use_count"`
}

type User struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Username    string    `json:"username" db:"username"`
	Email       string    `json:"email" db:"email"`
	Role        UserRole  `json:"role" db:"role"`
	Permissions []string  `json:"permissions" db:"permissions"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

type AuditLog struct {
	ID         uuid.UUID              `json:"id" db:"id"`
	ActorID    uuid.UUID              `json:"actor_id" db:"actor_id"`
	ActorName  string                 `json:"actor_name" db:"actor_name"`
	Action     string                 `json:"action" db:"action"`
	Resource   string                 `json:"resource" db:"resource"`
	ResourceID string                 `json:"resource_id" db:"resource_id"`
	Timestamp  time.Time              `json:"timestamp" db:"timestamp"`
	IPAddress  string                 `json:"ip_address" db:"ip_address"`
	UserAgent  string                 `json:"user_agent" db:"user_agent"`
	Success    bool                   `json:"success" db:"success"`
	ErrorMsg   string                 `json:"error_msg,omitempty" db:"error_msg"`
	Metadata   map[string]interface{} `json:"metadata,omitempty" db:"metadata"`
	Level      LogLevel               `json:"level" db:"level"`
}

type CreateKeyRequest struct {
	Alias     string            `json:"alias" binding:"required,min=3,max=50"`
	Algorithm KeyType           `json:"algorithm" binding:"required"`
	ExpiresIn int               `json:"expires_in"` // days
	Tags      map[string]string `json:"tags"`
}

type RotateKeyRequest struct {
	Force bool `json:"force"`
}

type KeyUsageResponse struct {
	Key      Key    `json:"key"`
	PlainKey []byte `json:"plain_key,omitempty"`
}
