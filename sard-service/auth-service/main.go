package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// ISO 27001 - Information Security Management
type SecurityLogger struct {
	mu sync.Mutex
}

func (sl *SecurityLogger) LogAccess(action, user, resource, ip string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	timestamp := time.Now().Format(time.RFC3339)
	log.Printf("[SECURITY-LOG] %s | IP: %s | User: %s | Action: %s | Resource: %s",
		timestamp, ip, user, action, resource)
}

func (sl *SecurityLogger) LogAudit(event, details, ip string) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	timestamp := time.Now().Format(time.RFC3339)
	log.Printf("[AUDIT-LOG] %s | IP: %s | Event: %s | Details: %s",
		timestamp, ip, event, details)
}

// JWT Claims for authentication
type Claims struct {
	EncryptedPayload string `json:"encrypted_payload"` // Encrypted actual payload
	jwt.RegisteredClaims
}

// Actual payload that will be encrypted
type UserPayload struct {
	UserID   string   `json:"user_id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	TokenID  string   `json:"token_id"` // Unique token identifier
	IssuedAt int64    `json:"issued_at"`
}

// PCI DSS - Payment Card Industry Data Security Standard
type PCICompliantVault struct {
	encryptionKey []byte
	gcm           cipher.AEAD
	tokenMap      map[string]string // token -> encrypted PAN
	panMap        map[string]string // PAN hash -> token
	mu            sync.RWMutex
	logger        *SecurityLogger
	jwtSecret     []byte
	payloadKey    []byte      // Separate key for payload encryption
	jwtGCM        cipher.AEAD // GCM for JWT payload encryption
}

func NewPCICompliantVault(masterKey, jwtSecret string) (*PCICompliantVault, error) {
	key := sha256.Sum256([]byte(masterKey))

	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	payloadKey := sha256.Sum256([]byte(jwtSecret + "_payload_encryption"))
	payloadBlock, err := aes.NewCipher(payloadKey[:])
	if err != nil {
		return nil, fmt.Errorf("failed to create payload cipher: %w", err)
	}

	jwtGCM, err := cipher.NewGCM(payloadBlock)
	if err != nil {
		return nil, fmt.Errorf("failed to create JWT GCM: %w", err)
	}

	return &PCICompliantVault{
		encryptionKey: key[:],
		gcm:           gcm,
		tokenMap:      make(map[string]string),
		panMap:        make(map[string]string),
		logger:        &SecurityLogger{},
		jwtSecret:     []byte(jwtSecret),
		payloadKey:    payloadKey[:],
		jwtGCM:        jwtGCM,
	}, nil
}

// ISO 16609 - Requirements for message authentication using cryptographic techniques
type ISO16609Token struct {
	TokenValue string    `json:"token_value"`
	TokenType  string    `json:"token_type"`
	ExpiryDate time.Time `json:"expiry_date"`
	CreatedAt  time.Time `json:"created_at"`
	MAC        string    `json:"mac"` // Message Authentication Code
}

// API Request/Response structures
type TokenizeRequest struct {
	PAN    string `json:"pan" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

type TokenizeResponse struct {
	Success bool           `json:"success"`
	Token   *ISO16609Token `json:"token,omitempty"`
	Message string         `json:"message,omitempty"`
	Error   string         `json:"error,omitempty"`
}

type DetokenizeRequest struct {
	Token  string `json:"token" binding:"required"`
	UserID string `json:"user_id" binding:"required"`
}

type DetokenizeResponse struct {
	Success bool   `json:"success"`
	PAN     string `json:"pan,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

type AuthRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Token   string `json:"token,omitempty"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

func (vault *PCICompliantVault) encryptPayload(payload *UserPayload) (string, error) {
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %w", err)
	}

	nonce := make([]byte, vault.jwtGCM.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := vault.jwtGCM.Seal(nonce, nonce, payloadJSON, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (vault *PCICompliantVault) decryptPayload(encryptedPayload string) (*UserPayload, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted payload: %w", err)
	}

	nonceSize := vault.jwtGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("encrypted payload too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := vault.jwtGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt payload: %w", err)
	}

	var payload UserPayload
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &payload, nil
}

func (vault *PCICompliantVault) generateTokenID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (vault *PCICompliantVault) JWTAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			vault.logger.LogAudit("UNAUTHORIZED_ACCESS", "Missing Authorization header", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authorization header required",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			vault.logger.LogAudit("INVALID_AUTH_FORMAT", "Invalid Authorization format", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid authorization format",
			})
			c.Abort()
			return
		}

		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return vault.jwtSecret, nil
		})

		if err != nil || !token.Valid {
			vault.logger.LogAudit("INVALID_JWT", fmt.Sprintf("Error: %v", err), c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid token claims",
			})
			c.Abort()
			return
		}

		payload, err := vault.decryptPayload(claims.EncryptedPayload)
		if err != nil {
			vault.logger.LogAudit("PAYLOAD_DECRYPTION_FAILED", fmt.Sprintf("Error: %v", err), c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Failed to decrypt token payload",
			})
			c.Abort()
			return
		}
		c.Set("user_id", payload.UserID)
		c.Set("username", payload.Username)
		c.Set("roles", payload.Roles)

		c.Next()
	}
}

