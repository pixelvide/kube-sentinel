package models

type UserIdentity struct {
	Model
	UserID     uint   `gorm:"not null;index" json:"user_id"`
	Provider   string `gorm:"not null;index:idx_provider_id,unique" json:"provider"`    // e.g., "oidc", "gitlab"
	ProviderID string `gorm:"not null;index:idx_provider_id,unique" json:"provider_id"` // The unique ID from the provider (e.g., OIDC 'sub')
}
