package handler

import (
	"net/http"
	"os"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

// InitCheck checks the initialization status of the application
func InitCheck(c *gin.Context) {
	step := 0

	// Count users
	var userCount int64
	if err := models.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check initialization status"})
		return
	}

	// If no users exist, we're at step 0 (need to create admin)
	if userCount == 0 {
		// Clear any existing auth token
		c.SetCookie("auth_token", "", -1, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{"initialized": false, "step": step})
		return
	}

	// Users exist, move to step 1
	step++

	// Check if OIDC is configured (ENV or DB)
	// We count ANY provider in DB (even disabled ones properly marked as skips)
	var oauthProviderCount int64
	models.DB.Model(&models.OAuthProvider{}).Count(&oauthProviderCount)

	oidcEnv := os.Getenv("OIDC_ISSUER") != ""

	if oidcEnv || oauthProviderCount > 0 {
		step++
	}

	// Initialization is complete when step >= 2 (Admin Created + OIDC Configured/Skipped)
	initialized := step >= 2
	c.JSON(http.StatusOK, gin.H{"initialized": initialized, "step": step})
}

// SkipOIDC creates a disabled OAuth provider to mark the step as skipped
func SkipOIDC(c *gin.Context) {
	// Check if we already have a provider (env or db)
	if os.Getenv("OIDC_ISSUER") != "" {
		c.JSON(http.StatusOK, gin.H{"message": "OIDC is already configured via environment variables"})
		return
	}

	var count int64
	models.DB.Model(&models.OAuthProvider{}).Count(&count)
	if count > 0 {
		c.JSON(http.StatusOK, gin.H{"message": "OIDC configuration or skip record already exists"})
		return
	}

	// Create a dummy disabled provider
	provider := models.OAuthProvider{
		Name:         "skipped",
		ClientID:     "skipped",
		ClientSecret: "skipped",
		Enabled:      false,
	}

	if err := models.DB.Create(&provider).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to skip OIDC step"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OIDC step skipped"})
}
