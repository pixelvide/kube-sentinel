package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pixelvide/kube-sentinel/pkg/model"
	"github.com/pixelvide/kube-sentinel/pkg/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type UpsertUserGitlabConfigReq struct {
	GitlabHostID uint   `json:"gitlab_host_id" binding:"required"`
	Token        string `json:"token" binding:"required"`
}

func ListUserGitlabConfigs(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	var configs []model.UserGitlabConfig
	if err := model.DB.Preload("GitlabHost").Where("user_id = ?", u.ID).Find(&configs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, configs)
}

func UpsertUserGitlabConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	var req UpsertUserGitlabConfigReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var config model.UserGitlabConfig
	err := model.DB.Where("user_id = ? AND gitlab_host_id = ?", u.ID, req.GitlabHostID).First(&config).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create
			config = model.UserGitlabConfig{
				UserID:       u.ID,
				GitlabHostID: req.GitlabHostID,
				Token:        req.Token,
			}
			if err := model.DB.Create(&config).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			model.DB.Preload("GitlabHost").First(&config, config.ID)
			c.JSON(http.StatusCreated, config)
			return
		}
		// Database Error
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update
	config.Token = req.Token
	// Reset validation status on token update
	config.IsValidated = false

	if err := model.DB.Save(&config).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	model.DB.Preload("GitlabHost").First(&config, config.ID)
	c.JSON(http.StatusOK, config)
}

func ValidateUserGitlabConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var config model.UserGitlabConfig
	if err := model.DB.Preload("GitlabHost").Where("id = ? AND user_id = ?", id, u.ID).First(&config).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	// Get User Config for Storage Namespace
	userConfig, err := model.GetUserConfig(u.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get user config"})
		return
	}

	glabConfigDir, err := utils.GetUserGlabConfigDir(userConfig.StorageNamespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get glab config dir"})
		return
	}

	if err := utils.GlabAuthLogin(config.GitlabHost.Host, config.Token, glabConfigDir); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validation successful
	config.IsValidated = true
	model.DB.Save(&config)

	c.JSON(http.StatusOK, gin.H{"message": "valid", "config": config})
}

func DeleteUserGitlabConfig(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	u := user.(model.User)

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	result := model.DB.Where("id = ? AND user_id = ?", id, u.ID).Delete(&model.UserGitlabConfig{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "config not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func ListGitlabHosts(c *gin.Context) {
	var hosts []model.GitlabHosts
	if err := model.DB.Find(&hosts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, hosts)
}

func RestoreGitlabConfigs() {
	var configs []model.UserGitlabConfig
	if err := model.DB.Preload("GitlabHost").Where("is_validated = ?", true).Find(&configs).Error; err != nil {
		klog.Errorf("Failed to fetch user gitlab configs for restoration: %v", err)
		return
	}

	for _, config := range configs {
		userConfig, err := model.GetUserConfig(config.UserID)
		if err != nil {
			klog.Errorf("Failed to get user config for user %d: %v", config.UserID, err)
			continue
		}

		glabConfigDir, err := utils.GetUserGlabConfigDir(userConfig.StorageNamespace)
		if err != nil {
			klog.Errorf("Failed to get glab config dir for user %d: %v", config.UserID, err)
			continue
		}

		if err := utils.GlabAuthLogin(config.GitlabHost.Host, config.Token, glabConfigDir); err != nil {
			klog.Errorf("Failed to restore gitlab auth for user %d host %s: %v", config.UserID, config.GitlabHost.Host, err)
			// Optional: Invalidate config if login fails?
			// config.IsValidated = false
			// model.DB.Save(&config)
		} else {
			klog.Infof("Restored gitlab auth for user %d host %s", config.UserID, config.GitlabHost.Host)
		}
	}
}
