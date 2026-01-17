package models

import (
	"errors"
)

type AppConfig struct {
	Model
	AppID uint   `gorm:"not null;uniqueIndex:idx_app_key" json:"app_id"`
	App   App    `json:"-"`
	Key   string `gorm:"not null;uniqueIndex:idx_app_key" json:"key"`
	Value string `json:"value"`
}

// IsLocalLoginEnabled checks if local login is enabled for the default app
func IsLocalLoginEnabled() (bool, error) {
	if GlobalApp == nil {
		return false, nil
	}

	var config AppConfig
	if err := DB.Where("app_id = ? AND key = ?", GlobalApp.ID, "LOCAL_LOGIN_ENABLED").First(&config).Error; err != nil {
		return false, err
	}

	return config.Value == "true", nil
}

// GetAppConfig retrieves an AppConfig by key for the global app
func GetAppConfig(key string) (*AppConfig, error) {
	if GlobalApp == nil {
		return nil, errors.New("GlobalApp is not initialized")
	}

	var config AppConfig
	if err := DB.Where("app_id = ? AND key = ?", GlobalApp.ID, key).First(&config).Error; err != nil {
		return nil, err
	}

	return &config, nil
}

// CreateAppConfig creates a new AppConfig for the global app
func CreateAppConfig(config *AppConfig) error {
	if GlobalApp == nil {
		return errors.New("GlobalApp is not initialized")
	}

	config.AppID = GlobalApp.ID
	return DB.Create(config).Error
}

// UpdateAppConfig updates an existing AppConfig for the global app
func UpdateAppConfig(config *AppConfig) error {
	if GlobalApp == nil {
		return errors.New("GlobalApp is not initialized")
	}

	return DB.Model(&AppConfig{}).Where("app_id = ? AND key = ?", GlobalApp.ID, config.Key).Update("value", config.Value).Error
}
