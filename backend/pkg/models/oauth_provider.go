package models

import (
	"strings"
)

type OAuthProvider struct {
	Model
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	ClientID     string `gorm:"not null" json:"client_id"`
	ClientSecret string `gorm:"not null" json:"-"` // Never expose via JSON
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
	Scopes       string `gorm:"default:'openid,profile,email'" json:"scopes"`
	Issuer       string `json:"issuer"`
	Enabled      bool   `gorm:"default:true" json:"enabled"`

	// Auto-generated redirect URL
	RedirectURL string `json:"-" gorm:"-"`
}

// GetEnabledOAuthProviders retrieves all enabled OAuth providers from the database
func GetEnabledOAuthProviders() ([]OAuthProvider, error) {
	var providers []OAuthProvider
	err := DB.Where("enabled = ?", true).Find(&providers).Error
	return providers, err
}

// GetAllOAuthProviders retrieves all OAuth providers (including disabled ones)
func GetAllOAuthProviders() ([]OAuthProvider, error) {
	var providers []OAuthProvider
	err := DB.Find(&providers).Error
	return providers, err
}

// GetOAuthProviderByName retrieves a provider by its name (case-insensitive)
func GetOAuthProviderByName(name string) (OAuthProvider, error) {
	var provider OAuthProvider
	name = strings.ToLower(name)
	err := DB.Where("name = ? AND enabled = ?", name, true).First(&provider).Error
	if err != nil {
		return OAuthProvider{}, err
	}
	return provider, nil
}

// CreateOAuthProvider creates a new OAuth provider in the database
func CreateOAuthProvider(provider *OAuthProvider) error {
	// Ensure name is lowercase
	provider.Name = strings.ToLower(provider.Name)
	return DB.Create(provider).Error
}

// UpdateOAuthProvider updates an existing OAuth provider
func UpdateOAuthProvider(provider *OAuthProvider, updates map[string]interface{}) error {
	return DB.Model(provider).Updates(updates).Error
}
