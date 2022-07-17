package v1

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/JIeeiroSst/user-service/common"
	"github.com/JIeeiroSst/user-service/model"
	"github.com/gin-gonic/gin"
)

func (h *Handler) initUserRoutes(api *gin.RouterGroup) {
	group := api.Group("/user")

	group.POST("/login", h.Login)
	group.POST("/sign-up", h.SignUp)
	group.PUT("/:id", h.UpdateProfile)
	group.PUT("/lock/:id", h.LockAccount)

	group.POST("/authentication", h.Authentication)
}

// Login godoc
// @Summary Login Account
// @Description login account
// @Accept  json
// @Produce  json
// @Param user  body model.Login true "user in json"
// @Success 200 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/login [post]
func (r *Handler) Login(c *gin.Context) {
	var login model.Login
	if err := c.ShouldBind(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user := model.Users{
		Username: login.Username,
		Password: login.Password,
	}
	id, token, err := r.usecase.Login(user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if len(token) == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"token": "couldn't find the token just created "})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      id,
		"token":   token,
		"message": "login success",
	})
}

// SignUp godoc
// @Summary SignUp Account
// @Description SignUp account
// @Accept  json
// @Produce  json
// @Param user body model.Users false "user in json login"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Router /api/v1/sign-up [post]
func (r *Handler) SignUp(c *gin.Context) {
	var user model.Users
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := r.usecase.SignUp(user)
	if errors.Is(err, common.PasswordFailed) {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if errors.Is(err, common.EmailFailed) {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if errors.Is(err, common.HashPasswordFailed) {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if errors.Is(err, common.UserAlready) {
		c.JSON(400, gin.H{
			"message": err.Error(),
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mesage": "signup success",
	})
}

// UpdateProfile godoc
// @Summary UpdateProfile Account
// @Description UpdateProfile account
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Param id path int false "id in json login"
// @Param user body model.Users false "user in json"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 500 {object} map[string]interface{}
// @Router /api/v1/update/profile [put]
func (r *Handler) UpdateProfile(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user model.Users
	user.UpdateTime = time.Now()
	if err := c.ShouldBind(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := r.usecase.UpdateProfile(id, user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mesage": "update profile user success",
	})
}

// LockAccount godoc
// @Summary LockAccount Account
// @Description LockAccount account
// @Accept  json
// @Produce  json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/lock_user [post]
func (r *Handler) LockAccount(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := r.usecase.LockAccount(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"mesage": "lock user success",
	})
}

// Authentication godoc
// @Summary Authentication Account
// @Description Authentication account
// @Accept  json
// @Produce  json
// @Param token body model.Token true "token"
// @Param username query string true "username"
// @Success 200 {object} map[string]interface{}
// @Success 400 {object} map[string]interface{}
// @Success 401 {object} map[string]interface{}
// @Router /api/v1/authentication [post]
func (r *Handler) Authentication(c *gin.Context) {
	username := c.Query("username")
	var token model.Token
	if err := c.ShouldBind(&token); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := r.usecase.Users.Authentication(token.EncodeToken, username)
	if errors.Is(err, common.FailedTokenUsername) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if errors.Is(err, common.FailedToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"mesage": "user authorized success",
	})
}
