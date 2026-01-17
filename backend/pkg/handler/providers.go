package handler

import (
	"net/http"
	"os"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

// GetProviders returns the list of configured OAuth providers
func GetProviders(c *gin.Context) {
	// Priority 1: Environment Variables (Single SSO)
	if os.Getenv("OIDC_ISSUER") != "" {
		c.JSON(http.StatusOK, []gin.H{{
			"name":      "SSO",
			"type":      "oidc",
			"login_url": "/api/v1/auth/login",
		}})
		return
	}

	// Priority 2: Database Providers
	providers, err := models.GetEnabledOAuthProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	var result []gin.H
	for _, p := range providers {
		// Skip the dummy "skipped" provider which is used to mark initialization step as complete
		if p.Name == "skipped" {
			continue
		}

		result = append(result, gin.H{
			"name":      p.Name,
			"type":      p.Type,
			"login_url": "/api/v1/auth/login", // Currently routes to the single active provider
		})
	}

	c.JSON(http.StatusOK, result)
}
