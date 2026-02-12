package model

import (
	"errors"

	"github.com/pixelvide/kube-sentinel/pkg/common"
	"gorm.io/gorm"
)

type UserAWSConfig struct {
	Model
	UserID             uint         `json:"user_id" gorm:"uniqueIndex:idx_user_aws_config_user_id;not null"`
	CredentialsContent SecretString `json:"credentials_content" gorm:"type:text"`

	// Relationships
	User User `json:"user" gorm:"foreignKey:UserID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (UserAWSConfig) TableName() string {
	return common.GetAppTableName("user_aws_configs")
}

// GetUserAWSConfig retrieves the user AWS config for a given user ID.
func GetUserAWSConfig(userID uint) (*UserAWSConfig, error) {
	var config UserAWSConfig
	err := DB.Where("user_id = ?", userID).First(&config).Error
	if err == nil {
		return &config, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	return nil, err
}
