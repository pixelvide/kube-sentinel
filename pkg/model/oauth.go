package model

import (
	"strings"

	"github.com/pixelvide/kube-sentinel/pkg/common"
)

type OAuthProvider struct {
	Model
	AppID        uint            `json:"appId" gorm:"not null;uniqueIndex:idx_oauth_app_name"`
	Name         LowerCaseString `json:"name" gorm:"type:varchar(100);uniqueIndex:idx_oauth_app_name;not null"`
	ClientID     string          `json:"clientId" gorm:"type:varchar(255);not null"`
	ClientSecret SecretString    `json:"clientSecret" gorm:"type:text;not null"`
	AuthURL      string          `json:"authUrl" gorm:"type:varchar(255)"`
	TokenURL     string          `json:"tokenUrl" gorm:"type:varchar(255)"`
	UserInfoURL  string          `json:"userInfoUrl" gorm:"type:varchar(255)"`
	Scopes       string          `json:"scopes" gorm:"type:varchar(255);default:'openid,profile,email'"`
	Issuer       string          `json:"issuer" gorm:"type:varchar(255)"`
	Enabled      bool            `json:"enabled" gorm:"type:boolean;default:true"`

	// Auto-generated redirect URL
	RedirectURL string `json:"-" gorm:"-"`

	App App `json:"-" gorm:"foreignKey:AppID"`
}

func (OAuthProvider) TableName() string {
	return common.GetCoreTableName("oauth_providers")
}

// GetAllOAuthProviders returns all OAuth providers from database
func GetAllOAuthProviders() ([]OAuthProvider, error) {
	var providers []OAuthProvider
	err := DB.Where("app_id = ?", CurrentApp.ID).Find(&providers).Error
	return providers, err
}

// GetEnabledOAuthProviders returns only enabled OAuth providers
func GetEnabledOAuthProviders() ([]OAuthProvider, error) {
	var providers []OAuthProvider
	err := DB.Where("app_id = ? AND enabled = ?", CurrentApp.ID, true).Find(&providers).Error
	return providers, err
}

// GetOAuthProviderByName returns an OAuth provider by name
func GetOAuthProviderByName(name string) (OAuthProvider, error) {
	var provider OAuthProvider
	name = strings.ToLower(name)
	err := DB.Where("app_id = ? AND name = ? AND enabled = ?", CurrentApp.ID, name, true).First(&provider).Error
	if err != nil {
		return OAuthProvider{}, err
	}
	return provider, nil
}

// CreateOAuthProvider creates a new OAuth provider
func CreateOAuthProvider(provider *OAuthProvider) error {
	provider.AppID = CurrentApp.ID
	return DB.Create(provider).Error
}

// UpdateOAuthProvider updates an existing OAuth provider
func UpdateOAuthProvider(provider *OAuthProvider, updates map[string]interface{}) error {
	return DB.Model(provider).Where("app_id = ?", CurrentApp.ID).Updates(updates).Error
}

// DeleteOAuthProvider deletes an OAuth provider by ID
func DeleteOAuthProvider(id uint) error {
	return DB.Where("app_id = ?", CurrentApp.ID).Delete(&OAuthProvider{}, id).Error
}
