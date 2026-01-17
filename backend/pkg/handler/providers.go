package handler

import (
	"net/http"
	"os"

	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
)

// GetProviders returns the list of configured OAuth providers
func GetProviders(c *gin.Context) {
	// Check if local login (password) is enabled
	localLoginEnabled, _ := models.IsLocalLoginEnabled()

	// Priority 1: Environment Variables (Single SSO)
	if os.Getenv("OIDC_ISSUER") != "" {
		providers := []string{"SSO"}
		if localLoginEnabled {
			providers = append(providers, "password")
		}
		c.JSON(http.StatusOK, providers)
		return
	}

	// Priority 2: Database Providers
	providers, err := models.GetEnabledOAuthProviders()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch providers"})
		return
	}

	var result []string
	for _, p := range providers {
		// Skip the dummy "skipped" provider which is used to mark initialization step as complete
		if p.Name == "skipped" {
			continue
		}

		result = append(result, p.Name)
	}

	// Manually add "password" provider if enabled
	if localLoginEnabled {
		result = append(result, "password")
	}

	c.JSON(http.StatusOK, result)
}
