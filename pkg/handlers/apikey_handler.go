package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/pixelvide/kube-sentinel/pkg/rbac"
)

type CreateAPIKeyRequest struct {
	Name      string `json:"name" binding:"required"`
	ExpiresAt string `json:"expiresAt"` // Optional, format: 2006-01-02
}

func ListAPIKeys(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	u := user.(model.User)

	var userID uint
	// If not admin, only show own tokens. If admin, show all (userID=0).
	if !rbac.UserHasRole(u, model.DefaultAdminRole.Name) {
		userID = u.ID
	}

	tokens, err := model.ListPersonalAccessTokens(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list API keys"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"apiKeys": tokens})
}

func CreateAPIKey(c *gin.Context) {
	var req CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	u := user.(model.User)

	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		t, err := time.Parse("2006-01-02", req.ExpiresAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expiry date format, use YYYY-MM-DD"})
			return
		}
		if t.After(time.Now().AddDate(0, 0, 365)) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "maximum expiry is 365 days"})
			return
		}
		expiresAt = &t
	}

	token, pat, err := model.NewPersonalAccessToken(u.ID, req.Name, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create API key: %v", err)})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"apiKey": pat,
		"token":  token, // Cleartext token ONLY on creation
	})
}

func DeleteAPIKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ID"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	u := user.(model.User)

	if err := model.DeletePersonalAccessToken(uint(id), u.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete API key"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}