// Rate limiting middleware (PCI DSS compliance)
func RateLimitMiddleware() gin.HandlerFunc {
	clients := make(map[string][]time.Time)
	var mu sync.Mutex

	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		clientIP := c.ClientIP()
		now := time.Now()

		// Clean old requests (older than 1 minute)
		if requests, exists := clients[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range requests {
				if now.Sub(reqTime) < time.Minute {
					validRequests = append(validRequests, reqTime)
				}
			}
			clients[clientIP] = validRequests
		}

		// Check rate limit (max 100 requests per minute)
		if len(clients[clientIP]) >= 100 {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Add current request
		clients[clientIP] = append(clients[clientIP], now)
		c.Next()
	}
}

// PCI DSS encryption methods
func (vault *PCICompliantVault) encryptPAN(pan string) (string, error) {
	nonce := make([]byte, vault.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	ciphertext := vault.gcm.Seal(nonce, nonce, []byte(pan), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (vault *PCICompliantVault) decryptPAN(encryptedPAN string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(encryptedPAN)
	if err != nil {
		return "", fmt.Errorf("failed to decode encrypted PAN: %w", err)
	}

	nonceSize := vault.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := vault.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt PAN: %w", err)
	}

	return string(plaintext), nil
}

// Luhn algorithm validation
func validateLuhn(number string) bool {
	var sum int
	alternate := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}

		sum += digit
		alternate = !alternate
	}

	return sum%10 == 0
}

// Generate format-preserving token
func (vault *PCICompliantVault) generateToken(pan string) (string, error) {
	if len(pan) < 12 {
		return "", errors.New("invalid PAN length")
	}

	bin := pan[:6]
	last4 := pan[len(pan)-4:]

	middleLength := len(pan) - 10
	middle := make([]byte, middleLength)

	for i := 0; i < middleLength; i++ {
		digit := make([]byte, 1)
		rand.Read(digit)
		middle[i] = '0' + (digit[0] % 10)
	}

	tokenWithoutCheck := bin + string(middle) + last4[:3]

	var sum int
	alternate := true
	for i := len(tokenWithoutCheck) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(tokenWithoutCheck[i]))
		if alternate {
			digit *= 2
			if digit > 9 {
				digit = (digit % 10) + 1
			}
		}
		sum += digit
		alternate = !alternate
	}

	checkDigit := (10 - (sum % 10)) % 10
	token := tokenWithoutCheck + strconv.Itoa(checkDigit)

	return token, nil
}

// Generate MAC for ISO 16609
func (vault *PCICompliantVault) generateMAC(data string) string {
	h := sha256.New()
	h.Write([]byte(data))
	h.Write(vault.encryptionKey)
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// API Endpoints

// Login endpoint
func (vault *PCICompliantVault) login(c *gin.Context) {
	var req AuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, AuthResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	// Simple authentication (in production, use proper user management)
	validUsers := map[string]string{
		"admin":    "admin123",
		"operator": "op123",
	}

	if password, exists := validUsers[req.Username]; !exists || password != req.Password {
		vault.logger.LogAudit("LOGIN_FAILED", fmt.Sprintf("Username: %s", req.Username), c.ClientIP())
		c.JSON(http.StatusUnauthorized, AuthResponse{
			Success: false,
			Error:   "Invalid credentials",
		})
		return
	}

	// Create user payload that will be encrypted
	tokenID := vault.generateTokenID()
	userPayload := &UserPayload{
		UserID:   req.Username + "_id",
		Username: req.Username,
		Roles:    []string{"token_operator"},
		TokenID:  tokenID,
		IssuedAt: time.Now().Unix(),
	}

	// Encrypt the payload
	encryptedPayload, err := vault.encryptPayload(userPayload)
	if err != nil {
		vault.logger.LogAudit("PAYLOAD_ENCRYPTION_FAILED", fmt.Sprintf("Username: %s, Error: %s", req.Username, err.Error()), c.ClientIP())
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Error:   "Failed to create secure token",
		})
		return
	}

	// Create JWT claims with encrypted payload
	claims := &Claims{
		EncryptedPayload: encryptedPayload,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   "encrypted_token", // Generic subject to avoid information disclosure
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(vault.jwtSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, AuthResponse{
			Success: false,
			Error:   "Failed to generate token",
		})
		return
	}

	vault.logger.LogAudit("LOGIN_SUCCESS", fmt.Sprintf("Username: %s, TokenID: %s", req.Username, tokenID), c.ClientIP())
	c.JSON(http.StatusOK, AuthResponse{
		Success: true,
		Token:   tokenString,
		Message: "Login successful",
	})
}

