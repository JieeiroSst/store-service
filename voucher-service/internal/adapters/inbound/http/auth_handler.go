package http

import (
	stdhttp "net/http"

	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/JIeeiroSst/voucher-service/internal/domain/user"
	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	svc authapp.AuthService
}

func NewAuthHandler(svc authapp.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

type registerRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	Role     string `json:"role"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	role := user.RoleRetail
	if req.Role != "" {
		role = user.Role(req.Role)
	}
	u, err := h.svc.Register(c.Request.Context(), authapp.RegisterInput{Email: req.Email, Password: req.Password, Role: role})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusCreated, gin.H{"id": u.ID.String(), "email": u.Email})
}

type loginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(stdhttp.StatusBadRequest, gin.H{"error": gin.H{"code": "validation_error", "message": err.Error()}})
		return
	}
	out, err := h.svc.Login(c.Request.Context(), authapp.LoginInput{Email: req.Email, Password: req.Password})
	if err != nil {
		mapError(c, err)
		return
	}
	c.JSON(stdhttp.StatusOK, out)
}
