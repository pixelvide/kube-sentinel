package handler

import (
	"net/http"

	"cloud-sentinel-k8s/pkg/common"
	"cloud-sentinel-k8s/pkg/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type CreateSuperUserInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name"`
}

// CreateSuperUser creates the first admin user during initialization
func CreateSuperUser(c *gin.Context) {
	// Check if any users already exist
	var userCount int64
	if err := models.DB.Model(&models.User{}).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check user count"})
		return
	}

	if userCount > 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "Super user already exists"})
		return
	}

	var input CreateSuperUserInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create super admin user
	user := models.User{
		Email:    input.Email,
		Password: string(hashedPassword),
		Name:     input.Name,
		Role:     "admin",
		Enabled:  true,
	}

	if err := models.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Grant access to the default app
	var app models.App
	if err := models.DB.Where("name = ?", common.AppName).First(&app).Error; err == nil {
		appUser := models.AppUser{
			UserID:  user.ID,
			AppID:   app.ID,
			Enabled: true,
		}
		models.DB.Create(&appUser)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Super user created successfully",
		"user": gin.H{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}