// Tokenize endpoint
func (vault *PCICompliantVault) tokenize(c *gin.Context) {
	var req TokenizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TokenizeResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	username := c.GetString("username")
	vault.logger.LogAccess("TOKENIZE_REQUEST", username, "PAN_TOKENIZATION", c.ClientIP())

	// Validate PAN
	if !validateLuhn(req.PAN) {
		vault.logger.LogAudit("INVALID_PAN", fmt.Sprintf("User: %s", username), c.ClientIP())
		c.JSON(http.StatusBadRequest, TokenizeResponse{
			Success: false,
			Error:   "Invalid PAN - failed Luhn check",
		})
		return
	}

	// Create token
	tokenData, err := vault.createToken(req.PAN, username, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusInternalServerError, TokenizeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TokenizeResponse{
		Success: true,
		Token:   tokenData,
		Message: "Token created successfully",
	})
}

// Detokenize endpoint
func (vault *PCICompliantVault) detokenize(c *gin.Context) {
	var req DetokenizeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, DetokenizeResponse{
			Success: false,
			Error:   "Invalid request format",
		})
		return
	}

	username := c.GetString("username")
	vault.logger.LogAccess("DETOKENIZE_REQUEST", username, "PAN_RETRIEVAL", c.ClientIP())

	pan, err := vault.detokenizePAN(req.Token, username, c.ClientIP())
	if err != nil {
		c.JSON(http.StatusNotFound, DetokenizeResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, DetokenizeResponse{
		Success: true,
		PAN:     pan,
		Message: "PAN retrieved successfully",
	})
}

// Health check endpoint
func (vault *PCICompliantVault) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "1.0.0",
	})
}

