package models

type OAuthProvider struct {
	Model
	Name         string `gorm:"uniqueIndex;not null" json:"name"`
	Type         string `gorm:"not null;default:'oidc'" json:"type"` // e.g., "oidc", "gitlab", "github"
	ClientID     string `gorm:"not null" json:"client_id"`
	ClientSecret string `gorm:"not null" json:"-"` // Never expose via JSON
	AuthURL      string `json:"auth_url"`
	TokenURL     string `json:"token_url"`
	UserInfoURL  string `json:"user_info_url"`
	Scopes       string `gorm:"default:'openid,profile,email'" json:"scopes"`
	Issuer       string `json:"issuer"`
	Enabled      bool   `gorm:"default:true" json:"enabled"`
}
