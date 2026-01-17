package handler

import (
	"cloud-sentinel-k8s/auth"
	"cloud-sentinel-k8s/pkg/common"
	"cloud-sentinel-k8s/pkg/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PasswordLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func PasswordLogin(c *gin.Context) {
	var req PasswordLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request payload"})
		return
	}

	// Check if local login is enabled
	enabled, err := models.IsLocalLoginEnabled()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check login configuration"})
		return
	}
	if !enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "Password login is disabled"})
		return
	}

	// Username in frontend maps to Email in User model
	user, err := models.GetUserByEmail(req.Username)
	if err != nil {
		// Return generic error to avoid user enumeration
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.CheckPassword(req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !user.Enabled {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is disabled"})
		return
	}

	if err := models.LoginUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update login status"})
		return
	}

	// Generate JWT
	token, err := auth.GenerateToken(user.ID, user.Email, user.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Set auth cookie
	c.SetCookie("auth_token", token, common.CookieExpirationSeconds, "/", "", false, true)

	c.Status(http.StatusNoContent)
}