// Core business logic methods
func (vault *PCICompliantVault) createToken(pan, userID, ip string) (*ISO16609Token, error) {
	vault.mu.Lock()
	defer vault.mu.Unlock()

	// Check if token already exists
	panHash := fmt.Sprintf("%x", sha256.Sum256([]byte(pan)))
	if existingToken, exists := vault.panMap[panHash]; exists {
		vault.logger.LogAudit("TOKEN_REUSE", fmt.Sprintf("User: %s", userID), ip)
		tokenData := &ISO16609Token{
			TokenValue: existingToken,
			TokenType:  "PAN_TOKEN",
			ExpiryDate: time.Now().Add(24 * time.Hour * 365),
			CreatedAt:  time.Now(),
		}
		tokenData.MAC = vault.generateMAC(tokenData.TokenValue + tokenData.TokenType)
		return tokenData, nil
	}

	// Generate new token
	token, err := vault.generateToken(pan)
	if err != nil {
		vault.logger.LogAudit("TOKEN_GENERATION_FAILED", fmt.Sprintf("User: %s, Error: %s", userID, err.Error()), ip)
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Encrypt PAN
	encryptedPAN, err := vault.encryptPAN(pan)
	if err != nil {
		vault.logger.LogAudit("ENCRYPTION_FAILED", fmt.Sprintf("User: %s, Error: %s", userID, err.Error()), ip)
		return nil, fmt.Errorf("failed to encrypt PAN: %w", err)
	}

	// Store mappings
	vault.tokenMap[token] = encryptedPAN
	vault.panMap[panHash] = token

	// Create token response
	tokenData := &ISO16609Token{
		TokenValue: token,
		TokenType:  "PAN_TOKEN",
		ExpiryDate: time.Now().Add(24 * time.Hour * 365),
		CreatedAt:  time.Now(),
	}
	tokenData.MAC = vault.generateMAC(tokenData.TokenValue + tokenData.TokenType)

	vault.logger.LogAudit("TOKEN_CREATED", fmt.Sprintf("User: %s, Token: %s", userID, token[:8]+"****"), ip)
	return tokenData, nil
}

func (vault *PCICompliantVault) detokenizePAN(token, userID, ip string) (string, error) {
	vault.mu.RLock()
	defer vault.mu.RUnlock()
	fmt.Println("===========================================", token, "===========", vault.tokenMap)

	encryptedPAN, exists := vault.tokenMap[token]
	if !exists {
		vault.logger.LogAudit("DETOKENIZE_FAILED", fmt.Sprintf("User: %s, Token: %s", userID, token[:8]+"****"), ip)
		return "", errors.New("token not found")
	}

	pan, err := vault.decryptPAN(encryptedPAN)
	if err != nil {
		vault.logger.LogAudit("DECRYPTION_FAILED", fmt.Sprintf("User: %s, Error: %s", userID, err.Error()), ip)
		return "", fmt.Errorf("failed to decrypt PAN: %w", err)
	}

	vault.logger.LogAudit("PAN_RETRIEVED", fmt.Sprintf("User: %s, Token: %s", userID, token[:8]+"****"), ip)
	return pan, nil
}

func main() {
	// Initialize vault
	vault, err := NewPCICompliantVault(
		"K9#mX7$qL2@nR5!wF8*jP1&vB4%cE6+h",
		"aB9#kM3$xZ7!qW2@nL5%pR8*jF4&vC6+hE1^yT0~uI9-sG3_dA7|bN5:oK2}mX8[",
	)
	if err != nil {
		log.Fatal("Failed to initialize vault:", err)
	}

	// Setup Gin router
	r := gin.New()

	// Global middleware
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(RateLimitMiddleware())

	// Security headers middleware
	r.Use(func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Content-Security-Policy", "default-src 'self'")
		c.Next()
	})

	// Public endpoints
	r.POST("/api/v1/auth/login", vault.login)
	r.GET("/api/v1/health", vault.healthCheck)

	// Protected endpoints
	protected := r.Group("/api/v1")
	protected.Use(vault.JWTAuthMiddleware())
	{
		protected.POST("/tokens/create", vault.tokenize)
		protected.POST("/tokens/detokenize", vault.detokenize)
	}

	fmt.Println("=== ISO Compliant Token API Server ===")
	fmt.Println("Server starting on :8081")
	fmt.Println("\nEndpoints:")
	fmt.Println("POST /api/v1/auth/login")
	fmt.Println("POST /api/v1/tokens/create")
	fmt.Println("POST /api/v1/tokens/detokenize")
	fmt.Println("GET  /api/v1/health")
	fmt.Println("\nCompliance Features:")
	fmt.Println("✓ ISO 27001: JWT Authentication, Access Logging, Audit Trails")
	fmt.Println("✓ PCI DSS: AES-256-GCM Encryption, Rate Limiting, Security Headers")
	fmt.Println("✓ ISO 16609: MAC Generation, Message Authentication")
	fmt.Println("✓ ENCRYPTED JWT: Payload encryption prevents information disclosure")
	fmt.Println("✓ SECURITY: Even jwt.io cannot decode the actual user information")

	log.Fatal(r.Run(":8081"))
}

/*
Example API Usage:

1. Login:
curl -X POST http://localhost:8081/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

2. Create Token:
curl -X POST http://localhost:8081/api/v1/tokens/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"pan":"4532015112830366","user_id":"admin"}'

3. Detokenize:
curl -X POST http://localhost:8081/api/v1/tokens/detokenize \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -d '{"token":"4532015830080369","user_id":"admin"}'

4. Health Check:
curl http://localhost:8081/api/v1/health

SECURITY FEATURE: JWT Payload Encryption
- The JWT token payload is encrypted using AES-256-GCM
- Even if someone copies the JWT token and uses jwt.io, they will only see:
  {
    "encrypted_payload": "base64_encrypted_data...",
    "sub": "encrypted_token",
    "exp": 1234567890,
    "iat": 1234567890
  }
- The actual user information (username, roles, user_id) is encrypted and cannot be decoded
- Only the server with the correct encryption key can decrypt the payload
- This prevents information disclosure even if JWT tokens are intercepted

Security Benefits:
✓ Zero-knowledge JWT: No sensitive data visible in token structure
✓ Prevents reconnaissance: Attackers cannot see user roles or permissions
✓ Compliance: Meets data protection requirements for sensitive applications
✓ Tamper-proof: Encrypted payload prevents payload manipulation attacks
*/
